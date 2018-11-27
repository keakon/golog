package golog

import (
	"bytes"
	"io"
)

// A Handler is a leveled log handler with a formatter and several writers.
type Handler struct {
	level     Level
	formatter *Formatter
	writers   []io.WriteCloser
}

// NewHandler creates a new Handler of the given level with the formatter.
// Records with the lower level than the handler will be ignored.
func NewHandler(level Level, formatter *Formatter) *Handler {
	return &Handler{
		level:     level,
		formatter: formatter,
	}
}

// AddWriter adds a writer to the Handler.
// The Write() method of the writer should be thread-safe.
func (h *Handler) AddWriter(w io.WriteCloser) {
	h.writers = append(h.writers, w)
}

// Handle formats a record using its formatter, then writes the formatted result to all of its writers.
// Returns true if it can handle the record.
// It's not thread-safe, concurrent record may be written in a random order through different writers.
// But two records won't be mixed in a single line.
func (h *Handler) Handle(r *Record) bool {
	if r.level >= h.level {
		buf := bufPool.Get().(*bytes.Buffer)
		buf.Reset()
		h.formatter.Format(r, buf)
		content := buf.Bytes()
		for _, w := range h.writers {
			_, err := w.Write(content)
			if err != nil {
				logError(err)
			}
		}
		bufPool.Put(buf)
		return true
	}
	return false
}

// Close closes all its writers.
// It's safe to call this method more than once,
// but it's unsafe to call its writers' Close() more than once.
func (h *Handler) Close() {
	for _, w := range h.writers {
		err := w.Close()
		if err != nil {
			logError(err)
		}
	}
	h.writers = nil
}
