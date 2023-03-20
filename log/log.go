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
	for level := golog.DebugLevel; level <= golog.CritLevel; level++ {
		if level < minLevel {
			switch level {
			case golog.DebugLevel:
				Debug = nop
				Debugf = nopf
			case golog.InfoLevel:
				Info = nop
				Infof = nopf
			case golog.WarnLevel:
				Warn = nop
				Warnf = nopf
			case golog.ErrorLevel:
				Error = nop
				Errorf = nopf
			case golog.CritLevel:
				Crit = nop
				Critf = nopf
			}
		} else {
			switch level {
			case golog.DebugLevel:
				Debug = logFuncs[level]
				Debugf = logfFuncs[level]
			case golog.InfoLevel:
				Info = logFuncs[level]
				Infof = logfFuncs[level]
			case golog.WarnLevel:
				Warn = logFuncs[level]
				Warnf = logfFuncs[level]
			case golog.ErrorLevel:
				Error = logFuncs[level]
				Errorf = logfFuncs[level]
			case golog.CritLevel:
				Crit = logFuncs[level]
				Critf = logfFuncs[level]
			}
		}
	}
}

func nop(args ...interface{})              {}
func nopf(msg string, args ...interface{}) {}

var (
	// Debug logs a _debug level message. It uses fmt.Fprint() to format args.
	Debug = nop
	// Info logs a _info level message. It uses fmt.Fprint() to format args.
	Info = nop
	// Warn logs a _warning level message. It uses fmt.Fprint() to format args.
	Warn = nop
	// Error logs an _error level message. It uses fmt.Fprint() to format args.
	Error = nop
	// Crit logs a _critical level message. It uses fmt.Fprint() to format args.
	Crit = nop

	// Debugf logs a _debug level message. It uses fmt.Fprintf() to format msg and args.
	Debugf = nopf
	// Infof logs a _info level message. It uses fmt.Fprintf() to format msg and args.
	Infof = nopf
	// Warnf logs a _warning level message. It uses fmt.Fprintf() to format msg and args.
	Warnf = nopf
	// Errorf logs a _error level message. It uses fmt.Fprintf() to format msg and args.
	Errorf = nopf
	// Critf logs a _critical level message. It uses fmt.Fprintf() to format msg and args.
	Critf = nopf

	logFuncs = []func(args ...interface{}){
		func(args ...interface{}) {
			file, line := golog.Caller(1) // deeper caller will be more expensive, so don't use init() to initialize those functions
			defaultLogger.Log(golog.DebugLevel, file, line, "", args...)
		},
		func(args ...interface{}) {
			file, line := golog.Caller(1)
			defaultLogger.Log(golog.InfoLevel, file, line, "", args...)
		},
		func(args ...interface{}) {
			file, line := golog.Caller(1)
			defaultLogger.Log(golog.WarnLevel, file, line, "", args...)
		},
		func(args ...interface{}) {
			file, line := golog.Caller(1)
			defaultLogger.Log(golog.ErrorLevel, file, line, "", args...)
		},
		func(args ...interface{}) {
			file, line := golog.Caller(1)
			defaultLogger.Log(golog.CritLevel, file, line, "", args...)
		},
	}

	logfFuncs = []func(msg string, args ...interface{}){
		func(msg string, args ...interface{}) {
			file, line := golog.Caller(1)
			defaultLogger.Log(golog.DebugLevel, file, line, msg, args...)
		},
		func(msg string, args ...interface{}) {
			file, line := golog.Caller(1)
			defaultLogger.Log(golog.InfoLevel, file, line, msg, args...)
		},
		func(msg string, args ...interface{}) {
			file, line := golog.Caller(1)
			defaultLogger.Log(golog.WarnLevel, file, line, msg, args...)
		},
		func(msg string, args ...interface{}) {
			file, line := golog.Caller(1)
			defaultLogger.Log(golog.ErrorLevel, file, line, msg, args...)
		},
		func(msg string, args ...interface{}) {
			file, line := golog.Caller(1)
			defaultLogger.Log(golog.CritLevel, file, line, msg, args...)
		},
	}
)
