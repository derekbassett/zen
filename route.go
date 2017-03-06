package zen

import (
	"net/http"
	"path"
)

const (
	// CONNECT : http method CONNECT
	CONNECT = "CONNECT"
	// DELETE : http method DELETE
	DELETE = "DELETE"
	// GET : http method GET
	GET = "GET"
	// HEAD : http method HEAD
	HEAD = "HEAD"
	// OPTIONS : http method OPTIONS
	OPTIONS = "OPTIONS"
	// PATCH : http method PATCH
	PATCH = "PATCH"
	// POST : http method POST
	POST = "POST"
	// PUT : http method PUT
	PUT = "PUT"
	// TRACE : http method TRACE
	TRACE = "TRACE"
)

func (s *Server) methodRouteTree(method string) *node {
	for _, t := range s.trees {
		if t.method == method {
			return t.node
		}
	}
	methodRoot := &methodTree{
		method: method,
		node:   new(node),
	}
	s.trees = append(s.trees, methodRoot)

	return methodRoot.node
}

// Route set handler for given pattern and method
func (s *Server) route(method string, path string, handlers Handlers) {
	assert(path[0] == '/', "path must begin with '/'")
	assert(len(method) > 0, "HTTP method can not be empty")
	assert(handlers != nil && len(handlers) > 0, "handler cannot be nil")
	h := make(Handlers, len(s.filters)+len(handlers))
	c := 0
	c += copy(h[c:], s.filters)
	c += copy(h[c:], handlers)
	root := s.methodRouteTree(method)
	root.addRoute(path, h)
}

// Filter add a global filter
func (s *Server) Filter(handler HandlerFunc) {
	s.filters = append(s.filters, handler)
}

// Static :Adds a new Route for Static http requests. Serves
// static files from the specified directory
func (s *Server) Static(staticpath string, dir string) {
	s.Route(GET, path.Join(staticpath, "/*filepath"), func(c *Context) {
		http.StripPrefix(staticpath, http.FileServer(http.Dir(dir))).ServeHTTP(c.rw, c.Req)
	})
}

// HandleNotFound set server's notFoundHandler
func (s *Server) HandleNotFound(handler HandlerFunc) {
	s.notFoundHandler = handler
}

// HandleNotAllowed set server's methodNotAllowed
func (s *Server) HandleNotAllowed(handler HandlerFunc) {
	s.methodNotAllowed = handler
}

// HandlePanic set server's panicHandler
func (s *Server) HandlePanic(handler PanicHandler) {
	s.panicHandler = handler
}

// handlePanic call server's panic handler
func (s *Server) handlePanic(c *Context) {

	if err := recover(); err != nil {
		if s.panicHandler != nil {
			s.panicHandler(c, err)
		} else {
			http.Error(c.rw, StatusText(StatusInternalServerError), StatusInternalServerError)
		}
	}
}

// handleNotFound call server's not found handler
func (s *Server) handleNotFound(c *Context) {

	if s.notFoundHandler != nil {
		s.notFoundHandler(c)
		return
	}

	http.NotFound(c.rw, c.Req)
}
