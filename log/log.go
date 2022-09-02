package log

import (
	"github.com/keakon/golog"
)

var defaultLogger *golog.Logger

var (
	logFuncs = map[golog.Level](func(args ...interface{})){
		golog.DebugLevel: _debug,
		golog.InfoLevel:  _info,
		golog.WarnLevel:  _warn,
		golog.ErrorLevel: _error,
		golog.CritLevel:  _crit,
	}
	logfFuncs = map[golog.Level](func(msg string, args ...interface{})){
		golog.DebugLevel: _debugf,
		golog.InfoLevel:  _infof,
		golog.WarnLevel:  _warnf,
		golog.ErrorLevel: _errorf,
		golog.CritLevel:  _critf,
	}
)

// SetLogFunc set the log function with specified level for the defaultLogger.
// This function should be called before SetDefaultLogger.
func SetLogFunc(f func(args ...interface{}), level golog.Level) {
	logFuncs[level] = f
}

// SetLogfFunc set the logf function with specified level for the defaultLogger.
// This function should be called before SetDefaultLogger.
func SetLogfFunc(f func(msg string, args ...interface{}), level golog.Level) {
	logfFuncs[level] = f
}

// SetDefaultLogger set the logger as the defaultLogger.
// The logging functions in this package use it as their logger.
// This function should be called before using below functions.
func SetDefaultLogger(l *golog.Logger) {
	defaultLogger = l

	minLevel := l.GetMinLevel()
	for level, f := range logFuncs {
		if minLevel > level {
			f = nop
		}
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
	for level, f := range logfFuncs {
		if minLevel > level {
			f = nopf
		}
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
}

func nop(args ...interface{})              {}
func nopf(msg string, args ...interface{}) {}

// Debug logs a _debug level message. It uses fmt.Fprint() to format args.
var Debug func(args ...interface{})

// Debugf logs a _debug level message. It uses fmt.Fprintf() to format msg and args.
var Debugf func(msg string, args ...interface{})

// Info logs a _info level message. It uses fmt.Fprint() to format args.
var Info func(args ...interface{})

// Infof logs a _info level message. It uses fmt.Fprintf() to format msg and args.
var Infof func(msg string, args ...interface{})

// Warn logs a _warning level message. It uses fmt.Fprint() to format args.
var Warn func(args ...interface{})

// Warnf logs a _warning level message. It uses fmt.Fprintf() to format msg and args.
var Warnf func(msg string, args ...interface{})

// Error logs an _error level message. It uses fmt.Fprint() to format args.
var Error func(args ...interface{})

// Errorf logs a _error level message. It uses fmt.Fprintf() to format msg and args.
var Errorf func(msg string, args ...interface{})

// Crit logs a _critical level message. It uses fmt.Fprint() to format args.
var Crit func(args ...interface{})

// Critf logs a _critical level message. It uses fmt.Fprintf() to format msg and args.
var Critf func(msg string, args ...interface{})

func _debug(args ...interface{}) {
	file, line := golog.Caller(1) // deeper caller will be more expensive
	defaultLogger.Log(golog.DebugLevel, file, line, "", args...)
}

func _debugf(msg string, args ...interface{}) {
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.DebugLevel, file, line, msg, args...)
}

func _info(args ...interface{}) {
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.InfoLevel, file, line, "", args...)
}

func _infof(msg string, args ...interface{}) {
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.InfoLevel, file, line, msg, args...)
}

func _warn(args ...interface{}) {
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.WarnLevel, file, line, "", args...)
}

func _warnf(msg string, args ...interface{}) {
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.WarnLevel, file, line, msg, args...)
}

func _error(args ...interface{}) {
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.ErrorLevel, file, line, "", args...)
}

func _errorf(msg string, args ...interface{}) {
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.ErrorLevel, file, line, msg, args...)
}

func _crit(args ...interface{}) {
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.CritLevel, file, line, "", args...)
}

func _critf(msg string, args ...interface{}) {
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.CritLevel, file, line, msg, args...)
}
