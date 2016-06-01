package mqtt

// Available options for MQTT client:
type ClientOptions struct {
	Address      string
	ClientID     string
	Keepalive    int
	CleanSession bool
}

func NewClientOptions(address string, clientID string) *ClientOptions {
	return &ClientOptions{Address: address, ClientID: clientID, Keepalive: 30, CleanSession: true}
}
