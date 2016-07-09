package packet

import "bytes"

// PDUConnAck is the PDU sent as a reply to CONNECT control packet.
// It contains the result of the CONNECT operation.
type PDUConnAck struct {
	ReturnCode int
}

// Marshall serializes a CONNACK struct as an MQTT control packet.
func (pdu PDUConnAck) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	reserved := 0

	variablePart.WriteByte(byte(reserved))
	variablePart.WriteByte(byte(pdu.ReturnCode))

	fixedHeader := (connackType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

// ============================================================================

type connAckDecoder struct{}

var pduConnAck connAckDecoder

func (connAckDecoder) decode(payload []byte) PDUConnAck {
	return PDUConnAck{
		ReturnCode: int(payload[1]),
	}
}
