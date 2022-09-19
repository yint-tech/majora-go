package infra

import (
	"testing"
)

func TestPing(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "baidu",
			args: args{url: "https://www.baidu.com"},
			want: true,
		},
		{
			name: "google",
			args: args{url: "https://www.taobao.com"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Ping(tt.args.url); got != tt.want {
				t.Errorf("Ping() = %v, want %v", got, tt.want)
			}
		})
	}
}
