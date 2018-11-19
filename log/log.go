package log

import (
	"runtime"

	"github.com/keakon/golog"
)

var defaultLogger *golog.Logger

func SetDefaultLogger(l *golog.Logger) {
	defaultLogger = l
}

func Debug(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1) // deeper caller will be more expensive
	defaultLogger.Log(golog.DebugLevel, file, line, "", args...)
}

func Debugf(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Log(golog.DebugLevel, file, line, msg, args...)
}

func Info(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Log(golog.InfoLevel, file, line, "", args...)
}

func Infof(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Log(golog.InfoLevel, file, line, msg, args...)
}

func Warn(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Log(golog.WarnLevel, file, line, "", args...)
}

func Warnf(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Log(golog.WarnLevel, file, line, msg, args...)
}

func Error(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Log(golog.ErrorLevel, file, line, "", args...)
}

func Errorf(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Log(golog.ErrorLevel, file, line, msg, args...)
}

func Crit(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Log(golog.CritLevel, file, line, "", args...)
}

func Critf(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Log(golog.CritLevel, file, line, msg, args...)
}
