//
// MQTT basic subscriber
//

package main

import (
	"log"
	"time"

	"github.com/processone/gomqtt/mqtt"
)

func main() {
	client := mqtt.New("localhost:1883")
	client.ClientID = "MQTT-Sub"
	log.Printf("Server to connect to: %s\n", client.Address)

	messages := make(chan mqtt.Message)

	postConnect := func(c *mqtt.Client) {
		name := "test/topic"
		topic := mqtt.Topic{Name: name, QOS: 1}
		c.Subscribe(topic)
	}

	client.Handler = autoReconnectHandler(client, messages, postConnect)
	connect(client, messages, postConnect)

	for m := range messages {
		log.Printf("Received message from MQTT server on topic %s: %+v\n", m.Topic, m.Payload)
	}
}

// postConnect function, if defined, is executed right after connection
// success (CONNACK).
type postConnect func(c *mqtt.Client)

// Connect loop
func connect(client *mqtt.Client, msgs chan mqtt.Message, pc postConnect) {
	var backoff mqtt.Backoff

	for {
		if err := client.Connect(msgs); err != nil {
			log.Printf("Connection error: %v\n", err)
			time.Sleep(backoff.Duration()) // Do we want a function backoff.Sleep() ?)
		} else {
			break
		}
	}

	if pc != nil {
		pc(client)
	}
}

func autoReconnectHandler(client *mqtt.Client, messages chan mqtt.Message, postConnect postConnect) mqtt.EventHandler {
	handler := func(e mqtt.Event) {
		if e.State == mqtt.StateDisconnected {
			connect(client, messages, postConnect)
		}
	}
	return handler
}
