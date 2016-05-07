package mqtt

// Packet interface shared by all MQTT control packets
type Packet interface {
	PacketType() int
}
