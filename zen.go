package zen

import (
	"net/http"
	"sync"
)

const (
	// Version is current version num
	Version = "v1.0.0Beta"
)

type (
	methodTree struct {
		method string
		node   *node
	}

	// Server struct
	Server struct {
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
		HandleOPTIONS   bool
		notFoundHandler HandlerFunc
		panicHandler    PanicHandler
		contextPool     sync.Pool
	}
)

// New will create a Server instance and return a pointer which point to it
func New() *Server {
	s := &Server{
		contextPool:            sync.Pool{},
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
	}

	s.contextPool.New = func() interface{} {
		c := Context{rw: &responseWriter{}}
		return &c
	}

	return s
}

// Required by http.Handler interface. This method is invoked by the
// http server and will handle all page routing
func (s *Server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// get context instance from pool
	c := s.getContext(rw, r)

	s.handleHTTPRequest(c)

	// put context back into pool
	s.putBackContext(c)
}

func (s *Server) handleHTTPRequest(c *Context) {
	// handle panic
	defer s.handlePanic(c)

	httpMethod := c.Req.Method
	path := c.Req.URL.Path

	for i := 0; i < len(s.trees); i++ {
		t := s.trees[i]
		if t.method == httpMethod {
			handlers, params, _ := t.node.getValue(path)
			c.params = params

			for _, h := range handlers {
				h(c)
				if c.rw.written {
					return
				}
			}

		}
	}

	s.handleNotFound(c)
}

// Run server on addr
func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s)
}

// RunTLS Run server on addr with tls
func (s *Server) RunTLS(addr string, certFile string, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, s)
}
