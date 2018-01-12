package ratelimit

import (
	"context"
	"time"

	"github.com/philchia/zen"
)

var _ Ratelimiter = (*ratelimiter)(nil)

// Ratelimiter ...
type Ratelimiter interface {
	Allow() bool

	Wait(context.Context) error
	WaitN(context.Context, int) error

	MiddleWare() zen.Middleware
}

type ratelimiter struct {
	limiter chan struct{}
	tick    *time.Ticker
}

// New create a new rate limiter
func New(n int) Ratelimiter {
	ret := &ratelimiter{}
	ret.limiter = make(chan struct{}, 1)
	ret.tick = time.NewTicker(time.Second / time.Duration(n))

	go ret.ticker()

	return ret
}

func (r *ratelimiter) ticker() {
	for range r.tick.C {
		select {
		case r.limiter <- struct{}{}:
		default:
		}
	}
}

// Allow request a token from limiter
func (r *ratelimiter) Allow() bool {
	var ret bool

	select {
	case <-r.limiter:
		ret = true
	default:
	}
	return ret
}

// Wait for token until ctx timeout
func (r *ratelimiter) Wait(ctx context.Context) error {
	var ret error
	select {
	case <-r.limiter:
		ret = nil
	case <-ctx.Done():
		ret = ctx.Err()
	}
	return ret
}

// WaitN for N token until ctx timeout
func (r *ratelimiter) WaitN(ctx context.Context, n int) error {
	for i := 0; i < n; i++ {
		select {
		case <-r.limiter:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (r *ratelimiter) MiddleWare() zen.Middleware {
	return r.Wrap
}

// Wrap is a middleware
func (r *ratelimiter) Wrap(handler zen.HandlerFunc) zen.HandlerFunc {
	return func(ctx *zen.Context) {
		if !r.Allow() {
			ctx.WriteStatus(zen.StatusGatewayTimeout)
			return
		}
		handler(ctx)
	}
}
