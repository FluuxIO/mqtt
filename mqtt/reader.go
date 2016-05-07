package mqtt

import (
	"errors"
	"fmt"
	"io"

	"github.com/processone/gomqtt/mqtt/packet"
)

func readPacket(r io.Reader) {
	fixedHeader := make([]byte, 1)
	io.ReadFull(r, fixedHeader)
	packetType := fixedHeader[0] >> 4
	// TODO decode flags, depending on packet type

	fmt.Printf("PacketType: %d\n", packetType)
	length, _ := readRemainingLength(r)
	fmt.Printf("Length: %d\n", length)
	payload := make([]byte, length)
	io.ReadFull(r, payload)
	payloadToStruct(int(packetType), payload)
}

// ReadRemainingLength decodes MQTT Packet remaining length field
// Reference: http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html#_Toc398718023
func readRemainingLength(r io.Reader) (int, error) {
	var multiplier uint32 = 1
	var value uint32
	var err error
	encodedByte := make([]byte, 1)
	for ok := true; ok; ok = (encodedByte[0]&128 != 0) {
		io.ReadFull(r, encodedByte)
		value += uint32(encodedByte[0]&127) * multiplier
		multiplier *= 128
		if multiplier > 128*128*128 {
			err = errors.New("mqtt: malformed remaining length")
			return 0, err
		}
	}

	return int(value), err
}

func payloadToStruct(packetType int, payload []byte) packet.Marshaller {
	switch packetType {
	case 2:
		return packet.DecodeConnAck(payload)
	default:
		fmt.Println("Unsupported MQTT packet type")
		return nil
	}
}
