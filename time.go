package utilz

import "time"

const (
	FilenameTimeFormat = "Mon02Jan2006_15.04.05"
)

// RetryExponentialBackoff executes task; if it returns an error,
// it will retry executing the task the specified number of times before giving up,
// and at each retry it will sleep double the amout of time of the previous retry;
// the initial sleep time is specified by the user.
func RetryExponentialBackoff(attempts int, initialSleep time.Duration, task func() error) []error {
	var errs []error
	for i := 1; i <= attempts; i++ {
		err := task()
		if err == nil {
			// if one of the runs succeeds, do not return any previous errors
			return nil
		}
		errs = append(errs, err)
		time.Sleep(initialSleep)
		initialSleep = initialSleep * 2
	}
	return errs
}

func RetryLinearBackoff(attempts int, sleep time.Duration, task func() error) []error {
	var errs []error
	for i := 1; i <= attempts; i++ {
		err := task()
		if err == nil {
			// if one of the runs succeeds, do not return any previous errors
			return nil
		}
		errs = append(errs, err)
		time.Sleep(sleep)
	}
	return errs
}
func FormatErrorArray(prefix string, errs []error) string {
	var res string
	for i, err := range errs {
		isLast := i == len(errs)-1

		newline := "\n"
		if isLast {
			newline = ""
		}
		res += Sf(
			"%s%v: %s%s",
			prefix,
			i+1,
			err,
			newline,
		)
	}
	return res
}
func KitchenTimeNow() string {
	return time.Now().Format("15:04:05")
}
func KitchenTimeMsNow() string {
	return time.Now().Format("15:04:05.999")
}

// NewTicker returns a time channel that ticks ever specified interval,
// with the initial tick righ when you start listening.
func NewTicker(duration time.Duration) <-chan time.Time {
	c := make(chan time.Time, 1)

	go func() {
		c <- time.Now()
		ticker := time.NewTicker(duration)

		for tick := range ticker.C {
			// Ticks immediately and every duration thereafter
			c <- tick
		}
	}()

	return c
}

/// ticker countdown
type Ticker interface {
	Duration() time.Duration
	Tick()
	Stop()
}

type ticker struct {
	*time.Ticker
	d time.Duration
}

func (t *ticker) Tick()                   { <-t.C }
func (t *ticker) Duration() time.Duration { return t.d }

func NewTickerCountdown(d time.Duration) Ticker {
	return &ticker{time.NewTicker(d), d}
}

type TickFunc func(d time.Duration)

func Countdown(ticker Ticker, duration time.Duration) chan time.Duration {
	remainingCh := make(chan time.Duration, 1)
	go func(ticker Ticker, dur time.Duration, remainingCh chan time.Duration) {
		for remaining := duration; remaining >= 0; remaining -= ticker.Duration() {
			remainingCh <- remaining
			ticker.Tick()
		}
		ticker.Stop()
		close(remainingCh)
	}(ticker, duration, remainingCh)
	return remainingCh
}
func NewTimer() func() time.Duration {
	start := time.Now()
	return func() time.Duration {
		return time.Now().Sub(start).Round(time.Second)
	}
}

func NewTimerRaw() func() time.Duration {
	start := time.Now()
	return func() time.Duration {
		return time.Now().Sub(start)
	}
}
