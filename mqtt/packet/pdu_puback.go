package packet

import (
	"bytes"
	"encoding/binary"
)

// PDUPubAck is the PDU sent by client or server as response to client PUBLISH,
// when QOS for publish is greater than 1.
type PDUPubAck struct {
	ID int
}

// Marshall serializes a PUBACK struct as an MQTT control packet.
func (s PDUPubAck) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	variablePart.Write(encodeUint16(uint16(s.ID)))

	fixedHeaderFlags := 0
	fixedHeader := (pubackType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

//==============================================================================

type pduPubAckDecoder struct{}

var pduPubAck pduPubAckDecoder

func (pduPubAckDecoder) decode(payload []byte) PDUPubAck {
	return PDUPubAck{
		ID: int(binary.BigEndian.Uint16(payload[:2])),
	}
}
