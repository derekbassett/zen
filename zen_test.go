package zen

import (
	"net/http"
	"testing"
	"time"
)

func TestServer_Run(t *testing.T) {
	server := New()
	go func() {
		time.Sleep(time.Millisecond)
		if err := server.Close(); err != nil {
			t.Error(err)
		}
	}()

	if err := server.Run(":8080"); err != nil {
		if err != http.ErrServerClosed {
			t.Error(err)
		}
	}
}

func TestServer_RunTLS(t *testing.T) {

	server := New()
	go func() {
		time.Sleep(time.Millisecond)
		if err := server.Close(); err != nil {
			t.Error(err)
		}
	}()

	if err := server.RunTLS(":8081", "./cert/server.crt", "./cert/server.key"); err != nil {
		if err != http.ErrServerClosed {
			t.Error(err)
		}
	}
}

func TestServerShutdown(t *testing.T) {
	server := New()
	server.ShutdownDuration = time.Second
	if err := server.Shutdown(); err != nil {
		t.Log(err.Error())
	}
}

func TestServerClose(t *testing.T) {
	server := New()
	if err := server.Close(); err != nil {
		t.Log(err.Error())
	}
}
