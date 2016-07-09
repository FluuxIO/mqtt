package packet

import "testing"

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
