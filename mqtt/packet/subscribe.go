package packet

import (
	"bytes"
	"encoding/binary"
)

type Subscribe struct {
	id     int
	topics []Topic
}

type Topic struct {
	Name string
	Qos  int
}

func (s *Subscribe) AddTopic(topic Topic) {
	s.topics = append(s.topics, topic)
}

func (s *Subscribe) PacketType() int {
	return subscribeType
}

func (s *Subscribe) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	// Empty topic list is incorrect. Server must disconnect.
	if len(s.topics) == 0 {
		return packet
	}

	variablePart.Write(encodeUint16(uint16(s.id)))

	for _, topic := range s.topics {
		variablePart.Write(encodeString(topic.Name))
		// TODO Check that QOS is valid
		variablePart.WriteByte(byte(topic.Qos))
	}

	fixedHeaderFlags := 2 // mandatory value
	fixedHeader := (subscribeType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

func decodeSubscribe(payload []byte) *Subscribe {
	subscribe := new(Subscribe)
	subscribe.id = int(binary.BigEndian.Uint16(payload[:2]))

	for remaining := payload[2:]; len(remaining) > 0; {
		topic := Topic{}
		var rest []byte
		topic.Name, rest = extractNextString(remaining)
		topic.Qos = int(rest[0])
		subscribe.AddTopic(topic)
		remaining = rest[1:]
	}

	return subscribe
}
