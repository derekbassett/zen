package zen

import (
	"fmt"
	"testing"
)

func Test_assert(t *testing.T) {
	type args struct {
		c   bool
		msg string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"true",
			args{
				true,
				"nil",
			},
		},
		{
			"false",
			args{
				false,
				"panic",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					if fmt.Sprint(err) != tt.args.msg {
						t.Errorf("assert want panic %s, got %s", tt.args.msg, fmt.Sprint(err))
					}
				}
			}()
			assert(tt.args.c, tt.args.msg)
		})
	}
}
