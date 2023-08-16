package server

import (
	"os"
	"path/filepath"

	"github.com/ipoluianov/gazer_node/system/system"
)

type Server struct {
	system      *system.System
	guestAccess map[string]bool
}

func CurrentExePath() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
}

func NewServer(sys *system.System) *Server {
	var c Server
	c.system = sys
	c.initApiAccess()
	return &c
}

func (c *Server) Start() {
}

func (c *Server) Stop() error {
	return nil
}
