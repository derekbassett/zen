package zen

import (
	"net/http"
	"net/http/pprof"
	"path/filepath"
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
func (s *Server) Route(method string, path string, handler HandlerFunc) {
	assert(path[0] == '/', "path must begin with '/'")
	assert(len(method) > 0, "HTTP method can not be empty")
	assert(handler != nil, "handler cannot be nil")

	root := s.methodRouteTree(method)
	root.addRoute(path, Handlers{handler})
}

// Get adds a new Route for GET requests.
func (s *Server) Get(path string, handler HandlerFunc) {
	s.Route(GET, path, handler)
}

// Post adds a new Route for POST requests.
func (s *Server) Post(path string, handler HandlerFunc) {
	s.Route(POST, path, handler)
}

// Put adds a new Route for PUT requests.
func (s *Server) Put(path string, handler HandlerFunc) {
	s.Route(PUT, path, handler)
}

// Del adds a new Route for DELETE requests.
func (s *Server) Del(path string, handler HandlerFunc) {
	s.Route(DELETE, path, handler)
}

// Patch adds a new Route for PATCH requests.
func (s *Server) Patch(path string, handler HandlerFunc) {
	s.Route(PATCH, path, handler)
}

// Head adds a new Route for HEAD requests.
func (s *Server) Head(path string, handler HandlerFunc) {
	s.Route(HEAD, path, handler)
}

// Options adds a new Route for OPTIONS requests.
func (s *Server) Options(path string, handler HandlerFunc) {
	s.Route(OPTIONS, path, handler)
}

// Connect adds a new Route for CONNECT requests.
func (s *Server) Connect(path string, handler HandlerFunc) {
	s.Route(CONNECT, path, handler)
}

// Trace adds a new Route for TRACE requests.
func (s *Server) Trace(path string, handler HandlerFunc) {
	s.Route(TRACE, path, handler)
}

// Any adds new Route for ALL method requests.
func (s *Server) Any(relativePath string, handler HandlerFunc) {
	s.Route("GET", relativePath, handler)
	s.Route("POST", relativePath, handler)
	s.Route("PUT", relativePath, handler)
	s.Route("PATCH", relativePath, handler)
	s.Route("HEAD", relativePath, handler)
	s.Route("OPTIONS", relativePath, handler)
	s.Route("DELETE", relativePath, handler)
	s.Route("CONNECT", relativePath, handler)
	s.Route("TRACE", relativePath, handler)
}

// Static :Adds a new Route for Static http requests. Serves
// static files from the specified directory
func (s *Server) Static(pattern string, dir string) {
	//append a regex to the param to match everything
	// that comes after the prefix
	pattern = pattern + "(.+)"
	s.Route(GET, pattern, func(c *Context) {
		path := filepath.Clean(c.Req.URL.Path)
		path = filepath.Join(dir, path)
		http.ServeFile(c.rw, c.Req, path)
	})
}

// PProf serve golang's pprof tool
func (s *Server) PProf(pattern string) {
	s.Get(pattern, wrapF(pprof.Index))
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
			http.Error(c.rw, "internal server error", http.StatusInternalServerError)
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
