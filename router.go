package zen

// Router ...
type Router interface {
	// Route set handler for given pattern and method
	Route(method string, path string, handler HandlerFunc)

	// Get adds a new Route for GET requests.
	Get(path string, handler HandlerFunc)

	// Post adds a new Route for POST requests.
	Post(path string, handler HandlerFunc)

	// Put adds a new Route for PUT requests.
	Put(path string, handler HandlerFunc)

	// Delete adds a new Route for DELETE requests.
	Delete(path string, handler HandlerFunc)

	// Patch adds a new Route for PATCH requests.
	Patch(path string, handler HandlerFunc)

	// Head adds a new Route for HEAD requests.
	Head(path string, handler HandlerFunc)

	// Options adds a new Route for OPTIONS requests.
	Options(path string, handler HandlerFunc)

	// Connect adds a new Route for CONNECT requests.
	Connect(path string, handler HandlerFunc)

	// Trace adds a new Route for TRACE requests.
	Trace(path string, handler HandlerFunc)

	// Any adds new Route for ALL method requests.
	Any(path string, handler HandlerFunc)

	// Filter add a filter for given path
	Filter(handler HandlerFunc)
}
