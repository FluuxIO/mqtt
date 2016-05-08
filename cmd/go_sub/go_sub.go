package main

import (
	"fmt"

	"github.com/processone/gomqtt/mqtt"
	"github.com/processone/gomqtt/mqtt/packet"
)

func main() {
	options := mqtt.ClientOptions{Address: "localhost:1883", Keepalive: 30}
	fmt.Printf("Server to connect to: %s\n", options.Address)
	client, _ := mqtt.NewClient(options)
	statusChan := client.Connect()
	<-statusChan
	topic := packet.Topic{Name: "test/topic"}
	client.Subscribe(topic)
	<-statusChan
}
