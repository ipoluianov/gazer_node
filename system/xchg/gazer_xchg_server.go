package xchg_server

import (
	"crypto/rsa"
	"encoding/base32"
	"errors"
	"io/ioutil"
	"os"

	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/utilities/logger"
	"github.com/ipoluianov/gomisc/crypt_tools"
	"github.com/ipoluianov/xchg/xchg"
	"github.com/ipoluianov/xchg/xchg_connections"
	"github.com/ipoluianov/xchg/xchg_network"
)

type GazerXchgServer struct {
	serverPrivateKey32 string
	path               string
	network            *xchg_network.Network
	serverConnection   *xchg_connections.ServerConnection
	requester          common_interfaces.Requester
}

func NewGazerXchgServer() *GazerXchgServer {
	var c GazerXchgServer
	return &c
}

func (c *GazerXchgServer) SetRequester(requester common_interfaces.Requester) {
	c.requester = requester
}

func (c *GazerXchgServer) Start(path string, network *xchg_network.Network) {
	c.path = path
	c.LoadKeys()
	c.serverConnection = xchg_connections.NewServerConnection()
	c.serverConnection.SetProcessor(c)
	c.network = network
	c.serverConnection.Start(c.serverPrivateKey32, c.network)
}

func (c *GazerXchgServer) Stop() {
	c.serverConnection.Stop()
}

func (c *GazerXchgServer) ServerProcessorAuth(authData []byte) (err error) {
	if string(authData) == "pass" {
		return nil
	}
	return errors.New(xchg.ERR_XCHG_ACCESS_DENIED)
}

func (c *GazerXchgServer) ServerProcessorCall(function string, parameter []byte) (response []byte, err error) {
	if c.requester == nil {
		return
	}
	response, err = c.requester.RequestJson(function, parameter, "", true)
	return
}

func (c *GazerXchgServer) LoadKeys() error {
	var privateKey *rsa.PrivateKey

	privateKeyFile := c.path + "/private_key.pem"
	publicKeyFile := c.path + "/public_key.pem"
	addressFile := c.path + "/address.txt"

	_, err := os.Stat(privateKeyFile)
	if os.IsNotExist(err) {
		privateKey, err = crypt_tools.GenerateRSAKey()
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(privateKeyFile, []byte(crypt_tools.RSAPrivateKeyToPem(privateKey)), 0666)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(publicKeyFile, []byte(crypt_tools.RSAPublicKeyToPem(&privateKey.PublicKey)), 0666)
		if err != nil {
			return err
		}
		logger.Println("Key saved to file")
	} else {
		var bs []byte
		bs, err = ioutil.ReadFile(privateKeyFile)
		if err != nil {
			return err
		}
		privateKey, err = crypt_tools.RSAPrivateKeyFromPem(string(bs))
		if err != nil {
			return err
		}
		logger.Println("Key loaded from file")
	}

	address := xchg.AddressForPublicKey(&privateKey.PublicKey)
	err = ioutil.WriteFile(addressFile, []byte(address), 0666)
	c.serverPrivateKey32 = base32.StdEncoding.EncodeToString([]byte(crypt_tools.RSAPrivateKeyToDer(privateKey)))
	logger.Println("ADDRESS:", address)
	return nil
}
