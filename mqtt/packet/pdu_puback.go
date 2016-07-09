package packet

import (
	"bytes"
	"encoding/binary"
)

// PubAck ...
type PDUPubAck struct {
	ID int
}

// Marshall ....
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
