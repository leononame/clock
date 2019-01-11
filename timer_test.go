package clock

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFakeTimer(t *testing.T) {
	tests := []struct {
		duration time.Duration
		count    int32
	}{
		{time.Second, 0},
		{time.Second * 59, 0},
		{time.Minute, 1},
		{time.Hour, 1},
	}
	for _, test := range tests {
		executed := int32(0)
		clock := NewMock()
		timer := clock.NewTimer(time.Minute)
		go func() {
			<-timer.Chan()
			atomic.AddInt32(&executed, 1)
		}()
		clock.Forward(test.duration)
		time.Sleep(time.Millisecond * 10)
		assert.Equal(t, test.count, atomic.LoadInt32(&executed))
	}
}

func TestFakeTimer_Stop(t *testing.T) {
	tests := []struct {
		duration time.Duration
		stopped  bool
		count    int32
	}{
		{time.Second, true, 0},
		{time.Second * 2, false, 1},
	}
	for _, test := range tests {
		executed := int32(0)
		clock := NewMock()
		timer := clock.NewTimer(time.Second * 2)
		go func() {
			<-timer.Chan()
			atomic.AddInt32(&executed, 1)
		}()
		clock.Forward(test.duration)
		assert.Equal(t, test.stopped, timer.Stop())
		assert.Equal(t, test.count, atomic.LoadInt32(&executed))
	}
}

func TestFakeTimer_Reset(t *testing.T) {
	tests := []struct {
		duration1 time.Duration
		duration2 time.Duration
		stopped   bool
		count1    int32
		count2    int32
	}{
		{time.Second * 59, time.Minute, true, 0, 1},
		{time.Second * 59, time.Second * 59, true, 0, 0},
		{time.Minute, time.Minute, false, 1, 2},
	}
	for _, test := range tests {
		executed := int32(0)
		clock := NewMock()
		timer := clock.NewTimer(time.Minute)
		go func() {
			for {
				<-timer.Chan()
				atomic.AddInt32(&executed, 1)
			}
		}()

		clock.Forward(test.duration1)
		time.Sleep(time.Millisecond * 10)

		assert.Equal(t, test.count1, atomic.LoadInt32(&executed))
		assert.Equal(t, test.stopped, timer.Reset(time.Minute))
		clock.Forward(test.duration2)
		assert.Equal(t, test.count2, atomic.LoadInt32(&executed))
	}
}
