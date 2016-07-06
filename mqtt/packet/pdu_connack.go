package packet

import "bytes"

// PDUConnAck is the PDU for ...
type PDUConnAck struct {
	ReturnCode int
}

// Marshall ...
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

type pdu_ConnAck struct{}

var pduConnAck pdu_ConnAck

func (pdu_ConnAck) decode(payload []byte) PDUConnAck {
	return PDUConnAck{
		ReturnCode: int(payload[1]),
	}
}
