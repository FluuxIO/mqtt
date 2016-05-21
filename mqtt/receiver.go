package mqtt

import (
	"fmt"
	"io"

	"github.com/processone/gomqtt/mqtt/packet"
)

// Receive, decode and dispatch messages to the message channel
func receiver(c *Client) {
	var p packet.Packet
	var err error
	conn := c.conn
Loop:
	for {
		if p, err = packet.Read(conn); err != nil {
			if err == io.EOF {
				fmt.Printf("Connection closed\n")
			}
			fmt.Printf("packet read error: %q\n", err)
			break Loop
		}
		fmt.Printf("Received: %+v\n", p)
		sendAck(c, p)
		// Only broadcast message back to client when we receive publish packets
		switch publish := p.(type) {
		case *packet.Publish:
			m := new(Message)
			m.Topic = publish.Topic
			m.Payload = publish.Payload
			c.message <- m
		default:
		}
	}

	// TODO Support ability to disable autoreconnect
	// Cleanup and reconnect
	conn.Close()
	c.keepaliveCtl <- keepaliveStop
	go c.connect(true)
}
