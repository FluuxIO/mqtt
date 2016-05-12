package main

import (
	"fmt"
	"time"

	"github.com/processone/gomqtt/mqtt"
	"github.com/processone/gomqtt/mqtt/packet"
)

func main() {
	options := mqtt.ClientOptions{Address: "localhost:1883", Keepalive: 30}
	fmt.Printf("Server to connect to: %s\n", options.Address)

	client, _ := mqtt.NewClient(options)
	statusChan := client.Connect()

	if s1 := <-statusChan; s1.Err != nil {
		fmt.Printf("Connection error: %q\n", s1.Err)
		return
	}

	name := "test/topic"
	topic := packet.Topic{Name: name, Qos: 1}
	client.Subscribe(topic)

	time.AfterFunc(time.Duration(15)*time.Second, func() {
		client.Unsubscribe(name)
	})

	for {
		if s2 := <-statusChan; s2.Err != nil {
			fmt.Printf("MQTT error: %q\n", s2.Err)
			break
		} else {
			fmt.Printf("Received packet from Server: %+v\n", s2.Packet)
		}
	}

}
