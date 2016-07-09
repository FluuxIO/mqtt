package packet

import "testing"

func TestSubscribeDecode(t *testing.T) {
	subscribe := PDUSubscribe{}
	subscribe.ID = 2

	t1 := Topic{Name: "test/topic", QOS: 0}
	subscribe.Topics = append(subscribe.Topics, t1)
	t2 := Topic{Name: "test2/*", QOS: 1}
	subscribe.Topics = append(subscribe.Topics, t2)

	buf := subscribe.Marshall()
	if packet, err := PacketRead(&buf); err != nil {
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
