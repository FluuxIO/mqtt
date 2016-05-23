package packet

import (
	"bytes"
	"errors"
	"fmt"
)

type ConnAck struct {
	ReturnCode int
}

const (
	ConnAccepted                     = 0x00
	ConnRefusedBadProtocolVersion    = 0x01
	ConnRefusedIDRejected            = 0x02
	ConnRefusedServerUnavailable     = 0x03
	ConnRefusedBadUsernameOrPassword = 0x04
	ConnRefusedNotAuthorised         = 0x05
)

func (c *ConnAck) PacketType() int {
	return connackType
}

func (c *ConnAck) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	reserved := 0

	variablePart.WriteByte(byte(reserved))
	variablePart.WriteByte(byte(c.ReturnCode))

	fixedHeader := (connackType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

func decodeConnAck(payload []byte) *ConnAck {
	connAck := new(ConnAck)
	// MQTT 3.1.1: payload[0] is reserved for future use
	connAck.ReturnCode = int(payload[1])
	fmt.Printf("Return Code: %d\n", connAck.ReturnCode)
	return connAck
}

func ConnAckError(returnCode int) error {
	switch returnCode {
	case ConnRefusedBadProtocolVersion:
		return errors.New("connection refused, unacceptable protocol version")
	case ConnRefusedIDRejected:
		return errors.New("connection refused, identifier rejected")
	case ConnRefusedServerUnavailable:
		return errors.New("connection refused, server unavailable")
	case ConnRefusedBadUsernameOrPassword:
		return errors.New("connection refused, bad user name or password")
	case ConnRefusedNotAuthorised:
		return errors.New("connection refused, not authorized")
	}
	return errors.New("connection refused, unknown error")
}
