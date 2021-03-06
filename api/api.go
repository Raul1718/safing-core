// Copyright Safing ICS Technologies GmbH. Use of this source code is governed by the AGPL license that can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/Safing/safing-core/log"
	"github.com/Safing/safing-core/modules"
)

var apiModule *modules.Module

func Start() {
	apiModule = modules.Register("Api", 32)

	go run()

	<-apiModule.Stop
	apiModule.StopComplete()
}

func run() {
	router := NewRouter()

	log.Infof("%s", http.ListenAndServe(":18", router))
}
