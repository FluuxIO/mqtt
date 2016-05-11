package packet

import (
	"bytes"
	"encoding/binary"
)

type UnsubAck struct {
	id int
}

func (u *UnsubAck) PacketType() int {
	return unsubackType
}

func (u *UnsubAck) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	variablePart.Write(encodeUint16(uint16(u.id)))

	fixedHeaderFlags := 2
	fixedHeader := (unsubackType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

func decodeUnsubAck(payload []byte) *UnsubAck {
	unsuback := new(UnsubAck)
	if len(payload) >= 2 {
		unsuback.id = int(binary.BigEndian.Uint16(payload[:2]))
	}
	return unsuback
}
