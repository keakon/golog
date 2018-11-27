package log

import (
	"runtime"

	"github.com/keakon/golog"
)

var defaultLogger *golog.Logger

// SetDefaultLogger set the logger as the defaultLogger.
// The logging functions in this package use it as their logger.
func SetDefaultLogger(l *golog.Logger) {
	defaultLogger = l
}

// Debug logs a debug level message. It uses fmt.Sprint() to format args.
func Debug(args ...interface{}) {
	if defaultLogger.IsEnabledFor(golog.DebugLevel) {
		_, file, line, _ := runtime.Caller(1) // deeper caller will be more expensive
		defaultLogger.Log(golog.DebugLevel, file, line, "", args...)
	}
}

// Debugf logs a debug level message. It uses fmt.Sprintf() to format msg and args.
func Debugf(msg string, args ...interface{}) {
	if defaultLogger.IsEnabledFor(golog.DebugLevel) {
		_, file, line, _ := runtime.Caller(1)
		defaultLogger.Log(golog.DebugLevel, file, line, msg, args...)
	}
}

// Info logs a info level message. It uses fmt.Sprint() to format args.
func Info(args ...interface{}) {
	if defaultLogger.IsEnabledFor(golog.InfoLevel) {
		_, file, line, _ := runtime.Caller(1)
		defaultLogger.Log(golog.InfoLevel, file, line, "", args...)
	}
}

// Infof logs a info level message. It uses fmt.Sprintf() to format msg and args.
func Infof(msg string, args ...interface{}) {
	if defaultLogger.IsEnabledFor(golog.InfoLevel) {
		_, file, line, _ := runtime.Caller(1)
		defaultLogger.Log(golog.InfoLevel, file, line, msg, args...)
	}
}

// Warn logs a warning level message. It uses fmt.Sprint() to format args.
func Warn(args ...interface{}) {
	if defaultLogger.IsEnabledFor(golog.WarnLevel) {
		_, file, line, _ := runtime.Caller(1)
		defaultLogger.Log(golog.WarnLevel, file, line, "", args...)
	}
}

// Warnf logs a warning level message. It uses fmt.Sprintf() to format msg and args.
func Warnf(msg string, args ...interface{}) {
	if defaultLogger.IsEnabledFor(golog.WarnLevel) {
		_, file, line, _ := runtime.Caller(1)
		defaultLogger.Log(golog.WarnLevel, file, line, msg, args...)
	}
}

// Error logs an error level message. It uses fmt.Sprint() to format args.
func Error(args ...interface{}) {
	if defaultLogger.IsEnabledFor(golog.ErrorLevel) {
		_, file, line, _ := runtime.Caller(1)
		defaultLogger.Log(golog.ErrorLevel, file, line, "", args...)
	}
}

// Errorf logs a error level message. It uses fmt.Sprintf() to format msg and args.
func Errorf(msg string, args ...interface{}) {
	if defaultLogger.IsEnabledFor(golog.ErrorLevel) {
		_, file, line, _ := runtime.Caller(1)
		defaultLogger.Log(golog.ErrorLevel, file, line, msg, args...)
	}
}

// Crit logs a critical level message. It uses fmt.Sprint() to format args.
func Crit(args ...interface{}) {
	if defaultLogger.IsEnabledFor(golog.CritLevel) {
		_, file, line, _ := runtime.Caller(1)
		defaultLogger.Log(golog.CritLevel, file, line, "", args...)
	}
}

// Critf logs a critical level message. It uses fmt.Sprintf() to format msg and args.
func Critf(msg string, args ...interface{}) {
	if defaultLogger.IsEnabledFor(golog.CritLevel) {
		_, file, line, _ := runtime.Caller(1)
		defaultLogger.Log(golog.CritLevel, file, line, msg, args...)
	}
}
