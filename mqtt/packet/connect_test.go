package packet

import "testing"

// TODO: Refactor test
// TODO: Test incorrect will QOS
func TestConnectFlag(t *testing.T) {
	c := Connect{}
	c.cleanSession = true
	if c.connectFlag() != 2 {
		t.Error("incorrect connect flag: cleanSession is not true (%d)", c.connectFlag())
	}
	c.SetWill("topic/a", "Disconnected", 0)
	if c.connectFlag() != 6 {
		t.Error("incorrect connect flag: willFlag is not true (%d)", c.connectFlag())
	}
	c.SetWill("topic/a", "Disconnected", 1)
	if c.connectFlag() != 14 {
		t.Error("incorrect connect flag: willQOS is not properly set (%d)", c.connectFlag())
	}
	c.SetWill("topic/a", "Disconnected", 2)
	if c.connectFlag() != 22 {
		t.Error("incorrect connect flag: willQOS is not properly set (%d)", c.connectFlag())
	}
	c.willRetain = true
	if c.connectFlag() != 54 {
		t.Error("incorrect connect flag: willRetain is not properly set (%d)", c.connectFlag())
	}
	c.username = "User1"
	if c.connectFlag() != 118 {
		t.Error("incorrect connect flag: usernameFlag is not properly set (%d)", c.connectFlag())
	}
	c.password = "Password"
	if c.connectFlag() != 246 {
		t.Error("incorrect connect flag: passwordFlag is not properly set (%d)", c.connectFlag())
	}
}
