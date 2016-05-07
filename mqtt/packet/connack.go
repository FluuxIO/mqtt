package packet

import (
	"bytes"
	"fmt"
)

type ConnAck struct {
	ReturnCode int
}

func (c *ConnAck) PacketType() int {
	return 2
}

// TODO Not yet implemented
func (c *ConnAck) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	reserved := 0

	variablePart.WriteByte(byte(reserved))
	variablePart.WriteByte(byte(c.ReturnCode))

	fixedHeader := (connackType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

func decodeConnAck(payload []byte) *ConnAck {
	connAck := new(ConnAck)
	// MQTT 3.1.1: payload[0] is reserved for future use
	connAck.ReturnCode = int(payload[1])
	fmt.Printf("Return Code: %d\n", connAck.ReturnCode)
	return connAck
}
