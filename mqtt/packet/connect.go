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
	keepalive    int
	clientID     string
	cleanSession bool
	// TODO: Will should be a sub-struct
	willFlag    bool
	willTopic   string
	willMessage string
	willQOS     int
	willRetain  bool
	username    string
	password    string
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

func (c *Connect) SetWill(topic string, message string, qos int) {
	c.willFlag = true
	c.willQOS = qos
	c.willTopic = topic
	c.willMessage = message
}

// Marshall return buffer containing serialized CONNECT MQTT control packet
func (c *Connect) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	connectFlags := c.connectFlag()
	keepalive := uint16(c.keepalive)

	variablePart.Write(encodeString(protocolName))
	variablePart.WriteByte(byte(protocolLevel))
	variablePart.WriteByte(byte(connectFlags))
	variablePart.Write(encodeUint16(keepalive))
	variablePart.Write(encodeString(defineClientID(c.clientID)))

	fixedHeader := (connectType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

func defineClientID(clientID string) string {
	if clientID == "" {
		return defaultClientID
	}
	return clientID
}

func (c *Connect) connectFlag() int {
	// Only set willFlag if there is actually a topic set
	willFlag := c.willFlag && len(c.willTopic) >= 0

	willQOS := 0
	willRetain := false
	if willFlag {
		if c.willQOS > 0 {
			willQOS = c.willQOS
		}
		if c.willRetain {
			willRetain = true
		}
	}

	usernameFlag, passwordFlag := false, false
	if len(c.username) > 0 {
		usernameFlag = true
		if len(c.password) > 0 {
			passwordFlag = true
		}
	}

	flag := (bool2int(passwordFlag)<<7 | bool2int(usernameFlag)<<6 | bool2int(willRetain)<<5 | willQOS<<3 |
		bool2int(willFlag)<<2 | bool2int(c.cleanSession)<<1)
	return flag
}
