package mqtt // import "gosrc.io/mqtt"

import "log"

// postConnect function, if defined, is executed right after connection
// success (CONNACK).
type postConnect func(c *Client) // TODO Should we not take an MQTT client, but an io.Writer ?

// ClientManager supervises an MQTT client connection. Its role is to handle connection events and
// apply reconnection strategy.
type ClientManager struct {
	Client      *Client
	PostConnect postConnect
	// TODO Handler func to rebroadcast MQTT event
	// Handler     mqtt.EventHandler
	// TODO: Configurable logger
}

// NewClientManager creates a new client manager structure, intended to support
// handling MQTT client state event changes and autotrigger connection reconnection
// based on ClientManager configuration.
func NewClientManager(client *Client, pc postConnect) *ClientManager {
	return &ClientManager{
		Client:      client,
		PostConnect: pc,
	}
}

// Start launch the connection loop
func (cm *ClientManager) Start() {
	// TODO Fix me: Ensure we do not override existing handler by supporting a list of handlers.
	cm.Client.Handler = func(e Event) {
		if e.State == StateDisconnected {
			cm.connect(cm.Client.Messages)
		}
	}
	cm.connect(cm.Client.Messages)
}

// Stop cancels pending operations and terminates existing MQTT client.
func (cm *ClientManager) Stop() {
	// Remove on disconnect handler to avoid triggering reconnect
	cm.Client.Handler = nil
	cm.Client.Disconnect()
}

// connect manages the reconnection loop and apply the define backoff to avoid overloading the server.
func (cm *ClientManager) connect(msgs chan<- Message) {
	var backoff Backoff // TODO Probably group backoff calculation features with connection manager.

	for {
		if err := cm.Client.Connect(msgs); err != nil {
			log.Printf("Connection error: %v\n", err)
			backoff.Wait()
		} else {
			break
		}
	}

	if cm.PostConnect != nil {
		cm.PostConnect(cm.Client)
	}
}
