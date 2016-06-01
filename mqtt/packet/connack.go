package packet

import (
	"bytes"
	"errors"
	"fmt"
)

const (
	ConnAccepted                     = 0x00
	ConnRefusedBadProtocolVersion    = 0x01
	ConnRefusedIDRejected            = 0x02
	ConnRefusedServerUnavailable     = 0x03
	ConnRefusedBadUsernameOrPassword = 0x04
	ConnRefusedNotAuthorized         = 0x05
)

var (
	ErrConnRefusedBadProtocolVersion    = errors.New("connection refused, unacceptable protocol version")
	ErrConnRefusedIDRejected            = errors.New("connection refused, identifier rejected")
	ErrConnRefusedServerUnavailable     = errors.New("connection refused, server unavailable")
	ErrConnRefusedBadUsernameOrPassword = errors.New("connection refused, bad user name or password")
	ErrConnRefusedNotAuthorized         = errors.New("connection refused, not authorized")
	ErrConnUnknown                      = errors.New("connection refused, unknown error")
)

type ConnAck struct {
	ReturnCode int
}

// ============================================================================

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
	// In MQTT 3.1.1, first byte (= payload[0]) is reserved for future use
	connAck.ReturnCode = int(payload[1])
	fmt.Printf("Return Code: %d\n", connAck.ReturnCode)
	return connAck
}

// ============================================================================

func ConnAckError(returnCode int) error {
	switch returnCode {
	case ConnRefusedBadProtocolVersion:
		return ErrConnRefusedBadProtocolVersion
	case ConnRefusedIDRejected:
		return ErrConnRefusedIDRejected
	case ConnRefusedServerUnavailable:
		return ErrConnRefusedServerUnavailable
	case ConnRefusedBadUsernameOrPassword:
		return ErrConnRefusedBadUsernameOrPassword
	case ConnRefusedNotAuthorized:
		return ErrConnRefusedNotAuthorized
	}
	return ErrConnUnknown
}
