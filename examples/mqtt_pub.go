//
// MQTT basic publisher
//

package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/processone/gomqtt/mqtt"
)

func main() {
	options := mqtt.NewClientOptions("localhost:1883", "MQTT-Pub")
	fmt.Printf("Server to connect to: %s\n", options.Address)

	client, _ := mqtt.NewClient(options)
	if err := client.Connect(); err != nil {
		fmt.Printf("Connection error: %q\n", err)
		return
	}

	// I use this to check number of go routines in memory
	// Can be commented out
	go quitDebugHandler()

	ticker := time.NewTicker(time.Duration(5) * time.Second)
	stop := make(chan bool)
	go tickLoop(client, ticker, stop)

	for {
		s2 := client.ReadNext()
		fmt.Printf("Received packet from Server: %+v\n", s2.Payload)
	}
}

func tickLoop(client *mqtt.Client, ticker *time.Ticker, stop <-chan bool) {
	for done := false; !done; {
		select {
		case <-ticker.C:
			client.Publish("test/topic", []byte("Hi, There !"))
		case <-stop:
			done = true
			break
		}
	}
}

func quitDebugHandler() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGQUIT)
	//	buf := make([]byte, 1<<20)
	for {
		<-sigs
		pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
	}
}
