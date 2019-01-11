package clock

import (
	"time"
)

type Ticker interface {
	Chan() <-chan time.Time
	Stop()
}

type realTicker struct {
	*time.Ticker
}

func (r *realTicker) Chan() <-chan time.Time {
	return r.C
}

type fakeTicker struct {
	ch      chan time.Time
	stop    chan struct{}
	changed chan time.Time
	clock   *Mock
	d       time.Duration
}

func (f *fakeTicker) Chan() <-chan time.Time {
	return f.ch
}

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
