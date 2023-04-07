package app

import (
	"fmt"

	"github.com/ipoluianov/gazer_node/application"
	"github.com/ipoluianov/gazer_node/utilities/logger"
	"github.com/ipoluianov/gazer_node/utilities/paths"
)

func RunDesktop() {
	if *runServerFlagPtr {
		logger.Init(paths.HomeFolder() + "/gazer/log_ui")
		start(application.ServerDataPathArgument)
		logger.Println("Started as console application")
		logger.Println("Press ENTER to stop")
		_, _ = fmt.Scanln()
		stop()
	}
}
