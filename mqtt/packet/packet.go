package packet

import (
	"bytes"
	"errors"
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

const (
	ConnAccepted                     = 0x00
	ConnRefusedBadProtocolVersion    = 0x01
	ConnRefusedIDRejected            = 0x02
	ConnRefusedServerUnavailable     = 0x03
	ConnRefusedBadUsernameOrPassword = 0x04
	ConnRefusedNotAuthorized         = 0x05
)

const (
	fixedHeaderFlags = 0
	protocolName     = "MQTT"
	protocolLevel    = 4 // This is MQTT v3.1.1
	defaultClientID  = "GoMQTT"
)

// =============================================================================

var (
	ErrMalformedLength                  = errors.New("malformed mqtt packet remaining length")
	ErrConnRefusedBadProtocolVersion    = errors.New("connection refused, unacceptable protocol version")
	ErrConnRefusedIDRejected            = errors.New("connection refused, identifier rejected")
	ErrConnRefusedServerUnavailable     = errors.New("connection refused, server unavailable")
	ErrConnRefusedBadUsernameOrPassword = errors.New("connection refused, bad user name or password")
	ErrConnRefusedNotAuthorized         = errors.New("connection refused, not authorized")
	ErrConnUnknown                      = errors.New("connection refused, unknown error")
)

// Marshaller interface is shared by all MQTT control packets
type Marshaller interface {
	Marshall() bytes.Buffer
}

// =============================================================================

// ConnAckError ...
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

// Decode returns parsed struct from byte array
func Decode(packetType int, fixedHeaderFlags int, payload []byte) Marshaller {
	switch packetType {
	case connectType:
		return pduConnect.decode(payload)
	case connackType:
		return pduConnAck.decode(payload)
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
