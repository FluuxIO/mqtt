package mqtt // import "fluux.io/gomqtt/mqtt"

import (
	"net"
)

// Sender need the following interface:
// - Net.conn to send TCP packets
// - Error send channel to trigger teardown on send error
// - SendChannel receiving []byte
// - KeepaliveCtl to reset keepalive packet timer after a send
// - Way to stop the sender when client wants to stop / disconnect

type sender struct {
	done <-chan struct{}
	out  chan<- []byte
	quit chan<- struct{}
}

func initSender(conn net.Conn, keepalive int) sender {
	tearDown := make(chan struct{})
	out := make(chan []byte)
	quit := make(chan struct{})

	// Start go routine that manage keepalive timer:
	var keepaliveCtl chan int
	if keepalive > 0 {
		keepaliveCtl = startKeepalive(keepalive, func() {
			pingReq := PingReqPacket{}
			buf := pingReq.Marshall()
			conn.Write(buf)
		})
	}

	s := sender{done: tearDown, out: out, quit: quit}
	go senderLoop(conn, keepaliveCtl, out, quit, tearDown)
	return s
}

func senderLoop(conn net.Conn, keepaliveCtl chan int, out <-chan []byte, quit <-chan struct{}, tearDown chan<- struct{}) {
Loop:
	for {
		select {
		case buf := <-out:
			conn.Write(buf) // TODO Trigger teardown and stop on write error
			keepaliveSignal(keepaliveCtl, keepaliveReset)
		case <-quit:
			// Client want this sender to terminate
			terminateSender(conn, keepaliveCtl)
			break Loop
		}
	}
}

func (s sender) send(buf []byte) {
	s.out <- buf
}

// clean-up:
func terminateSender(conn net.Conn, keepaliveCtl chan int) {
	keepaliveSignal(keepaliveCtl, keepaliveStop)
	conn.Close()
}

// keepaliveSignal sends keepalive commands on keepalive channel (if
// keepalive is not disabled).
func keepaliveSignal(keepaliveCtl chan<- int, signal int) {
	if keepaliveCtl == nil {
		return
	}
	keepaliveCtl <- signal
}
