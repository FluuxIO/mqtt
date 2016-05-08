package mqtt

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/processone/gomqtt/mqtt/packet"
)

// Client is the main structure use to connect as a client on an MQTT
// server.
type Client struct {
	// Store user defined options
	options ClientOptions
	// TCP level connection / can be replace by a TLS session after starttls
	conn      net.Conn
	status    chan Status
	pingTimer *time.Timer
}

type Status struct {
	Packet packet.Marshaller
	Err    error
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
	c.status = make(chan Status)

	go func() {
		var err error
		c.conn, err = net.DialTimeout("tcp", c.options.Address, 5*time.Second)
		if err != nil {
			c.status <- Status{Err: err}
			return
		}
		// Send connect packet
		connectPacket := packet.NewConnect()
		connectPacket.SetKeepalive(c.options.Keepalive)
		buf := connectPacket.Marshall()
		buf.WriteTo(c.conn)

		// TODO Check connack value before sending status to channel
		packet.Read(c.conn)
		c.status <- Status{}

		// TODO create ping go routine trigger by keepalive timer
		c.pingTimer = time.NewTimer(time.Duration(c.options.Keepalive) * time.Second)
		go pinger(c)

		// Status routine to receive incoming data
		go receive(c)
	}()

	return c.status
}

// TODO Send back packet to client through channel
func receive(c *Client) {
	var p packet.Marshaller
	var err error
	for {
		if p, err = packet.Read(c.conn); err != nil {
			fmt.Printf("packet read error: %q\n", err)
			break
		}
		fmt.Printf("Received: %+v\n", p)
	}
}

func pinger(c *Client) {
	for {
		<-c.pingTimer.C
		pingReq := packet.NewPingReq()
		buf := pingReq.Marshall()
		buf.WriteTo(c.conn)
		c.pingTimer.Reset(time.Duration(c.options.Keepalive) * time.Second)
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
