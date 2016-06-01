package packet

import (
	"bytes"
	"encoding/binary"
)

type PubAck struct {
	id int
}

func (s *PubAck) Marshall() bytes.Buffer {
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

func decodePubAck(payload []byte) *PubAck {
	puback := new(PubAck)
	puback.id = int(binary.BigEndian.Uint16(payload[:2]))
	return puback
}
