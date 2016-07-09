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

// MQTT error codes returned on CONNECT.
const (
	ConnAccepted                     = 0x00
	ConnRefusedBadProtocolVersion    = 0x01
	ConnRefusedIDRejected            = 0x02
	ConnRefusedServerUnavailable     = 0x03
	ConnRefusedBadUsernameOrPassword = 0x04
	ConnRefusedNotAuthorized         = 0x05
)

// Default protocol values
const (
	fixedHeaderFlags = 0
	ProtocolName     = "MQTT"
	ProtocolLevel    = 4 // This is MQTT v3.1.1
	DefaultClientID  = "GoMQTT"
)

// =============================================================================

// Errors MQTT client can return.
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

// ConnAckError translates an MQTT ConnAck error into a Go error.
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

// Decode returns parsed struct from byte array. It assumes payload does not contain
// MQTT control packet fixed header, as parsing fixed header is needed to extract
// the packet type code we have to decode.
func Decode(packetType int, fixedHeaderFlags int, payload []byte) Marshaller {
	switch packetType {
	case connectType:
		return pduConnect.decode(payload)
	case connackType:
		return pduConnAck.decode(payload)
	case publishType:
		return pduPublish.decode(fixedHeaderFlags, payload)
	case pubackType:
		return pduPubAck.decode(payload)
	case subscribeType:
		return pduSubscribe.decode(payload)
	case subackType:
		return pduSubAck.decode(payload)
	case unsubscribeType:
		return pduUnsubscribe.decode(payload)
	case unsubackType:
		return pduUnsubAck.decode(payload)
	case pingreqType:
		return pduPingReq.decode(payload)
	case pingrespType:
		return pduPingResp.decode(payload)
	case disconnectType:
		return pduDisconnect.decode(payload)
	default: // Unsupported MQTT packet type
		return nil
	}
}
