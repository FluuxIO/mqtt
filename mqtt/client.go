package mqtt

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/processone/gomqtt/mqtt/packet"
)

const (
	timerReset = 0
)

// Client is the main structure use to connect as a client on an MQTT
// server.
type Client struct {
	// Store user defined options
	options ClientOptions
	// TCP level connection / can be replace by a TLS session after starttls
	conn         net.Conn
	status       chan Status
	pingTimer    *time.Timer
	pingTimerCtl chan int
}

type Status struct {
	Packet packet.Packet
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
	c.pingTimerCtl = make(chan int)

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

		// Ping go routine to manage keepalive timer
		go pinger(c)

		// Status routine to receive incoming data
		go receiver(c)
	}()

	return c.status
}

// TODO Serialize packet send into its own channel / go routine
// FIXME packet.Topic does not seem a good name
func (c *Client) Subscribe(topic packet.Topic) {
	subscribe := packet.NewSubscribe()
	subscribe.AddTopic(topic)
	buf := subscribe.Marshall()
	c.send(&buf)
}

func (c *Client) Unsubscribe(topic string) {
	unsubscribe := packet.NewUnsubscribe()
	unsubscribe.AddTopic(topic)
	buf := unsubscribe.Marshall()
	c.send(&buf)
}

func (c *Client) Publish(topic string, payload []byte) {
	publish := packet.NewPublish()
	publish.SetTopic(topic)
	publish.SetPayload(payload)
	buf := publish.Marshall()
	c.send(&buf)
}

// Disconnect sends DISCONNECT MQTT packet to other party
func (c *Client) Disconnect() {
	buf := packet.NewDisconnect().Marshall()
	c.send(&buf)
}

func (c *Client) send(buf *bytes.Buffer) {
	buf.WriteTo(c.conn)
	c.resetTimer()
}

func (c *Client) resetTimer() {
	c.pingTimerCtl <- timerReset
}

// Receive, decode and dispatch messages to Status channel
func receiver(c *Client) {
	var p packet.Packet
	var err error
	for {
		if p, err = packet.Read(c.conn); err != nil {
			fmt.Printf("packet read error: %q\n", err)
			break
		}
		fmt.Printf("Received: %+v\n", p)
		sendAck(c, p)
		// For now, only broadcast publish packets back to client
		if p.PacketType() == 3 { // TODO refactor not to hardcode that value
			c.status <- Status{Packet: p}
		}
	}
}

// TODO Move to another source file
func pinger(c *Client) {
	c.pingTimer = time.NewTimer(time.Duration(c.options.Keepalive) * time.Second)
	for {
		select {
		case <-c.pingTimer.C:
			pingReq := packet.NewPingReq()
			buf := pingReq.Marshall()
			buf.WriteTo(c.conn)
			c.pingTimer.Reset(time.Duration(c.options.Keepalive) * time.Second)
		case msg := <-c.pingTimerCtl:
			switch msg {
			case timerReset:
				c.pingTimer.Reset(time.Duration(c.options.Keepalive) * time.Second)
			default:
			}
		}
	}
}

// Send acks if needed, depending on packet QOS
func sendAck(c *Client, pkt packet.Packet) {
	switch p := pkt.(type) {
	case *packet.Publish:
		if p.Qos == 1 {
			puback := packet.NewPubAck(p.ID)
			buf := puback.Marshall()
			c.send(&buf)
		}
	}
}
