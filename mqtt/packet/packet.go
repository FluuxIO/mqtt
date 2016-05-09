package packet

import (
	"bytes"
	"fmt"
)

const (
	reserved1Type   = iota
	connectType     = iota
	connackType     = iota
	publishType     = iota
	pubackType      = iota
	pubrecType      = iota
	pubrelType      = iota
	pubcompType     = iota
	subscribeType   = iota
	subackType      = iota
	unsubscribeType = iota
	unsubackType    = iota
	pingreqType     = iota
	pingrespType    = iota
	disconnectType  = iota
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

// NewSubscribe creates an empty SUBSCRIBE packet. You need to add at
// least one topic to create a valid subscribe packet.
func NewSubscribe() *Subscribe {
	return new(Subscribe)
}

// NewPingReq creates a PINGREQ packet
func NewPingReq() *PingReq {
	return new(PingReq)
}

// NewPingResp creates a PINGRESP packet
func NewPingResp() *PingResp {
	return new(PingResp)
}

// Decode returns parsed struct from byte array
func Decode(packetType int, fixedHeaderFlags int, payload []byte) Marshaller {
	fmt.Printf("Decoding packet type: %d\n", packetType)
	switch packetType {
	case connackType:
		return decodeConnAck(payload)
	case publishType:
		return decodePublish(fixedHeaderFlags, payload)
	case subscribeType:
		return decodeSubscribe(payload)
	case subackType:
		return decodeSubAck(payload)
	case pingreqType:
		return decodePingReq(payload)
	case pingrespType:
		return decodePingResp(payload)
	default:
		fmt.Println("Unsupported MQTT packet type")
		return nil
	}
}
