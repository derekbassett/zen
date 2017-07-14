package zen

import (
	"testing"
	"time"
)

func TestServer_Run(t *testing.T) {
	server := New()

	go func() {
		server.Get("/example", func(c *Context) { c.RawStr("ok") })
		if err := server.Run(":8080"); err != nil {
			t.Error("Run server got error", err)
		}
	}()

	time.Sleep(5 * time.Millisecond)
	if err := server.Run(":8080"); err == nil {
		t.Error("Run server got nil, want port already used")
	}
}

func TestServer_RunTLS(t *testing.T) {
	server := New()

	server.Get("/example", func(c *Context) { c.RawStr("ok") })

	if err := server.RunTLS(":8080", "", ""); err == nil {
		t.Error("Run server got nil, want port already used")
	}
}

func TestServerShutdown(t *testing.T) {
	server := New()
	server.ShutdownDuration = time.Second
	if err := server.Shutdown(); err != nil {
		t.Log(err.Error())
	}
}
