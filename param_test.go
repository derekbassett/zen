package zen

import "testing"

func TestParams_ByName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		ps   Params
		args args
		want string
	}{
		{
			"case1",
			Params{Param{Key: "name", Value: "zen"}, Param{Key: "version", Value: Version}},
			args{"name"},
			"zen",
		},
		{
			"case2",
			Params{Param{Key: "name", Value: "zen"}, Param{Key: "version", Value: Version}},
			args{"age"},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ps.ByName(tt.args.name); got != tt.want {
				t.Errorf("Params.ByName() = %v, want %v", got, tt.want)
			}
		})
	}
}
