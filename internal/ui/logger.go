// Package ui provides logging utilities with Charm Log
package ui

import (
	"os"

	"charm.land/log/v2"
	"github.com/charmbracelet/colorprofile"
)

// Logger is an interface for logging
type Logger interface {
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})
	Fatal(msg string, keyvals ...interface{})
	With(keyvals ...interface{}) Logger
}

// charmLogger implements Logger using Charm Log
type charmLogger struct {
	logger *log.Logger
}

// NewLogger creates a new logger with Charm Log
func NewLogger(verbose bool) Logger {
	level := log.InfoLevel
	if verbose {
		level = log.DebugLevel
	}

	l := log.NewWithOptions(colorprofile.NewWriter(os.Stderr, os.Environ()), log.Options{
		Level:           level,
		ReportCaller:    verbose,
		ReportTimestamp: true,
		Prefix:          "aceplay",
	})

	return &charmLogger{logger: l}
}

// NewDefaultLogger creates a logger with default configuration
func NewDefaultLogger() Logger {
	return NewLogger(false)
}

func (c *charmLogger) Debug(msg string, keyvals ...interface{}) {
	c.logger.Debug(msg, keyvals...)
}

func (c *charmLogger) Info(msg string, keyvals ...interface{}) {
	c.logger.Info(msg, keyvals...)
}

func (c *charmLogger) Warn(msg string, keyvals ...interface{}) {
	c.logger.Warn(msg, keyvals...)
}

func (c *charmLogger) Error(msg string, keyvals ...interface{}) {
	c.logger.Error(msg, keyvals...)
}

func (c *charmLogger) Fatal(msg string, keyvals ...interface{}) {
	c.logger.Fatal(msg, keyvals...)
}

func (c *charmLogger) With(keyvals ...interface{}) Logger {
	return &charmLogger{logger: c.logger.With(keyvals...)}
}

// SetLevel changes the logging level
func SetLevel(l Logger, verbose bool) {
	if cl, ok := l.(*charmLogger); ok {
		level := log.InfoLevel
		if verbose {
			level = log.DebugLevel
		}
		cl.logger.SetLevel(level)
	}
}
