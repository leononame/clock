package clock

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMock_Forward(t *testing.T) {
	c := NewMock()
	n := c.Now()
	diff := time.Hour
	c.Forward(diff)
	assert.Equal(t, c.Now(), n.Add(diff))
}

func TestMock_Set(t *testing.T) {
	c := NewMock()
	target := time.Unix(1000, 0)
	assert.NotEqual(t, c.Now(), target)
	c.Set(target)
	time.Sleep(time.Millisecond * 10)
	assert.Equal(t, c.Now(), target)
}

func TestMock_Since(t *testing.T) {
	c := NewMock()
	target := time.Unix(1000, 0)
	diff := time.Hour
	c.Set(target.Add(diff))
	assert.Equal(t, c.Since(target), diff)
}

func TestMock_Until(t *testing.T) {
	c := NewMock()
	target := time.Unix(1000, 0)
	diff := time.Hour
	c.Set(target)
	assert.Equal(t, c.Until(target.Add(diff)), diff)
}

func TestMock_After(t *testing.T) {
	received := int32(0)
	clock := NewMock()
	ch := clock.After(time.Minute)
	go func() {
		<-ch
		atomic.AddInt32(&received, 1)
	}()

	clock.Forward(time.Second * 59)
	assert.Zero(t, atomic.LoadInt32(&received))
	clock.Forward(time.Second)
	assert.NotZero(t, atomic.LoadInt32(&received))
}

func TestMock_AfterFunc(t *testing.T) {
	received := int32(0)
	clock := NewMock()
	fn := func() {
		atomic.AddInt32(&received, 1)
	}
	clock.AfterFunc(time.Minute, fn)

	clock.Forward(time.Second * 59)
	assert.Zero(t, atomic.LoadInt32(&received))
	clock.Forward(time.Second)
	assert.NotZero(t, atomic.LoadInt32(&received))

}

func TestMock_Sleep(t *testing.T) {
	received := int32(0)
	clock := NewMock()
	go func() {
		clock.Sleep(time.Hour + time.Second)
		atomic.AddInt32(&received, 1)
	}()
	sched()

	clock.Forward(time.Hour)
	assert.Zero(t, atomic.LoadInt32(&received))
	clock.Forward(time.Second)
	// Go to sleep just in case the goroutine wasn't scheduled yet
	time.Sleep(time.Millisecond)
	assert.NotZero(t, atomic.LoadInt32(&received))
}
