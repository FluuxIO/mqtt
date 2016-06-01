package packet

import "bytes"

type Disconnect struct {
}

func (d *Disconnect) Marshall() bytes.Buffer {
	var packet bytes.Buffer
	fixedHeader := (disconnectType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(0))
	return packet
}

func decodeDisconnect(payload []byte) *Disconnect {
	disconnect := new(Disconnect)
	return disconnect
}
