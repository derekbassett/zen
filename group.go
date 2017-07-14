package zen

import "path"

type group struct {
	base    string
	filters Handlers
	server  *Server
}

// Group create a group router with base url and shared filter
func (s *Server) Group(base string, filters ...HandlerFunc) Router {
	return &group{
		base:    base,
		filters: filters,
		server:  s,
	}
}

// route set handler for given pattern and method
func (g *group) Route(method string, path string, handler HandlerFunc) {
	path = joinPath(g.base, path)
	handlers := g.combineHandlers(handler)

	g.server.route(method, path, handlers)
}

// Get adds a new route for GET requests.
func (g *group) Get(path string, handler HandlerFunc) {
	g.Route(GET, path, handler)
}

// Post adds a new route for POST requests.
func (g *group) Post(path string, handler HandlerFunc) {
	g.Route(POST, path, handler)
}

// Put adds a new route for PUT requests.
func (g *group) Put(path string, handler HandlerFunc) {
	g.Route(PUT, path, handler)
}

// Delete adds a new route for DELETE requests.
func (g *group) Delete(path string, handler HandlerFunc) {
	g.Route(DELETE, path, handler)
}

// Patch adds a new route for PATCH requests.
func (g *group) Patch(path string, handler HandlerFunc) {
	g.Route(PATCH, path, handler)
}

// Head adds a new route for HEAD requests.
func (g *group) Head(path string, handler HandlerFunc) {
	g.Route(HEAD, path, handler)
}

// Options adds a new route for OPTIONS requests.
func (g *group) Options(path string, handler HandlerFunc) {
	g.Route(OPTIONS, path, handler)
}

// Connect adds a new route for CONNECT requests.
func (g *group) Connect(path string, handler HandlerFunc) {
	g.Route(CONNECT, path, handler)
}

// Trace adds a new route for TRACE requests.
func (g *group) Trace(path string, handler HandlerFunc) {
	g.Route(TRACE, path, handler)
}

// Any adds new route for ALL method requests.
func (g *group) Any(path string, handler HandlerFunc) {
	g.Route(GET, path, handler)
	g.Route(POST, path, handler)
	g.Route(PUT, path, handler)
	g.Route(PATCH, path, handler)
	g.Route(HEAD, path, handler)
	g.Route(OPTIONS, path, handler)
	g.Route(DELETE, path, handler)
	g.Route(CONNECT, path, handler)
	g.Route(TRACE, path, handler)
}

// Filter add a filter in group
func (g *group) Filter(handler HandlerFunc) {
	g.filters = append(g.filters, handler)
}

func joinPath(base, sub string) string {
	ret := path.Join(base, sub)
	if sub[len(sub)-1] == '/' {
		ret = ret + "/"
	}
	return ret
}

func (g *group) combineHandlers(handler HandlerFunc) Handlers {
	return append(g.filters, handler)
}
