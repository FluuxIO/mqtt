package mqtt

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

var (
	// ErrIncorrectConnectResponse is triggered on CONNECT when server
	// does not reply with CONNACK packet.
	ErrIncorrectConnectResponse = errors.New("incorrect connect response")
)

const (
	// DefaultMQTTServer is a shortcut to define connection to local
	// server
	DefaultMQTTServer = "localhost:1883"
)

//=============================================================================

// OptConnect defines optional MQTT connection parameters.
// MQTT client libraries will default to sensible values.
type OptConnect struct {
	ProtocolLevel int
	ClientID      string
	Keepalive     int // TODO Keepalive should also probably be a time.Duration for more flexibility
	CleanSession  bool
}

// OptTCP defines TCP/IP related parameters. They are used to
// configure TCP client connection.
type OptTCP struct {
	ConnectTimeout time.Duration
}

// Config provides a data structure of required configuration
// parameters for MQTT connection
type Config struct {
	Address string

	// *************************************************************************
	// ** Not Required, optional                                              **
	// *************************************************************************
	OptConnect
	OptTCP
}

//=============================================================================

// Message encapsulates Publish MQTT payload from the MQTT client perspective.
// Message is used to abstract the detail of the MQTT protocol to the developer.
type Message struct {
	Topic   string
	Payload []byte
}

// This is a the list of events happening on the connection that the
// client can be notified about.
const (
	EventDisconnected = iota
)

// Event is a structure use to convey event changes related to client state. This
// is for example used to notify the client when the client get disconnected.
type Event struct {
	Type        int
	Description string
}

// EventHandler is use to pass events about state of the connection to
// client implementation.
type EventHandler func(Event)

//=============================================================================

// Client is the main structure use to connect as a client on an MQTT
// server.
type Client struct {
	Config

	Handler  EventHandler
	Messages chan<- Message

	mu      sync.RWMutex
	backoff Backoff
	sender  sender
}

// New generates a new MQTT client with default parameters. Address
// must be set as we cannot find relevant default value for server.
// TODO: Should messages channel be set on New ?
func New(address string) *Client {
	return &Client{
		Config: Config{
			Address: address,

			// As default we do not want to use a persistent session:
			OptConnect: OptConnect{
				ProtocolLevel: ProtocolLevel,
				Keepalive:     30,
				CleanSession:  true,
			},
			OptTCP: OptTCP{
				ConnectTimeout: 30 * time.Second,
			},
		},
	}
}

// ============================================================================

// Connect initiates synchronous connection to MQTT server and
// performs MQTT connect handshake.
//
// We must have a default channel for the client to work: If the
// connection is persistent, it is possible that we receive messages
// coming from previous connection even if we do not subscribe to
// anything in that session of the client. Having a default channel
// makes sure we always have a way to receive all messages.
//
// The channel will be closed when the session is closed and no
// further automatic reconnection will be attempted. You can use that
// close signal to reconnect the client if you wish to, immediately or
// after a delay.
//
// The channel is expected to be passed by the caller because it
// allows the caller to pass a channel with a buffer size suiting its
// own use case and expected throughput.
func (c *Client) Connect(defaultMsgChannel chan<- Message) error {
	c.Messages = defaultMsgChannel
	return c.connect()
}

// Disconnect sends DISCONNECT MQTT packet to other party and clean up
// the client state.
func (c *Client) Disconnect() {
	packet := PDUDisconnect{}
	c.send(&packet)
	// TODO Properly terminates receiver and sender and close message channel
}

// ============================================================================

// Subscribe sends SUBSCRIBE MQTT control packet.  At the moment
// suscribe are not kept in client state and are lost on reconnection.
func (c *Client) Subscribe(topic Topic) {
	subscribe := PDUSubscribe{}
	subscribe.Topics = append(subscribe.Topics, topic)
	c.send(&subscribe)
}

// Unsubscribe sends UNSUBSCRIBE MQTT control packet.
func (c *Client) Unsubscribe(topic string) {
	unsubscribe := PDUUnsubscribe{}
	unsubscribe.Topics = append(unsubscribe.Topics, topic)
	c.send(&unsubscribe)
}

// ============================================================================

// Publish sends PUBLISH MQTT control packet.
func (c *Client) Publish(topic string, payload []byte) {
	publish := PDUPublish{}
	publish.Topic = topic
	publish.Payload = payload
	c.send(&publish)
}

// ============================================================================
// Internal

// TODO remove retry parameter as it is not used
func (c *Client) connect() error {
	log.Println("Connecting to server")
	conn, err := net.DialTimeout("tcp", c.Address, 5*time.Second)
	if err != nil {
		return err
	}

	// 1. Open session - Login
	// Send connect packet
	connectPacket := PDUConnect{ProtocolLevel: c.ProtocolLevel, ProtocolName: "MQTT"}
	connectPacket.Keepalive = c.Keepalive
	connectPacket.ClientID = c.ClientID
	connectPacket.CleanSession = c.CleanSession
	buf := connectPacket.Marshall()
	buf.WriteTo(conn)

	conn.SetReadDeadline(time.Now().Add(c.ConnectTimeout))
	var connack Marshaller
	if connack, err = PacketRead(conn); err != nil {
		return err
	}

	switch p := connack.(type) {
	case PDUConnAck:
		switch p.ReturnCode {
		case ConnAccepted:
		default:
			return ConnAckError(p.ReturnCode)
		}
	default:
		return ErrIncorrectConnectResponse
	}

	conn.SetReadDeadline(time.Time{})

	// 2. Connected. We set environment up
	c.backoff.Reset()

	c.setSender(initSender(conn, c.Keepalive))
	// Start routine to receive incoming data
	tearDown := initReceiver(conn, c.Messages, c.sender)
	// Routine to watch for disconnect event and trigger reconnect
	go c.disconnected(tearDown, c.sender.done, c.Messages)
	return nil
}

// get receiver tearDown signal, clean client state and trigger reconnect
func (c *Client) disconnected(receiverDone <-chan struct{}, senderDone <-chan struct{}, messageChannel chan<- Message) {
	select {
	case <-receiverDone:
		c.sender.quit <- struct{}{}
	case <-senderDone:
		// We do nothing for now: As the sender closes socket, this should
		// be enough to have read Loop fail and properly shutdown process.
	}

	if c.Handler != nil {
		c.Handler(Event{Type: EventDisconnected})
	}
}

// ============================================================================

func (c *Client) send(packet Marshaller) {
	buf := packet.Marshall()
	sender := c.getSender()
	sender.send(&buf)
}

// ============================================================================
// sender setter / getter
// TODO: Probably it is not needed as we probably do not need to really reset
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
