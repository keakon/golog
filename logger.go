package golog

import (
	"io"
	"runtime"
	"time"
)

type Level uint8

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	CritLevel
)

var (
	levelNames = []byte("DIWEC")

	internalLogger *Logger
)

type Record struct {
	Level   Level
	Time    time.Time
	File    string
	Line    int
	Message string
	Args    []interface{}
}

type Logger struct {
	level    Level
	handlers []*Handler
}

func (l *Logger) AddHandler(h *Handler) {
	l.handlers = append(l.handlers, h)
}

func (l *Logger) Log(lv Level, file string, line int, msg string, args ...interface{}) {
	if lv < l.level || lv > CritLevel {
		return
	}

	r := &Record{
		Level:   lv,
		Time:    now(),
		File:    file,
		Line:    line,
		Message: msg,
		Args:    args,
	}

	for _, handler := range l.handlers {
		handler.Handle(r)
	}
}

func (l *Logger) Close() {
	for _, h := range l.handlers {
		h.Close()
	}
}

func (l *Logger) Debug(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1) // deeper caller will be more expensive
	l.Log(DebugLevel, file, line, "", args...)
}

func (l *Logger) Debugf(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	l.Log(DebugLevel, file, line, msg, args...)
}

func (l *Logger) Info(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	l.Log(InfoLevel, file, line, "", args...)
}

func (l *Logger) Infof(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	l.Log(InfoLevel, file, line, msg, args...)
}

func (l *Logger) Warn(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	l.Log(WarnLevel, file, line, "", args...)
}

func (l *Logger) Warnf(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	l.Log(WarnLevel, file, line, msg, args...)
}

func (l *Logger) Error(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	l.Log(ErrorLevel, file, line, "", args...)
}

func (l *Logger) Errorf(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	l.Log(ErrorLevel, file, line, msg, args...)
}

func (l *Logger) Crit(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	l.Log(CritLevel, file, line, "", args...)
}

func (l *Logger) Critf(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	l.Log(CritLevel, file, line, msg, args...)
}

func NewLoggerWithWriter(w io.WriteCloser) *Logger {
	h := NewHandler(InfoLevel, DefaultFormatter)
	h.AddWriter(w)
	l := &Logger{level: InfoLevel}
	l.AddHandler(h)
	return l
}

func NewStdoutLogger() *Logger {
	return NewLoggerWithWriter(NewStdoutWriter())
}

func NewStderrLogger() *Logger {
	return NewLoggerWithWriter(NewStderrWriter())
}

func SetInternalLogger(l *Logger) {
	internalLogger = l
}
