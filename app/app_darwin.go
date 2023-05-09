package app

import (
	"fmt"
	"os"

	"github.com/ipoluianov/gazer_node/application"
	"github.com/ipoluianov/gazer_node/utilities/logger"
	"github.com/ipoluianov/gazer_node/utilities/paths"
)

func RunDesktop() {
	logger.Init(paths.HomeFolder() + "/gazer/log_ui")
	if len(os.Args) == 1 {
		return
	}

	start(application.ServerDataPathArgument)
	logger.Println("Started as console application")
	logger.Println("Press ENTER to stop")
	_, _ = fmt.Scanln()
	stop()
}
