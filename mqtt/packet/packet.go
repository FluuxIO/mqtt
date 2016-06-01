package packet

import (
	"bytes"
	"errors"
)

var (
	ErrMalformedLength = errors.New("malformed mqtt packet remaining length")
)

const (
	reserved1Type = iota
	connectType
	connackType
	publishType
	pubackType
	pubrecType
	pubrelType
	pubcompType
	subscribeType
	subackType
	unsubscribeType
	unsubackType
	pingreqType
	pingrespType
	disconnectType
	reserved2Type
)

// Marshaller interface is shared by all MQTT control packets
type Marshaller interface {
	Marshall() bytes.Buffer
}

// NewConnect creates a CONNECT packet with default values
func NewConnect() *Connect {
	connect := new(Connect)
	connect.keepalive = 30
	connect.protocolName = protocolName
	connect.protocolLevel = protocolLevel
	return connect
}

// NewPublish creates an empty PUBLISH packet with default value.
// You need at least to set a topic to make a valid packet.
func NewPublish() *Publish {
	return new(Publish)
}

// NewPubAck creates a valid PUBACK packet with id
func NewPubAck(id int) *PubAck {
	puback := new(PubAck)
	puback.id = id
	return puback
}

// NewSubscribe creates an empty SUBSCRIBE packet. You need to add at
// least one topic to create a valid subscribe packet.
func NewSubscribe() *Subscribe {
	return new(Subscribe)
}

// NewUnsubscribe creates an empty UNSUBSCRIBE packet. You need to add at
// least one topic to create a valid unsubscribe packet.
func NewUnsubscribe() *Unsubscribe {
	return new(Unsubscribe)
}

// NewPingReq creates a PINGREQ packet
func NewPingReq() *PingReq {
	return new(PingReq)
}

// NewPingResp creates a PINGRESP packet
func NewPingResp() *PingResp {
	return new(PingResp)
}

// NewDisconnect creates a DISCONNECT packet
func NewDisconnect() *Disconnect {
	return new(Disconnect)
}

// TODO Should probably go in a decode.go file
// Decode returns parsed struct from byte array
func Decode(packetType int, fixedHeaderFlags int, payload []byte) Marshaller {
	switch packetType {
	case connectType:
		return decodeConnect(payload)
	case connackType:
		return decodeConnAck(payload)
	case publishType:
		return decodePublish(fixedHeaderFlags, payload)
	case pubackType:
		return decodePubAck(payload)
	case subscribeType:
		return decodeSubscribe(payload)
	case subackType:
		return decodeSubAck(payload)
	case unsubscribeType:
		return decodeUnsubscribe(payload)
	case unsubackType:
		return decodeUnsubAck(payload)
	case pingreqType:
		return decodePingReq(payload)
	case pingrespType:
		return decodePingResp(payload)
	case disconnectType:
		return decodeDisconnect(payload)
	default: // Unsupported MQTT packet type
		return nil
	}
}
