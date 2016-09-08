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
	done := make(chan struct{})
	ready := make(chan struct{})
	go mqttServerMock(t, done, ready, func(t *testing.T, c net.Conn) { return })
	defer close(done)

	<-ready

	// Test / Check result
	client := mqtt.New(testMQTTAddress)
	client.ConnectTimeout = 100 * time.Millisecond

	if err := client.Connect(nil); err != nil {
		fmt.Println("connect error:", err)
		if neterr, ok := err.(net.Error); ok && !neterr.Timeout() {
			t.Error("MQTT connection should timeout")
		}
	}
}

func TestClient_Connect(t *testing.T) {
	// Setup Mock server
	done := make(chan struct{})
	ready := make(chan struct{})
	go mqttServerMock(t, done, ready, handlerConnackSuccess)
	defer close(done)

	<-ready

	// Test / Check result
	client := mqtt.New(testMQTTAddress)
	client.ConnectTimeout = 30 * time.Second
	if err := client.Connect(nil); err != nil {
		t.Errorf("MQTT connection failed: %s", err)
	}
}

func TestClient_Unauthorized(t *testing.T) {
	// Setup Mock server
	done := make(chan struct{})
	ready := make(chan struct{})
	go mqttServerMock(t, done, ready, handlerUnauthorized)
	defer close(done)

	<-ready

	// Test / Check result
	client := mqtt.New(testMQTTAddress)
	client.ClientID = "testClientID"
	if err := client.Connect(nil); err == nil {
		t.Error("MQTT connection should have failed")
	}
}

func TestClient_KeepAliveDisable(t *testing.T) {
	// Setup Mock server
	done := make(chan struct{})
	ready := make(chan struct{})
	go mqttServerMock(t, done, ready, handlerConnackSuccess)
	defer close(done)

	<-ready

	// Test / Check result
	client := mqtt.New(testMQTTAddress)
	client.Keepalive = 0
	if err := client.Connect(nil); err != nil {
		t.Error("MQTT connection failed")
	}
	// TODO Check that client does not send PINGREQ to server when keep alive is 0.
	// keepalive 0 should disable keep alive.
}

//=============================================================================
// Mock MQTT server for testing client

type testHandler func(t *testing.T, conn net.Conn)

func mqttServerMock(t *testing.T, done <-chan struct{}, ready chan<- struct{}, handler testHandler) {
	l, err := net.Listen("tcp", testMQTTAddress)
	if err != nil {
		t.Errorf("mqttServerMock cannot listen on address: %q", testMQTTAddress)
		return
	}

	stopAccept := make(chan struct{})
	go mqttServerMockDone(l, done, stopAccept)
	close(ready)

	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-stopAccept:
				return
			default:
			}

			t.Error("mqttServerMock accept error:", err.Error())
			l.Close()
			return
		}
		go handler(t, conn) // TODO Create and pass a stop channel to stop them
	}
}

func mqttServerMockDone(listener net.Listener, done <-chan struct{}, stopAccept chan<- struct{}) {
	select {
	case <-done:
		close(stopAccept)
		listener.Close()
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

// To fix test we need to make sure that we do not have race conditions:
// - Check start mock / listen in a synchronous way
// - Check that close is also synchronous to also avoid to block next listen
//
// + Setup SemaphoreCI with Coveralls: https://github.com/mattn/goveralls
