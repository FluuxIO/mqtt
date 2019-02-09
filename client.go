package mqtt // import "gosrc.io/mqtt"

import "C"
import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"
)

var (
	// ErrIncorrectConnectResponse is triggered on CONNECT when server
	// does not reply with CONNACK packet.
	ErrIncorrectConnectResponse = errors.New("incorrect mqtt connect response")
)

const (
	// DefaultMQTTServer is a shortcut to define connection to local
	// server
	DefaultMQTTServer = "tcp://localhost:1883"
)

//=============================================================================

// OptConnect defines optional MQTT connection parameters.
// MQTT client libraries will default to sensible values.
// TODO Should this be called OptMQTT?
type OptConnect struct {
	ProtocolLevel int
	ClientID      string
	Keepalive     int // TODO Keepalive should also probably be a time.Duration for more flexibility
	CleanSession  bool
	Username      string
	Password      string
}

// OptTCP defines TCP/IP related parameters. They are used to
// configure low level TCP client connection. Default should be fine
// for standard cases.
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

//=============================================================================

// ConnState represents the current connection state.
type ConnState int

// This is a the list of events happening on the connection that the
// client can be notified about.
const (
	StateDisconnected ConnState = iota
)

// Event is a structure use to convey event changes related to client state. This
// is for example used to notify the client when the client get disconnected.
type Event struct {
	State       ConnState
	Description string
}

// EventHandler is use to pass events about state of the connection to
// client implementation.
type EventHandler func(Event)

//=============================================================================
// State management

// Keep trakc on acknowledged subscriptions
type Subscriptions map[string]int

// Keeps track of inflight requests subscriptions, etc
// State
type inflight map[int]QOSOutPacket

type qosState struct {
	qosResponse   chan<- QOSResponse
	Subscriptions Subscriptions
	inflight      inflight
}

//=============================================================================

// Client is the main structure use to connect as a client on an MQTT
// server.
type Client struct {
	Config

	Handler  EventHandler
	Messages chan<- Message

	qosState

	mu       sync.RWMutex
	sender   sender
	packetID int
}

// New generates a new MQTT client with default parameters. Address
// must be set as we cannot find relevant default value for server.
// address is of the form tcp://hostname:port for cleartext connection
// or tls://hostname:port for TLS connection.
// TODO: Should messages channel be set on New ?
func NewClient(address string) *Client {
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
		qosState: qosState{
			Subscriptions: make(Subscriptions),
			inflight:      make(inflight),
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
	packet := DisconnectPacket{}
	c.send(packet)
	c.sender.quit <- struct{}{}

	// Terminate client receive channel
	// TODO Should we really close the channel or let it live in case client reconnects ?
	close(c.Messages)
	c.Messages = nil
	// TODO Properly terminates receiver and sender
}

// ============================================================================

// Subscribe sends SUBSCRIBE MQTT control packet.  At the moment
// subscription state is not kept in client state and are lost on reconnection.
func (c *Client) Subscribe(topic Topic) {
	c.packetID++
	subscribe := SubscribePacket{ID: c.packetID}
	subscribe.Topics = append(subscribe.Topics, topic)
	c.send(subscribe)
}

// Unsubscribe sends UNSUBSCRIBE MQTT control packet.
func (c *Client) Unsubscribe(topic string) {
	c.packetID++
	unsubscribe := UnsubscribePacket{ID: c.packetID}
	unsubscribe.Topics = append(unsubscribe.Topics, topic)
	c.send(unsubscribe)
}

// ============================================================================

// Publish sends PUBLISH MQTT control packet.
func (c *Client) Publish(topic string, payload []byte) {
	c.packetID++
	publish := PublishPacket{ID: c.packetID}
	publish.Topic = topic
	publish.Payload = payload
	c.send(publish)
}

// Format printable version of client state
func (c *Client) String() string {
	str := fmt.Sprintf(`
Subscription: %v
Inflight: %v`, c.Subscriptions, c.inflight)
	return str
}

// ============================================================================
// Internal

func (c *Client) connect() (err error) {
	// Parse address string
	uri, err := url.Parse(c.Address)
	if err != nil {
		return err
	}

	var conn net.Conn
	switch uri.Scheme {
	case "tcp":
		conn, err = net.DialTimeout("tcp", uri.Host, 5*time.Second)
		if err != nil {
			return err
		}
		return c.login(conn)
	case "tls":
		conn, err = net.DialTimeout("tcp", uri.Host, 5*time.Second)
		if err != nil {
			return err
		}
		config := tls.Config{ServerName: uri.Hostname()}
		tlsConn := tls.Client(conn, &config)
		err = tlsConn.Handshake()
		if err != nil {
			return err
		}
		return c.login(tlsConn)
	default:
		return errors.New("url scheme must be tcp or tls")
	}
}

func (c *Client) login(conn net.Conn) (err error) {
	// 1. Open session - Login
	// Send connect packet
	connectPacket := ConnectPacket{ProtocolLevel: c.ProtocolLevel, ProtocolName: "MQTT"}
	connectPacket.Keepalive = c.Keepalive
	connectPacket.ClientID = c.ClientID
	connectPacket.CleanSession = c.CleanSession
	connectPacket.Username = c.Username
	connectPacket.Password = c.Password
	buf := connectPacket.Marshall()
	if _, err = conn.Write(buf); err != nil {
		return err
	}

	// 2. Check login result
	if err = conn.SetReadDeadline(time.Now().Add(c.ConnectTimeout)); err != nil {
		return err
	}
	var connack Marshaller
	if connack, err = PacketRead(conn); err != nil {
		return err
	}

	switch p := connack.(type) {
	case ConnAckPacket:
		switch p.ReturnCode {
		case ConnAccepted:
		default:
			return ConnAckError(p.ReturnCode)
		}
	default:
		return ErrIncorrectConnectResponse
	}

	if err = conn.SetReadDeadline(time.Time{}); err != nil {
		return err
	}

	// 3. Configure sender and receiver
	c.setSender(initSender(conn, c.Keepalive))
	// Start routine to receive incoming data
	receiverChannel := spawnReceiver(conn, c.Messages, c.sender)
	// Routine to maintain client state based on event from receiver and sender (disconnect signal, QOS / Ack messages, etc)
	go c.stateLoop(receiverChannel, c.sender.done, c.Messages)
	return nil
}

// Go routine used to coordinates client state management loop.
// Routine to maintain client state based on event from receiver and sender (disconnect signal, QOS / Ack messages, etc)
// It updates the state of inflight messages, but also track disconnect event to shutdown properly.
func (c *Client) stateLoop(receiverChannel <-chan QOSResponse, senderDone <-chan struct{}, messageChannel chan<- Message) {
Loop:
	for {
		select {
		case qosResponse, ok := <-receiverChannel:
			if !ok { // Receiver terminated
				c.sender.quit <- struct{}{}
				break Loop
			}
			c.handleQOSResponse(qosResponse)
		case <-senderDone:
			// We do nothing for now: As the sender closes socket, this should
			// be enough to have read Loop fail and properly shutdown process.

			// TODO: Handle the case when the client is done ?
			break Loop
		}
	}

	if c.Handler != nil {
		c.Handler(Event{State: StateDisconnected})
	}
}

// QOS packets includes all packets that are acked (so this includes subscribe and unsubscribe).

type QOSOutPacket interface {
	PacketID() int
}

type QOSResponse interface {
	ResponseID() int
}

// TODO: Refactor in smaller functions
func (c *Client) handleQOSResponse(qosResponse QOSResponse) {
	id := qosResponse.ResponseID()
	if originalPacket, found := c.inflight[id]; found {
		switch resp := qosResponse.(type) {
		case SubAckPacket:
			// Ack only contains the response code for each topic: We need to merge the result with the
			// original subscription request.
			if sub, ok := originalPacket.(SubscribePacket); ok {
				for i, topic := range sub.Topics {
					if resp.ReturnCodes[i] == 0x80 {
						fmt.Printf("Subscription failed for topic %s\n", topic.Name)
						continue
					}
					topic.QOS = resp.ReturnCodes[i]
					c.Subscriptions[topic.Name] = topic.QOS
				}
			} else {
				fmt.Printf("SubAck received, but packet %d is not a subscribe packet\n", id)
			}
		case UnsubAckPacket:
			// When the ack is received, delete all our subscriptions from the local list.
			if unsub, ok := originalPacket.(UnsubscribePacket); ok {
				for _, topic := range unsub.Topics {
					delete(c.Subscriptions, topic)
				}
			}
		}
		c.deleteInflight(id)
	}
}

func (c *Client) addToInflight(packet Marshaller) {
	if qosPacket, ok := packet.(QOSOutPacket); ok {
		c.addInflight(qosPacket)
	}
}

// ============================================================================

func (c *Client) send(packet Marshaller) {
	c.addToInflight(packet)
	buf := packet.Marshall()
	out := c.getSender()
	out.send(buf)
}

// ============================================================================
// sender setter / getter
// TODO: Probably it is not needed as we probably do not need to really reset
//   sender on reconnect

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

// Delete or remove packets from inflight packet queue
func (c *Client) addInflight(p QOSOutPacket) {
	c.mu.Lock()
	{
		c.inflight[p.PacketID()] = p
	}
	c.mu.Unlock()
}

func (c *Client) deleteInflight(id int) {
	c.mu.Lock()
	{
		delete(c.inflight, id)
	}
	c.mu.Unlock()
}
