package internal

import (
	"os"

	"github.com/rs/zerolog"
)

type ConsoleLevelWriter struct{}

func (l ConsoleLevelWriter) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

func (l ConsoleLevelWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	if level <= zerolog.WarnLevel {
		return zerolog.ConsoleWriter{Out: os.Stdout}.Write(p)
	}
	return zerolog.ConsoleWriter{Out: os.Stderr}.Write(p)
}
