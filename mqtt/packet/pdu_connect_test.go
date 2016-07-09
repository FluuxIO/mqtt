package packet

import "testing"

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
	if packet, err := PacketRead(&buf); err != nil {
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
