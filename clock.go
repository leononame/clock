// Package clock provides an abstraction layer around tickers. These abstraction layers can be used during testing
// tock mock ticking behaviour.
package clock

import (
	"runtime"
	"sort"
	"sync"
	"time"
)

// Executer is an interface that allows the fake clock implementation to abstract Timer and Ticker. This way,
// both types can be united in the same slice and thus be sorted by their next execution
type Executer interface {
	// NextExecution returns the next execution time
	NextExecution() time.Time
	// Execute executes the Executer object
	Execute(time.Time)
}

// Clock provides an abstraction for often-used time functions
type Clock interface {
	// After waits for the duration to elapse and then sends the current time on the returned channel.
	After(d time.Duration) <-chan time.Time
	// AfterFunc waits for the duration to elapse and then executes a function.
	// A Timer is returned that can be stopped.
	AfterFunc(d time.Duration, fn func()) Timer
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

// AfterFunc waits for the duration to elapse and then executes a function.
// A Timer is returned that can be stopped.
func (c *clock) AfterFunc(d time.Duration, fn func()) Timer { return &realTimer{time.AfterFunc(d, fn)} }

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
	timers  []Executer
}

// Len returns the number of internal Timers or Tickers that are being tracked.
func (m *Mock) Len() int { return len(m.timers) }

// Swap swaps the elements at i and j in the internal tracker for Timers and Tickers
func (m *Mock) Swap(i, j int) { m.timers[i], m.timers[j] = m.timers[j], m.timers[i] }

// Less indicates whether Executer at position i should be executed before Executer at position j
func (m *Mock) Less(i, j int) bool {
	return m.timers[i].NextExecution().Before(m.timers[j].NextExecution())
}

// Forward moves the internal time forward by the amount specified. Any timers or tickers that fire during that time
// period will be activated
func (m *Mock) Forward(d time.Duration) {
	m.mu.Lock()
	t := m.now.Add(d)
	m.now = t
	m.mu.Unlock()
	m.tick(t)
	sched()
}

// Set sets the internal time to a specific point in time. Any timers or tickers that fire during that time
// period will be activated
func (m *Mock) Set(t time.Time) {
	m.mu.Lock()
	m.now = t
	m.mu.Unlock()
	m.tick(t)
	sched()
}

// RunUntilDone sets the internal time to the point in time where the latest timer will be fired.
// This means, all After() and AfterFunc() calls will have fired.
// Since tickers potentially run forever, they aren't included.
func (m *Mock) RunUntilDone() {
	var last time.Time
	m.mu.RLock()
	for _, t := range m.timers {
		timer, ok := t.(*fakeTimer)
		if ok && timer.NextExecution().After(last) {
			last = timer.NextExecution()
		}
	}
	m.mu.RUnlock()
	m.Set(last)
}

// tick sends an event to all tickers and timers informing them that time has changed.
func (m *Mock) tick(t time.Time) {
	for m.tickNext(t) {
	}
}

// tickNext executes the next Timer or Ticker in the queue
func (m *Mock) tickNext(t time.Time) bool {
	m.mu.Lock()
	sort.Sort(m)
	if len(m.timers) == 0 {
		m.mu.Unlock()
		return false
	}
	n := m.timers[0]
	if n.NextExecution().After(t) {
		m.mu.Unlock()
		return false
	}
	m.mu.Unlock()
	n.Execute(t)
	return true
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

// AfterFunc waits for the duration to elapse and then executes a function.
// A Timer is returned that can be stopped.
func (m *Mock) AfterFunc(d time.Duration, fn func()) Timer {
	t := m.fakeTimer(d)
	t.fn = fn
	sched()
	return t
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
	t.ch = make(chan time.Time, 1)
	t.clock = m
	t.d = d
	t.next = m.Now().Add(d)
	m.addTimer(&t)
	return &t
}

// NewTimer creates a new Timer that will send
// the current time on its channel after at least duration d.
func (m *Mock) NewTimer(d time.Duration) Timer {
	t := m.fakeTimer(d)
	t.ch = make(chan time.Time, 1)
	return t
}

// fakeTimer returns a fakeTimer object with some standard setup
func (m *Mock) fakeTimer(d time.Duration) *fakeTimer {
	t := fakeTimer{}
	// Set this to nil expressively to show that this Timer will not do anything
	t.ch = nil
	t.fn = nil
	t.due = m.Now().Add(d)
	t.clock = m

	m.addTimer(&t)
	return &t
}

// removeTimer removes a given Executer from the list of timers
func (m *Mock) removeTimer(t Executer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, ti := range m.timers {
		if t == ti {
			m.timers[i] = m.timers[len(m.timers)-1]
			m.timers[len(m.timers)-1] = nil
			m.timers = m.timers[:len(m.timers)-1]
			return
		}
	}
}

// addTimer adds an Executer to the list of timers
func (m *Mock) addTimer(t Executer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.timers = append(m.timers, t)
}

// sched calls the go scheduler. Implementation might change to time.Sleep(time.Millisecond).
func sched() {
	runtime.Gosched()
}
