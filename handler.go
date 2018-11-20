package golog

import (
	"bytes"
	"io"
)

// A Handler is a level log handler with a formatter and several writers.
type Handler struct {
	level     Level
	formatter *Formatter
	writers   []io.WriteCloser
}

// NewHandler creates a new Handler.
func NewHandler(level Level, formatter *Formatter) *Handler {
	return &Handler{
		level:     level,
		formatter: formatter,
	}
}

// AddWriter adds a writer to the Handler.
func (h *Handler) AddWriter(w io.WriteCloser) {
	h.writers = append(h.writers, w)
}

// Handle processes a record, formats it using the formatter, then writes to all the writers.
func (h *Handler) Handle(r *Record) {
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
	}
}

// Close closes all its writers, it shouldn't be called twice.
func (h *Handler) Close() {
	for _, w := range h.writers {
		err := w.Close()
		if err != nil {
			logError(err)
		}
	}
	h.writers = nil
}
