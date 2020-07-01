package utilz

import (
	"os"
	"os/signal"
	"sync"
	"time"
)

type Handler func(os.Signal) bool

// Notify calls handler when gets specified signals and pass given signal to
// handler. If handler returns false, notify stops waiting for signals.
func Notify(handler Handler, signals ...os.Signal) {
	pipe := make(chan os.Signal, 1)
	signal.Notify(pipe, signals...)

	for sign := range pipe {
		if !handler(sign) {
			return
		}
	}
}
func DefaultNotify(handler func(os.Signal) bool) {
	Notify(handler, os.Interrupt, os.Kill)
}

// WaitHasTimedout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func WaitHasTimedout(wg *sync.WaitGroup, timeout time.Duration) (timedout bool) {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
