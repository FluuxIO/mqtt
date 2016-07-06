// Direct conversion from my Elixir implementation
package packet

import (
	"bytes"
	"encoding/binary"
)

// PDUConnect is the PDU for ...
type PDUConnect struct {
	ProtocolName  string
	ProtocolLevel int
	Keepalive     int
	ClientID      string
	CleanSession  bool

	// TODO: Will should be a sub-struct
	WillFlag    bool
	WillTopic   string
	WillMessage string
	WillQOS     int
	WillRetain  bool
	Username    string
	Password    string
}

// Marshall return buffer containing serialized CONNECT MQTT control packet.
func (pdu PDUConnect) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	variablePart.Write(encodeProtocolName(pdu.ProtocolName))
	variablePart.WriteByte(encodeProtocolLevel(pdu.ProtocolLevel))
	variablePart.WriteByte(byte(pdu.connectFlag()))
	variablePart.Write(encodeUint16(uint16(pdu.Keepalive)))
	variablePart.Write(encodeClientID(pdu.ClientID))

	if pdu.WillFlag && len(pdu.WillTopic) > 0 {
		variablePart.Write(encodeString(pdu.WillTopic))
		variablePart.Write(encodeString(pdu.WillMessage))
	}

	if len(pdu.Username) > 0 {
		variablePart.Write(encodeString(pdu.Username))
		if len(pdu.Password) > 0 {
			variablePart.Write(encodeString(pdu.Password))
		}
	}

	fixedHeader := (connectType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

func (pdu PDUConnect) connectFlag() int {

	// Only set willFlag if there is actually a topic set.
	willFlag := pdu.WillFlag && len(pdu.WillTopic) >= 0

	willQOS := 0
	willRetain := false
	if willFlag {
		if pdu.WillQOS > 0 {
			willQOS = pdu.WillQOS
		}
		if pdu.WillRetain {
			willRetain = true
		}
	}

	usernameFlag, passwordFlag := false, false
	if len(pdu.Username) > 0 {
		usernameFlag = true
		if len(pdu.Password) > 0 {
			passwordFlag = true
		}
	}

	return (bool2int(passwordFlag)<<7 | bool2int(usernameFlag)<<6 | bool2int(willRetain)<<5 | willQOS<<3 |
		bool2int(willFlag)<<2 | bool2int(pdu.CleanSession)<<1)
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

//==============================================================================

type pdu_Connect struct{}

var pduConnect pdu_Connect

func (pdu_Connect) decode(payload []byte) PDUConnect {
	var pdu PDUConnect
	var rest []byte

	pdu.ProtocolName, rest = extractNextString(payload)
	pdu.ProtocolLevel = int(rest[0])

	flag := rest[1]
	pdu.CleanSession = int2bool(int((flag & 2) >> 1))
	if pdu.WillFlag = int2bool(int((flag & 4) >> 2)); pdu.WillFlag {
		pdu.WillQOS = int((flag & 24) >> 3)
		pdu.WillRetain = int2bool(int((flag & 32) >> 5))
	}
	usernameFlag := int2bool(int((flag & 64) >> 6))
	passwordFlag := int2bool(int((flag & 128) >> 7))

	pdu.Keepalive = int(binary.BigEndian.Uint16(rest[2:4]))
	payload = rest[4:]
	pdu.ClientID, payload = extractNextString(payload)

	if pdu.WillFlag {
		pdu.WillTopic, payload = extractNextString(payload)
		pdu.WillMessage, payload = extractNextString(payload)
	}

	if usernameFlag {
		pdu.Username, payload = extractNextString(payload)
	}
	if passwordFlag {
		pdu.Password, payload = extractNextString(payload)
	}

	return pdu
}
