package mqtt

import "time"

const (
	timerReset = iota
	timerStop  = iota
)

type keepaliveAction func()

func pinger(keepalive int, pingTimerCtl chan int, action keepaliveAction) {
	pingTimer := time.NewTimer(time.Duration(keepalive) * time.Second)
	defer pingTimer.Stop()
Loop:
	for {
		select {
		case <-pingTimer.C:
			action()
			pingTimer.Reset(time.Duration(keepalive) * time.Second)
		case msg := <-pingTimerCtl:
			switch msg {
			case timerReset:
				pingTimer.Reset(time.Duration(keepalive) * time.Second)
			case timerStop:
				pingTimer.Stop()
				break Loop
			}
		}
	}
}

func startKeepalive(c *Client, action keepaliveAction) chan int {
	channel := make(chan int)
	go pinger(c.options.Keepalive, channel, action)
	return channel
}
