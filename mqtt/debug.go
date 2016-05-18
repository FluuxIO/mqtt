package mqtt

/*
When debug mode is enabled in MQTT client, this handler goroutine is installed.
It will dump state of the system and keep the system running when we do:

    pkill -SIGQUIT prog_name
*/

import (
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
)

func QuitDebugHandler() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGQUIT)
	//	buf := make([]byte, 1<<20)
	for {
		<-sigs
		pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
	}
}
