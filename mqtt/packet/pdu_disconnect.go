package packet

import "bytes"

// PDUDisconnect is the PDU sent from client to notify disconnection from server.
type PDUDisconnect struct {
}

// Marshall serializes a DISCONNECT struct as an MQTT control packet.
func (d PDUDisconnect) Marshall() bytes.Buffer {
	var packet bytes.Buffer

	fixedHeader := (disconnectType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(0))
	return packet
}

//==============================================================================

type pduDisconnectDecoder struct{}

var pduDisconnect pduDisconnectDecoder

func (pduDisconnectDecoder) decode(payload []byte) PDUDisconnect {
	var disconnect PDUDisconnect
	return disconnect
}
