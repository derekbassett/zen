package zen

import (
	"net/http"
)

type (
	// HandlerFunc is a type alias for handler
	HandlerFunc func(Context)
	// Middleware accept a HandlerFunc and return another HandlerFunc
	Middleware func(HandlerFunc) HandlerFunc
	// Middlewares contains list of Middleware
	Middlewares []Middleware
)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	UnWrapF(f).ServeHTTP(w, r)
}

// WrapF wrap a http handlerfunc into HandlerFunc
func WrapF(h http.HandlerFunc) HandlerFunc {
	return func(ctx Context) {
		h(ctx.Rw, ctx.Req)
	}
}

// UnWrapF ...
func UnWrapF(h HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		ctx := getContext(rw, req)
		h(ctx)
	}
}

// WrapStdMiddleware add support for 3rd party http middleware
func WrapStdMiddleware(middleware func(http.HandlerFunc) http.HandlerFunc) Middleware {
	return func(h HandlerFunc) HandlerFunc {
		return WrapF(middleware(UnWrapF(h)))
	}
}

// Wrap middlewares into HandlerFunc
func (m Middlewares) Wrap(h HandlerFunc) HandlerFunc {
	var ret = h
	for _, w := range m {
		ret = w(ret)
	}
	return ret
}
