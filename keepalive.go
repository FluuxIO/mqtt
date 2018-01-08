package mqtt // import "fluux.io/mqtt"

import "time"

const (
	keepaliveReset = iota
	keepaliveStop
)

type keepaliveAction func()

func startKeepalive(keepaliveDuration int, action keepaliveAction) chan int {
	channel := make(chan int)
	go keepalive(keepaliveDuration, channel, action)
	return channel
}

func keepalive(keepalive int, pingTimerCtl chan int, action keepaliveAction) {
	timer := time.NewTimer(time.Duration(keepalive) * time.Second)
	defer timer.Stop()
Loop:
	for {
		select {
		case <-timer.C:
			action()
			timer.Reset(time.Duration(keepalive) * time.Second)
		case msg := <-pingTimerCtl:
			switch msg {
			case keepaliveReset:
				timer.Reset(time.Duration(keepalive) * time.Second)
			case keepaliveStop:
				timer.Stop()
				break Loop
			}
		}
	}
}
