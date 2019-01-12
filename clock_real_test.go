package clock

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClock_Now(t *testing.T) {
	count := 0
	for i := 0; i < 10; i++ {
		t1 := time.Now().Round(time.Second)
		t2 := New().Now().Round(time.Second)
		if t1 == t2 {
			count++
		}
	}
	// At least 9 times the time should match. We introduce this
	// because of the slight chance that one of the comparisons
	// is done just when the second changes
	assert.True(t, count > 8)
}

func TestClock_Since(t *testing.T) {
	c := New()
	since := c.Now().Add(time.Minute * -1)
	count := 0
	for i := 0; i < 10; i++ {
		t1 := time.Since(since).Round(time.Second)
		t2 := c.Since(since).Round(time.Second)
		if t1 == t2 {
			count++
		}
	}
	// At least 9 times the time should match. We introduce this
	// because of the slight chance that one of the comparisons
	// is done just when the second changes
	assert.True(t, count > 8)
}

func TestClock_Until(t *testing.T) {
	c := New()
	until := c.Now().Add(time.Minute)
	count := 0
	for i := 0; i < 10; i++ {
		t1 := time.Until(until).Round(time.Second)
		t2 := c.Until(until).Round(time.Second)
		if t1 == t2 {
			count++
		}
	}
	// At least 9 times the time should match. We introduce this
	// because of the slight chance that one of the comparisons
	// is done just when the second changes
	assert.True(t, count > 8)
}

func TestClock_Sleep(t *testing.T) {
	var count int32
	go func() {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&count, 1)
	}()
	go func() {
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&count, 1)

	}()
	sched()

	New().Sleep(25 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&count))
}

func TestClock_After(t *testing.T) {
	var count int32
	go incUponReceive(New().After(25*time.Millisecond), &count)
	time.Sleep(40 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&count))
}

func TestClock_AfterFunc(t *testing.T) {
	var count int32
	New().AfterFunc(time.Millisecond*25, func() { atomic.AddInt32(&count, 1) })
	time.Sleep(40 * time.Millisecond)
	assert.Equal(t, int32(1), count)
}

func TestClock_NewTimer(t *testing.T) {
	var count int32
	go incUponReceive(New().NewTimer(25*time.Millisecond).Chan(), &count)
	time.Sleep(40 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&count))
}

func TestClock_NewTicker(t *testing.T) {
	var count int32
	go incUponReceive(New().NewTicker(25*time.Millisecond).Chan(), &count)
	time.Sleep(120 * time.Millisecond)
	assert.Equal(t, int32(4), atomic.LoadInt32(&count))
}
