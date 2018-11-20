package golog

import (
	"io"
	"runtime"
	"time"
)

// Level specifies the log level.
type Level uint8

// All the log levels.
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

// A Record is an item which contains enough context for the logger.
type Record struct {
	level   Level
	time    time.Time
	file    string
	line    int
	message string
	args    []interface{}
}

// A Logger is a leveled logger with several handlers.
type Logger struct {
	level    Level
	handlers []*Handler
}

// NewLogger creates a new Logger.
func NewLogger(lv Level) *Logger {
	return &Logger{level: lv}
}

// AddHandler adds a Handler to the Logger.
func (l *Logger) AddHandler(h *Handler) {
	l.handlers = append(l.handlers, h)
}

// Log logs a message with context.
func (l *Logger) Log(lv Level, file string, line int, msg string, args ...interface{}) {
	if lv < l.level || lv > CritLevel {
		return
	}

	r := recordPool.Get().(*Record)
	r.level = lv
	r.time = now()
	r.file = file
	r.line = line
	r.message = msg
	r.args = args

	for _, handler := range l.handlers {
		handler.Handle(r)
	}

	recordPool.Put(r)
}

// Close closes its handlers, it shouldn't be called twice.
func (l *Logger) Close() {
	for _, h := range l.handlers {
		h.Close()
	}
	l.handlers = nil
}

// Debug logs a debug level message. It use fmt.Sprint() to format args.
func (l *Logger) Debug(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1) // deeper caller will be more expensive
	l.Log(DebugLevel, file, line, "", args...)
}

// Debugf logs a debug level message. It use fmt.Sprintf() to format msg and args.
func (l *Logger) Debugf(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	l.Log(DebugLevel, file, line, msg, args...)
}

// Info logs a info level message. It use fmt.Sprint() to format args.
func (l *Logger) Info(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	l.Log(InfoLevel, file, line, "", args...)
}

// Infof logs a info level message. It use fmt.Sprintf() to format msg and args.
func (l *Logger) Infof(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	l.Log(InfoLevel, file, line, msg, args...)
}

// Warn logs a warning level message. It use fmt.Sprint() to format args.
func (l *Logger) Warn(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	l.Log(WarnLevel, file, line, "", args...)
}

// Warnf logs a warning level message. It use fmt.Sprintf() to format msg and args.
func (l *Logger) Warnf(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	l.Log(WarnLevel, file, line, msg, args...)
}

// Error logs an error level message. It use fmt.Sprint() to format args.
func (l *Logger) Error(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	l.Log(ErrorLevel, file, line, "", args...)
}

// Errorf logs a error level message. It use fmt.Sprintf() to format msg and args.
func (l *Logger) Errorf(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	l.Log(ErrorLevel, file, line, msg, args...)
}

// Crit logs a critical level message. It use fmt.Sprint() to format args.
func (l *Logger) Crit(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	l.Log(CritLevel, file, line, "", args...)
}

// Critf logs a critical level message. It use fmt.Sprintf() to format msg and args.
func (l *Logger) Critf(msg string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	l.Log(CritLevel, file, line, msg, args...)
}

// NewLoggerWithWriter creates an info level logger with a writer.
func NewLoggerWithWriter(w io.WriteCloser) *Logger {
	h := NewHandler(InfoLevel, DefaultFormatter)
	h.AddWriter(w)
	l := &Logger{level: InfoLevel}
	l.AddHandler(h)
	return l
}

// NewStdoutLogger creates a logger with a stdout writer.
func NewStdoutLogger() *Logger {
	return NewLoggerWithWriter(NewStdoutWriter())
}

// NewStderrLogger creates a logger with a stderr writer.
func NewStderrLogger() *Logger {
	return NewLoggerWithWriter(NewStderrWriter())
}

// SetInternalLogger sets the internalLogger which used to log internal errors.
func SetInternalLogger(l *Logger) {
	internalLogger = l
}
