package packet

import (
	"bytes"
	"encoding/binary"
)

type Unsubscribe struct {
	id     int
	topics []string
}

func (u *Unsubscribe) AddTopic(topic string) {
	u.topics = append(u.topics, topic)
}

func (u *Unsubscribe) PacketType() int {
	return unsubscribeType
}

func (u *Unsubscribe) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	// Empty topic list is incorrect. Server must disconnect.
	if len(u.topics) == 0 {
		return packet
	}

	variablePart.Write(encodeUint16(uint16(u.id)))

	for _, topic := range u.topics {
		variablePart.Write(encodeString(topic))
	}

	fixedHeaderFlags := 2 // mandatory value
	fixedHeader := (unsubscribeType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

func decodeUnsubscribe(payload []byte) *Unsubscribe {
	unsubscribe := new(Unsubscribe)
	unsubscribe.id = int(binary.BigEndian.Uint16(payload[:2]))

	for remaining := payload[2:]; len(remaining) > 0; {
		var topic string
		topic, remaining = extractNextString(remaining)
		unsubscribe.AddTopic(topic)
	}

	return unsubscribe
}
