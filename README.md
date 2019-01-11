# clock

Clock is a library for mocking the time package in go. It supports mocking time.Timer and time.Ticker for testing purposes.

This library is heavily inspired by [benbjohnson/clock](https://github.com/benbjohnson/clock) and [jonboulle/clockwork](https://github.com/jonboulle/clockwork).

## Usage

Instead of `time`, use the `clock.Clock` interface. In your real application, you can call `clock.New()` to get the actual instance. For testing purposes, call `clock.NewMock()`.

