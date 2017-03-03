package zen

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type mockResponseWriter struct {
	code int
	body bytes.Buffer
}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return m.body.Write(p)
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return m.body.WriteString(s)
}

func (m *mockResponseWriter) WriteHeader(code int) {
	m.code = code
}

func TestParams(t *testing.T) {
	ps := Params{
		Param{"param1", "value1"},
		Param{"param2", "value2"},
		Param{"param3", "value3"},
	}
	for i := range ps {
		if val := ps.ByName(ps[i].Key); val != ps[i].Value {
			t.Errorf("Wrong value for %s: Got %s; Want %s", ps[i].Key, val, ps[i].Value)
		}
	}
	if val := ps.ByName("noKey"); val != "" {
		t.Errorf("Expected empty string for not found key; got: %s", val)
	}
}

func TestRouter(t *testing.T) {
	router := New()

	routed := false
	router.Route("GET", "/user/:name", func(c *Context) {
		routed = true
		want := Params{Param{"name", "gopher"}}
		if !reflect.DeepEqual(c.params, want) {
			t.Fatalf("wrong wildcard values: want %v, got %v", want, c.params)
		}
	})

	w := new(mockResponseWriter)

	req, _ := http.NewRequest("GET", "/user/gopher", nil)
	router.ServeHTTP(w, req)

	if !routed {
		t.Fatal("routing failed")
	}
}

func TestRouterAPI(t *testing.T) {
	var get, head, options, post, put, patch, delete, connect, trace, handlerFunc bool

	router := New()
	router.Get("/GET", func(c *Context) {
		get = true
	})
	router.Head("/GET", func(c *Context) {
		head = true
	})
	router.Options("/GET", func(c *Context) {
		options = true
	})
	router.Post("/POST", func(c *Context) {
		post = true
	})
	router.Put("/PUT", func(c *Context) {
		put = true
	})
	router.Patch("/PATCH", func(c *Context) {
		patch = true
	})
	router.Delete("/DELETE", func(c *Context) {
		delete = true
	})
	router.Connect("/CONNECT", func(c *Context) {
		connect = true
	})
	router.Trace("/TRACE", func(c *Context) {
		trace = true
	})
	router.Route("GET", "/HandlerFunc", func(c *Context) {
		handlerFunc = true
	})

	w := new(mockResponseWriter)

	r, _ := http.NewRequest("GET", "/GET", nil)
	router.ServeHTTP(w, r)
	if !get {
		t.Error("routing GET failed")
	}

	r, _ = http.NewRequest("HEAD", "/GET", nil)
	router.ServeHTTP(w, r)
	if !head {
		t.Error("routing HEAD failed")
	}

	r, _ = http.NewRequest("OPTIONS", "/GET", nil)
	router.ServeHTTP(w, r)
	if !options {
		t.Error("routing OPTIONS failed")
	}

	r, _ = http.NewRequest("POST", "/POST", nil)
	router.ServeHTTP(w, r)
	if !post {
		t.Error("routing POST failed")
	}

	r, _ = http.NewRequest("PUT", "/PUT", nil)
	router.ServeHTTP(w, r)
	if !put {
		t.Error("routing PUT failed")
	}

	r, _ = http.NewRequest("PATCH", "/PATCH", nil)
	router.ServeHTTP(w, r)
	if !patch {
		t.Error("routing PATCH failed")
	}

	r, _ = http.NewRequest("DELETE", "/DELETE", nil)
	router.ServeHTTP(w, r)
	if !delete {
		t.Error("routing DELETE failed")
	}

	r, _ = http.NewRequest("CONNECT", "/CONNECT", nil)
	router.ServeHTTP(w, r)
	if !connect {
		t.Error("routing CONNECT failed")
	}

	r, _ = http.NewRequest("TRACE", "/TRACE", nil)
	router.ServeHTTP(w, r)
	if !trace {
		t.Error("routing TRACE failed")
	}

	r, _ = http.NewRequest("GET", "/HandlerFunc", nil)
	router.ServeHTTP(w, r)
	if !handlerFunc {
		t.Error("routing HandlerFunc failed")
	}
}

func TestRouterAPIAny(t *testing.T) {
	var get, head, options, post, put, patch, delete, connect, trace, filter, filterGroup, groupHandler bool

	router := New()
	router.Filter(func(c *Context) {
		filter = true
	})

	group := router.Group("/GROUP")

	group.Filter(func(c *Context) {
		filterGroup = true
	})

	group.Get("/ANY", func(c *Context) {
		groupHandler = true
	})

	router.Any("/ANY", func(c *Context) {
		switch c.Req.Method {
		case GET:
			get = true
		case HEAD:
			head = true
		case OPTIONS:
			options = true
		case POST:
			post = true
		case PUT:
			put = true
		case PATCH:
			patch = true
		case DELETE:
			delete = true
		case CONNECT:
			connect = true
		case TRACE:
			trace = true
		}
	})

	w := new(mockResponseWriter)

	r, _ := http.NewRequest("GET", "/ANY", nil)
	router.ServeHTTP(w, r)
	if !get {
		t.Error("routing GET failed")
	}

	r, _ = http.NewRequest("HEAD", "/ANY", nil)
	router.ServeHTTP(w, r)
	if !head {
		t.Error("routing HEAD failed")
	}

	r, _ = http.NewRequest("OPTIONS", "/ANY", nil)
	router.ServeHTTP(w, r)
	if !options {
		t.Error("routing OPTIONS failed")
	}

	r, _ = http.NewRequest("POST", "/ANY", nil)
	router.ServeHTTP(w, r)
	if !post {
		t.Error("routing POST failed")
	}

	r, _ = http.NewRequest("PUT", "/ANY", nil)
	router.ServeHTTP(w, r)
	if !put {
		t.Error("routing PUT failed")
	}

	r, _ = http.NewRequest("PATCH", "/ANY", nil)
	router.ServeHTTP(w, r)
	if !patch {
		t.Error("routing PATCH failed")
	}

	r, _ = http.NewRequest("DELETE", "/ANY", nil)
	router.ServeHTTP(w, r)
	if !delete {
		t.Error("routing DELETE failed")
	}

	r, _ = http.NewRequest("CONNECT", "/ANY", nil)
	router.ServeHTTP(w, r)
	if !connect {
		t.Error("routing CONNECT failed")
	}

	r, _ = http.NewRequest("TRACE", "/ANY", nil)
	router.ServeHTTP(w, r)
	if !trace {
		t.Error("routing TRACE failed")
	}

	if !filter {
		t.Error("routing filter failed")
	}

	r, _ = http.NewRequest("GET", "/GROUP/ANY", nil)
	router.ServeHTTP(w, r)
	if !filterGroup {
		t.Error("routing group filter failed")
	}
	if !groupHandler {
		t.Error("routing group handler failed")
	}
}

func TestRouterRoot(t *testing.T) {
	router := New()
	recv := catchPanic(func() {
		router.Get("noSlashRoot", nil)
	})
	if recv == nil {
		t.Fatal("registering path not beginning with '/' did not panic")
	}
}

func TestRouterOPTIONS(t *testing.T) {
	handlerFunc := func(c *Context) {}

	router := New()
	router.Post("/path", handlerFunc)

	// test not allowed
	// * (server)
	r, _ := http.NewRequest("OPTIONS", "*", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusOK) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// path
	r, _ = http.NewRequest("OPTIONS", "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusOK) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	r, _ = http.NewRequest("OPTIONS", "/doesnotexist", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusNotFound) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	}

	// add another method
	router.Get("/path", handlerFunc)

	// test again
	// * (server)
	r, _ = http.NewRequest("OPTIONS", "*", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusOK) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "POST, GET, OPTIONS" && allow != "GET, POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// path
	r, _ = http.NewRequest("OPTIONS", "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusOK) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "POST, GET, OPTIONS" && allow != "GET, POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// custom handler
	var custom bool
	router.Options("/path", func(c *Context) {
		custom = true
	})

	// test again
	// * (server)
	r, _ = http.NewRequest("OPTIONS", "*", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusOK) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "POST, GET, OPTIONS" && allow != "GET, POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}
	if custom {
		t.Error("custom handler called on *")
	}

	// path
	r, _ = http.NewRequest("OPTIONS", "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusOK) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	}
	if !custom {
		t.Error("custom handler not called")
	}
}

func TestRouterNotAllowed(t *testing.T) {
	handlerFunc := func(c *Context) {}

	router := New()
	router.Post("/path", handlerFunc)

	// test not allowed
	r, _ := http.NewRequest("GET", "/path", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusMethodNotAllowed) {
		t.Errorf("NotAllowed handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// add another method
	router.Delete("/path", handlerFunc)
	router.Options("/path", handlerFunc) // must be ignored

	// test again
	r, _ = http.NewRequest("GET", "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusMethodNotAllowed) {
		t.Errorf("NotAllowed handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "POST, DELETE, OPTIONS" && allow != "DELETE, POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// test custom handler
	w = httptest.NewRecorder()
	responseText := "custom method"
	router.HandleNotAllowed(wrapF(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(responseText))
	}))
	router.ServeHTTP(w, r)
	if got := w.Body.String(); !(got == responseText) {
		t.Errorf("unexpected response got %q want %q", got, responseText)
	}
	if w.Code != http.StatusTeapot {
		t.Errorf("unexpected response code %d want %d", w.Code, http.StatusTeapot)
	}
	if allow := w.Header().Get("Allow"); allow != "POST, DELETE, OPTIONS" && allow != "DELETE, POST, OPTIONS" {
		t.Error("unexpected Allow header value: " + allow)
	}
}

func TestRouterNotFound(t *testing.T) {
	handlerFunc := func(c *Context) {}

	router := New()
	router.Get("/path", handlerFunc)
	router.Get("/dir/", handlerFunc)
	router.Get("/", handlerFunc)

	testRoutes := []struct {
		route  string
		code   int
		header string
	}{
		{"/path/", 301, "map[Location:[/path]]"},   // TSR -/
		{"/dir", 301, "map[Location:[/dir/]]"},     // TSR +/
		{"", 301, "map[Location:[/]]"},             // TSR +/
		{"/PATH", 301, "map[Location:[/path]]"},    // Fixed Case
		{"/DIR/", 301, "map[Location:[/dir/]]"},    // Fixed Case
		{"/PATH/", 301, "map[Location:[/path]]"},   // Fixed Case -/
		{"/DIR", 301, "map[Location:[/dir/]]"},     // Fixed Case +/
		{"/../path", 301, "map[Location:[/path]]"}, // CleanPath
		{"/nope", 404, ""},                         // NotFound
	}
	for _, tr := range testRoutes {
		r, _ := http.NewRequest("GET", tr.route, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if !(w.Code == tr.code && (w.Code == 404 || fmt.Sprint(w.Header()) == tr.header)) {
			t.Errorf("NotFound handling route %s failed: Code=%d, Header=%v", tr.route, w.Code, w.Header())
		}
	}

	// Test custom not found handler
	var notFound bool
	router.HandleNotFound(wrapF(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(404)
		notFound = true
	}))
	r, _ := http.NewRequest("GET", "/nope", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == 404 && notFound == true) {
		t.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", w.Code, w.Header())
	}

	// Test other method than GET (want 307 instead of 301)
	router.Patch("/path", handlerFunc)
	r, _ = http.NewRequest("PATCH", "/path/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == 307 && fmt.Sprint(w.Header()) == "map[Location:[/path]]") {
		t.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", w.Code, w.Header())
	}

	// Test special case where no node for the prefix "/" exists
	router = New()
	router.Get("/a", handlerFunc)
	r, _ = http.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == 404) {
		t.Errorf("NotFound handling route / failed: Code=%d", w.Code)
	}
}

func TestRouterPanicHandler(t *testing.T) {
	router := New()
	panicHandled := false

	router.HandlePanic(func(c *Context, _ interface{}) {
		panicHandled = true
	})

	router.Route("PUT", "/user/:name", func(c *Context) {
		panic("oops!")
	})

	w := new(mockResponseWriter)
	req, _ := http.NewRequest("PUT", "/user/gopher", nil)

	router.ServeHTTP(w, req)

	if !panicHandled {
		t.Fatal("simulating failed")
	}
}

func TestRouterDefaultPanicHandler(t *testing.T) {
	router := New()

	router.Route("PUT", "/user/:name", func(c *Context) {
		panic("oops!")
	})

	w := new(mockResponseWriter)
	req, _ := http.NewRequest("PUT", "/user/gopher", nil)

	router.ServeHTTP(w, req)

	if w.code != http.StatusInternalServerError {
		t.Fatal("simulating failed")
	}
}

func TestRouterLookup(t *testing.T) {
	routed := false
	wantHandle := func(c *Context) {
		routed = true
	}
	wantParams := Params{Param{"name", "gopher"}}

	router := New()

	// try empty router first
	handle, _, tsr := router.Lookup("GET", "/nope")
	if handle != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", handle)
	}
	if tsr {
		t.Error("Got wrong TSR recommendation!")
	}

	// insert route and try again
	router.Get("/user/:name", wantHandle)

	handles, params, tsr := router.Lookup("GET", "/user/gopher")
	if handles == nil {
		t.Fatal("Got no handle!")
	} else {
		handles[0](nil)
		if !routed {
			t.Fatal("Routing failed!")
		}
	}

	if !reflect.DeepEqual(params, wantParams) {
		t.Fatalf("Wrong parameter values: want %v, got %v", wantParams, params)
	}

	handle, _, tsr = router.Lookup("GET", "/user/gopher/")
	if handle != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", handle)
	}
	if !tsr {
		t.Error("Got no TSR recommendation!")
	}

	handle, _, tsr = router.Lookup("GET", "/nope")
	if handle != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", handle)
	}
	if tsr {
		t.Error("Got wrong TSR recommendation!")
	}
}

func TestServer_Static(t *testing.T) {
	testRoot, _ := os.Getwd()
	f, err := ioutil.TempFile(testRoot, "")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("zen")
	f.Close()

	dir, filename := filepath.Split(f.Name())
	fmt.Println(dir, filename)
	server := New()
	server.Static("/static", dir)

	req := mustNewRequest("GET", "/static/"+filename, nil)
	rw := new(mockResponseWriter)
	server.ServeHTTP(rw, req)
	if rw.code != 200 {
		t.Error("Get static file code != 200")
	}
	if rw.body.String() != "zen" {
		t.Errorf("Get static file body want %s got %s", "zen", rw.body.String())
	}
}
