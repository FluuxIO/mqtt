package mqtt

// Available options for MQTT client:
type ClientOptions struct {
	Address           string
	ClientID          string
	Keepalive         int
	PersistentSession bool
}
