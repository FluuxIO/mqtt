package packet

import (
	"bytes"
	"fmt"
)

// Packet interface shared by all MQTT control packets
type Marshaller interface {
	Marshall() bytes.Buffer
	PacketType() int
}

func NewConnect() *Connect {
	connect := new(Connect)
	connect.keepalive = 30
	return connect
}

func Decode(packetType int, payload []byte) Marshaller {
	switch packetType {
	case 2:
		return decodeConnAck(payload)
	default:
		fmt.Println("Unsupported MQTT packet type")
		return nil
	}
}
