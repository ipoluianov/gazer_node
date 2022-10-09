package system

import (
	"crypto/rsa"
	"errors"

	"github.com/gazercloud/gazernode/common_interfaces"
	"github.com/ipoluianov/xchg/xchg"
)

type XchgServer struct {
	serverConnection *xchg.Peer
	requester        common_interfaces.Requester
}

func NewXchgServer(privateKey *rsa.PrivateKey) *XchgServer {
	var c XchgServer
	c.serverConnection = xchg.NewPeer(privateKey)
	c.serverConnection.SetProcessor(&c)
	return &c
}

func (c *XchgServer) Start() {
	c.serverConnection.Start()
}

func (c *XchgServer) Stop() {
	c.serverConnection.Stop()
}

func (c *XchgServer) SetRequester(requester common_interfaces.Requester) {
	c.requester = requester
}

func (c *XchgServer) ServerProcessorAuth(authData []byte) (err error) {
	if string(authData) == "pass" {
		return nil
	}
	return errors.New(xchg.ERR_XCHG_ACCESS_DENIED)
}

func (c *XchgServer) ServerProcessorCall(function string, parameter []byte) (response []byte, err error) {
	response, err = c.requester.RequestJson(function, parameter, "", false)
	return
}
