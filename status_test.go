package zen

import "testing"

func TestStatusText(t *testing.T) {
	type args struct {
		code int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"200",
			args{
				200,
			},
			"OK",
		},
		{
			"404",
			args{
				404,
			},
			"Not Found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatusText(tt.args.code); got != tt.want {
				t.Errorf("StatusText() = %v, want %v", got, tt.want)
			}
		})
	}
}
