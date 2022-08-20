package app

import (
	"fmt"
	"github.com/ipoluianov/gazer_node/application"
	"github.com/ipoluianov/gazer_node/cmd"
	"github.com/ipoluianov/gazer_node/utilities/logger"
	"github.com/ipoluianov/gazer_node/utilities/paths"
	"os"
)

func RunDesktop() {
	logger.Init(paths.HomeFolder() + "/gazer/log_ui")
	if len(os.Args) == 1 {
		cmd.Console()
		return
	}

	start(application.ServerDataPathArgument)
	logger.Println("Started as console application")
	logger.Println("Press ENTER to stop")
	_, _ = fmt.Scanln()
	stop()
}
