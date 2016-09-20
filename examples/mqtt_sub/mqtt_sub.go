//
// MQTT basic subscriber
//

package main

import (
	"log"

	"github.com/processone/gomqtt/mqtt"
)

func main() {
	client := mqtt.New("localhost:1883")
	client.ClientID = "MQTT-Sub"
	log.Printf("Server to connect to: %s\n", client.Address)

	messages := make(chan mqtt.Message)
	client.Messages = messages

	postConnect := func(c *mqtt.Client) {
		name := "test/topic"
		topic := mqtt.Topic{Name: name, QOS: 1}
		c.Subscribe(topic)
	}

	cm := mqtt.NewClientManager(client, postConnect)
	cm.Start()

	for m := range messages {
		log.Printf("Received message from MQTT server on topic %s: %+v\n", m.Topic, m.Payload)
	}
}
