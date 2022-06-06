package xchg

import (
	"encoding/json"
	"github.com/gazercloud/gazernode/common_interfaces"
)

type Point struct {
	id        string
	client    *Client
	requester common_interfaces.Requester
}

func NewPoint() *Point {
	var c Point
	c.client = NewClient("node", c.onRcv)
	return &c
}

func (c *Point) onRcv(frame []byte) (response []byte, err error) {
	var f Frame
	json.Unmarshal(frame, &f)
	response, err = c.requester.RequestJson(f.Function, f.Data, "", true)
	return
}

func (c *Point) Start() {
}

func (c *Point) Stop() {
}

func (c *Point) SetRequester(requester common_interfaces.Requester) {
	c.requester = requester
}
