package mqtt

import (
	"testing"
	"time"
)

func TestDurationForAttempt_NoJitter(t *testing.T) {
	b := Backoff{Base: 25}
	if b.DurationForAttempt(0) != time.Duration(b.Base) {
		t.Errorf("incorrect default duration for attempt #0 (%d) = %d", b.DurationForAttempt(0), b.Base)
	}
	var prevDuration, d time.Duration = 0, 0
	for i := 0; i < 10; i++ {
		d = b.DurationForAttempt(i)
		if !(d >= prevDuration) {
			t.Errorf("duration should be increasing between attempts. #%d (%d) > %d", i, d, prevDuration)
		}
		prevDuration = d
	}
}
