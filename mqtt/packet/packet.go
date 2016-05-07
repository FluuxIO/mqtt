package packet

import "bytes"

// Packet interface shared by all MQTT control packets
type Marshaller interface {
	Marshall() bytes.Buffer
	PacketType() int
}

// TODO interface should be packet.Marshaller ?

func NewConnect() *Connect {
	return new(Connect)
}
