package mqtt

import (
	"log"
	"time"
)

// postConnect function, if defined, is executed right after connection
// success (CONNACK).
type postConnect func(c *Client) // TODO Should not take an MQTT client, but an io.Writer

// ClientManager supervises a connection to handle connection events and
// reconnection strategy.
type ClientManager struct {
	client      Client
	PostConnect postConnect
	// TODO Handler func to rebroadcast MQTT event
	// Handler     mqtt.EventHandler
}

// NewClientManager creates a new client manager structure, intended to support
// handling MQTT client state event changes and autotrigger connection reconnection
// based on ClientManager configuration.
func NewClientManager(client *Client, pc postConnect) *ClientManager {
	return &ClientManager{
		client:      *client,
		PostConnect: pc,
	}
}

// Start launch the connection loop
func (cm ClientManager) Start() error {
	// TODO Fix me: Ensure we do not override existing handler by supporting a list of handlers.
	cm.client.Handler = func(e Event) {
		if e.State == StateDisconnected {
			cm.connect(cm.client.Messages)
		}
	}

	return cm.connect(cm.client.Messages)
}

// Stop cancels pending operations and terminates existing MQTT client.
func (cm ClientManager) Stop() {
	// TODO
}

// Connect loop with backoff
func (cm ClientManager) connect(msgs chan<- Message) error {
	var backoff Backoff // TODO Probably group backoff calculation features with connection manager.

	for {
		if err := cm.client.Connect(msgs); err != nil {
			log.Printf("Connection error: %v\n", err)
			time.Sleep(backoff.Duration()) // Do we want a function backoff.Sleep() ?)
		} else {
			break
		}
	}

	if cm.PostConnect != nil {
		cm.PostConnect(&cm.client)
	}

	return nil
}
