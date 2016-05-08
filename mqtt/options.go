package mqtt

// Available options for MQTT client
type ClientOptions struct {
	Address   string
	ClientID  string
	Keepalive int
}

// TODO set default value for keepalive
