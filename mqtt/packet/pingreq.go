package packet

import "bytes"

type PingReq struct {
}

func (p *PingReq) PacketType() int {
	return pingreqType
}

func (c *PingReq) Marshall() bytes.Buffer {
	var packet bytes.Buffer
	fixedHeaderFlags := 0

	fixedHeader := (pingreqType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(0))

	return packet
}

func decodePingReq(payload []byte) *PingReq {
	ping := new(PingReq)
	return ping
}
