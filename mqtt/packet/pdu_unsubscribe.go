package packet

import (
	"bytes"
	"encoding/binary"
)

// PDUUnsubscribe is the PDU sent by client to unsubscribe from one or more topics.
type PDUUnsubscribe struct {
	ID     int
	Topics []string
}

// Marshall serializes a UNSUBSCRIBE struct as an MQTT control packet.
func (u PDUUnsubscribe) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	// Empty topic list is incorrect. Server must disconnect.
	if len(u.Topics) == 0 {
		return packet
	}

	variablePart.Write(encodeUint16(uint16(u.ID)))

	for _, topic := range u.Topics {
		variablePart.Write(encodeString(topic))
	}

	fixedHeaderFlags := 2 // mandatory value
	fixedHeader := (unsubscribeType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

//==============================================================================

type pduUnsubscribeDecoder struct{}

var pduUnsubscribe pduUnsubscribeDecoder

func (pduUnsubscribeDecoder) decode(payload []byte) PDUUnsubscribe {
	unsubscribe := PDUUnsubscribe{}
	unsubscribe.ID = int(binary.BigEndian.Uint16(payload[:2]))

	for remaining := payload[2:]; len(remaining) > 0; {
		var topic string
		topic, remaining = extractNextString(remaining)
		unsubscribe.Topics = append(unsubscribe.Topics, topic)
	}

	return unsubscribe
}
