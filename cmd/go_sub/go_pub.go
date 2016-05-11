package main

import (
	"fmt"

	"github.com/processone/gomqtt/mqtt"
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

	client.Publish("test/topic", []byte("Hi, There !"))

	for {
		if s2 := <-statusChan; s2.Err != nil {
			fmt.Printf("MQTT error: %q\n", s2.Err)
			break
		} else {
			fmt.Printf("Received packet from Server: %+v\n", s2.Packet)
		}
	}

}
