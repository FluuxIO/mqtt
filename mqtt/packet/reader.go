package packet

import (
	"errors"
	"fmt"
	"io"
)

// Read returns unmarshalled packet from io.Reader stream
func Read(r io.Reader) (Marshaller, error) {
	var err error
	fixedHeader := make([]byte, 1)

	if _, err = io.ReadFull(r, fixedHeader); err != nil {
		if err == io.EOF {
			fmt.Printf("Connection closed\n")
		}
		return nil, err
	}

	packetType := fixedHeader[0] >> 4
	fixedHeaderFlags := fixedHeader[0] & 15 // keep only last 4 bits

	fmt.Printf("PacketType: %d\n", packetType)
	length, _ := readRemainingLength(r)
	fmt.Printf("Length: %d\n", length)
	payload := make([]byte, length)
	if _, err = io.ReadFull(r, payload); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			fmt.Printf("Connection closed unexpectedly\n")
		}
		return nil, err
	}
	return Decode(int(packetType), int(fixedHeaderFlags), payload), err
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
