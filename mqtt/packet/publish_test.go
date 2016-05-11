package packet

import (
	"bytes"
	"testing"
)

func TestPublishDecode(t *testing.T) {
	publish := NewPublish()
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
		case *Publish:
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
