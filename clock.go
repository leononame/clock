// Package clock provides an abstraction layer around tickers. These abstraction layers can be used during testing
// tock mock ticking behaviour.
package clock

import (
	"runtime"
	"sync"
	"time"
)

type Clock interface {
	After(d time.Duration) <-chan time.Time
	Now() time.Time
	Since(time.Time) time.Duration
	Until(time.Time) time.Duration
	Sleep(time.Duration)
	NewTicker(d time.Duration) Ticker
	NewTimer(d time.Duration) Timer
}

func New() Clock {
	return &clock{}
}

func NewMock() *Mock {
	m := &Mock{}
	m.changed = make(chan time.Time)
	return m
}

type clock struct{}

func (c *clock) After(d time.Duration) <-chan time.Time { return time.After(d) }
func (c *clock) Now() time.Time                         { return time.Now() }
func (c *clock) Since(t time.Time) time.Duration        { return time.Since(t) }
func (c *clock) Until(t time.Time) time.Duration        { return time.Until(t) }
func (c *clock) Sleep(d time.Duration)                  { time.Sleep(d) }
func (c *clock) NewTicker(d time.Duration) Ticker       { return &realTicker{time.NewTicker(d)} }
func (c *clock) NewTimer(d time.Duration) Timer         { return &realTimer{time.NewTimer(d)} }

type Mock struct {
	mu      sync.RWMutex
	now     time.Time
	changed chan time.Time
	timers  []*fakeTimer
	tickers []*fakeTicker
}

func (m *Mock) Forward(d time.Duration) {
	m.mu.Lock()
	t := m.now.Add(d)
	m.now = t
	m.tick(t)
	m.mu.Unlock()
	sched()
}

func (m *Mock) Set(t time.Time) {
	m.mu.Lock()
	m.now = t
	m.tick(t)
	m.mu.Unlock()
	sched()
}

func (m *Mock) tick(t time.Time) {
	for _, ti := range m.tickers {
		ti.changed <- t
	}
	for _, ti := range m.timers {
		ti.changed <- t
	}
}

func (m *Mock) Now() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.now
}

func (m *Mock) After(d time.Duration) <-chan time.Time {
	t := m.NewTimer(d)
	return t.Chan()
}
func (m *Mock) Since(t time.Time) time.Duration { return m.Now().Sub(t) }
func (m *Mock) Until(t time.Time) time.Duration { return t.Sub(m.Now()) }
func (m *Mock) Sleep(d time.Duration) {
	<-m.After(d)
}

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

func sched() {
	runtime.Gosched()
}
