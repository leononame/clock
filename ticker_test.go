package clock

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func incUponReceive(ch <-chan time.Time, counter *int32) {
	for {
		<-ch
		atomic.AddInt32(counter, 1)
	}
}

func TestFakeTicker(t *testing.T) {
	tests := []struct {
		duration time.Duration
		count    int32
	}{
		{time.Hour, 1},
		{time.Hour * 2, 2},
		{time.Hour * 5, 5},
		{time.Hour * 24, 24},
		{time.Minute * 59, 0},
		{time.Hour + time.Minute*59, 1},
	}
	for _, test := range tests {
		executed := int32(0)
		clock := NewMock()
		ch := clock.NewTicker(time.Hour).Chan()
		go incUponReceive(ch, &executed)
		sched()
		clock.Forward(test.duration)
		// go to sleep because we need our goroutine to block on channel read before we compare results
		time.Sleep(time.Millisecond * 10)
		assert.Equal(t, test.count, atomic.LoadInt32(&executed))
	}
}

func TestFakeTicker_Stop(t *testing.T) {
	tests := []struct {
		beforeStop time.Duration
		afterStop  time.Duration
		count      int32
	}{
		{time.Hour, time.Hour, 1},
		{time.Hour * 2, time.Hour, 2},
		{time.Hour * 2, time.Hour * 2, 2},
		{time.Minute, time.Hour * 24, 0},
	}
	for _, test := range tests {
		executed := int32(0)
		clock := NewMock()
		ticker := clock.NewTicker(time.Hour)
		go incUponReceive(ticker.Chan(), &executed)
		sched()
		clock.Forward(test.beforeStop)
		ticker.Stop()
		clock.Forward(test.afterStop)
		// go to sleep because we need our goroutine to block on channel read before we compare results
		time.Sleep(time.Millisecond * 10)
		assert.Equal(t, test.count, atomic.LoadInt32(&executed))
	}

}

// Make sure the ticker won't block when not read
func TestFakeTicker_Unread(t *testing.T) {
	clock := NewMock()
	ticker := clock.NewTicker(time.Microsecond)
	clock.Forward(10 * time.Microsecond)
	ticker.Stop()
}

func TestFakeTicker_Multiple(t *testing.T) {
	clock := NewMock()
	var executions [10]int32

	for i := 0; i < 10; i++ {
		ticker := clock.NewTicker(time.Second * time.Duration(i+1))
		go func(i int) {
			for {
				<-ticker.Chan()
				atomic.AddInt32(&executions[i], 1)
			}
		}(i)
	}
	sched()
	clock.Forward(20 * time.Second)
	time.Sleep(time.Microsecond * 100)
	for i := 0; i < 10; i++ {
		assert.Equal(t, int32(20/(i+1)), atomic.LoadInt32(&executions[i]))
	}
}
