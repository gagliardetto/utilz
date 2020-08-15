package utilz

import (
	"fmt"
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

func WaitSystemSignal(messages ...string) {
	DefaultNotify(func(_ os.Signal) bool {
		defer func() {
			fmt.Fprintln(os.Stderr, toInterfaceArray(messages)...)
		}()
		return false
	})
}

// HardWaitSystemSignal calls the callback at first signal, and waits for
// the callback to finish; when the callback finishes, HardWaitSystemSignal also finishes.
// If a second sygnal is sent before the callback finishes, a hard os.Exit is done.
func HardWaitSystemSignal(callback func(), messages ...string) {
	signalNum := 0

	pipe := make(chan os.Signal, 1)
	signal.Notify(pipe, os.Interrupt, os.Kill)

Loop:
	for {
		select {
		case _, ok := <-pipe:
			if !ok {
				return
			}
			if signalNum == 0 {
				signalNum++

				fmt.Fprintln(os.Stderr, toInterfaceArray(messages)...)

				go func() {
					defer close(pipe)
					callback()
				}()
				continue Loop
			} else {
				Ln("Forcing exit")
				os.Exit(1)
			}

		}
	}

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
