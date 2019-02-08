package mqtt // import "gosrc.io/mqtt"

import (
	"encoding/binary"
)

// ============================================================================
// CONNECT
// ============================================================================

// ConnectPacket is the control packet sent from client to log into an MQTT server.
type ConnectPacket struct {
	ProtocolName  string
	ProtocolLevel int
	Keepalive     int
	ClientID      string
	CleanSession  bool

	// TODO: Should 'Will' be a sub-struct ?
	WillFlag    bool
	WillTopic   string
	WillMessage string
	WillQOS     int
	WillRetain  bool
	Username    string
	Password    string
}

// SetWill defines all the will values connect control packet at once,
// for consistency.
func (connect *ConnectPacket) SetWill(topic string, message string, qos int) {
	connect.WillFlag = true
	connect.WillQOS = qos
	connect.WillTopic = topic
	connect.WillMessage = message
}

// PayloadSize calculates variable length part of CONNECT MQTT packets.
func (connect ConnectPacket) PayloadSize() int {
	length := 10 // This is the base length for variable part, without optional fields

	// TODO This is just formal / cosmetic, but it would be nice to support any protocol names
	length += stringSize(connect.ClientID)
	if connect.WillFlag {
		length += stringSize(connect.WillTopic)
		length += stringSize(connect.WillMessage)
	}
	if len(connect.Username) > 0 {
		length += stringSize(connect.Username)
		length += stringSize(connect.Password)
	}
	return length
}

// Marshall serializes a CONNECT struct as an MQTT control packet.
func (connect ConnectPacket) Marshall() []byte {
	headerSize := 2
	buf := make([]byte, headerSize+connect.PayloadSize())

	// Fixed headers
	buf[0] = connectType << 4
	buf[1] = byte(connect.PayloadSize())

	// Variable headers
	copy(buf[2:8], encodeProtocolName(connect.ProtocolName))
	buf[8] = encodeProtocolLevel(connect.ProtocolLevel)
	buf[9] = byte(connect.connectFlag())
	binary.BigEndian.PutUint16(buf[10:12], uint16(connect.Keepalive))
	nextPos := 12 + stringSize(connect.ClientID) // TODO Does not work for custom protocol name as position could be different
	copy(buf[12:nextPos], encodeClientID(connect.ClientID))

	if connect.WillFlag && len(connect.WillTopic) > 0 {
		nextPos = copyBufferString(buf, nextPos, connect.WillTopic)
		nextPos = copyBufferString(buf, nextPos, connect.WillMessage)
	}

	if len(connect.Username) > 0 {
		nextPos = copyBufferString(buf, nextPos, connect.Username)
		if len(connect.Password) > 0 {
			copyBufferString(buf, nextPos, connect.Password)
		}
	}

	return buf
}

func (connect ConnectPacket) connectFlag() int {
	// Only set willFlag if there is actually a topic set.
	willFlag := connect.WillFlag && len(connect.WillTopic) >= 0

	willQOS := 0
	willRetain := false
	if willFlag {
		if connect.WillQOS > 0 {
			willQOS = connect.WillQOS
		}
		willRetain = connect.WillRetain
	}

	usernameFlag, passwordFlag := false, false
	if len(connect.Username) > 0 {
		usernameFlag = true
		if len(connect.Password) > 0 {
			passwordFlag = true
		}
	}

	return bool2int(passwordFlag)<<7 | bool2int(usernameFlag)<<6 | bool2int(willRetain)<<5 | willQOS<<3 |
		bool2int(willFlag)<<2 | bool2int(connect.CleanSession)<<1
}

func encodeClientID(clientID string) []byte {
	id := defaultValue(clientID, DefaultClientID)
	return encodeString(id)
}

func encodeProtocolName(name string) []byte {
	n := defaultValue(name, ProtocolName)
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
		level = ProtocolLevel
	}
	return byte(level)
}

//==============================================================================

type connectDecoder struct{}

var connectPacket connectDecoder

func (connectDecoder) decode(payload []byte) ConnectPacket {
	var connect ConnectPacket
	var rest []byte

	connect.ProtocolName, rest = extractNextString(payload)
	connect.ProtocolLevel = int(rest[0])

	flag := rest[1]
	connect.CleanSession = int2bool(int((flag & 2) >> 1))
	if connect.WillFlag = int2bool(int((flag & 4) >> 2)); connect.WillFlag {
		connect.WillQOS = int((flag & 24) >> 3)
		connect.WillRetain = int2bool(int((flag & 32) >> 5))
	}
	usernameFlag := int2bool(int((flag & 64) >> 6))
	passwordFlag := int2bool(int((flag & 128) >> 7))

	connect.Keepalive = int(binary.BigEndian.Uint16(rest[2:4]))
	payload = rest[4:]
	connect.ClientID, payload = extractNextString(payload)

	if connect.WillFlag {
		connect.WillTopic, payload = extractNextString(payload)
		connect.WillMessage, payload = extractNextString(payload)
	}

	if usernameFlag {
		connect.Username, payload = extractNextString(payload)
	}
	if passwordFlag {
		connect.Password, _ = extractNextString(payload)
	}

	return connect
}

// ============================================================================
// CONNACK
// ============================================================================

// ConnAckPacket is the control packet sent as a reply to CONNECT packet.
// It contains the result of the CONNECT operation.
type ConnAckPacket struct {
	SessionPresent bool
	ReturnCode     int
}

func (connack ConnAckPacket) PayloadSize() int {
	return 2
}

// Marshall serializes a CONNACK struct as an MQTT control packet.
func (connack ConnAckPacket) Marshall() []byte {
	fixedLength := 2
	buf := make([]byte, connack.PayloadSize()+fixedLength)

	buf[0] = connackType << 4
	buf[1] = byte(connack.PayloadSize())
	// TODO support Session Present flag:
	buf[2] = 0 // reserved
	buf[3] = byte(connack.ReturnCode)

	return buf
}

// ============================================================================

type connAckDecoder struct{}

var connAckPacket connAckDecoder

func (connAckDecoder) decode(payload []byte) ConnAckPacket {
	return ConnAckPacket{
		ReturnCode: int(payload[1]),
	}
}

// ============================================================================
// DISCONNECT
// ============================================================================

// DisconnectPacket is the control packet sent from client to notify
// disconnection from server.
type DisconnectPacket struct{}

// Marshall serializes a DISCONNECT struct as an MQTT control packet.
func (DisconnectPacket) Marshall() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength)

	buf[0] = disconnectType << 4
	buf[1] = 0
	return buf
}

//==============================================================================

type disconnectDecoder struct{}

var disconnectPacket disconnectDecoder

func (disconnectDecoder) decode(payload []byte) DisconnectPacket {
	var disconnect DisconnectPacket
	return disconnect
}

// ============================================================================
// PUBLISH
// ============================================================================

// PublishPacket is the control packet sent by client or server to initiate or
// deliver payload broadcast.
type PublishPacket struct {
	ID      int
	Dup     bool
	Qos     int
	Retain  bool
	Topic   string
	Payload []byte
}

// TODO Find a better name
// From spec, Size is not size of the payload but of the variable header
func (publish PublishPacket) PayloadSize() int {
	length := stringSize(publish.Topic)
	if publish.Qos == 1 || publish.Qos == 2 {
		length += 2
	}
	length += len(publish.Payload)
	return length
}

// Marshall serializes a PUBLISH struct as an MQTT control packet.
func (publish PublishPacket) Marshall() []byte {
	headerSize := 2
	buf := make([]byte, headerSize+publish.PayloadSize())

	// Header
	buf[0] = byte(publishType<<4 | bool2int(publish.Dup)<<3 | publish.Qos<<1 | bool2int(publish.Retain))
	buf[1] = byte(publish.PayloadSize())

	// Topic
	nextPos := copyBufferString(buf, 2, publish.Topic)

	// Packet ID
	if publish.Qos == 1 || publish.Qos == 2 {
		// Packet ID (it must be non zero, so we use 1 if value is zero to generate a valid packet)
		id := 1
		if publish.ID > id {
			id = publish.ID
		}
		binary.BigEndian.PutUint16(buf[nextPos:nextPos+2], uint16(id))
		nextPos = nextPos + 2
	}

	// Published message payload
	payloadSize := len(publish.Payload)
	copy(buf[nextPos:nextPos+payloadSize], publish.Payload)

	return buf
}

//==============================================================================

type publishDecoder struct{}

var publishPacket publishDecoder

func (publishDecoder) decode(fixedHeaderFlags int, payload []byte) PublishPacket {
	var publish PublishPacket

	publish.Dup = int2bool(fixedHeaderFlags >> 3)
	publish.Qos = (fixedHeaderFlags & 6) >> 1
	publish.Retain = int2bool(fixedHeaderFlags & 1)
	var rest []byte
	publish.Topic, rest = extractNextString(payload)
	var index int
	if len(rest) > 0 {
		if publish.Qos == 1 || publish.Qos == 2 {
			offset := 2
			publish.ID = int(binary.BigEndian.Uint16(rest[:offset]))
			index = offset
		}
		if len(rest) > index {
			publish.Payload = rest[index:]
		}
	}
	return publish
}

// ============================================================================
// PUBACK
// ============================================================================

// PubAckPacket is the control packet sent by client or server as response to
// client PUBLISH, when QOS for publish is greater than 1.
type PubAckPacket struct {
	ID int
}

func (puback PubAckPacket) PayloadSize() int {
	return 2
}

// Marshall serializes a PUBACK struct as an MQTT control packet.
func (puback PubAckPacket) Marshall() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength+puback.PayloadSize())

	// Header
	buf[0] = pubackType << 4
	buf[1] = byte(puback.PayloadSize())

	// Packet ID
	binary.BigEndian.PutUint16(buf[2:4], uint16(puback.ID))

	return buf
}

//==============================================================================

type pubAckDecoder struct{}

var pubAckPacket pubAckDecoder

func (pubAckDecoder) decode(payload []byte) PubAckPacket {
	return PubAckPacket{
		ID: int(binary.BigEndian.Uint16(payload[:2])),
	}
}

// ============================================================================
// SUBSCRIBE
// ============================================================================

// Topic is a channel used for publish and subscribe in MQTT protocol.
type Topic struct {
	Name string
	QOS  int
}

// SubscribePacket is the control packet sent by client to subscribe to one or
// more topics.
type SubscribePacket struct {
	ID     int
	Topics []Topic
}

func (subscribe SubscribePacket) PayloadSize() int {
	length := 2
	for _, topic := range subscribe.Topics {
		length += stringSize(topic.Name) + 1
	}
	return length
}

// Marshall serializes a SUBSCRIBE struct as an MQTT control packet.
func (subscribe SubscribePacket) Marshall() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength+subscribe.PayloadSize())

	// Header
	fixedHeaderFlags := 2 // mandatory value
	buf[0] = byte(subscribeType<<4 | fixedHeaderFlags)
	buf[1] = byte(subscribe.PayloadSize())

	// Packet ID (it must be non zero, so we use 1 if value is zero to generate a valid packet)
	id := 1
	if subscribe.ID > id {
		id = subscribe.ID
	}
	binary.BigEndian.PutUint16(buf[2:4], uint16(id))

	// Topic filters
	nextPos := 4
	for _, topic := range subscribe.Topics {
		nextPos = copyBufferString(buf, nextPos, topic.Name)
		buf[nextPos] = byte(topic.QOS)
		nextPos++
	}

	return buf
}

func (subscribe SubscribePacket) PacketID() int {
	return subscribe.ID
}

//==============================================================================

type subscribeDecoder struct{}

var subscribePacket subscribeDecoder

func (subscribeDecoder) decode(payload []byte) SubscribePacket {
	subscribe := SubscribePacket{}
	subscribe.ID = int(binary.BigEndian.Uint16(payload[:2]))

	for remaining := payload[2:]; len(remaining) > 0; {
		topic := Topic{}
		var rest []byte
		topic.Name, rest = extractNextString(remaining)
		topic.QOS = int(rest[0])
		subscribe.Topics = append(subscribe.Topics, topic)
		remaining = rest[1:]
	}

	return subscribe
}

// ============================================================================
// SUBACK
// ============================================================================

// SubAckPacket is the control packet sent by server to acknowledge client
// SUBSCRIBE.
type SubAckPacket struct {
	ID          int
	ReturnCodes []int
}

func (suback SubAckPacket) PayloadSize() int {
	length := 2
	length += len(suback.ReturnCodes)
	return length
}

// Marshall serializes a SUBACK struct as an MQTT control packet.
func (suback SubAckPacket) Marshall() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength+suback.PayloadSize())

	// Header
	buf[0] = byte(subackType << 4)
	buf[1] = byte(suback.PayloadSize())

	// Packet ID
	binary.BigEndian.PutUint16(buf[2:4], uint16(suback.ID))

	// Return codes
	nextPos := 4
	for _, rc := range suback.ReturnCodes {
		buf[nextPos] = byte(rc)
		nextPos++
	}

	return buf
}

func (suback SubAckPacket) ResponseID() int {
	return suback.ID
}

//==============================================================================

type subAckDecoder struct{}

var subAckPacket subAckDecoder

// We likely want to keep in memory current subscription with their state
// Client could read the current subscription state map to read the status of each subscription.
// We should probably return error if a subscription is rejected or if
// one of the QOS is lower than the level we asked for.
func (subAckDecoder) decode(payload []byte) SubAckPacket {
	var suback SubAckPacket

	suback.ID = int(binary.BigEndian.Uint16(payload[:2]))
	if len(payload) >= 2 {
		for _, v := range payload[2:] {
			suback.ReturnCodes = append(suback.ReturnCodes, int(v))
		}
	}
	return suback
}

// ============================================================================
// UNSUBSCRIBE
// ============================================================================

// UnsubscribePacket is the control packet sent by client to unsubscribe from
// one or more topics.
type UnsubscribePacket struct {
	ID     int
	Topics []string
}

func (unsubscribe UnsubscribePacket) PayloadSize() int {
	length := 2
	for _, topic := range unsubscribe.Topics {
		length += stringSize(topic)
	}
	return length
}

// Marshall serializes a UNSUBSCRIBE struct as an MQTT control packet.
func (unsubscribe UnsubscribePacket) Marshall() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength+unsubscribe.PayloadSize())

	// Header
	fixedHeaderFlags := 2 // mandatory value
	buf[0] = byte(unsubscribeType<<4 | fixedHeaderFlags)
	buf[1] = byte(unsubscribe.PayloadSize())

	// Packet ID (it must be non zero, so we use 1 if value is zero to generate a valid packet)
	id := 1
	if unsubscribe.ID > id {
		id = unsubscribe.ID
	}
	binary.BigEndian.PutUint16(buf[2:4], uint16(id))

	// Topics name
	nextPos := 4
	for _, topic := range unsubscribe.Topics {
		nextPos = copyBufferString(buf, nextPos, topic)
	}

	return buf
}

func (unsubscribe UnsubscribePacket) PacketID() int {
	return unsubscribe.ID
}

//==============================================================================

type unsubscribeDecoder struct{}

var unsubscribePacket unsubscribeDecoder

func (unsubscribeDecoder) decode(payload []byte) UnsubscribePacket {
	unsubscribe := UnsubscribePacket{}
	unsubscribe.ID = int(binary.BigEndian.Uint16(payload[:2]))

	for remaining := payload[2:]; len(remaining) > 0; {
		var topic string
		topic, remaining = extractNextString(remaining)
		unsubscribe.Topics = append(unsubscribe.Topics, topic)
	}

	return unsubscribe
}

// ============================================================================
// UNSUBACK
// ============================================================================

// UnsubAckPacket is the control packet sent by server to acknowledge client
// UNSUBSCRIBE.
type UnsubAckPacket struct {
	ID int
}

func (unsub UnsubAckPacket) PayloadSize() int {
	return 2
}

// Marshall serializes a UNSUBACK struct as an MQTT control packet.
func (unsub UnsubAckPacket) Marshall() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength+unsub.PayloadSize())

	// Header
	fixedHeaderFlags := 2 // Mandatory value
	buf[0] = byte(unsubackType<<4 | fixedHeaderFlags)
	buf[1] = byte(unsub.PayloadSize())

	// Packet ID
	binary.BigEndian.PutUint16(buf[2:4], uint16(unsub.ID))

	return buf
}

func (unsub UnsubAckPacket) ResponseID() int {
	return unsub.ID
}

//==============================================================================

type unsubAckDecoder struct{}

var unsubAckPacket unsubAckDecoder

func (unsubAckDecoder) decode(payload []byte) UnsubAckPacket {
	unsuback := UnsubAckPacket{}
	if len(payload) >= 2 {
		unsuback.ID = int(binary.BigEndian.Uint16(payload[:2]))
	}
	return unsuback
}

// ============================================================================
// PINGREQ
// ============================================================================

// PingReqPacket is the control packet sent from client for connection
////// keepalive. Client expects to receive a PingRespPacket
type PingReqPacket struct{}

// Marshall serializes a PINGREQ struct as an MQTT control packet.
func (pingreq PingReqPacket) Marshall() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength)

	// Header
	buf[0] = byte(pingreqType << 4)
	buf[1] = byte(0)

	return buf
}

//==============================================================================

type pingReqDecoder struct{}

var pingReqPacket pingReqDecoder

func (pingReqDecoder) decode(payload []byte) PingReqPacket {
	var ping PingReqPacket
	return ping
}

// ============================================================================
// PINGRESP
// ============================================================================

// PingRespPacket is the control packet sent by server as response to client
// PINGREQ.
type PingRespPacket struct {
}

// Marshall serializes a PINGRESP struct as an MQTT control packet.
func (pdu PingRespPacket) Marshall() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength)

	// Header
	buf[0] = byte(pingrespType << 4)
	buf[1] = byte(0)

	return buf
}

//==============================================================================

type pingRespDecoder struct{}

var pingRespPacket pingRespDecoder

func (pingRespDecoder) decode(payload []byte) PingRespPacket {
	var ping PingRespPacket
	return ping
}
