package packet

import "bytes"

// PDUDisconnect ...
type PDUDisconnect struct {
}

// Marshall ...
func (d *PDUDisconnect) Marshall() bytes.Buffer {
	var packet bytes.Buffer

	fixedHeader := (disconnectType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(0))
	return packet
}

//==============================================================================

// decodeDisconnect ...
func decodeDisconnect(payload []byte) *PDUDisconnect {
	disconnect := new(PDUDisconnect)
	return disconnect
}
