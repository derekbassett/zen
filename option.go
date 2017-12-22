package zen

import (
	"time"
)

// Option use to customize Server
type Option func(*Server)

// SetTimeout return Option for set timeout
func SetTimeout(t time.Duration) Option {
	return func(s *Server) {
		s.ReadTimeout = t
		s.WriteTimeout = t
		s.ReadHeaderTimeout = t
	}
}

// SetReadTimeout return Option for set read timeout
func SetReadTimeout(t time.Duration) Option {
	return func(s *Server) {
		s.ReadTimeout = t
	}
}

// SetWriteTimeout return Option for set write timeout
func SetWriteTimeout(t time.Duration) Option {
	return func(s *Server) {
		s.WriteTimeout = t
	}
}

// SetReadHeaderTimeout return Option for set read header timeout
func SetReadHeaderTimeout(t time.Duration) Option {
	return func(s *Server) {
		s.ReadHeaderTimeout = t
	}
}

// SetRedirectTrailingSlash return Option for set RedirectTrailingSlash
func SetRedirectTrailingSlash(b bool) Option {
	return func(s *Server) {
		s.RedirectTrailingSlash = b
	}
}

// SetRedirectFixedPath return Option for set RedirectFixedPath
func SetRedirectFixedPath(b bool) Option {
	return func(s *Server) {
		s.RedirectFixedPath = b
	}
}

// SetHandleMethodNotAllowed return Option for set HandleMethodNotAllowed
func SetHandleMethodNotAllowed(b bool) Option {
	return func(s *Server) {
		s.HandleMethodNotAllowed = b
	}
}

// SetHandleOPTIONS return Option for set HandleOPTIONS
func SetHandleOPTIONS(b bool) Option {
	return func(s *Server) {
		s.HandleOPTIONS = b
	}
}
