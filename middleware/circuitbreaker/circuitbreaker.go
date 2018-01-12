package circuitbreaker

import (
	"sync/atomic"
	"time"
)

// State ...
type State uint32

const (
	// StatusClose cb is closed
	StatusClose State = iota
	// StatusHalfOpen cb is halfopen
	StatusHalfOpen
	// StatusOpen cb is open
	StatusOpen
)

func (s State) String() string {
	switch s {
	case StatusClose:
		return "close"
	case StatusHalfOpen:
		return "halfopen"
	case StatusOpen:
		return "open"
	}

	return ""
}

type count struct {
	total                uint64
	totalSuccesses       uint64
	totalFailures        uint64
	consecutiveSuccesses uint64
	consecutiveFailures  uint64
}

func (c *count) onReqest() {
	atomic.AddUint64(&c.total, 1)
}

func (c *count) onSuccess() {
	atomic.AddUint64(&c.totalSuccesses, 1)
	atomic.AddUint64(&c.consecutiveSuccesses, 1)
	atomic.StoreUint64(&c.consecutiveFailures, 0)
}

func (c *count) onFailure() {
	atomic.AddUint64(&c.totalFailures, 1)
	atomic.AddUint64(&c.consecutiveFailures, 1)
	atomic.StoreUint64(&c.consecutiveSuccesses, 0)
}

func (c *count) reset() {
	atomic.StoreUint64(&c.total, 0)
	atomic.StoreUint64(&c.totalSuccesses, 0)
	atomic.StoreUint64(&c.totalFailures, 0)
	atomic.StoreUint64(&c.consecutiveSuccesses, 0)
	atomic.StoreUint64(&c.consecutiveFailures, 0)
}

// CircuitBreaker is a state machine to prevent sending requests that are likely to fail.
type CircuitBreaker struct {
	Name string

	Timeout time.Duration

	State State

	count count
}

func (cd *CircuitBreaker) Exec(f func() (interface{}, error)) (interface{}, error) {
	return f()
}
