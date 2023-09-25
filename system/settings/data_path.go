package settings

import (
	"os/user"
	"path/filepath"
	"strings"

	"github.com/ipoluianov/gazer_node/utilities/logger"
)

type Settings struct {
	serverDataPath string
}

func NewSettings() *Settings {
	var c Settings
	c.serverDataPath = "~/gazernode"
	return &c
}

func (c *Settings) SetServerDataPath(path string) {

	usr, _ := user.Current()
	dir := usr.HomeDir

	if path == "~" {
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(dir, path[2:])
	}

	c.serverDataPath = path
	logger.Println("Server Path:", c.serverDataPath)
}

func (c *Settings) ServerDataPath() string {
	return c.serverDataPath
}
