package packet

import "testing"

func TestConnAckEncodeDecode(t *testing.T) {
	rc := 1
	ca := NewConnAck()
	ca.ReturnCode = rc
	buf := ca.Marshall()
	if packet, err := Read(&buf); err != nil {
		t.Error("cannot decode connack packet")
	} else {
		switch p := packet.(type) {
		case *ConnAck:
			if p.ReturnCode != rc {
				t.Errorf("incorrect result code (%d) = %d", p.ReturnCode, rc)
			}
		}
	}
}
