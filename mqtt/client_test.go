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
	done := make(chan struct{})
	go mqttServerMock(t, done, func(c net.Conn) { return })
	defer close(done)

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

func mqttServerMock(t *testing.T, done <-chan struct{}, handler testHandler) {
	l, err := net.Listen("tcp", testMQTTAddress)
	if err != nil {
		t.Errorf("mqttServerMock cannot listen on address: %q", testMQTTAddress)
		return
	}

	stopAccept := make(chan struct{})
	go mqttServerMockDone(l, done, stopAccept)

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
		go handler(conn)
	}
}

func mqttServerMockDone(listener net.Listener, done <-chan struct{}, stopAccept chan<- struct{}) {
	select {
	case <-done:
		close(stopAccept)
		listener.Close()
	}
}
