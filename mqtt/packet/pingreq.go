package packet

import "bytes"

type PingReq struct {
}

func NewPingReq() PingReq {
	return PingReq{}
}

func (c *PingReq) Marshall() bytes.Buffer {
	var packet bytes.Buffer
	fixedHeaderFlags := 0

	fixedHeader := (pingreqType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(0))

	return packet
}

func decodePingReq(payload []byte) PingReq {
	var ping PingReq
	return ping
}
