package mqtt // import "fluux.io/gomqtt/mqtt"

import (
	"bytes"
	"encoding/binary"
)

// ============================================================================
// CONNECT
// ============================================================================

// PDUConnect is the PDU sent from client to log into an MQTT server.
type PDUConnect struct {
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
func (pdu *PDUConnect) SetWill(topic string, message string, qos int) {
	pdu.WillFlag = true
	pdu.WillQOS = qos
	pdu.WillTopic = topic
	pdu.WillMessage = message
}

// Marshall serializes a CONNECT struct as an MQTT control packet.
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

	fixedHeader := connectType<<4 | fixedHeaderFlags
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

// Size calculates variable length part of CONNECT MQTT packets.
func (pdu PDUConnect) PayloadSize() int {
	length := 10 // This is the base length for variable part, without optional fields

	// TODO This is just formal / cosmetic, but it would be nice to support any protocol names
	length += stringSize(pdu.ClientID)
	if pdu.WillFlag {
		length += stringSize(pdu.WillTopic)
		length += stringSize(pdu.WillMessage)
	}
	if len(pdu.Username) > 0 {
		length += stringSize(pdu.Username)
		length += stringSize(pdu.Password)
	}
	return length
}

// TODO Compare number of allocations to check if we can still reduce further
// allocation numbers.
func (pdu PDUConnect) Marshall2() []byte {
	headerSize := 2
	buf := make([]byte, headerSize+pdu.PayloadSize())

	// Fixed headers
	buf[0] = connectType << 4
	buf[1] = byte(pdu.PayloadSize())

	// Variable headers
	copy(buf[2:8], encodeProtocolName(pdu.ProtocolName))
	buf[8] = encodeProtocolLevel(pdu.ProtocolLevel)
	buf[9] = byte(pdu.connectFlag())
	binary.BigEndian.PutUint16(buf[10:12], uint16(pdu.Keepalive))
	nextPos := 12 + stringSize(pdu.ClientID) // TODO Does not work for custom protocol name as position could be different
	copy(buf[12:nextPos], encodeClientID(pdu.ClientID))

	if pdu.WillFlag && len(pdu.WillTopic) > 0 {
		nextPos = copyBufferString(buf, nextPos, pdu.WillTopic)
		nextPos = copyBufferString(buf, nextPos, pdu.WillMessage)
	}

	if len(pdu.Username) > 0 {
		nextPos = copyBufferString(buf, nextPos, pdu.Username)
		if len(pdu.Password) > 0 {
			copyBufferString(buf, nextPos, pdu.Password)
		}
	}

	return buf
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
		willRetain = pdu.WillRetain
	}

	usernameFlag, passwordFlag := false, false
	if len(pdu.Username) > 0 {
		usernameFlag = true
		if len(pdu.Password) > 0 {
			passwordFlag = true
		}
	}

	return bool2int(passwordFlag)<<7 | bool2int(usernameFlag)<<6 | bool2int(willRetain)<<5 | willQOS<<3 |
		bool2int(willFlag)<<2 | bool2int(pdu.CleanSession)<<1
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

type pduConnectDecoder struct{}

var pduConnect pduConnectDecoder

func (pduConnectDecoder) decode(payload []byte) PDUConnect {
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

// ============================================================================
// CONNACK
// ============================================================================

// PDUConnAck is the PDU sent as a reply to CONNECT control packet.
// It contains the result of the CONNECT operation.
type PDUConnAck struct {
	SessionPresent bool
	ReturnCode     int
}

// Marshall serializes a CONNACK struct as an MQTT control packet.
func (pdu PDUConnAck) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	reserved := 0

	variablePart.WriteByte(byte(reserved))
	variablePart.WriteByte(byte(pdu.ReturnCode))

	fixedHeader := connackType<<4 | fixedHeaderFlags
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

func (pdu PDUConnAck) PayloadSize() int {
	return 2
}

func (pdu PDUConnAck) Marshall2() []byte {
	fixedLength := 2
	buf := make([]byte, pdu.PayloadSize()+fixedLength)

	buf[0] = connackType << 4
	buf[1] = byte(pdu.PayloadSize())
	// TODO support Session Present flag:
	buf[2] = 0 // reserved
	buf[3] = byte(pdu.ReturnCode)

	return buf
}

// ============================================================================

type connAckDecoder struct{}

var pduConnAck connAckDecoder

func (connAckDecoder) decode(payload []byte) PDUConnAck {
	return PDUConnAck{
		ReturnCode: int(payload[1]),
	}
}

// ============================================================================
// DISCONNECT
// ============================================================================

// PDUDisconnect is the PDU sent from client to notify disconnection from server.
type PDUDisconnect struct{}

// Marshall serializes a DISCONNECT struct as an MQTT control packet.
func (PDUDisconnect) Marshall() bytes.Buffer {
	var packet bytes.Buffer

	fixedHeader := disconnectType<<4 | fixedHeaderFlags
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(0))
	return packet
}

func (PDUDisconnect) Marshall2() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength)

	buf[0] = disconnectType << 4
	buf[1] = 0
	return buf
}

//==============================================================================

type pduDisconnectDecoder struct{}

var pduDisconnect pduDisconnectDecoder

func (pduDisconnectDecoder) decode(payload []byte) PDUDisconnect {
	var disconnect PDUDisconnect
	return disconnect
}

// ============================================================================
// PUBLISH
// ============================================================================

// PDUPublish is the PDU sent by client or server to initiate or deliver
// payload broadcast.
type PDUPublish struct {
	ID      int
	Dup     bool
	Qos     int
	Retain  bool
	Topic   string
	Payload []byte
}

// Marshall serializes a PUBLISH struct as an MQTT control packet.
func (pdu PDUPublish) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	variablePart.Write(encodeString(pdu.Topic))
	if pdu.Qos == 1 || pdu.Qos == 2 {
		variablePart.Write(encodeUint16(uint16(pdu.ID)))
	}
	variablePart.Write(pdu.Payload)

	fixedHeader := publishType<<4 | bool2int(pdu.Dup)<<3 | pdu.Qos<<1 | bool2int(pdu.Retain)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

// TODO Find a better name
// From spec, Size is not size of the payload but of the variable header
func (pdu PDUPublish) PayloadSize() int {
	length := stringSize(pdu.Topic)
	if pdu.Qos == 1 || pdu.Qos == 2 {
		length += 2
	}
	length += len(pdu.Payload)
	return length
}

func (pdu PDUPublish) Marshall2() []byte {
	headerSize := 2
	buf := make([]byte, headerSize+pdu.PayloadSize())

	// Header
	buf[0] = byte(publishType<<4 | bool2int(pdu.Dup)<<3 | pdu.Qos<<1 | bool2int(pdu.Retain))
	buf[1] = byte(pdu.PayloadSize())

	// Topic
	nextPos := copyBufferString(buf, 2, pdu.Topic)

	// Packet ID
	if pdu.Qos == 1 || pdu.Qos == 2 {
		binary.BigEndian.PutUint16(buf[nextPos:nextPos+2], uint16(pdu.ID))
		nextPos = nextPos + 2
	}

	// Published message payload
	payloadSize := len(pdu.Payload)
	copy(buf[nextPos:nextPos+payloadSize], pdu.Payload)

	return buf
}

//==============================================================================

type pduPublishDecoder struct{}

var pduPublish pduPublishDecoder

func (pduPublishDecoder) decode(fixedHeaderFlags int, payload []byte) PDUPublish {
	var publish PDUPublish

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

// PDUPubAck is the PDU sent by client or server as response to client PUBLISH,
// when QOS for publish is greater than 1.
type PDUPubAck struct {
	ID int
}

// Marshall serializes a PUBACK struct as an MQTT control packet.
func (pdu PDUPubAck) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	variablePart.Write(encodeUint16(uint16(pdu.ID)))

	fixedHeaderFlags := 0
	fixedHeader := pubackType<<4 | fixedHeaderFlags
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

func (pdu PDUPubAck) PayloadSize() int {
	return 2
}

func (pdu PDUPubAck) Marshall2() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength+pdu.PayloadSize())

	// Header
	buf[0] = pubackType << 4
	buf[1] = byte(pdu.PayloadSize())

	// Packet ID
	binary.BigEndian.PutUint16(buf[2:4], uint16(pdu.ID))

	return buf
}

//==============================================================================

type pduPubAckDecoder struct{}

var pduPubAck pduPubAckDecoder

func (pduPubAckDecoder) decode(payload []byte) PDUPubAck {
	return PDUPubAck{
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

// PDUSubscribe is the PDU sent by client to subscribe to one or more topics.
type PDUSubscribe struct {
	ID     int
	Topics []Topic
}

// Marshall serializes a SUBSCRIBE struct as an MQTT control packet.
func (pdu PDUSubscribe) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	// Empty topic list is incorrect. Server must disconnect.
	if len(pdu.Topics) == 0 {
		return packet
	}

	variablePart.Write(encodeUint16(uint16(pdu.ID)))

	for _, topic := range pdu.Topics {
		variablePart.Write(encodeString(topic.Name))
		// TODO Check that QOS is valid
		variablePart.WriteByte(byte(topic.QOS))
	}

	fixedHeaderFlags := 2 // mandatory value
	fixedHeader := subscribeType<<4 | fixedHeaderFlags
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

func (pdu PDUSubscribe) PayloadSize() int {
	length := 2
	for _, topic := range pdu.Topics {
		length += stringSize(topic.Name) + 1
	}
	return length
}

func (pdu PDUSubscribe) Marshall2() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength+pdu.PayloadSize())

	// Header
	fixedHeaderFlags := 2 // mandatory value
	buf[0] = byte(subscribeType<<4 | fixedHeaderFlags)
	buf[1] = byte(pdu.PayloadSize())

	// Packet ID
	binary.BigEndian.PutUint16(buf[2:4], uint16(pdu.ID))

	// Topic filters
	nextPos := 4
	for _, topic := range pdu.Topics {
		nextPos = copyBufferString(buf, nextPos, topic.Name)
		buf[nextPos] = byte(topic.QOS)
		nextPos += 1
	}

	return buf
}

//==============================================================================

type pduSubscribeDecoder struct{}

var pduSubscribe pduSubscribeDecoder

func (pduSubscribeDecoder) decode(payload []byte) PDUSubscribe {
	subscribe := PDUSubscribe{}
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

// PDUSubAck is the PDU sent by server to acknowledge client SUBSCRIBE.
type PDUSubAck struct {
	ID          int
	ReturnCodes []int
}

// Marshall serializes a SUBACK struct as an MQTT control packet.
func (pdu PDUSubAck) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	variablePart.Write(encodeUint16(uint16(pdu.ID)))

	for _, rc := range pdu.ReturnCodes {
		variablePart.WriteByte(byte(rc))
	}

	fixedHeaderFlags := 0
	fixedHeader := subackType<<4 | fixedHeaderFlags
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

func (pdu PDUSubAck) PayloadSize() int {
	length := 2
	length += len(pdu.ReturnCodes)
	return length
}

func (pdu PDUSubAck) Marshall2() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength+pdu.PayloadSize())

	// Header
	buf[0] = byte(subackType << 4)
	buf[1] = byte(pdu.PayloadSize())

	// Packet ID
	binary.BigEndian.PutUint16(buf[2:4], uint16(pdu.ID))

	// Return codes
	nextPos := 4
	for _, rc := range pdu.ReturnCodes {
		buf[nextPos] = byte(rc)
		nextPos += 1
	}

	return buf
}

//==============================================================================

type pduSubAckDecoder struct{}

var pduSubAck pduSubAckDecoder

// We likely want to keep in memory current subscription with their state
// Client could read the current subscription state map to read the status of each subscription.
// We should probably return error if a subscription is rejected or if
// one of the QOS is lower than the level we asked for.
//
// TODO How to return all code backs to client using the library ?
func (pduSubAckDecoder) decode(payload []byte) PDUSubAck {
	var suback PDUSubAck

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

// PDUUnsubscribe is the PDU sent by client to unsubscribe from one or more topics.
type PDUUnsubscribe struct {
	ID     int
	Topics []string
}

// Marshall serializes a UNSUBSCRIBE struct as an MQTT control packet.
func (pdu PDUUnsubscribe) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	// Empty topic list is incorrect. Server must disconnect.
	if len(pdu.Topics) == 0 {
		return packet
	}

	variablePart.Write(encodeUint16(uint16(pdu.ID)))

	for _, topic := range pdu.Topics {
		variablePart.Write(encodeString(topic))
	}

	fixedHeaderFlags := 2 // mandatory value
	fixedHeader := unsubscribeType<<4 | fixedHeaderFlags
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

func (pdu PDUUnsubscribe) PayloadSize() int {
	length := 2
	for _, topic := range pdu.Topics {
		length += stringSize(topic)
	}
	return length
}

func (pdu PDUUnsubscribe) Marshall2() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength+pdu.PayloadSize())

	// Header
	fixedHeaderFlags := 2 // mandatory value
	buf[0] = byte(unsubscribeType<<4 | fixedHeaderFlags)
	buf[1] = byte(pdu.PayloadSize())

	// Packet ID
	binary.BigEndian.PutUint16(buf[2:4], uint16(pdu.ID))

	// Topics name
	nextPos := 4
	for _, topic := range pdu.Topics {
		nextPos = copyBufferString(buf, nextPos, topic)
	}

	return buf
}

//==============================================================================

type pduUnsubscribeDecoder struct{}

var pduUnsubscribe pduUnsubscribeDecoder

func (pduUnsubscribeDecoder) decode(payload []byte) PDUUnsubscribe {
	unsubscribe := PDUUnsubscribe{}
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

// PDUUnsubAck is the PDU sent by server to acknowledge client UNSUBSCRIBE.
type PDUUnsubAck struct {
	ID int
}

// Marshall serializes a UNSUBACK struct as an MQTT control packet.
func (pdu PDUUnsubAck) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	variablePart.Write(encodeUint16(uint16(pdu.ID)))

	fixedHeaderFlags := 2 // Mandatory value
	fixedHeader := unsubackType<<4 | fixedHeaderFlags
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

func (pdu PDUUnsubAck) PayloadSize() int {
	return 2
}

func (pdu PDUUnsubAck) Marshall2() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength+pdu.PayloadSize())

	// Header
	fixedHeaderFlags := 2 // Mandatory value
	buf[0] = byte(unsubackType<<4 | fixedHeaderFlags)
	buf[1] = byte(pdu.PayloadSize())

	// Packet ID
	binary.BigEndian.PutUint16(buf[2:4], uint16(pdu.ID))

	return buf
}

//==============================================================================

type pduUnsubAckDecoder struct{}

var pduUnsubAck pduUnsubAckDecoder

func (pduUnsubAckDecoder) decode(payload []byte) PDUUnsubAck {
	unsuback := PDUUnsubAck{}
	if len(payload) >= 2 {
		unsuback.ID = int(binary.BigEndian.Uint16(payload[:2]))
	}
	return unsuback
}

// ============================================================================
// PINGREQ
// ============================================================================

// PDUPingReq is the PDU sent from client for connection keepalive. Client expects to
// receive a PDUPingResp
type PDUPingReq struct {
}

// Marshall serializes a PINGREQ struct as an MQTT control packet.
func (pdu PDUPingReq) Marshall() bytes.Buffer {
	var packet bytes.Buffer
	fixedHeaderFlags := 0
	fixedHeader := pingreqType<<4 | fixedHeaderFlags
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(0))

	return packet
}

func (pdu PDUPingReq) Marshall2() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength)

	// Header
	buf[0] = byte(pingreqType << 4)
	buf[1] = byte(0)

	return buf
}

//==============================================================================

type pduPingReqDecoder struct{}

var pduPingReq pduPingReqDecoder

func (pduPingReqDecoder) decode(payload []byte) PDUPingReq {
	var ping PDUPingReq
	return ping
}

// ============================================================================
// PINGRESP
// ============================================================================

// PDUPingResp is the PDU sent by server as response to client PINGREQ.
type PDUPingResp struct {
}

// Marshall serializes a PINGRESP struct as an MQTT control packet.
func (pdu PDUPingResp) Marshall() bytes.Buffer {
	var packet bytes.Buffer
	fixedHeaderFlags := 0

	fixedHeader := pingrespType<<4 | fixedHeaderFlags
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(0))

	return packet
}

func (pdu PDUPingResp) Marshall2() []byte {
	fixedLength := 2
	buf := make([]byte, fixedLength)

	// Header
	buf[0] = byte(pingrespType << 4)
	buf[1] = byte(0)

	return buf
}

//==============================================================================

type pduPingRespDecoder struct{}

var pduPingResp pduPingRespDecoder

func (pduPingRespDecoder) decode(payload []byte) *PDUPingResp {
	ping := new(PDUPingResp)
	return ping
}
