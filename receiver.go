package mqtt // import "gosrc.io/mqtt"

import (
	"io"
	"log"
)

type receiver struct {
	// Connection to read the MQTT data from
	conn io.Reader
	// sender is the struct managing the go routine to send
	sender sender
	// Channel to send back message received (PUBLISH control packets) to the client using the library
	messageChannel chan<- Message
	// Channel to send back QOS packet (acks) to the internal client process.
	qosChannel chan<- QOSResponse
}

// Receiver actually need:
// - Net.conn
// - Sender (to send ack packet when packets requiring acks are received)
// - Error send channel to trigger teardown
// - MessageSendChannel to dispatch messages to client
// Returns teardown channel used to notify when the receiver terminates.
func spawnReceiver(conn io.Reader, messageChannel chan<- Message, s sender) <-chan QOSResponse {
	qosChannel := make(chan QOSResponse)
	go receiverLoop(conn, qosChannel, messageChannel, s)
	return qosChannel
}

// Receive, decode and dispatch messages to the message channel
func receiverLoop(conn io.Reader, qosChannel chan<- QOSResponse, message chan<- Message, s sender) {
	var p Marshaller
	var err error

Loop:
	for {
		if p, err = PacketRead(conn); err != nil {
			if err == io.EOF {
				log.Printf("Connection closed\n")
			}
			log.Printf("packet read error: %q\n", err)
			break Loop
		}
		// fmt.Printf("Received: %+v\n", p)

		sendAckIfNeeded(p, s)

		// Only broadcast message back to client when we receive publish packets
		switch packetType := p.(type) {
		case PublishPacket:
			m := Message{}
			m.Topic = packetType.Topic
			m.Payload = packetType.Payload
			message <- m // TODO Back pressure. We may block on processing message if client does not read fast enough. Make sure we can quit.
		default:
			if ResponsePacket, ok := p.(QOSResponse); ok {
				qosChannel <- ResponsePacket
			}
		}
	}

	// Loop ended, send receiver close signal
	close(qosChannel)
}

// Send acks if needed, depending on packet QOS
func sendAckIfNeeded(pkt Marshaller, s sender) {
	switch p := pkt.(type) {
	case PublishPacket:
		if p.Qos == 1 {
			puback := PubAckPacket{ID: p.ID}
			buf := puback.Marshall()
			s.send(buf)
		}
	}
}
