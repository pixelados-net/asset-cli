// Package logger creates the process logger.
package logger

import (
	"fmt"

	"go.uber.org/zap/zapcore"
)

// Format identifies the Zap encoder format.
type Format string

const (
	// FormatConsole enables human-readable structured logs.
	FormatConsole Format = "console"
	// FormatJSON enables machine-readable JSON logs.
	FormatJSON Format = "json"
)

// Config contains process logger settings.
type Config struct {
	// Level is the minimum enabled Zap log level.
	Level zapcore.Level `env:"LEVEL" envDefault:"info"`
	// Format selects the console or JSON encoder.
	Format Format `env:"FORMAT" envDefault:"console"`
}

// UnmarshalText validates and stores a logger format.
func (format *Format) UnmarshalText(value []byte) error {
	parsed := Format(value)
	if parsed != FormatConsole && parsed != FormatJSON {
		return fmt.Errorf("unsupported log format %q", value)
	}
	*format = parsed
	return nil
}
