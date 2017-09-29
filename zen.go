package zen

import (
	"net/http"
	"sync"
	"time"
)

const (
	// Version is current version num
	Version = "2.0"
)

// ensure Server implement http.Handler
var _ http.Handler = (*Server)(nil)

// global contextPool to reuse context
var contextPool = &sync.Pool{
	New: func() interface{} {
		c := Context{rw: &responseWriter{}}
		return &c
	},
}

type (
	methodTree struct {
		method string
		node   *node
	}

	// StatsHandler used to handle req statistic
	StatsHandler func(path string, begin time.Time, end time.Time)

	// Server struct
	Server struct {

		// internal http server
		once   sync.Once
		server http.Server
		// Router
		Router
		// global middleware
		interceptors Handlers
		// tier tree store all handlers
		trees []*methodTree

		// Enables automatic redirection if the current route can't be matched but a
		// handler for the path with (without) the trailing slash exists.
		// For example if /foo/ is requested but a route only exists for /foo, the
		// client is redirected to /foo with http status code 301 for GET requests
		// and 307 for all other request methods.
		RedirectTrailingSlash bool

		// If enabled, the router tries to fix the current request path, if no
		// handle is registered for it.
		// First superfluous path elements like ../ or // are removed.
		// Afterwards the router does a case-insensitive lookup of the cleaned path.
		// If a handle can be found for this route, the router makes a redirection
		// to the corrected path with status code 301 for GET requests and 307 for
		// all other request methods.
		// For example /FOO and /..//Foo could be redirected to /foo.
		// RedirectTrailingSlash is independent of this option.
		RedirectFixedPath bool

		// If enabled, the router checks if another method is allowed for the
		// current route, if the current request can not be routed.
		// If this is the case, the request is answered with 'Method Not Allowed'
		// and HTTP status code 405.
		// If no other Method is allowed, the request is delegated to the NotFound
		// handler.
		HandleMethodNotAllowed bool

		// If enabled, the router automatically replies to OPTIONS requests.
		// Custom OPTIONS handlers take priority over automatic replies.
		HandleOPTIONS bool

		// timeout config
		ReadTimeout       time.Duration
		WriteTimeout      time.Duration
		ReadHeaderTimeout time.Duration
		ShutdownDuration  time.Duration

		// notFoundHandler handle 404
		notFoundHandler HandlerFunc
		// panicHandler handle internal panic
		panicHandler PanicHandler
		// methodNotAllowed handle method not allowed
		methodNotAllowed HandlerFunc
		// StatsHandler handle every single request
		StatsHandler StatsHandler
	}
)

// New will create a Server instance and return a pointer which point to it
func New() *Server {
	s := &Server{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
	}

	s.Router = s.Group("")

	return s
}

// Lookup allows the manual lookup of a method + path combo.
// This is e.g. useful to build a framework around this router.
// If the path was found, it returns the handle function and the path parameter
// values. Otherwise the third return value indicates whether a redirection to
// the same path with an extra / without the trailing slash should be performed.
func (s *Server) Lookup(method, path string) (Handlers, Params, bool) {
	for i := range s.trees {
		if s.trees[i].method == method {
			return s.trees[i].node.getValue(path)
		}
	}
	return nil, nil, false
}

func (s *Server) allowed(path, reqMethod string) (allow string) {
	if path == "*" { // server-wide

		for i := range s.trees {
			if s.trees[i].method == "OPTIONS" {
				continue
			}

			// add request method to list of allowed methods
			if len(allow) == 0 {
				allow = s.trees[i].method
			} else {
				allow += ", " + s.trees[i].method
			}
		}
	} else { // specific path
		for i := range s.trees {
			// Skip the requested method - we already tried this one
			if s.trees[i].method == reqMethod || s.trees[i].method == "OPTIONS" {
				continue
			}

			handles, _, _ := s.trees[i].node.getValue(path)
			if handles != nil {
				// add request method to list of allowed methods
				if len(allow) == 0 {
					allow = s.trees[i].method
				} else {
					allow += ", " + s.trees[i].method
				}
			}
		}
	}
	if len(allow) > 0 {
		allow += ", OPTIONS"
	}
	return
}

// Required by http.Handler interface. This method is invoked by the
// http server and will handle all page routing
func (s *Server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	begin := time.Now()
	defer func() {
		end := time.Now()
		if s.StatsHandler != nil {
			s.StatsHandler(r.RequestURI, begin, end)
		}
	}()
	// get context instance from pool
	c := s.getContext(rw, r)
	s.handleHTTPRequest(c)

	// put context back into pool
	s.putBackContext(c)
}

func (s *Server) handleHTTPRequest(ctx *Context) {
	// handle panic
	defer s.handlePanic(ctx)

	httpMethod := ctx.Req.Method
	path := ctx.Req.URL.Path

	for i := 0; i < len(s.trees); i++ {
		t := s.trees[i]
		if t.method == httpMethod {
			if handlers, params, tsr := t.node.getValue(path); handlers != nil {
				ctx.params = params
				for _, h := range handlers {
					h(ctx)
					if ctx.rw.written {
						return
					}
				}
				return
			} else if ctx.Req.Method != "CONNECT" && path != "/" {
				code := 301 // Permanent redirect, request with GET method
				if ctx.Req.Method != "GET" {
					// Temporary redirect, request with same method
					// As of Go 1.3, Go does not support status code 308.
					code = 307
				}

				if tsr && s.RedirectTrailingSlash {
					if len(path) > 1 && path[len(path)-1] == '/' {
						ctx.Req.URL.Path = path[:len(path)-1]
					} else {
						ctx.Req.URL.Path = path + "/"
					}
					http.Redirect(ctx.rw, ctx.Req, ctx.Req.URL.String(), code)
					return
				}

				// Try to fix the request path
				if s.RedirectFixedPath {
					fixedPath, found := t.node.findCaseInsensitivePath(
						CleanPath(path),
						s.RedirectTrailingSlash,
					)
					if found {
						ctx.Req.URL.Path = string(fixedPath)
						http.Redirect(ctx.rw, ctx.Req, ctx.Req.URL.String(), code)
						return
					}
				}
			}

		}
	}

	if ctx.Req.Method == "OPTIONS" {
		// Handle OPTIONS requests
		if s.HandleOPTIONS {
			if allow := s.allowed(path, ctx.Req.Method); len(allow) > 0 {
				ctx.WriteHeader("Allow", allow)
				return
			}
		}
	} else {
		// Handle 405
		if s.HandleMethodNotAllowed {
			if allow := s.allowed(path, ctx.Req.Method); len(allow) > 0 {
				ctx.WriteHeader("Allow", allow)
				if s.methodNotAllowed != nil {
					s.methodNotAllowed(ctx)
				} else {
					http.Error(ctx.rw,
						StatusText(StatusMethodNotAllowed),
						StatusMethodNotAllowed,
					)
				}
				return
			}
		}
	}

	s.handleNotFound(ctx)
}

// Run server on addr
func (s *Server) Run(addr string) error {
	var err error
	once(&s.once, func() {
		s.server = http.Server{Handler: s, Addr: addr, ReadTimeout: s.ReadTimeout, ReadHeaderTimeout: s.ReadHeaderTimeout, WriteTimeout: s.WriteTimeout}
		err = s.server.ListenAndServe()
	})
	return err
}

// RunTLS Run server on addr with tls
func (s *Server) RunTLS(addr string, certFile string, keyFile string) error {
	var err error
	once(&s.once, func() {
		s.server = http.Server{Handler: s, Addr: addr, ReadTimeout: s.ReadTimeout, ReadHeaderTimeout: s.ReadHeaderTimeout, WriteTimeout: s.WriteTimeout}
		err = s.server.ListenAndServeTLS(certFile, keyFile)
	})
	return err
}

// Shutdown gracefully with deadline
func (s *Server) Shutdown() error {
	ctx := s.getContext(nil, nil)
	if s.ShutdownDuration > 0 {
		ctx, _ = ctx.WithDeadline(time.Now().Add(s.ShutdownDuration))
	}
	return s.server.Shutdown(ctx)
}

// Close shutdown server forcibly
func (s *Server) Close() error {
	return s.server.Close()
}

func once(once *sync.Once, closure func()) {
	once.Do(closure)
}
