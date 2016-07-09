package packet

import (
	"bytes"
	"encoding/binary"
)

type Topic struct {
	Name string
	QOS  int
}

type PDUSubscribe struct {
	ID     int
	Topics []Topic
}

func (s PDUSubscribe) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	// Empty topic list is incorrect. Server must disconnect.
	if len(s.Topics) == 0 {
		return packet
	}

	variablePart.Write(encodeUint16(uint16(s.ID)))

	for _, topic := range s.Topics {
		variablePart.Write(encodeString(topic.Name))
		// TODO Check that QOS is valid
		variablePart.WriteByte(byte(topic.QOS))
	}

	fixedHeaderFlags := 2 // mandatory value
	fixedHeader := (subscribeType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

//==============================================================================

type pduSubscribeDecoder struct{}

var pduSubscribe pduSubscribeDecoder

func (pduSubscribeDecoder) decode(payload []byte) PDUSubscribe {
	subscribe := PDUSubscribe{}
	subscribe.ID = int(binary.BigEndian.Uint16(payload[:2]))

	for remaining := payload[2:]; len(remaining) > 0; {
		topic := Topic{}
		var rest []byte
		topic.Name, rest = extractNextString(remaining)
		topic.QOS = int(rest[0])
		subscribe.Topics = append(subscribe.Topics, topic)
		remaining = rest[1:]
	}

	return subscribe
}
