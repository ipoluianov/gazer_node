// Copyright 2021-2023 Ivan Poluianov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.
// The list of authors can be found in the AUTHORS file

package main

import (
	"github.com/ipoluianov/gazer_node/app"
	"github.com/ipoluianov/gazer_node/application"
)

func main() {
	//utilities.TS()
	//return
	// imagegenerator.Generate()
	// return

	application.Name = "Gazer"
	application.ServiceName = "Gazer"
	application.ServiceDisplayName = "Gazer"
	application.ServiceDescription = "Gazer Service"
	application.ServiceRunFunc = app.RunAsService
	application.ServiceStopFunc = app.StopService

	if !application.TryService() {
		app.RunDesktop()
	}
}
