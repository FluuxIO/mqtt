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
	protocolName  string
	protocolLevel int
	keepalive     int
	clientID      string
	cleanSession  bool
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

func (c *Connect) SetCleanSession(flag bool) {
	c.cleanSession = flag
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

	variablePart.Write(encodeProtocolName(c.protocolName))
	variablePart.WriteByte(encodeProtocolLevel(c.protocolLevel))
	variablePart.WriteByte(byte(connectFlags))
	variablePart.Write(encodeUint16(keepalive))
	variablePart.Write(encodeClientID(c.clientID))

	if c.willFlag && len(c.willTopic) > 0 {
		variablePart.Write(encodeString(c.willTopic))
		variablePart.Write(encodeString(c.willMessage))
	}

	if len(c.username) > 0 {
		variablePart.Write(encodeString(c.username))
		if len(c.password) > 0 {
			variablePart.Write(encodeString(c.password))
		}
	}

	fixedHeader := (connectType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

func encodeClientID(clientID string) []byte {
	id := defaultValue(clientID, defaultClientID)
	return encodeString(id)
}

func encodeProtocolName(name string) []byte {
	n := defaultValue(name, protocolName)
	return encodeString(n)
}

func defaultValue(val string, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}

func encodeProtocolLevel(level int) byte {
	if level == 0 {
		level = protocolLevel
	}
	return byte(level)
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

func decodeConnect(payload []byte) *Connect {
	connect := NewConnect()
	return connect
}

/*
// TODO
func decodeConnect(payload []byte) *Connect {
	connect := NewConnect()
	connect.Dup = int2bool(fixedHeaderFlags >> 3)
	connect.Qos = int((fixedHeaderFlags & 6) >> 1)
	connect.Retain = int2bool((fixedHeaderFlags & 1))
	var rest []byte
	connect.Topic, rest = extractNextString(payload)
	var index int
	if len(rest) > 0 {
		if connect.Qos == 1 || connect.Qos == 2 {
			offset := 2
			connect.ID = int(binary.BigEndian.Uint16(rest[:offset]))
			index = offset
		}
		if len(rest) > index {
			connect.Payload = rest[index:]
		}
	}
	return connect
}
*/
