package mqtt

import (
	"bytes"
	"net"

	"github.com/processone/gomqtt/mqtt/packet"
)

// Sender need the following interface:
// - Net.conn to send TCP packets
// - Error send channel to trigger teardown on send error
// - SendChannel receiving *bytes.Buffer
// - KeepaliveCtl to reset keepalive packet timer after a send
// - Way to stop the sender when client wants to stop / disconnect

type sender struct {
	done <-chan struct{}
	out  chan<- *bytes.Buffer
	quit chan<- struct{}
}

func initSender(conn net.Conn, keepalive int) sender {
	tearDown := make(chan struct{})
	out := make(chan *bytes.Buffer)
	quit := make(chan struct{})

	// Start go routine that manage keepalive timer:
	keepaliveCtl := startKeepalive(keepalive, func() {
		pingReq := packet.NewPingReq()
		buf := pingReq.Marshall()
		buf.WriteTo(conn)
	})

	s := sender{done: tearDown, out: out, quit: quit}
	go senderLoop(conn, keepaliveCtl, out, quit, tearDown)
	return s
}

func senderLoop(conn net.Conn, keepaliveCtl chan int, out <-chan *bytes.Buffer, quit <-chan struct{}, tearDown chan<- struct{}) {
Loop:
	for {
		select {
		case buf := <-out:
			buf.WriteTo(conn) // TODO Trigger teardown and stop on write error
			keepaliveCtl <- keepaliveReset
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

// Clean-up:
func terminateSender(conn net.Conn, keepaliveCtl chan int) {
	keepaliveCtl <- keepaliveStop
	conn.Close()
}
