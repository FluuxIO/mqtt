package packet

import "bytes"

type PDUPingResp struct {
}

func (c PDUPingResp) Marshall() bytes.Buffer {
	var packet bytes.Buffer
	fixedHeaderFlags := 0

	fixedHeader := (pingrespType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(0))

	return packet
}

//==============================================================================

type pdu_PingResp struct{}

var pduPingResp pdu_PingResp

func (pdu_PingResp) decode(payload []byte) *PDUPingResp {
	ping := new(PDUPingResp)
	return ping
}
