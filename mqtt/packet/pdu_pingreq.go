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

type pduPingReqDecoder struct{}

var pduPingReq pduPingReqDecoder

func (pduPingReqDecoder) decode(payload []byte) PDUPingReq {
	var ping PDUPingReq
	return ping
}
