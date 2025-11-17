package server_test

import (
	"simpleGo/100gomistakes/011_functional_options_pattern/server"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name string
		host string
		opts []server.Option
		want *server.Server
	}{
		{
			name: "Valid options",
			host: "localhost",
			opts: []server.Option{
				server.WithPort(8080),
				server.WithTimeout(5 * time.Second),
				server.WithTLS(),
			},
			want: &server.Server{
				Host:    "localhost",
				Port:    8080,
				Timeout: 5 * time.Second,
				UseTLS:  true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := server.NewServer(tt.host, tt.opts...)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewServer() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
