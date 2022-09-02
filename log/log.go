package log

import (
	"github.com/keakon/golog"
)

var defaultLogger *golog.Logger

// SetLogFunc set the log function with specified level for the defaultLogger.
// This function should be called before SetDefaultLogger.
func SetLogFunc(f func(args ...interface{}), level golog.Level) {
	switch level {
	case golog.DebugLevel:
		Debug = f
	case golog.InfoLevel:
		Info = f
	case golog.WarnLevel:
		Warn = f
	case golog.ErrorLevel:
		Error = f
	case golog.CritLevel:
		Crit = f
	}
}

// SetLogfFunc set the logf function with specified level for the defaultLogger.
// This function should be called before SetDefaultLogger.
func SetLogfFunc(f func(msg string, args ...interface{}), level golog.Level) {
	switch level {
	case golog.DebugLevel:
		Debugf = f
	case golog.InfoLevel:
		Infof = f
	case golog.WarnLevel:
		Warnf = f
	case golog.ErrorLevel:
		Errorf = f
	case golog.CritLevel:
		Critf = f
	}
}

// SetDefaultLogger set the logger as the defaultLogger.
// The logging functions in this package use it as their logger.
// This function should be called before using below functions.
func SetDefaultLogger(l *golog.Logger) {
	defaultLogger = l
	minLevel := l.GetMinLevel()
	for level := golog.DebugLevel; level < minLevel; level++ {
		switch level {
		case golog.DebugLevel:
			Debug = nop
		case golog.InfoLevel:
			Info = nop
		case golog.WarnLevel:
			Warn = nop
		case golog.ErrorLevel:
			Error = nop
		case golog.CritLevel:
			Crit = nop
		}
	}
	for level := golog.DebugLevel; level < minLevel; level++ {
		switch level {
		case golog.DebugLevel:
			Debugf = nopf
		case golog.InfoLevel:
			Infof = nopf
		case golog.WarnLevel:
			Warnf = nopf
		case golog.ErrorLevel:
			Errorf = nopf
		case golog.CritLevel:
			Critf = nopf
		}
	}
}

func nop(args ...interface{})              {}
func nopf(msg string, args ...interface{}) {}

// Debug logs a _debug level message. It uses fmt.Fprint() to format args.
var Debug = func(args ...interface{}) {
	file, line := golog.Caller(1) // deeper caller will be more expensive
	defaultLogger.Log(golog.DebugLevel, file, line, "", args...)
}

// Debugf logs a _debug level message. It uses fmt.Fprintf() to format msg and args.
var Debugf = func(msg string, args ...interface{}) {
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.DebugLevel, file, line, msg, args...)
}

// Info logs a _info level message. It uses fmt.Fprint() to format args.
var Info = func(args ...interface{}) {
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.InfoLevel, file, line, "", args...)
}

// Infof logs a _info level message. It uses fmt.Fprintf() to format msg and args.
var Infof = func(msg string, args ...interface{}) {
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.InfoLevel, file, line, msg, args...)
}

// Warn logs a _warning level message. It uses fmt.Fprint() to format args.
var Warn = func(args ...interface{}) {
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.WarnLevel, file, line, "", args...)
}

// Warnf logs a _warning level message. It uses fmt.Fprintf() to format msg and args.
var Warnf = func(msg string, args ...interface{}) {
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.WarnLevel, file, line, msg, args...)
}

// Error logs an _error level message. It uses fmt.Fprint() to format args.
var Error = func(args ...interface{}) {
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.ErrorLevel, file, line, "", args...)
}

// Errorf logs a _error level message. It uses fmt.Fprintf() to format msg and args.
var Errorf = func(msg string, args ...interface{}) {
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.ErrorLevel, file, line, msg, args...)
}

// Crit logs a _critical level message. It uses fmt.Fprint() to format args.
var Crit = func(args ...interface{}) {
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.CritLevel, file, line, "", args...)
}

// Critf logs a _critical level message. It uses fmt.Fprintf() to format msg and args.
var Critf = func(msg string, args ...interface{}) {
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.CritLevel, file, line, msg, args...)
}
