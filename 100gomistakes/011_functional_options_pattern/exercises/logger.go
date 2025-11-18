package exercises

import (
	"io"
	"os"
)

type Logr struct {
	Level     string
	Output    io.Writer
	Timestamp bool
}

type Options func(*Logr)

func WithLevel(level string) Options {
	return func(l *Logr) {
		l.Level = level
	}
}

func NewLogger(options ...Options) *Logr {
	logger := &Logr{
		Level:     "INFO",
		Output:    os.Stdout,
		Timestamp: true,
	}

	for _, o := range options {
		o(logger)
	}

	return logger

}
