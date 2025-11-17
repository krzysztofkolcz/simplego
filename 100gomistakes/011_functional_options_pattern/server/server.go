package server

import (
	"fmt"
	"time"
)

type Server struct {
	Host    string
	Port    int
	Timeout time.Duration
	UseTLS  bool
}

// Typ opcji
type Option func(*Server)

// Opcje
func WithPort(p int) Option {
	return func(s *Server) {
		s.Port = p
	}
}

func WithTimeout(t time.Duration) Option {
	return func(s *Server) {
		s.Timeout = t
	}
}

func WithTLS() Option {
	return func(s *Server) {
		s.UseTLS = true
	}
}

// Konstruktor
func NewServer(host string, opts ...Option) *Server {
	s := &Server{
		Host:    host,
		Port:    80,
		Timeout: 10 * time.Second,
		UseTLS:  false,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func main() {
	server := NewServer(
		"localhost",
		WithPort(8080),
		WithTimeout(5*time.Second),
		WithTLS(),
	)

	fmt.Printf("%+v\n", server)
}
