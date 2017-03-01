package zen

import (
	"net/http"
	"sync"
	"testing"
)

func TestServer_getContext(t *testing.T) {
	type fields struct {
		notFoundHandler HandlerFunc
		panicHandler    PanicHandler
		filters         []HandlerFunc
		contextPool     *sync.Pool
	}
	type args struct {
		rw  http.ResponseWriter
		req *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantNil bool
	}{
		{"case1",
			fields{
				contextPool: &sync.Pool{
					New: func() interface{} {
						c := Context{
							params: Params{},
							rw:     &responseWriter{},
						}
						return &c
					},
				},
			},
			args{
				nil, nil,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				notFoundHandler: tt.fields.notFoundHandler,
				panicHandler:    tt.fields.panicHandler,
				contextPool:     *tt.fields.contextPool,
			}
			if got := s.getContext(tt.args.rw, tt.args.req); (got == nil) != tt.wantNil {
				t.Errorf("Server.getContext() = %v, want nil? %v", got, tt.wantNil)
			} else {
				s.putBackContext(got)
			}
		})
	}
}

func BenchmarkGetContext(b *testing.B) {
	s := &Server{

		contextPool: sync.Pool{
			New: func() interface{} {
				c := Context{
					params: Params{},
					rw:     &responseWriter{},
				}
				return &c
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := s.getContext(nil, nil)
		s.putBackContext(c)
	}
}
