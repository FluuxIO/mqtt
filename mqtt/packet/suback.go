package packet

import (
	"bytes"
	"encoding/binary"
)

type SubAck struct {
	id          int
	returnCodes []int
}

func (s *SubAck) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	variablePart.Write(encodeUint16(uint16(s.id)))

	for _, rc := range s.returnCodes {
		variablePart.WriteByte(byte(rc))
	}

	fixedHeaderFlags := 0
	fixedHeader := (subackType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

// TODO How to return all code backs to client using the library ?
// We likely want to keep in memory current subscription with their state
// Client could read the current subscription state map to read the status of each subscription.
// We should probably return error if a subscription is rejected or if
// one of the QOS is lower than the level we asked for.
func decodeSubAck(payload []byte) *SubAck {
	suback := new(SubAck)
	if len(payload) >= 2 {
		suback.id = int(binary.BigEndian.Uint16(payload[:2]))
		for b := range payload[2:] {
			suback.returnCodes = append(suback.returnCodes, int(b))
		}
	}
	return suback
}
