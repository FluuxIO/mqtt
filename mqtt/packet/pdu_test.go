package packet

import (
	"bytes"
	"testing"
)

// ============================================================================
// CONNECT
// ============================================================================

// TODO: Test incorrect will QOS
func TestIncrementalConnectFlag(t *testing.T) {
	c := PDUConnect{}
	assertConnectFlagValue(t, "incorrect connect flag: default connect flag should be empty (%d)", c.connectFlag(), 0)

	c.CleanSession = true
	assertConnectFlagValue(t, "incorrect connect flag: cleanSession is not true (%d)", c.connectFlag(), 2)

	c.SetWill("topic/a", "Disconnected", 0)
	assertConnectFlagValue(t, "incorrect connect flag: willFlag is not true (%d)", c.connectFlag(), 6)

	c.SetWill("topic/a", "Disconnected", 1)
	assertConnectFlagValue(t, "incorrect connect flag: willQOS is not properly set (%d)", c.connectFlag(), 14)

	c.SetWill("topic/a", "Disconnected", 2)
	assertConnectFlagValue(t, "incorrect connect flag: willQOS is not properly set (%d)", c.connectFlag(), 22)

	c.WillRetain = true
	assertConnectFlagValue(t, "incorrect connect flag: willRetain is not properly set (%d)", c.connectFlag(), 54)

	c.Username = "User1"
	assertConnectFlagValue(t, "incorrect connect flag: usernameFlag is not properly set (%d)", c.connectFlag(), 118)

	c.Password = "Password"
	assertConnectFlagValue(t, "incorrect connect flag: passwordFlag is not properly set (%d)", c.connectFlag(), 246)
}

func TestConnectDecode(t *testing.T) {
	connect := PDUConnect{ProtocolLevel: ProtocolLevel, ProtocolName: ProtocolName}
	connect.CleanSession = true
	connect.WillFlag = true
	connect.WillQOS = 1
	connect.WillRetain = true
	connect.Keepalive = 42
	connect.ClientID = "TestClientID"
	connect.WillTopic = "test/will"
	connect.WillMessage = "test message"
	connect.Username = "testuser"
	connect.Password = "testpass"

	buf := connect.Marshall()
	if packet, err := Read(&buf); err != nil {
		t.Errorf("cannot decode connect packet: %q", err)
	} else {
		switch p := packet.(type) {
		case PDUConnect:
			if p != connect {
				t.Errorf("unmarshalled connect does not match original (%+v) = %+v", p, connect)
			}
		}
	}
}

// Helpers

func assertConnectFlagValue(t *testing.T, message string, flag int, expected int) {
	if flag != expected {
		t.Errorf(message, flag)
	}
}

// ============================================================================
// CONNACK
// ============================================================================

func TestConnAckEncodeDecode(t *testing.T) {
	returnCode := 1
	ca := &PDUConnAck{}
	ca.ReturnCode = returnCode
	buf := ca.Marshall()
	if packet, err := Read(&buf); err != nil {
		t.Error("cannot decode connack control packet")
	} else {
		switch p := packet.(type) {
		case PDUConnAck:
			if p.ReturnCode != returnCode {
				t.Errorf("incorrect result code (%d) = %d", p.ReturnCode, returnCode)
			}
		}
	}
}

// ============================================================================
// PUBLISH
// ============================================================================

func TestPublishDecode(t *testing.T) {
	publish := PDUPublish{}
	publish.ID = 1
	publish.Dup = false
	publish.Qos = 1
	publish.Retain = false
	publish.Topic = "test/1"
	publish.Payload = []byte("Hi")

	buf := publish.Marshall()
	if packet, err := Read(&buf); err != nil {
		t.Errorf("cannot decode publish packet: %q", err)
	} else {
		switch p := packet.(type) {
		case PDUPublish:
			if p.Dup != publish.Dup {
				t.Errorf("incorrect dup flag (%t) = %t", p.Dup, publish.Dup)
			}
			if p.Qos != publish.Qos {
				t.Errorf("incorrect qos flag (%d) = %d", p.Qos, publish.Qos)
			}
			if p.Retain != publish.Retain {
				t.Errorf("incorrect retain flag (%t) = %t", p.Retain, publish.Retain)
			}
			if p.Topic != publish.Topic {
				t.Errorf("incorrect topic (%q) = %q", p.Topic, publish.Topic)
			}
			if p.ID != publish.ID {
				t.Errorf("incorrect id (%d) = %d", p.ID, publish.ID)
			}
			if bytes.Compare(p.Payload, publish.Payload) != 0 {
				t.Errorf("incorrect payload (%q) = %q", p.Payload, publish.Payload)
			}

		default:
			t.Error("incorrect packet type for publish")
		}
	}
}

// ============================================================================
// SUBSCRIBE
// ============================================================================

func TestSubscribeDecode(t *testing.T) {
	subscribe := PDUSubscribe{}
	subscribe.ID = 2

	t1 := Topic{Name: "test/topic", QOS: 0}
	subscribe.Topics = append(subscribe.Topics, t1)
	t2 := Topic{Name: "test2/*", QOS: 1}
	subscribe.Topics = append(subscribe.Topics, t2)

	buf := subscribe.Marshall()
	if packet, err := Read(&buf); err != nil {
		t.Errorf("cannot decode subscribe packet: %q", err)
	} else {
		switch p := packet.(type) {
		case PDUSubscribe:
			if p.ID != subscribe.ID {
				t.Errorf("incorrect id (%d) = %d", p.ID, subscribe.ID)
			}
			if len(p.Topics) != 2 {
				t.Errorf("incorrect topics length (%d) = %d", len(p.Topics), 2)
			}
			parsedt1 := p.Topics[0]
			if parsedt1.Name != t1.Name {
				t.Errorf("incorrect topic name (%q) = %q", parsedt1.Name, t1.Name)
			}
			if parsedt1.QOS != t1.QOS {
				t.Errorf("incorrect topic qos (%q) = %q", parsedt1.QOS, t1.QOS)
			}

			parsedt2 := p.Topics[1]
			if parsedt2.QOS != t2.QOS {
				t.Errorf("incorrect topic qos (%q) = %q", parsedt2.QOS, t2.QOS)
			}
		default:
			t.Error("Incorrect packet type for subscribe")
		}
	}
}
