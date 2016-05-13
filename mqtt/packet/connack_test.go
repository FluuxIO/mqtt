package packet

import "testing"

func TestConnAckEncodeDecode(t *testing.T) {
	returnCode := 1
	ca := NewConnAck()
	ca.ReturnCode = returnCode
	buf := ca.Marshall()
	if packet, err := Read(&buf); err != nil {
		t.Error("cannot decode connack packet")
	} else {
		switch p := packet.(type) {
		case *ConnAck:
			if p.ReturnCode != returnCode {
				t.Errorf("incorrect result code (%d) = %d", p.ReturnCode, returnCode)
			}
		}
	}
}
