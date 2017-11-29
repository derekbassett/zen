package zen

import (
	"net/http"
	"testing"
)

func Test_WrapF(t *testing.T) {
	type args struct {
		h http.HandlerFunc
	}
	tests := []struct {
		name    string
		args    args
		wantNil bool
	}{
		{
			"case1",
			args{
				http.NotFound,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WrapF(tt.args.h); (got == nil) != tt.wantNil {
				t.Errorf("WrapF() = %v, want nil ? %v", got, tt.wantNil)
			}
		})
	}
}

func TestUnWrapF(t *testing.T) {
	type args struct {
		h HandlerFunc
	}
	tests := []struct {
		name    string
		args    args
		wantNil bool
	}{
		{
			"case1",
			args{
				func(*Context) {},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UnWrapF(tt.args.h); (got == nil) != tt.wantNil {
				t.Errorf("UnWrapF() = %v, wantNil %v", got, tt.wantNil)
			}
		})
	}
}
