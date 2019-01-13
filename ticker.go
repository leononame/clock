package clock

import (
	"sync"
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
	mu      sync.RWMutex
	ch      chan time.Time
	clock   *Mock
	d       time.Duration
	next    time.Time
	stopped bool
}

// Chan returns the readonly channel of the ticker.
func (f *fakeTicker) Chan() <-chan time.Time {
	return f.ch
}

// Stop stops the ticker. No more events will be sent through the channel
func (f *fakeTicker) Stop() {
	f.mu.Lock()
	f.stopped = true
	f.mu.Unlock()
	f.clock.removeTimer(f)
}

// Execute executes the Ticker
func (f *fakeTicker) Execute(t time.Time) {
	f.mu.RLock()
	next := f.next
	stopped := f.stopped
	f.mu.RUnlock()

	if stopped {
		return
	}

	f.mu.Lock()
	f.next = next.Add(f.d)
	f.mu.Unlock()

	select {
	case f.ch <- next:
	default:
	}
	sched()
}

// NextExecution returns the next execution time
func (f *fakeTicker) NextExecution() time.Time {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.next
}
