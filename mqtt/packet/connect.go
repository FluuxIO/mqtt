// Direct conversion from my Elixir implementation
package packet

import "bytes"

const (
	fixedHeaderFlags = 0
	protocolName     = "MQTT"
	protocolLevel    = 4 // This is MQTT v3.1.1
	defaultClientID  = "GoMQTT"
)

// Connect MQTT 3.1.1 control packet
type Connect struct {
	keepalive int
	clientID  string
}

// PacketType returns packet type numerical value
func (c *Connect) PacketType() int {
	return connectType
}

func (c *Connect) SetKeepalive(keepalive int) {
	c.keepalive = keepalive
}

func (c *Connect) SetClientID(clientID string) {
	c.clientID = clientID
}

// Marshall return buffer containing serialized CONNECT MQTT control packet
func (c *Connect) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	connectFlags := 0 // TODO: support connect flag definition
	keepalive := uint16(c.keepalive)

	variablePart.Write(encodeString(protocolName))
	variablePart.WriteByte(byte(protocolLevel))
	variablePart.WriteByte(byte(connectFlags))
	variablePart.Write(encodeUint16(keepalive))
	variablePart.Write(encodeString(setDefaultClientID(c.clientID)))

	fixedHeader := (connectType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

func setDefaultClientID(clientID string) string {
	if clientID == "" {
		return defaultClientID
	}
	return clientID
}
