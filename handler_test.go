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
