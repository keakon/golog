package golog

import (
	"io"
	"runtime"
	"sort"
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

	disabledLevel Level = 255
)

var (
	levelNames = []byte("DIWEC")

	internalLogger *Logger
)

// A Record is an item which contains required context for the logger.
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
	level    Level // the lowest acceptable level
	minLevel Level // the min level of its handlers
	handlers []*Handler
}

// NewLogger creates a new Logger of the given level.
// Messages with lower level than the logger will be ignored.
func NewLogger(lv Level) *Logger {
	return &Logger{level: lv, minLevel: disabledLevel} // disable all levels for empty logger
}

// AddHandler adds a Handler to the Logger.
// A handler with lower level than the logger will be ignored.
func (l *Logger) AddHandler(h *Handler) {
	if h.level < l.level {
		return
	}

	l.handlers = append(l.handlers, h)
	if len(l.handlers) > 1 {
		if h.level < l.minLevel {
			l.minLevel = h.level
		}
		sort.Slice(l.handlers, func(i, j int) bool {
			return l.handlers[i].level < l.handlers[j].level
		})
	} else {
		l.minLevel = h.level
	}
}

// IsEnabledFor returns whether it's enabled for the level
func (l *Logger) IsEnabledFor(level Level) bool {
	return l.minLevel <= level
}

// GetMinLevel returns its minLevel.
// Records lower than its minLevel will be ignored.
func (l *Logger) GetMinLevel() Level {
	return l.minLevel
}

// Log logs a message with context.
// A logger should check the message level before call its Log().
// The line param should be uint32.
// It's not thread-safe, concurrent messages may be written in a random order
// through different handlers or writers.
// But two messages won't be mixed in a single line.
func (l *Logger) Log(lv Level, file string, line int, msg string, args ...interface{}) {
	r := recordPool.Get().(*Record)
	r.level = lv
	r.time = now()
	r.file = file
	r.line = line
	r.message = msg
	r.args = args

	for _, h := range l.handlers {
		if !h.Handle(r) {
			break
		}
	}

	recordPool.Put(r)
}

// Close closes its handlers.
// It's safe to call this method more than once.
func (l *Logger) Close() {
	for _, h := range l.handlers {
		h.Close()
	}
	l.handlers = nil
}

// Debug logs a debug level message. It uses fmt.Sprint() to format args.
func (l *Logger) Debug(args ...interface{}) {
	if l.IsEnabledFor(DebugLevel) {
		_, file, line, _ := runtime.Caller(1) // deeper caller will be more expensive
		l.Log(DebugLevel, file, line, "", args...)
	}
}

// Debugf logs a debug level message. It uses fmt.Sprintf() to format msg and args.
func (l *Logger) Debugf(msg string, args ...interface{}) {
	if l.IsEnabledFor(DebugLevel) {
		_, file, line, _ := runtime.Caller(1)
		l.Log(DebugLevel, file, line, msg, args...)
	}
}

// Info logs a info level message. It uses fmt.Sprint() to format args.
func (l *Logger) Info(args ...interface{}) {
	if l.IsEnabledFor(InfoLevel) {
		_, file, line, _ := runtime.Caller(1)
		l.Log(InfoLevel, file, line, "", args...)
	}
}

// Infof logs a info level message. It uses fmt.Sprintf() to format msg and args.
func (l *Logger) Infof(msg string, args ...interface{}) {
	if l.IsEnabledFor(InfoLevel) {
		_, file, line, _ := runtime.Caller(1)
		l.Log(InfoLevel, file, line, msg, args...)
	}
}

// Warn logs a warning level message. It uses fmt.Sprint() to format args.
func (l *Logger) Warn(args ...interface{}) {
	if l.IsEnabledFor(WarnLevel) {
		_, file, line, _ := runtime.Caller(1)
		l.Log(WarnLevel, file, line, "", args...)
	}
}

// Warnf logs a warning level message. It uses fmt.Sprintf() to format msg and args.
func (l *Logger) Warnf(msg string, args ...interface{}) {
	if l.IsEnabledFor(WarnLevel) {
		_, file, line, _ := runtime.Caller(1)
		l.Log(WarnLevel, file, line, msg, args...)
	}
}

// Error logs an error level message. It uses fmt.Sprint() to format args.
func (l *Logger) Error(args ...interface{}) {
	if l.IsEnabledFor(ErrorLevel) {
		_, file, line, _ := runtime.Caller(1)
		l.Log(ErrorLevel, file, line, "", args...)
	}
}

// Errorf logs a error level message. It uses fmt.Sprintf() to format msg and args.
func (l *Logger) Errorf(msg string, args ...interface{}) {
	if l.IsEnabledFor(ErrorLevel) {
		_, file, line, _ := runtime.Caller(1)
		l.Log(ErrorLevel, file, line, msg, args...)
	}
}

// Crit logs a critical level message. It uses fmt.Sprint() to format args.
func (l *Logger) Crit(args ...interface{}) {
	if l.IsEnabledFor(CritLevel) {
		_, file, line, _ := runtime.Caller(1)
		l.Log(CritLevel, file, line, "", args...)
	}
}

// Critf logs a critical level message. It uses fmt.Sprintf() to format msg and args.
func (l *Logger) Critf(msg string, args ...interface{}) {
	if l.IsEnabledFor(CritLevel) {
		_, file, line, _ := runtime.Caller(1)
		l.Log(CritLevel, file, line, msg, args...)
	}
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
