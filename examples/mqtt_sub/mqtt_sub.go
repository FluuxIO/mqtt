//
// MQTT basic subscriber
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
	messages := make(chan *mqtt.Message)
	client := mqtt.New("localhost:1883", messages)
	client.ClientID = "MQTT-Sub"
	fmt.Printf("Server to connect to: %s\n", client.Address)

	if err := client.Connect(); err != nil {
		fmt.Printf("Connection error: %q\n", err)
		return
	}

	name := "test/topic"
	topic := mqtt.Topic{Name: name, QOS: 1}
	client.Subscribe(topic)

	time.AfterFunc(15*time.Second, func() {
		client.Unsubscribe(name)
	})

	// I use this to check number of go routines in memory
	// Can be commented out
	go quitDebugHandler()

	for m := range messages {
		fmt.Printf("Received message from MQTT server on topic %s: %+v\n", m.Topic, m.Payload)
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
