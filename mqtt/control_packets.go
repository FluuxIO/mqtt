package mqtt // import "fluux.io/gomqtt/mqtt"

import (
	"encoding/binary"
)

// ============================================================================
// CONNECT
// ============================================================================

// CPConnect is the PDU sent from client to log into an MQTT server.
type CPConnect struct {
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
func (cp *CPConnect) SetWill(topic string, message string, qos int) {
	cp.WillFlag = true
	cp.WillQOS = qos
	cp.WillTopic = topic
	cp.WillMessage = message
}

// Size calculates variable length part of CONNECT MQTT packets.
func (cp CPConnect) PayloadSize() int {
	length := 10 // This is the base length for variable part, without optional fields

	// TODO This is just formal / cosmetic, but it would be nice to support any protocol names
	length += stringSize(cp.ClientID)
	if cp.WillFlag {
		length += stringSize(cp.WillTopic)
		length += stringSize(cp.WillMessage)
	}
	if len(cp.Username) > 0 {
		length += stringSize(cp.Username)
		length += stringSize(cp.Password)
	}
	return length
}

// Marshall serializes a CONNECT struct as an MQTT control packet.
func (cp CPConnect) Marshall() []byte {
	headerSize := 2
	buf := make([]byte, headerSize+cp.PayloadSize())

	// Fixed headers
	buf[0] = connectType << 4
	buf[1] = byte(cp.PayloadSize())

	// Variable headers
	copy(buf[2:8], encodeProtocolName(cp.ProtocolName))
	buf[8] = encodeProtocolLevel(cp.ProtocolLevel)
	buf[9] = byte(cp.connectFlag())
	binary.BigEndian.PutUint16(buf[10:12], uint16(cp.Keepalive))
	nextPos := 12 + stringSize(cp.ClientID) // TODO Does not work for custom protocol name as position could be different
	copy(buf[12:nextPos], encodeClientID(cp.ClientID))

	if cp.WillFlag && len(cp.WillTopic) > 0 {
		nextPos = copyBufferString(buf, nextPos, cp.WillTopic)
		nextPos = copyBufferString(buf, nextPos, cp.WillMessage)
	}

	if len(cp.Username) > 0 {
		nextPos = copyBufferString(buf, nextPos, cp.Username)
		if len(cp.Password) > 0 {
			copyBufferString(buf, nextPos, cp.Password)
		}
	}

	return buf
}

func (cp CPConnect) connectFlag() int {
	// Only set willFlag if there is actually a topic set.
	willFlag := cp.WillFlag && len(cp.WillTopic) >= 0

	willQOS := 0
	willRetain := false
	if willFlag {
		if cp.WillQOS > 0 {
			willQOS = cp.WillQOS
		}
		willRetain = cp.WillRetain
	}

	usernameFlag, passwordFlag := false, false
	if len(cp.Username) > 0 {
		usernameFlag = true
		if len(cp.Password) > 0 {
			passwordFlag = true
		}
	}

	return bool2int(passwordFlag)<<7 | bool2int(usernameFlag)<<6 | bool2int(willRetain)<<5 | willQOS<<3 |
		bool2int(willFlag)<<2 | bool2int(cp.CleanSession)<<1
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

type cpConnectDecoder struct{}

var cpConnect cpConnectDecoder

func (cpConnectDecoder) decode(payload []byte) CPConnect {
	var cp CPConnect
	var rest []byte

	cp.ProtocolName, rest = extractNextString(payload)
	cp.ProtocolLevel = int(rest[0])

	flag := rest[1]
	cp.CleanSession = int2bool(int((flag & 2) >> 1))
	if cp.WillFlag = int2bool(int((flag & 4) >> 2)); cp.WillFlag {
		cp.WillQOS = int((flag & 24) >> 3)
		cp.WillRetain = int2bool(int((flag & 32) >> 5))
	}
	usernameFlag := int2bool(int((flag & 64) >> 6))
	passwordFlag := int2bool(int((flag & 128) >> 7))

	cp.Keepalive = int(binary.BigEndian.Uint16(rest[2:4]))
	payload = rest[4:]
	cp.ClientID, payload = extractNextString(payload)

	if cp.WillFlag {
		cp.WillTopic, payload = extractNextString(payload)
		cp.WillMessage, payload = extractNextString(payload)
	}

	if usernameFlag {
		cp.Username, payload = extractNextString(payload)
	}
	if passwordFlag {
		cp.Password, payload = extractNextString(payload)
	}

	return cp
}

// ============================================================================
// CONNACK
// ============================================================================

// CPConnAck is the PDU sent as a reply to CONNECT control packet.
// It contains the result of the CONNECT operation.
type CPConnAck struct {
	SessionPresent bool
	ReturnCode     int
}

func (cp CPConnAck) PayloadSize() int {
	return 2
}

// Marshall serializes a CONNACK struct as an MQTT control packet.
func (cp CPConnAck) Marshall() []byte {
	fixedLength := 2
	buf := make([]byte, cp.PayloadSize()+fixedLength)

	buf[0] = connackType << 4
	buf[1] = byte(cp.PayloadSize())
	// TODO support Session Present flag:
	buf[2] = 0 // reserved
	buf[3] = byte(cp.ReturnCode)

	return buf
}

// ============================================================================

type connAckDecoder struct{}

var cpConnAck connAckDecoder

func (connAckDecoder) decode(payload []byte) CPConnAck {
	return CPConnAck{
		ReturnCode: int(payload[1]),
	}
}

// ============================================================================
// DISCONNECT
// ============================================================================

// CPDisconnect is the PDU sent from client to notify disconnection from server.
type CPDisconnect struct{}

// Marshall serializes a DISCONNECT struct as an MQTT control packet.
func (CPDisconnect) Marshall() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength)

	buf[0] = disconnectType << 4
	buf[1] = 0
	return buf
}

//==============================================================================

type cpDisconnectDecoder struct{}

var cpDisconnect cpDisconnectDecoder

func (cpDisconnectDecoder) decode(payload []byte) CPDisconnect {
	var disconnect CPDisconnect
	return disconnect
}

// ============================================================================
// PUBLISH
// ============================================================================

// CPPublish is the PDU sent by client or server to initiate or deliver
// payload broadcast.
type CPPublish struct {
	ID      int
	Dup     bool
	Qos     int
	Retain  bool
	Topic   string
	Payload []byte
}

// TODO Find a better name
// From spec, Size is not size of the payload but of the variable header
func (cp CPPublish) PayloadSize() int {
	length := stringSize(cp.Topic)
	if cp.Qos == 1 || cp.Qos == 2 {
		length += 2
	}
	length += len(cp.Payload)
	return length
}

// Marshall serializes a PUBLISH struct as an MQTT control packet.
func (cp CPPublish) Marshall() []byte {
	headerSize := 2
	buf := make([]byte, headerSize+cp.PayloadSize())

	// Header
	buf[0] = byte(publishType<<4 | bool2int(cp.Dup)<<3 | cp.Qos<<1 | bool2int(cp.Retain))
	buf[1] = byte(cp.PayloadSize())

	// Topic
	nextPos := copyBufferString(buf, 2, cp.Topic)

	// Packet ID
	if cp.Qos == 1 || cp.Qos == 2 {
		binary.BigEndian.PutUint16(buf[nextPos:nextPos+2], uint16(cp.ID))
		nextPos = nextPos + 2
	}

	// Published message payload
	payloadSize := len(cp.Payload)
	copy(buf[nextPos:nextPos+payloadSize], cp.Payload)

	return buf
}

//==============================================================================

type cpPublishDecoder struct{}

var cpPublish cpPublishDecoder

func (cpPublishDecoder) decode(fixedHeaderFlags int, payload []byte) CPPublish {
	var publish CPPublish

	publish.Dup = int2bool(fixedHeaderFlags >> 3)
	publish.Qos = int((fixedHeaderFlags & 6) >> 1)
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

// CPPubAck is the PDU sent by client or server as response to client PUBLISH,
// when QOS for publish is greater than 1.
type CPPubAck struct {
	ID int
}

func (cp CPPubAck) PayloadSize() int {
	return 2
}

// Marshall serializes a PUBACK struct as an MQTT control packet.
func (cp CPPubAck) Marshall() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength+cp.PayloadSize())

	// Header
	buf[0] = pubackType << 4
	buf[1] = byte(cp.PayloadSize())

	// Packet ID
	binary.BigEndian.PutUint16(buf[2:4], uint16(cp.ID))

	return buf
}

//==============================================================================

type cpPubAckDecoder struct{}

var cpPubAck cpPubAckDecoder

func (cpPubAckDecoder) decode(payload []byte) CPPubAck {
	return CPPubAck{
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

// CPSubscribe is the PDU sent by client to subscribe to one or more topics.
type CPSubscribe struct {
	ID     int
	Topics []Topic
}

func (cp CPSubscribe) PayloadSize() int {
	length := 2
	for _, topic := range cp.Topics {
		length += stringSize(topic.Name) + 1
	}
	return length
}

// Marshall serializes a SUBSCRIBE struct as an MQTT control packet.
func (cp CPSubscribe) Marshall() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength+cp.PayloadSize())

	// Header
	fixedHeaderFlags := 2 // mandatory value
	buf[0] = byte(subscribeType<<4 | fixedHeaderFlags)
	buf[1] = byte(cp.PayloadSize())

	// Packet ID
	binary.BigEndian.PutUint16(buf[2:4], uint16(cp.ID))

	// Topic filters
	nextPos := 4
	for _, topic := range cp.Topics {
		nextPos = copyBufferString(buf, nextPos, topic.Name)
		buf[nextPos] = byte(topic.QOS)
		nextPos += 1
	}

	return buf
}

//==============================================================================

type cpSubscribeDecoder struct{}

var cpSubscribe cpSubscribeDecoder

func (cpSubscribeDecoder) decode(payload []byte) CPSubscribe {
	subscribe := CPSubscribe{}
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

// CPSubAck is the PDU sent by server to acknowledge client SUBSCRIBE.
type CPSubAck struct {
	ID          int
	ReturnCodes []int
}

func (cp CPSubAck) PayloadSize() int {
	length := 2
	length += len(cp.ReturnCodes)
	return length
}

// Marshall serializes a SUBACK struct as an MQTT control packet.
func (cp CPSubAck) Marshall() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength+cp.PayloadSize())

	// Header
	buf[0] = byte(subackType << 4)
	buf[1] = byte(cp.PayloadSize())

	// Packet ID
	binary.BigEndian.PutUint16(buf[2:4], uint16(cp.ID))

	// Return codes
	nextPos := 4
	for _, rc := range cp.ReturnCodes {
		buf[nextPos] = byte(rc)
		nextPos += 1
	}

	return buf
}

//==============================================================================

type cpSubAckDecoder struct{}

var cpSubAck cpSubAckDecoder

// We likely want to keep in memory current subscription with their state
// Client could read the current subscription state map to read the status of each subscription.
// We should probably return error if a subscription is rejected or if
// one of the QOS is lower than the level we asked for.
//
// TODO How to return all code backs to client using the library ?
func (cpSubAckDecoder) decode(payload []byte) CPSubAck {
	var suback CPSubAck

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

// CPUnsubscribe is the PDU sent by client to unsubscribe from one or more topics.
type CPUnsubscribe struct {
	ID     int
	Topics []string
}

func (cp CPUnsubscribe) PayloadSize() int {
	length := 2
	for _, topic := range cp.Topics {
		length += stringSize(topic)
	}
	return length
}

// Marshall serializes a UNSUBSCRIBE struct as an MQTT control packet.
func (cp CPUnsubscribe) Marshall() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength+cp.PayloadSize())

	// Header
	fixedHeaderFlags := 2 // mandatory value
	buf[0] = byte(unsubscribeType<<4 | fixedHeaderFlags)
	buf[1] = byte(cp.PayloadSize())

	// Packet ID
	binary.BigEndian.PutUint16(buf[2:4], uint16(cp.ID))

	// Topics name
	nextPos := 4
	for _, topic := range cp.Topics {
		nextPos = copyBufferString(buf, nextPos, topic)
	}

	return buf
}

//==============================================================================

type cpUnsubscribeDecoder struct{}

var cpUnsubscribe cpUnsubscribeDecoder

func (cpUnsubscribeDecoder) decode(payload []byte) CPUnsubscribe {
	unsubscribe := CPUnsubscribe{}
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

// CPUnsubAck is the PDU sent by server to acknowledge client UNSUBSCRIBE.
type CPUnsubAck struct {
	ID int
}

func (cp CPUnsubAck) PayloadSize() int {
	return 2
}

// Marshall serializes a UNSUBACK struct as an MQTT control packet.
func (cp CPUnsubAck) Marshall() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength+cp.PayloadSize())

	// Header
	fixedHeaderFlags := 2 // Mandatory value
	buf[0] = byte(unsubackType<<4 | fixedHeaderFlags)
	buf[1] = byte(cp.PayloadSize())

	// Packet ID
	binary.BigEndian.PutUint16(buf[2:4], uint16(cp.ID))

	return buf
}

//==============================================================================

type cpUnsubAckDecoder struct{}

var cpUnsubAck cpUnsubAckDecoder

func (cpUnsubAckDecoder) decode(payload []byte) CPUnsubAck {
	unsuback := CPUnsubAck{}
	if len(payload) >= 2 {
		unsuback.ID = int(binary.BigEndian.Uint16(payload[:2]))
	}
	return unsuback
}

// ============================================================================
// PINGREQ
// ============================================================================

// CPPingReq is the PDU sent from client for connection keepalive. Client expects to
// receive a CPPingResp
type CPPingReq struct {
}

// Marshall serializes a PINGREQ struct as an MQTT control packet.
func (cp CPPingReq) Marshall() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength)

	// Header
	buf[0] = byte(pingreqType << 4)
	buf[1] = byte(0)

	return buf
}

//==============================================================================

type cpPingReqDecoder struct{}

var cpPingReq cpPingReqDecoder

func (cpPingReqDecoder) decode(payload []byte) CPPingReq {
	var ping CPPingReq
	return ping
}

// ============================================================================
// PINGRESP
// ============================================================================

// CPPingResp is the PDU sent by server as response to client PINGREQ.
type CPPingResp struct {
}

// Marshall serializes a PINGRESP struct as an MQTT control packet.
func (cp CPPingResp) Marshall() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength)

	// Header
	buf[0] = byte(pingrespType << 4)
	buf[1] = byte(0)

	return buf
}

//==============================================================================

type cpPingRespDecoder struct{}

var cpPingResp cpPingRespDecoder

func (cpPingRespDecoder) decode(payload []byte) CPPingResp {
	var ping CPPingResp
	return ping
}
