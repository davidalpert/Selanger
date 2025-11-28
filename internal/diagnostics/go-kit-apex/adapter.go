package go_kit_apex

import (
	"errors"
	"fmt"
	"github.com/go-kit/log/level"

	apexlog "github.com/apex/log"
	"github.com/go-kit/log"
)

type Logger struct {
	field  *apexlog.Entry
	level  *apexlog.Level
	prefix string
}

type Option func(*Logger)

var errMissingValue = errors.New("(MISSING)")

// NewLogger returns a Go kit log.Logger that sends log events to an go-kit-apex.Logger.
func NewLogger(logger *apexlog.Entry, options ...Option) log.Logger {
	l := &Logger{
		field:  logger,
		prefix: "go-kit",
	}

	for _, optFunc := range options {
		optFunc(l)
	}

	return l
}

func levelPtr(level apexlog.Level) *apexlog.Level {
	return &level
}

// WithPrefix configures a go-kit-apex logger to log at level for all events.
func WithPrefix(prefix string) Option {
	return func(c *Logger) {
		c.prefix = prefix
	}
}

// WithDefaultLevel configures a go-kit-apex logger to log all events at a default level
func WithDefaultLevel(level apexlog.Level) Option {
	return func(c *Logger) {
		c.level = levelPtr(level)
	}
}

func (l Logger) Log(kv ...interface{}) error {
	fields := apexlog.Fields{}
	for i := 0; i < len(kv); i += 2 {
		if i+1 < len(kv) {
			fields[fmt.Sprint(kv[i])] = kv[i+1]
		} else {
			fields[fmt.Sprint(kv[i])] = errMissingValue
		}
	}

	var logLevel *apexlog.Level = nil
	if msgLevel, found := fields["level"]; found {
		switch msgLevel {
		case level.DebugValue():
			logLevel = levelPtr(apexlog.DebugLevel)
		case level.InfoValue():
			logLevel = levelPtr(apexlog.InfoLevel)
		case level.WarnValue():
			logLevel = levelPtr(apexlog.WarnLevel)
		case level.ErrorValue():
			logLevel = levelPtr(apexlog.ErrorLevel)
		default:
			fmt.Printf("UNMATCHED msgLeval %#v\n", msgLevel)
			logLevel = l.level
		}
	}

	message := l.prefix
	if msg, found := fields["event"]; found {
		message = fmt.Sprintf("%s: %s", l.prefix, msg)
	} else if msg, found = fields["message"]; found {
		message = fmt.Sprintf("%s: %s", l.prefix, msg)
	}

	if logLevel != nil {
		switch *logLevel {
		case apexlog.InfoLevel:
			l.field.WithFields(fields).Info(message)
		case apexlog.ErrorLevel:
			l.field.WithFields(fields).Error(message)
		case apexlog.DebugLevel:
			l.field.WithFields(fields).Debug(message)
		case apexlog.WarnLevel:
			l.field.WithFields(fields).Warn(message)
		default:
			l.field.WithFields(fields).Debug(message)
		}
	}

	return nil
}
