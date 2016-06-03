/*
MQTT package implements MQTT protocol. It can be use as a client library to write MQTT clients in Go.
*/
package mqtt

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/processone/gomqtt/mqtt/packet"
)

const (
	stateConnecting = iota
	stateConnected
	stateReconnecting
	stateDisconnected
)

var (
	ErrWrongConnectResponse = errors.New("incorrect connect response")
)

//=============================================================================

// OptConnect defines optional MQTT connection parameters.
// MQTT client libraries will default to sensible values.
type OptConnect struct {
	ClientID     string
	Keepalive    int
	CleanSession bool
}

// Config provides a data structure of required configuration parameters for MQTT connection
type Config struct {
	Address string

	// *************************************************************************
	// ** Not Required, optional                                              **
	// *************************************************************************
	OptConnect
}

//=============================================================================

// Message encapsulates Publish MQTT payload from the MQTT client perspective.
type Message struct {
	Topic   string
	Payload []byte
}

//=============================================================================

// Client is the main structure use to connect as a client on an MQTT
// server.
type Client struct {
	Config

	mu      sync.RWMutex
	backoff backoff
	message chan *Message
	sender  sender
}

// TODO split channel between status signals (informing about the state of the client) and message received (informing
// about the publish we have received)
// We also should abstract the Message to hide the details of the protocol from the developer client: MQTT protocol could
// change on the wire, but we can likely keep the same internal format for publish messages received.

// NewClient generates a new MQTT client with default parameters. Server must be set as we cannot find relevant default
// value for server
func New(address string) *Client {
	return &Client{
		Config: Config{
			Address: address,
			// As default we do not want to use a persistent session:
			OptConnect: OptConnect{
				Keepalive:    30,
				CleanSession: true,
			},
		},
	}
}

// ============================================================================

// Connect initiates synchronous connection to MQTT server and performs MQTT
// connect handshake.
func (c *Client) Connect() error {
	return c.connect(false)
}

// Disconnect sends DISCONNECT MQTT packet to other party and clean up the client
// state.
func (c *Client) Disconnect() {
	c.send(packet.NewDisconnect())
	// TODO Properly terminates receiver and sender and close message channel
}

// ============================================================================

// TODO Serialize packet send into its own channel / go routine
//
// FIXME(mr) packet.Topic does not seem a good name
func (c *Client) Subscribe(topic packet.Topic) {
	subscribe := packet.NewSubscribe()
	subscribe.AddTopic(topic)
	c.send(subscribe)
}

func (c *Client) Unsubscribe(topic string) {
	unsubscribe := packet.NewUnsubscribe()
	unsubscribe.AddTopic(topic)
	c.send(unsubscribe)
}

// ============================================================================

func (c *Client) Publish(topic string, payload []byte) {
	publish := packet.NewPublish()
	publish.SetTopic(topic)
	publish.SetPayload(payload)
	c.send(publish)
}

// ReadNext can be called from client to readNext message
func (c *Client) ReadNext() *Message {
	return <-c.message
}

// ============================================================================
// Internal

func (c *Client) connect(retry bool) error {
	fmt.Println("Trying to connect")
	conn, err := net.DialTimeout("tcp", c.Address, 5*time.Second)
	if err != nil {
		if !retry {
			return err
		}
		// Sleep with exponential backoff (and jitter) before triggering reconnect:
		time.AfterFunc(c.backoff.Duration(), func() { c.connect(retry) })
		return nil
	}

	// 1. Open session - Login
	// Send connect packet
	connectPacket := packet.NewConnect()
	// FIXME: client does not work properly if keepalive is 0
	connectPacket.SetKeepalive(c.Keepalive)
	connectPacket.SetClientID(c.ClientID)
	connectPacket.SetCleanSession(c.CleanSession)
	buf := connectPacket.Marshall()
	buf.WriteTo(conn)

	if connack, err := packet.Read(conn); err != nil {
		return err
	} else {
		switch p := connack.(type) {
		case *packet.ConnAck:
			switch p.ReturnCode {
			case packet.ConnAccepted:
			default:
				return packet.ConnAckError(p.ReturnCode)
			}
		default:
			return ErrWrongConnectResponse
		}
	}

	// 2. Connected. We set environment up
	c.backoff.Reset()
	// For now we do not need to change the message channel
	if c.message == nil {
		c.message = make(chan *Message)
	}

	c.setSender(initSender(conn, c.Keepalive))
	// Start routine to receive incoming data
	tearDown := initReceiver(conn, c.message, c.sender)
	// Routine to watch for disconnect event and trigger reconnect
	go c.disconnected(tearDown, c.sender.done)
	return nil
}

// get receiver tearDown signal, clean client state and trigger reconnect
func (c *Client) disconnected(receiverDone <-chan struct{}, senderDone <-chan struct{}) {
	select {
	case <-receiverDone:
		c.sender.quit <- struct{}{}
	case <-senderDone:
		// We do nothing for now: As the sender closes socket, this should be enough to have read Loop
		// fail and properly shutdown process.
	}
	go c.connect(true)
}

// ============================================================================

func (c *Client) send(packet packet.Marshaller) {
	buf := packet.Marshall()
	sender := c.getSender()
	sender.send(&buf)
}

// ============================================================================
// sender setter / getter
// TODO: Probably it is not sended as we probably do not need to really reset
// sender on reconnect

// setSender is used to protect against race on reconnect.
func (c *Client) setSender(sender sender) {
	c.mu.Lock()
	{
		c.sender = sender
	}
	c.mu.Unlock()
}

// getSender is used to protect against race on reconnect.
func (c *Client) getSender() sender {
	var s sender
	c.mu.RLock()
	{
		s = c.sender
	}
	c.mu.RUnlock()
	return s
}
