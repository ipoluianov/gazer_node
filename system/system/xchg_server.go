package system

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/gazercloud/gazernode/common_interfaces"
	"github.com/ipoluianov/xchg/xchg"
)

type XchgServer struct {
	serverConnection *xchg.Peer
	masterKey        string
	requester        common_interfaces.Requester
}

func NewXchgServer(privateKey *rsa.PrivateKey, masterKey string) *XchgServer {
	var c XchgServer
	c.masterKey = masterKey
	c.serverConnection = xchg.NewPeer(privateKey)
	c.serverConnection.SetProcessor(&c)
	serverAddress := xchg.AddressForPublicKey(&privateKey.PublicKey)

	bs, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	hh := hex.EncodeToString(bs)
	fmt.Println(hh)

	fmt.Println(privateKey.PublicKey.N)
	fmt.Println(serverAddress)
	return &c
}

func (c *XchgServer) Start() {
	c.serverConnection.StartUDPOnly()
}

func (c *XchgServer) Stop() {
	c.serverConnection.Stop()
}

func (c *XchgServer) SetRequester(requester common_interfaces.Requester) {
	c.requester = requester
}

func (c *XchgServer) ServerProcessorAuth(authData []byte) (err error) {
	if string(authData) == c.masterKey {
		return nil
	}
	return errors.New(xchg.ERR_XCHG_ACCESS_DENIED)
}

func (c *XchgServer) ServerProcessorCall(function string, parameter []byte) (response []byte, err error) {
	response, err = c.requester.RequestJson(function, parameter, "", false)
	return
}
