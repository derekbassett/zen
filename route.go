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
func (s *Server) route(method string, path string, handler HandlerFunc) {
	assert(path[0] == '/', "path must begin with '/'")
	assert(len(method) > 0, "HTTP method can not be empty")
	assert(handler != nil, "handler cannot be nil")

	handler = s.interceptors.Wrap(handler)
	root := s.methodRouteTree(method)
	root.addRoute(path, handler)
}

// AddInterceptor add a global interceptor
func (s *Server) AddInterceptor(handler Middleware) {
	s.Router.AddInterceptor(handler)
}

// Static :Adds a new Route for Static http requests. Serves
// static files from the specified directory
func (s *Server) Static(staticpath string, dir string) {
	s.Route(GET, path.Join(staticpath, "/*filepath"), func(ctx *Context) {
		http.StripPrefix(staticpath, http.FileServer(http.Dir(dir))).ServeHTTP(ctx.rw, ctx.Req)
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

// handleNotFound call server's not found handler
func (s *Server) handleNotFound(ctx *Context) {
	if s.notFoundHandler != nil {
		s.notFoundHandler(ctx)
		return
	}

	http.NotFound(ctx.rw, ctx.Req)
}
