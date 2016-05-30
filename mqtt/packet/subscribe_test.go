package packet

import "testing"

func TestSubscribeDecode(t *testing.T) {
	subscribe := NewSubscribe()
	subscribe.id = 2

	t1 := Topic{Name: "test/topic", QOS: 0}
	subscribe.AddTopic(t1)
	t2 := Topic{Name: "test2/*", QOS: 1}
	subscribe.AddTopic(t2)

	buf := subscribe.Marshall()
	if packet, err := Read(&buf); err != nil {
		t.Errorf("cannot decode subscribe packet: %q", err)
	} else {
		switch p := packet.(type) {
		case *Subscribe:
			if p.id != subscribe.id {
				t.Errorf("incorrect id (%d) = %d", p.id, subscribe.id)
			}
			if len(p.topics) != 2 {
				t.Errorf("incorrect topics length (%d) = %d", len(p.topics), 2)
			}
			parsedt1 := p.topics[0]
			if parsedt1.Name != t1.Name {
				t.Errorf("incorrect topic name (%q) = %q", parsedt1.Name, t1.Name)
			}
			if parsedt1.QOS != t1.QOS {
				t.Errorf("incorrect topic qos (%q) = %q", parsedt1.QOS, t1.QOS)
			}

			parsedt2 := p.topics[1]
			if parsedt2.QOS != t2.QOS {
				t.Errorf("incorrect topic qos (%q) = %q", parsedt2.QOS, t2.QOS)
			}
		default:
			t.Error("Incorrect packet type for subscribe")
		}
	}
}
