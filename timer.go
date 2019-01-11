package clock

import (
	"sync"
	"time"
)

type Timer interface {
	Chan() <-chan time.Time
	Stop() bool
	Reset(time.Duration) bool
}

type realTimer struct {
	*time.Timer
}

func (r *realTimer) Chan() <-chan time.Time {
	return r.C
}

type fakeTimer struct {
	mu      sync.RWMutex
	ch      chan time.Time
	due     time.Time
	clock   *Mock
	changed chan time.Time
	stopped bool
}

func (f *fakeTimer) Chan() <-chan time.Time {
	return f.ch
}

func (f *fakeTimer) Stop() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.stopped {
		return false
	}
	f.stopped = true
	return true
}

func (f *fakeTimer) Reset(d time.Duration) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.due = f.clock.Now().Add(d)

	if f.stopped {
		f.stopped = false
		return false
	}
	return true
}

func (f *fakeTimer) tick() {
	for {
		t := <-f.changed

		f.mu.RLock()
		if f.stopped || f.due.After(t) {
			f.mu.RUnlock()
			continue
		}
		f.mu.RUnlock()

		f.mu.Lock()
		f.ch <- f.due
		f.stopped = true
		f.mu.Unlock()
	}
}
