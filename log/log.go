package log

import (
	"github.com/keakon/golog"
)

var defaultLogger *golog.Logger

func nop(args ...interface{})              {}
func nopf(msg string, args ...interface{}) {}

var (
	// Debug logs a debug level message. It uses fmt.Fprint() to format args.
	Debug = nop
	// Info logs an info level message. It uses fmt.Fprint() to format args.
	Info = nop
	// Warn logs a warning level message. It uses fmt.Fprint() to format args.
	Warn = nop
	// Error logs an error level message. It uses fmt.Fprint() to format args.
	Error = nop
	// Crit logs a critical level message. It uses fmt.Fprint() to format args.
	Crit = nop

	// Debugf logs a debug level message. It uses fmt.Fprintf() to format msg and args.
	Debugf = nopf
	// Infof logs an info level message. It uses fmt.Fprintf() to format msg and args.
	Infof = nopf
	// Warnf logs a warning level message. It uses fmt.Fprintf() to format msg and args.
	Warnf = nopf
	// Errorf logs an error level message. It uses fmt.Fprintf() to format msg and args.
	Errorf = nopf
	// Critf logs a critical level message. It uses fmt.Fprintf() to format msg and args.
	Critf = nopf

	// logVars / logfVars index by log level so SetDefaultLogger and SetLogFunc /
	// SetLogfFunc can rewrite the matching package-level dispatch variable directly,
	// without a switch.
	logVars = [5]*func(args ...interface{}){
		&Debug, &Info, &Warn, &Error, &Crit,
	}
	logfVars = [5]*func(msg string, args ...interface{}){
		&Debugf, &Infof, &Warnf, &Errorf, &Critf,
	}

	logFuncs = [5]func(args ...interface{}){
		func(args ...interface{}) {
			file, line := golog.Caller(1) // deeper caller would be more expensive; do not init these via a loop
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

	logfFuncs = [5]func(msg string, args ...interface{}){
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

	// logFuncsNoCaller / logfFuncsNoCaller mirror the variants above but skip the
	// Caller() stack walk. They are selected by SetDefaultLogger when the logger's
	// format does not render the source location (no %s/%S), avoiding ~50% of the
	// per-call cost on those formats.
	logFuncsNoCaller = [5]func(args ...interface{}){
		func(args ...interface{}) {
			defaultLogger.Log(golog.DebugLevel, "", 0, "", args...)
		},
		func(args ...interface{}) {
			defaultLogger.Log(golog.InfoLevel, "", 0, "", args...)
		},
		func(args ...interface{}) {
			defaultLogger.Log(golog.WarnLevel, "", 0, "", args...)
		},
		func(args ...interface{}) {
			defaultLogger.Log(golog.ErrorLevel, "", 0, "", args...)
		},
		func(args ...interface{}) {
			defaultLogger.Log(golog.CritLevel, "", 0, "", args...)
		},
	}

	logfFuncsNoCaller = [5]func(msg string, args ...interface{}){
		func(msg string, args ...interface{}) {
			defaultLogger.Log(golog.DebugLevel, "", 0, msg, args...)
		},
		func(msg string, args ...interface{}) {
			defaultLogger.Log(golog.InfoLevel, "", 0, msg, args...)
		},
		func(msg string, args ...interface{}) {
			defaultLogger.Log(golog.WarnLevel, "", 0, msg, args...)
		},
		func(msg string, args ...interface{}) {
			defaultLogger.Log(golog.ErrorLevel, "", 0, msg, args...)
		},
		func(msg string, args ...interface{}) {
			defaultLogger.Log(golog.CritLevel, "", 0, msg, args...)
		},
	}
)

// SetLogFunc sets the log function with the specified level for the defaultLogger.
// This function should be called before SetDefaultLogger.
func SetLogFunc(f func(args ...interface{}), level golog.Level) {
	if int(level) < len(logVars) {
		*logVars[level] = f
	}
}

// SetLogfFunc sets the logf function with the specified level for the defaultLogger.
// This function should be called before SetDefaultLogger.
func SetLogfFunc(f func(msg string, args ...interface{}), level golog.Level) {
	if int(level) < len(logfVars) {
		*logfVars[level] = f
	}
}

// SetDefaultLogger sets the logger as the defaultLogger.
// The logging functions in this package use it as their logger.
// This function should be called before using the logging functions below.
func SetDefaultLogger(l *golog.Logger) {
	defaultLogger = l
	if l == nil {
		for level := golog.DebugLevel; level <= golog.CritLevel; level++ {
			*logVars[level] = nop
			*logfVars[level] = nopf
		}
		return
	}
	minLevel := l.GetMinLevel()
	needsCaller := l.NeedsCaller()
	for level := golog.DebugLevel; level <= golog.CritLevel; level++ {
		if level < minLevel {
			*logVars[level] = nop
			*logfVars[level] = nopf
		} else if needsCaller {
			*logVars[level] = logFuncs[level]
			*logfVars[level] = logfFuncs[level]
		} else {
			*logVars[level] = logFuncsNoCaller[level]
			*logfVars[level] = logfFuncsNoCaller[level]
		}
	}
}

// IsEnabledFor returns whether the default logger is enabled for the level.
func IsEnabledFor(level golog.Level) bool {
	return defaultLogger != nil && defaultLogger.IsEnabledFor(level)
}
