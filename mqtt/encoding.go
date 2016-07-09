package mqtt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// MQTT Control Packet types
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

//==============================================================================

func encodeString(str string) []byte {
	length := make([]byte, 2)
	binary.BigEndian.PutUint16(length, uint16(len(str)))
	return append(length, []byte(str)...)
}

func encodeUint16(num uint16) []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, num)
	return bytes
}

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func int2bool(i int) bool {
	if i == 1 {
		return true
	}
	return false
}

//==============================================================================

// PacketRead returns unmarshalled packet from io.Reader stream
func PacketRead(r io.Reader) (Marshaller, error) {
	var err error
	fixedHeader := make([]byte, 1)

	if _, err = io.ReadFull(r, fixedHeader); err != nil {
		fmt.Printf("Read error %q", err.Error())
		return nil, err
	}

	packetType := fixedHeader[0] >> 4
	fixedHeaderFlags := fixedHeader[0] & 15 // keep only last 4 bits

	fmt.Printf("PacketType: %d\n", packetType)
	length, _ := readRemainingLength(r)
	fmt.Printf("Length: %d\n", length)
	payload := make([]byte, length)
	if _, err = io.ReadFull(r, payload); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			fmt.Printf("Connection closed unexpectedly\n")
		}
		return nil, err
	}
	return Decode(int(packetType), int(fixedHeaderFlags), payload), err
}

// ReadRemainingLength decodes MQTT Packet remaining length field
// Reference: http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html#_Toc398718023
func readRemainingLength(r io.Reader) (int, error) {
	var multiplier uint32 = 1
	var value uint32
	var err error
	encodedByte := make([]byte, 1)
	for ok := true; ok; ok = (encodedByte[0]&128 != 0) {
		io.ReadFull(r, encodedByte)
		value += uint32(encodedByte[0]&127) * multiplier
		multiplier *= 128
		if multiplier > 128*128*128 {
			err = ErrMalformedLength
			return 0, err
		}
	}

	return int(value), err
}

func extractNextString(data []byte) (string, []byte) {
	offset := 2
	length := int(binary.BigEndian.Uint16(data[:offset]))
	return string(data[offset : length+offset]), data[length+offset:]
}
