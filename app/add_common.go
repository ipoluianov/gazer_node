package app

import (
	"flag"

	"github.com/ipoluianov/gazer_node/application"
	"github.com/ipoluianov/gazer_node/system/httpserver"
	"github.com/ipoluianov/gazer_node/system/settings"
	"github.com/ipoluianov/gazer_node/system/system"
	"github.com/ipoluianov/gazer_node/utilities/hostid"
	"github.com/ipoluianov/gazer_node/utilities/logger"
)

var httpServer *httpserver.HttpServer
var sys *system.System
var runServerFlagPtr = flag.Bool("server", false, "Run server")

func start(dataPath string) {
	hostid.InitHostId()

	ss := settings.NewSettings()
	ss.SetServerDataPath(dataPath)

	sys = system.NewSystem(ss)
	httpServer = httpserver.NewHttpServer(sys)
	sys.SetRequester(httpServer)
	sys.Start()
	httpServer.Start()
}

func stop() {
	if sys != nil {
		sys.Stop()
	}
	if httpServer != nil {
		_ = httpServer.Stop()
	}
}

func RunAsService() error {
	logger.Init(application.ServerDataPathArgument + "/log_service")
	logger.Println("Started as Service")
	start(application.ServerDataPathArgument)
	return nil
}

func StopService() {
	logger.Println("Service stopped")
	stop()
}
