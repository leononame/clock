package clock

import (
	"sync"
	"time"
)

// Timer is an abstraction for the type time.Timer that can be mocked for tests.
type Timer interface {
	// Chan returns the readonly channel of the ticker
	Chan() <-chan time.Time

	// Stop prevents the Timer from firing.
	// It returns true if the call stops the timer, false if the timer has already
	// expired or been stopped.
	// Stop does not close the channel, to prevent a read from the channel succeeding
	// incorrectly.
	//
	// From the official docs:
	// To prevent a timer created with NewTimer from firing after a call to Stop,
	// check the return value and drain the channel.
	// For example, assuming the program has not received from t.C already:
	//
	// 	if !t.Stop() {
	// 		<-t.C
	// 	}
	//
	// This cannot be done concurrent to other receives from the Timer's
	// channel.
	//
	// For a timer created with AfterFunc(d, f), if t.Stop returns false, then the timer
	// has already expired and the function f has been started in its own goroutine;
	// Stop does not wait for f to complete before returning.
	// If the caller needs to know whether f is completed, it must coordinate
	// with f explicitly.
	Stop() bool

	// Reset changes the timer to expire after duration d.
	// It returns true if the timer had been active, false if the timer had
	// expired or been stopped.
	//
	// From the official docs:
	// Resetting a timer must take care not to race with the send into t.C
	// that happens when the current timer expires.
	// If a program has already received a value from t.C, the timer is known
	// to have expired, and t.Reset can be used directly.
	// If a program has not yet received a value from t.C, however,
	// the timer must be stopped and—if Stop reports that the timer expired
	// before being stopped—the channel explicitly drained:
	//
	// 	if !t.Stop() {
	// 		<-t.C
	// 	}
	// 	t.Reset(d)
	//
	// This should not be done concurrent to other receives from the Timer's
	// channel.
	//
	// Note that it is not possible to use Reset's return value correctly, as there
	// is a race condition between draining the channel and the new timer expiring.
	// Reset should always be invoked on stopped or expired channels, as described above.
	// The return value exists to preserve compatibility with existing programs.
	Reset(time.Duration) bool
}

// realTimer is just the type time.Timer and implements the Timer interface.
type realTimer struct {
	*time.Timer
}

// Chan returns the readonly channel of the Timer.
func (r *realTimer) Chan() <-chan time.Time {
	return r.C
}

// fakeTimer is an implementation of Timer that's based on the time mocking done in Mock.
type fakeTimer struct {
	mu      sync.RWMutex
	ch      chan time.Time
	fn      func()
	due     time.Time
	clock   *Mock
	stopped bool
}

// Chan returns the readonly channel of the Timer.
func (f *fakeTimer) Chan() <-chan time.Time {
	return f.ch
}

// Stop prevents the Timer from firing.
// It returns true if the call stops the timer, false if the timer has already
// expired or been stopped.
// Stop does not close the channel, to prevent a read from the channel succeeding
// incorrectly.
//
// From the official docs:
// To prevent a timer created with NewTimer from firing after a call to Stop,
// check the return value and drain the channel.
// For example, assuming the program has not received from t.C already:
//
// 	if !t.Stop() {
// 		<-t.C
// 	}
//
// This cannot be done concurrent to other receives from the Timer's
// channel.
//
// For a timer created with AfterFunc(d, f), if t.Stop returns false, then the timer
// has already expired and the function f has been started in its own goroutine;
// Stop does not wait for f to complete before returning.
// If the caller needs to know whether f is completed, it must coordinate
// with f explicitly.
func (f *fakeTimer) Stop() bool {
	f.clock.removeTimer(f)
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.stopped {
		return false
	}
	f.stopped = true
	return true
}

// Reset changes the timer to expire after duration d.
// It returns true if the timer had been active, false if the timer had
// expired or been stopped.
//
// From the official docs:
// Resetting a timer must take care not to race with the send into t.C
// that happens when the current timer expires.
// If a program has already received a value from t.C, the timer is known
// to have expired, and t.Reset can be used directly.
// If a program has not yet received a value from t.C, however,
// the timer must be stopped and—if Stop reports that the timer expired
// before being stopped—the channel explicitly drained:
//
// 	if !t.Stop() {
// 		<-t.C
// 	}
// 	t.Reset(d)
//
// This should not be done concurrent to other receives from the Timer's
// channel.
//
// Note that it is not possible to use Reset's return value correctly, as there
// is a race condition between draining the channel and the new timer expiring.
// Reset should always be invoked on stopped or expired channels, as described above.
// The return value exists to preserve compatibility with existing programs.
func (f *fakeTimer) Reset(d time.Duration) bool {
	now := f.clock.Now()
	f.mu.Lock()
	f.due = now.Add(d)

	if f.stopped {
		f.stopped = false
		f.clock.addTimer(f)
		f.mu.Unlock()
		return false
	}
	f.mu.Unlock()

	return true
}

// Execute executes the Timer object
func (f *fakeTimer) Execute(t time.Time) {
	f.mu.RLock()
	if f.stopped {
		f.mu.RUnlock()
		return
	}
	f.mu.RUnlock()

	f.mu.Lock()
	if f.ch == nil {
		f.fn()
	} else {
		f.ch <- f.due
	}
	f.stopped = true
	f.mu.Unlock()
	f.clock.removeTimer(f)
}

// NextExecution returns the next execution time
func (f *fakeTimer) NextExecution() time.Time {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.due
}
