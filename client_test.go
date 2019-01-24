package mqtt_test // import "gosrc.io/mqtt"

import (
	"log"
	"net"
	"testing"
	"time"

	"gosrc.io/mqtt"
)

const (
	acceptTimeout = 300 * time.Millisecond
)

// TestClient_ConnectTimeout checks that connect will properly timeout and not
// block forever if server never send CONNACK.
func TestClient_ConnectTimeout(t *testing.T) {
	// Setup Mock server
	mock := MQTTServerMock{}
	mock.Start(t, func(t *testing.T, c net.Conn) { return })

	// Test / Check result
	client := mqtt.New(testMQTTAddress)
	client.ConnectTimeout = 100 * time.Millisecond

	if err := client.Connect(nil); err != nil {
		if neterr, ok := err.(net.Error); ok && !neterr.Timeout() {
			t.Error("MQTT connection should timeout")
		}
	}
	mock.Stop()
}

// TestClient_Connect checks that we can connect to MQTT server and
// get no error when we receive CONNACK.
func TestClient_Connect(t *testing.T) {
	// Setup Mock server
	mock := MQTTServerMock{}
	mock.Start(t, handlerConnackSuccess)

	// Test / Check result
	client := mqtt.New(testMQTTAddress)
	client.ConnectTimeout = 30 * time.Second
	if err := client.Connect(nil); err != nil {
		t.Errorf("MQTT connection failed: %s", err)
	}
	mock.Stop()
}

// TestClient_Unauthorized checks that MQTT connect fails when we
// received unauthorized response.
func TestClient_Unauthorized(t *testing.T) {
	// Setup Mock server
	mock := MQTTServerMock{}
	mock.Start(t, handlerUnauthorized)

	// Test / Check result
	client := mqtt.New(testMQTTAddress)
	client.ClientID = "testClientID"
	if err := client.Connect(nil); err == nil {
		t.Error("MQTT connection should have failed")
	}
	mock.Stop()
}

// TestClient_KeepAliveDisable checks that we can connect successfully
// without keepalive.
func TestClient_KeepAliveDisable(t *testing.T) {
	// Setup Mock server
	mock := MQTTServerMock{}
	mock.Start(t, handlerConnackSuccess)

	// Test / Check result
	client := mqtt.New(testMQTTAddress)
	client.Keepalive = 0
	if err := client.Connect(nil); err != nil {
		t.Error("MQTT connection failed")
	}
	// TODO Check that client does not send PINGREQ to server when keep alive is 0.
	// keepalive 0 should disable keep alive.
	mock.Stop()
}

//=============================================================================
// Mock MQTT server for testing client

const (
	// Default port is not standard MQTT port to avoid interfering
	// with local running MQTT server
	testMQTTAddress = "localhost:10883"
)

type testHandler func(t *testing.T, conn net.Conn)

type MQTTServerMock struct {
	t           *testing.T
	handler     testHandler
	listener    net.Listener
	connections []net.Conn
	done        chan struct{}
	cleanup     chan struct{}
}

func (mock *MQTTServerMock) Start(t *testing.T, handler testHandler) {
	mock.t = t
	mock.handler = handler
	if err := mock.init(); err != nil {
		return
	}
	go mock.loop()
}

func (mock *MQTTServerMock) Stop() {
	close(mock.done)

	// Check that main MQTT mock server loop is actually terminated
	<-mock.cleanup

	// Shutdown server mock
	if mock.listener != nil {
		if err := mock.listener.Close(); err != nil {
			log.Println("cannot close listener", err)
		}
	}
}

func (mock *MQTTServerMock) init() error {
	mock.done = make(chan struct{})
	mock.cleanup = make(chan struct{})

	l, err := net.Listen("tcp", testMQTTAddress)
	if err != nil {
		mock.t.Errorf("mqttServerMock cannot listen on address: %s (%s)", testMQTTAddress, err)
		return err
	}
	mock.listener = l
	return nil
}

func (mock *MQTTServerMock) loop() {
	defer mock.loopCleanup()
	listener := mock.listener
	for {
		if l, ok := listener.(*net.TCPListener); ok {
			l.SetDeadline(time.Now().Add(acceptTimeout))
		}
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-mock.done:
				return
			default:
				if err, ok := err.(net.Error); ok && err.Timeout() {
					// timeout error
					return
				}
				mock.t.Error("mqttServerMock accept error:", err.Error())
			}
			return
		}
		mock.connections = append(mock.connections, conn)
		// TODO Create and pass a context to cancel the handler if they are still around = avoid possible leak on complex handlers
		go mock.handler(mock.t, conn)
	}
}

func (mock *MQTTServerMock) loopCleanup() {
	// Close all existing connections
	for _, c := range mock.connections {
		if err := c.Close(); err != nil {
			log.Println("Cannot close connection", c)
		}
	}
	close(mock.cleanup)
}

//=============================================================================
// Basic MQTT Server Mock Handlers.

// handlerConnackSuccess sends connack to client without even reading from socket.
func handlerConnackSuccess(_ *testing.T, c net.Conn) {
	ack := mqtt.ConnAckPacket{}
	buf := ack.Marshall()
	c.Write(buf)
}

func handlerUnauthorized(t *testing.T, c net.Conn) {
	var p mqtt.Marshaller
	var err error

	// Only wait for client response for a small amount of time
	c.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	if p, err = mqtt.PacketRead(c); err != nil {
		t.Error("did not receive anything from client")
	}
	c.SetReadDeadline(time.Time{})

	switch pType := p.(type) {
	case mqtt.ConnectPacket:
		if pType.ClientID != "testClientID" {
			t.Error("connect packet is not properly parsed")
		}
		ack := mqtt.ConnAckPacket{ReturnCode: mqtt.ConnRefusedBadUsernameOrPassword}
		buf := ack.Marshall()
		c.Write(buf)
	default:
	}
	log.Println("Unauthorized handler done")
}
