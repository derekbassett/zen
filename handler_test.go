package zen

import (
	"net/http"
	"testing"
)

func Test_wrapF(t *testing.T) {
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
			if got := wrapF(tt.args.h); (got == nil) != tt.wantNil {
				t.Errorf("wrapF() = %v, want nil ? %v", got, tt.wantNil)
			}
		})
	}
}
