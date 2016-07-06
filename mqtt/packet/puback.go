package packet

import (
	"bytes"
	"encoding/binary"
)

// PubAck ...
type PubAck struct {
	id int
}

// Marshall ....
func (s PubAck) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	variablePart.Write(encodeUint16(uint16(s.id)))

	fixedHeaderFlags := 0
	fixedHeader := (pubackType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

//==============================================================================

type puback struct{}

var pubAck puback

func (puback) decode(payload []byte) PubAck {
	return PubAck{
		id: int(binary.BigEndian.Uint16(payload[:2])),
	}
}
