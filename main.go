// Copyright Safing ICS Technologies GmbH. Use of this source code is governed by the AGPL license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/Safing/safing-core/api"
	_ "github.com/Safing/safing-core/configuration"
	"github.com/Safing/safing-core/firewall"
	_ "github.com/Safing/safing-core/firewall/inspection/tls"
	"github.com/Safing/safing-core/log"
	"github.com/Safing/safing-core/modules"
	"github.com/Safing/safing-core/nameserver"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU() * 4)

	modules.RegisterLogger(log.Logger)

	// go intel.Start()
	go nameserver.Start()
	go firewall.Start()
	go api.Start()

	// SHUTDOWN
	// catch interrupt for clean shutdown
	signalCh := make(chan os.Signal)
	signal.Notify(
		signalCh,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	select {
	case <-signalCh:
		log.Warning("program was interrupted, shutting down.")
		modules.InitiateFullShutdown()
	case <-modules.GlobalShutdown:
	}

	// wait for shutdown to complete, panic after timeout
	time.Sleep(5 * time.Second)
	fmt.Println("===== TAKING TOO LONG FOR SHUTDOWN - PRINTING STACK TRACES =====")
	pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
	os.Exit(1)

}
