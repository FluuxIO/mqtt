/*
Interesting reference on backoff:
- Exponential Backoff And Jitter (AWS Blog):
  https://www.awsarchitectureblog.com/2015/03/backoff.html

We use Jitter as a default for exponential backoff, as the goal of
this module is not to provide precise 'ticks', but good behaviour to
implement retries that are helping the server to recover faster in
case of congestion.

It can be used in several ways:
- Using duration to get next sleep time.
- Using ticker channel to trigger callback function on tick

The functions for Backoff are not threadsafe, but you can:
- Keep the attempt counter on your end and use DurationForAttempt(int)
- Use lock in your own code to protect the Backoff structure.

TODO: Implement Backoff Ticker channel
*/

package mqtt

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

const (
	defaultBase   int = 20 // Backoff base, in ms
	defaultFactor int = 2
	defaultCap    int = 180000 // 3 minutes
)

type Backoff struct {
	NoJitter     bool
	Base         int
	Factor       int
	Cap          int
	LastDuration int
	attempt      int
}

func (b *Backoff) Duration() time.Duration {
	d := b.DurationForAttempt(b.attempt)
	b.attempt++
	fmt.Printf("Duration for backoff: %d", d)
	return d
}

func (b *Backoff) DurationForAttempt(attempt int) time.Duration {
	fmt.Printf("Calculating duration for attempt: %d", attempt)
	b.setDefault()
	expBackoff := math.Min(float64(b.Cap), float64(b.Base)*math.Pow(float64(b.Factor), float64(b.attempt)))
	d := int(math.Trunc(expBackoff))
	if !b.NoJitter {
		d = rand.Intn(d)
	}
	return time.Duration(d) * time.Millisecond
}

func (b *Backoff) Reset() {
	b.attempt = 0
}

func (b *Backoff) setDefault() {
	if b.Base == 0 {
		b.Base = defaultBase
	}

	if b.Cap == 0 {
		b.Cap = defaultCap
	}

	if b.Factor == 0 {
		b.Factor = defaultFactor
	}
}

/*
We use full jitter for now as it seems to provide good behaviour for reconnect.

base is the default interval between attempts (if backoff factor was equal to 1)

attempt is the number of retry for operation. If we start attempt at 0, first sleep equals base.

cap is the maximum sleep time duration we tolerate between attempts
*/
