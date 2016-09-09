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
	messages := make(chan mqtt.Message)
	client := mqtt.New("localhost:1883")
	client.ClientID = "MQTT-Pub"
	fmt.Printf("Server to connect to: %s\n", client.Address)

	if err := client.Connect(messages); err != nil {
		fmt.Printf("Connection error: %q\n", err)
		return
	}

	// I use this to check number of go routines in memory
	// Can be commented out
	go quitDebugHandler()

	ticker := time.NewTicker(5 * time.Second)
	stop := make(chan bool)
	go tickLoop(client, ticker, stop)

	for m := range messages {
		fmt.Printf("Received message from MQTT server on topic %s: %+v\n", m.Topic, m.Payload)
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
	for {
		<-sigs
		pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
	}
}
