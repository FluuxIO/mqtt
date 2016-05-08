package packet

import "bytes"

type PingResp struct {
}

func (p *PingResp) PacketType() int {
	return pingrespType
}

func (c *PingResp) Marshall() bytes.Buffer {
	var packet bytes.Buffer
	fixedHeaderFlags := 0

	fixedHeader := (pingrespType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(0))

	return packet
}

func decodePingResp(payload []byte) *PingResp {
	ping := new(PingResp)
	return ping
}
