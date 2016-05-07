package packet

import (
	"bytes"
	"fmt"
)

const (
	reserved1Type = iota
	connectType   = iota
	connackType   = iota
)

// Packet interface shared by all MQTT control packets
type Marshaller interface {
	Marshall() bytes.Buffer
	PacketType() int
}

// NewConnect creates a CONNECT packet with default values
func NewConnect() *Connect {
	connect := new(Connect)
	connect.keepalive = 30
	return connect
}

// NewConnAck creates a CONNACK packet with default values
func NewConnAck() *ConnAck {
	return new(ConnAck)
}

// Decode returns parsed struct from byte array
func Decode(packetType int, payload []byte) Marshaller {
	switch packetType {
	case connackType:
		return decodeConnAck(payload)
	default:
		fmt.Println("Unsupported MQTT packet type")
		return nil
	}
}
