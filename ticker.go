package clock

import (
	"time"
)

// Ticker is an abstraction for the type time.Ticker that can be mocked for tests.
type Ticker interface {
	// Chan returns the readonly channel of the ticker
	Chan() <-chan time.Time
	// Stop stops the ticker. No more events will be sent through the channel
	Stop()
}

// realTicker is just the type time.Ticker and implements the Ticker interface.
type realTicker struct {
	*time.Ticker
}

// Chan returns the readonly channel of the ticker
func (r *realTicker) Chan() <-chan time.Time {
	return r.C
}

// fakeTicker is a fake implementation of Ticker based on the time mocking in Mock.
type fakeTicker struct {
	ch      chan time.Time
	stop    chan struct{}
	changed chan time.Time
	clock   *Mock
	d       time.Duration
}

// Chan returns the readonly channel of the ticker.
func (f *fakeTicker) Chan() <-chan time.Time {
	return f.ch
}

// Stop stops the ticker. No more events will be sent through the channel
func (f *fakeTicker) Stop() {
	f.clock.mu.Lock()
	f.stop <- struct{}{}
	for i, t := range f.clock.tickers {
		if t == f {
			f.clock.tickers = append(f.clock.tickers[:i], f.clock.tickers[i+1:]...)
		}
	}
	f.clock.mu.Unlock()
}

// tick is a function checking the ticker. It should be started as a goroutine. Every time
// the fakeTicker receives a new time bump, it will check whether and how often it should tick.
func (f *fakeTicker) tick() {
	next := f.clock.Now().Add(f.d)
	for {
		select {
		case <-f.changed:
			n := f.clock.Now()
			for {
				if !next.After(n) {
					f.ch <- next
					next = next.Add(f.d)
				} else {
					break
				}
			}
		case <-f.stop:
			return
		}
	}
}
