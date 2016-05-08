package main

import (
	"fmt"

	"github.com/processone/gomqtt/mqtt"
)

func main() {
	options := mqtt.ClientOptions{Address: "localhost:1883"}
	fmt.Printf("Server to connect to: %s\n", options.Address)
	client, _ := mqtt.NewClient(options)
	statusChan := client.Connect()
	<-statusChan
	<-statusChan
}
