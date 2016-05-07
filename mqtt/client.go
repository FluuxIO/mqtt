package mqtt

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// Client is the main structure use to connect as a client on an MQTT
// server.
type Client struct {
	// Store user defined options
	options ClientOptions
	// TCP level connection / can be replace by a TLS session after starttls
	conn net.Conn
}

type Status struct {
	Err error
}

// NewClient generates a new XMPP client, based on Options passed as parameters.
// Default the port to 1883.
func NewClient(options ClientOptions) (c *Client, err error) {
	if options.Address, err = checkAddress(options.Address); err != nil {
		return
	}

	c = new(Client)
	c.options = options

	return
}

func checkAddress(addr string) (string, error) {
	var err error
	hostport := strings.Split(addr, ":")
	if len(hostport) > 2 {
		err = errors.New("too many colons in xmpp server address")
		return addr, err
	}

	// Address is composed of two parts, we are good
	if len(hostport) == 2 && hostport[1] != "" {
		return addr, err
	}

	// Port was not passed, we append default MQTT port:
	return strings.Join([]string{hostport[0], "1883"}, ":"), err
}

// Connect initiates asynchronous connection to MQTT server
func (c *Client) Connect() <-chan Status {
	out := make(chan Status)

	go func() {
		var err error
		c.conn, err = net.DialTimeout("tcp", c.options.Address, 5*time.Second)
		if err != nil {
			out <- Status{Err: err}
			return
		}
		// Send connect packet
		buf := connect()
		buf.WriteTo(c.conn)

		// TODO Check connack value before sending status to channel
		readPacket(c.conn)

		// TODO Go routine to receive incoming data

		out <- Status{}
	}()

	// Connection is ok, we now open MQTT session
	/*	if c.conn, c.Session, err = NewSession(c.conn, c.options); err != nil {
			return err
		}
	*/

	return out
}

func readPacket(r io.Reader) {
	fixedHeader := make([]byte, 1)
	io.ReadFull(r, fixedHeader)
	packetType := fixedHeader[0] >> 4
	// TODO decode flags, depending on packet type

	fmt.Printf("PacketType: %d\n", packetType)
	length, _ := ReadRemainingLength(r)
	fmt.Printf("Length: %d\n", length)
	payload := make([]byte, length)
	io.ReadFull(r, payload)
	payloadToStruct(int(packetType), payload)
}

func encodeString(str string) []byte {
	length := make([]byte, 2)
	binary.BigEndian.PutUint16(length, uint16(len(str)))
	return append(length, []byte(str)...)
}

func encodeUint16(num uint16) []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, num)
	return bytes
}

// ReadRemainingLength decodes MQTT Packet remaining length field
// Reference: http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html#_Toc398718023
func ReadRemainingLength(r io.Reader) (int, error) {
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

func payloadToStruct(packetType int, payload []byte) Packet {
	switch packetType {
	case 2:
		return decodeConnAck(payload)
	default:
		fmt.Println("Unsupported MQTT packet type")
		return nil
	}
}

/* TODO refactor to be able to test that way:
func test() {
	cp = packet.NewConnect()
	cp.usernmane = "mickael"
	buf, err = cp.Marshal()
	cp = packet.Read(buf)
}
*/
