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

	var backoff mqtt.Backoff
	for {
		messages := make(chan *mqtt.Message)
		if err := client.Connect(messages); err != nil {
			log.Printf("Connection error: %q\n", err)
			time.Sleep(backoff.Duration())
			continue
		}
		backoff.Reset()

		name := "test/topic"
		topic := mqtt.Topic{Name: name, QOS: 1}
		client.Subscribe(topic)

		time.AfterFunc(15*time.Second, func() {
			client.Unsubscribe(name)
		})

		for m := range messages {
			log.Printf("Received message from MQTT server on topic %s: %+v\n", m.Topic, m.Payload)
		}
		log.Printf("message channel closed, we have to reconnect")
	}
}
