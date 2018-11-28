package log

import (
	"runtime"

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
	logPtrs = map[golog.Level](*func(args ...interface{})){
		golog.DebugLevel: &Debug,
		golog.InfoLevel:  &Info,
		golog.WarnLevel:  &Warn,
		golog.ErrorLevel: &Error,
		golog.CritLevel:  &Crit,
	}
	logfPtrs = map[golog.Level](*func(msg string, args ...interface{})){
		golog.DebugLevel: &Debugf,
		golog.InfoLevel:  &Infof,
		golog.WarnLevel:  &Warnf,
		golog.ErrorLevel: &Errorf,
		golog.CritLevel:  &Critf,
	}
)

// SetDefaultLogger set the logger as the defaultLogger.
// The logging functions in this package use it as their logger.
// This function should be called before using the others.
func SetDefaultLogger(l *golog.Logger) {
	defaultLogger = l

	minLevel := l.GetMinLevel()
	for level, f := range logFuncs {
		if minLevel <= level {
			*logPtrs[level] = f
		} else {
			*logPtrs[level] = nop
		}
	}
	for level, f := range logfFuncs {
		if minLevel <= level {
			*logfPtrs[level] = f
		} else {
			*logfPtrs[level] = nopf
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
	_, file, line, _ := runtime.Caller(1) // deeper caller will be more expensive
	defaultLogger.Log(golog.DebugLevel, file, line, "", args...)
}

func _debugf(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Log(golog.DebugLevel, file, line, msg, args...)
}

func _info(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Log(golog.InfoLevel, file, line, "", args...)
}

func _infof(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Log(golog.InfoLevel, file, line, msg, args...)
}

func _warn(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Log(golog.WarnLevel, file, line, "", args...)
}

func _warnf(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Log(golog.WarnLevel, file, line, msg, args...)
}

func _error(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Log(golog.ErrorLevel, file, line, "", args...)
}

func _errorf(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Log(golog.ErrorLevel, file, line, msg, args...)
}

func _crit(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Log(golog.CritLevel, file, line, "", args...)
}

func _critf(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Log(golog.CritLevel, file, line, msg, args...)
}
