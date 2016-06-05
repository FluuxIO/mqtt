package mqtt

import (
	"net"
	"testing"
	"time"
)

const (
	testMQTTAddress = "localhost:10883"
)

// TestClient_ConnectTimeout checks that connect will properly timeout and not
// block forever if server never send CONNACK.
func TestClient_ConnectTimeout(t *testing.T) {
	go mqttServerMock(t, func(c net.Conn) { return })
	client := New(testMQTTAddress)
	client.ConnectTimeout = time.Duration(100) * time.Millisecond

	if err := client.Connect(); err != nil {
		if neterr, ok := err.(net.Error); ok && !neterr.Timeout() {
			t.Error("MQTT connection should timeout")
		}
	}
}

//=============================================================================
// Mock MQTT server for testing client

type testHandler func(conn net.Conn)

func mqttServerMock(t *testing.T, handler testHandler) {
	l, err := net.Listen("tcp", testMQTTAddress)
	if err != nil {
		t.Errorf("mqttServerMock cannot listen on address: %q", testMQTTAddress)
		return
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			t.Error("mqttServerMock accept error:", err.Error())
			l.Close()
			return
		}
		go handler(conn)
	}
}
