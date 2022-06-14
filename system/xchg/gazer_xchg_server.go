package xchg

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/gazercloud/gazernode/common_interfaces"
)

type GazerXchgServer struct {
	id        string
	client    *Server
	requester common_interfaces.Requester
}

func NewPoint() *GazerXchgServer {
	var c GazerXchgServer
	c.client = NewServer("node", c.onRcv)
	return &c
}

func (c *GazerXchgServer) onRcv(frame []byte) (response []byte, err error) {
	if len(frame) < 4 || len(frame) > 10*1024*1024 {
		return nil, errors.New("wrong message size")
	}
	unencryptedMessageSize := binary.LittleEndian.Uint32(frame)
	if unencryptedMessageSize+4 > uint32(len(frame)) {
		return nil, errors.New("wrong unencrypted message size")
	}

	frame = frame[4+unencryptedMessageSize:]

	var f Frame
	json.Unmarshal(frame, &f)
	response, err = c.requester.RequestJson(f.Function, f.Data, "", true)
	return
}

func (c *GazerXchgServer) Start() {
}

func (c *GazerXchgServer) Stop() {
}

func (c *GazerXchgServer) SetRequester(requester common_interfaces.Requester) {
	c.requester = requester
}
