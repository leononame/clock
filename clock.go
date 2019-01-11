// Package clock provides an abstraction layer around tickers. These abstraction layers can be used during testing
// tock mock ticking behaviour.
package clock

import (
	"runtime"
	"sync"
	"time"
)

// Clock provides an abstraction for often-used time functions
type Clock interface {
	// After waits for the duration to elapse and then sends the current time on the returned channel.
	After(d time.Duration) <-chan time.Time
	// Now returns the current local time.
	Now() time.Time
	// Since returns the time elapsed since t.
	Since(time.Time) time.Duration
	// Until returns the duration until t.
	Until(time.Time) time.Duration
	// Sleep pauses the current goroutine for at least the duration d.
	// A negative or zero duration causes Sleep to return immediately.
	Sleep(time.Duration)
	// NewTicker returns a new Ticker containing a channel that will send the
	// time with a period specified by the duration argument.
	NewTicker(d time.Duration) Ticker
	// NewTimer creates a new Timer that will send
	// the current time on its channel after at least duration d.
	NewTimer(d time.Duration) Timer
}

// New returns a Clock implementation based on the time package and is good for usage in deployed applications.
func New() Clock {
	return &clock{}
}

// NewMock returns a Mock which implements Clock. It can be used to mock the current time in tests.
// When a new Mock is created, it starts with Unix timestamp 0.
func NewMock() *Mock {
	m := &Mock{}
	m.changed = make(chan time.Time)
	m.now = time.Unix(0, 0)
	return m
}

// clock is a wrapper type that implements the standard time functions.
type clock struct{}

// After waits for the duration to elapse and then sends the current time on the returned channel.
func (c *clock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// Now returns the current local time.
func (c *clock) Now() time.Time { return time.Now() }

// Since returns the time elapsed since t.
func (c *clock) Since(t time.Time) time.Duration { return time.Since(t) }

// Until returns the duration until t.
func (c *clock) Until(t time.Time) time.Duration { return time.Until(t) }

// Sleep pauses the current goroutine for at least the duration d.
// A negative or zero duration causes Sleep to return immediately.
func (c *clock) Sleep(d time.Duration) { time.Sleep(d) }

// NewTicker returns a new Ticker containing a channel that will send the
// time with a period specified by the duration argument.
func (c *clock) NewTicker(d time.Duration) Ticker { return &realTicker{time.NewTicker(d)} }

// NewTimer creates a new Timer that will send
// the current time on its channel after at least duration d.
func (c *clock) NewTimer(d time.Duration) Timer { return &realTimer{time.NewTimer(d)} }

// Mock is a type used for mocking the time package during tests.
type Mock struct {
	mu      sync.RWMutex
	now     time.Time
	changed chan time.Time
	timers  []*fakeTimer
	tickers []*fakeTicker
}

// Forward moves the internal time forward by the amount specified. Any timers or tickers that fire during that time
// period will be activated
func (m *Mock) Forward(d time.Duration) {
	m.mu.Lock()
	t := m.now.Add(d)
	m.now = t
	m.tick(t)
	m.mu.Unlock()
	sched()
}

// Set sets the internal time to a specific point in time. Any timers or tickers that fire during that time
// period will be activated
func (m *Mock) Set(t time.Time) {
	m.mu.Lock()
	m.now = t
	m.tick(t)
	m.mu.Unlock()
	sched()
}

// tick sends an event to all tickers and timers informing them that time has changed.
// tick is not thread-safe and Mock should be locked before calling this.
func (m *Mock) tick(t time.Time) {
	for _, ti := range m.tickers {
		ti.changed <- t
	}
	for _, ti := range m.timers {
		ti.changed <- t
	}
}

// Now returns the current internal time as either set by Set() or forwarded by Forward().
func (m *Mock) Now() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.now
}

// After behaves like time.After. However, it only fires when the internal time is forwarded by at least d.
func (m *Mock) After(d time.Duration) <-chan time.Time {
	t := m.NewTimer(d)
	return t.Chan()
}

// Since returns the time elapsed since t in comparison to the internal time.
func (m *Mock) Since(t time.Time) time.Duration { return m.Now().Sub(t) }

// Until returns the duration until t in comparison to the internal time.
func (m *Mock) Until(t time.Time) time.Duration { return t.Sub(m.Now()) }

// Sleep pauses the current goroutine for at least the duration d in comparison to the internal time.
func (m *Mock) Sleep(d time.Duration) {
	<-m.After(d)
}

// NewTicker returns a new Ticker containing a channel that will send the
// time with a period specified by the duration argument.
func (m *Mock) NewTicker(d time.Duration) Ticker {
	t := fakeTicker{}
	t.ch = make(chan time.Time)
	// Set stop channel size to 1 so stopping a ticker doesn't block if it isn't read
	t.stop = make(chan struct{}, 1)
	t.changed = make(chan time.Time)
	t.clock = m
	t.d = d
	m.mu.Lock()
	m.tickers = append(m.tickers, &t)
	m.mu.Unlock()
	go t.tick()
	sched()
	return &t
}

// NewTimer creates a new Timer that will send
// the current time on its channel after at least duration d.
func (m *Mock) NewTimer(d time.Duration) Timer {
	t := fakeTimer{}
	t.ch = make(chan time.Time)
	t.due = m.Now().Add(d)
	t.clock = m
	t.changed = make(chan time.Time)
	m.mu.Lock()
	m.timers = append(m.timers, &t)
	m.mu.Unlock()
	go t.tick()
	sched()
	return &t
}

// sched calls the go scheduler. Implementation might change to time.Sleep(time.Millisecond).
func sched() {
	runtime.Gosched()
}
