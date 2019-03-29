# clock  [![GoDoc](https://godoc.org/github.com/leononame/clock?status.svg)](https://godoc.org/github.com/leononame/clock) [![Build Status](https://cloud.drone.io/api/badges/leononame/clock/status.svg)](https://cloud.drone.io/leononame/clock) [![codecov](https://codecov.io/gh/leononame/clock/branch/master/graph/badge.svg)](https://codecov.io/gh/leononame/clock)

Clock is a library for mocking the time package in go. It supports mocking time.Timer and time.Ticker for testing purposes.

This library is heavily inspired by [benbjohnson/clock](https://github.com/benbjohnson/clock) and [jonboulle/clockwork](https://github.com/jonboulle/clockwork).

## Usage

Instead of `time`, use the `clock.Clock` interface. In your real application, you can call `clock.New()` to get the actual instance. For testing purposes, call `clock.NewMock()`.

```go
import (
	"testing"

	"github.com/leononame/clock"
)
func TestApp_Do(t *testing.T) {
    clock := clock.NewMock()
    context := Context{Clock: clock}
}
```

## 

The interface exposes standard `time` functions that can be used like usual. 

```go
import (
    "time"

    "github.com/leononame/clock"
)

func timer() {
    c := clock.NewMock() // or clock.New()
    t := c.NewTicker(time.Second)
    now := c.Now()
}
```