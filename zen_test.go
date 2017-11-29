package zen

import (
	"testing"
	"time"
)

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
