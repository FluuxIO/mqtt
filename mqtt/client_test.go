package mqtt_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/processone/gomqtt/mqtt"
)

const (
	// Default port is not standard MQTT port to avoid interfering
	// with local running MQTT server
	testMQTTAddress = "localhost:10883"
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
		fmt.Println("connect error:", err)
		if neterr, ok := err.(net.Error); ok && !neterr.Timeout() {
			t.Error("MQTT connection should timeout")
		}
	}
	mock.Stop()
}

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

type testHandler func(t *testing.T, conn net.Conn)

type MQTTServerMock struct {
	t        *testing.T
	handler  testHandler
	listener net.Listener
	done     chan struct{}
}

func (m *MQTTServerMock) Start(t *testing.T, handler testHandler) {
	m.t = t
	m.handler = handler
	m.init()
	go m.loop()
}

func (m *MQTTServerMock) Stop() {
	close(m.done)
	m.listener.Close()
}

func (m *MQTTServerMock) init() {
	l, err := net.Listen("tcp", testMQTTAddress)
	if err != nil {
		m.t.Errorf("mqttServerMock cannot listen on address: %q", testMQTTAddress)
		return
	}
	m.listener = l
	m.done = make(chan struct{})
	return
}

func (m *MQTTServerMock) loop() {
	for {
		conn, err := m.listener.Accept()
		if err != nil {
			select {
			case <-m.done:
				return
			default:
				m.t.Error("mqttServerMock accept error:", err.Error())
			}
			return
		}
		// TODO Create and pass a context to cancel the handler if they are still around = avoid possible leak on complex handlers
		go m.handler(m.t, conn)
	}
}

//=============================================================================
// Basic MQTT Server Mock Handlers.

// handlerConnackSuccess sends connack to client without even reading from socket.
func handlerConnackSuccess(t *testing.T, c net.Conn) {
	ack := mqtt.PDUConnAck{}
	buf := ack.Marshall()
	buf.WriteTo(c)
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
	case mqtt.PDUConnect:
		if pType.ClientID != "testClientID" {
			t.Error("connect packet is not properly parsed")
		}
		ack := mqtt.PDUConnAck{ReturnCode: mqtt.ConnRefusedBadUsernameOrPassword}
		buf := ack.Marshall()
		buf.WriteTo(c)
	default:
	}
}
