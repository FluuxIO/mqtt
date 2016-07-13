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
TODO: Implement throttler interface. Throttler could be used to implement various reconnect strategies.
*/

package mqtt

import (
	"math"
	"math/rand"
	"time"
)

const (
	defaultBase   int = 20 // Backoff base, in ms
	defaultFactor int = 2
	defaultCap    int = 180000 // 3 minutes
)

// Backoff can provide increasing duration with the number of attempt
// performed. The structure is used to support exponential backoff on
// connection attempts to avoid hammering the server we are connecting
// to.
type Backoff struct {
	noJitter     bool
	base         int
	factor       int
	cap          int
	lastDuration int
	attempt      int
}

func (b *Backoff) Duration() time.Duration {
	d := b.DurationForAttempt(b.attempt)
	b.attempt++
	return d
}

func (b *Backoff) DurationForAttempt(attempt int) time.Duration {
	b.setDefault()
	expBackoff := math.Min(float64(b.cap), float64(b.base)*math.Pow(float64(b.factor), float64(b.attempt)))
	d := int(math.Trunc(expBackoff))
	if !b.noJitter {
		d = rand.Intn(d)
	}
	return time.Duration(d) * time.Millisecond
}

func (b *Backoff) Reset() {
	b.attempt = 0
}

func (b *Backoff) setDefault() {
	if b.base == 0 {
		b.base = defaultBase
	}

	if b.cap == 0 {
		b.cap = defaultCap
	}

	if b.factor == 0 {
		b.factor = defaultFactor
	}
}

/*
We use full jitter as default for now as it seems to provide good behaviour for reconnect.

base is the default interval between attempts (if backoff factor was equal to 1)

attempt is the number of retry for operation. If we start attempt at 0, first sleep equals base.

cap is the maximum sleep time duration we tolerate between attempts
*/
