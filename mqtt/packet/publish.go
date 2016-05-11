package packet

import (
	"bytes"
	"encoding/binary"
)

type Publish struct {
	id      int
	dup     bool
	qos     int
	retain  bool
	topic   string
	payload []byte
}

func (p *Publish) PacketType() int {
	return publishType
}

func (p *Publish) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	variablePart.Write(encodeString(p.topic))
	if p.qos == 1 || p.qos == 2 {
		variablePart.Write(encodeUint16(uint16(p.id)))
	}
	variablePart.Write([]byte(p.payload))

	fixedHeader := (publishType<<4 | bool2int(p.dup)<<3 | p.qos<<1 | bool2int(p.retain))
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

// Write unit test on decode / Marshall to check possible mistake in conversion
func decodePublish(fixedHeaderFlags int, payload []byte) *Publish {
	publish := NewPublish()
	publish.dup = int2bool(fixedHeaderFlags >> 3)
	publish.qos = int((fixedHeaderFlags & 6) >> 1)
	publish.retain = int2bool((fixedHeaderFlags & 1))
	var rest []byte
	publish.topic, rest = extractNextString(payload)
	var index int
	if len(rest) > 0 {
		if publish.qos == 1 || publish.qos == 2 {
			offset := 2
			publish.id = int(binary.BigEndian.Uint16(rest[:offset]))
			index = offset
		}
		if len(rest) > index {
			publish.payload = rest[index:]
		}
	}
	return publish
}

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func int2bool(i int) bool {
	if i == 1 {
		return true
	}
	return false
}

func extractNextString(data []byte) (string, []byte) {
	offset := 2
	length := int(binary.BigEndian.Uint16(data[:offset]))
	return string(data[offset : length+offset]), data[length+offset:]
}
