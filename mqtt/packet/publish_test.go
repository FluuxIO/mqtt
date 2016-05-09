package packet

import "testing"

func TestPublishDecode(t *testing.T) {
	publish := NewPublish()
	publish.id = 1
	publish.dup = false
	publish.qos = 1
	publish.retain = false
	publish.topic = "test/1"
	publish.payload = "Hi"

	buf := publish.Marshall()
	if packet, err := Read(&buf); err != nil {
		t.Errorf("cannot decode publish packet: %q", err)
	} else {
		switch p := packet.(type) {
		case *Publish:
			if p.dup != publish.dup {
				t.Errorf("incorrect dup flag (%t) = %t", p.dup, publish.dup)
			}
			if p.qos != publish.qos {
				t.Errorf("incorrect qos flag (%d) = %d", p.qos, publish.qos)
			}
			if p.retain != publish.retain {
				t.Errorf("incorrect retain flag (%t) = %t", p.retain, publish.retain)
			}
			if p.topic != publish.topic {
				t.Errorf("incorrect topic (%q) = %q", p.topic, publish.topic)
			}
			if p.id != publish.id {
				t.Errorf("incorrect id (%d) = %d", p.id, publish.id)
			}
			if p.payload != publish.payload {
				t.Errorf("incorrect payload (%q) = %q", p.payload, publish.payload)
			}

		default:
			t.Error("incorrect packet type for publish")
		}
	}
}
