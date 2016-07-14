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

	messages := make(chan *mqtt.Message)

	f := func(e mqtt.Event) {
		if e.Type == mqtt.EventDisconnected {
			connect(client, messages)
		}
	}
	client.Handler = f
	connect(client, messages)

	for m := range messages {
		log.Printf("Received message from MQTT server on topic %s: %+v\n", m.Topic, m.Payload)
	}
}

func connect(client *mqtt.Client, msgs chan *mqtt.Message) {
	var backoff mqtt.Backoff

	for {
		if err := client.Connect(msgs); err != nil {
			log.Printf("Connection error: %q\n", err)
			time.Sleep(backoff.Duration())
		} else {
			break
		}
	}

	// TODO Move this is a Connected EventHandler
	name := "test/topic"
	topic := mqtt.Topic{Name: name, QOS: 1}
	client.Subscribe(topic)
}
