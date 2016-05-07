package packet

import "testing"

func TestConnAckEncodeDecode(t *testing.T) {
	rc := 1
	ca := NewConnAck()
	ca.ReturnCode = rc
	buf := ca.Marshall()
	packet := Read(&buf)
	switch p := packet.(type) {
	case *ConnAck:
		if p.ReturnCode != rc {
			t.Errorf("incorrect result code (%d) = %d", p.ReturnCode, rc)
		}
	}
}
