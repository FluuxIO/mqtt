package mqtt // import "fluux.io/gomqtt/mqtt"

import (
	"bytes"
	"io"
	"net"
)

// Sender need the following interface:
// - Net.conn to send TCP packets
// - Error send channel to trigger teardown on send error
// - SendChannel receiving *bytes.Buffer
// - KeepaliveCtl to reset keepalive packet timer after a send
// - Way to stop the sender when client wants to stop / disconnect

type sender struct {
	done <-chan struct{}
	out  chan<- io.WriterTo
	quit chan<- struct{}
}

func initSender(conn net.Conn, keepalive int) sender {
	tearDown := make(chan struct{})
	out := make(chan io.WriterTo)
	quit := make(chan struct{})

	// Start go routine that manage keepalive timer:
	var keepaliveCtl chan int
	if keepalive > 0 {
		keepaliveCtl = startKeepalive(keepalive, func() {
			pingReq := PDUPingReq{}
			buf := pingReq.Marshall()
			buf.WriteTo(conn)
		})
	}

	s := sender{done: tearDown, out: out, quit: quit}
	go senderLoop(conn, keepaliveCtl, out, quit, tearDown)
	return s
}

func senderLoop(conn net.Conn, keepaliveCtl chan int, out <-chan io.WriterTo, quit <-chan struct{}, tearDown chan<- struct{}) {
Loop:
	for {
		select {
		case buf := <-out:
			buf.WriteTo(conn) // TODO Trigger teardown and stop on write error
			keepaliveSignal(keepaliveCtl, keepaliveReset)
		case <-quit:
			// Client want this sender to terminate
			terminateSender(conn, keepaliveCtl)
			break Loop
		}
	}
}

func (s sender) send(buf *bytes.Buffer) {
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
