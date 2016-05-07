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
	var buf bytes.Buffer
	return buf
}

func DecodeConnAck(payload []byte) *ConnAck {
	connAck := new(ConnAck)
	// MQTT 3.1.1: payload[0] is reserved for future use
	connAck.ReturnCode = int(payload[1])
	fmt.Printf("Return Code: %d\n", connAck.ReturnCode)
	return connAck
}
