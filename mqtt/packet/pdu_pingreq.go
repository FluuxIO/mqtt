package packet

import "bytes"

type PDUPingReq struct {
}

func (c PDUPingReq) Marshall() bytes.Buffer {
	var packet bytes.Buffer
	fixedHeaderFlags := 0

	fixedHeader := (pingreqType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(0))

	return packet
}

//==============================================================================

type pdu_PingReq struct{}

var pduPingReq pdu_PingReq

func (pdu_PingReq) decode(payload []byte) PDUPingReq {
	var ping PDUPingReq
	return ping
}
