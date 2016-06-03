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
	"github.com/processone/gomqtt/mqtt/packet"
)

func main() {
	client := mqtt.New("localhost:1883")
	client.ClientID = "MQTT-Sub"
	fmt.Printf("Server to connect to: %s\n", client.Address)

	if err := client.Connect(); err != nil {
		fmt.Printf("Connection error: %q\n", err)
		return
	}

	name := "test/topic"
	topic := packet.Topic{Name: name, QOS: 1}
	client.Subscribe(topic)

	time.AfterFunc(time.Duration(15)*time.Second, func() {
		client.Unsubscribe(name)
	})

	// I use this to check number of go routines in memory
	// Can be commented out
	go quitDebugHandler()

	for {
		s2 := client.ReadNext()
		fmt.Printf("Received packet from Server on %s: %+v\n", s2.Topic, s2.Payload)
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
