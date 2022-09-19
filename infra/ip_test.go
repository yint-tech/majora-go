package infra

import (
	"testing"
)

func TestGetIpByName(t *testing.T) {
	type args struct {
		netname string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "en0",
			args: args{netname: "en0"},
			want: "192.168.0.101",
		},
		{
			name: "lo0",
			args: args{netname: "lo0"},
			want: "127.0.0.1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetIpByName(tt.args.netname); got != tt.want {
				t.Errorf("GetIpByName() = %v, want %v", got, tt.want)
			}
		})
	}
}
