package golog

import (
	"bytes"
	"io"
)

type Handler struct {
	level     Level
	formatter *Formatter
	writers   []io.WriteCloser
}

func NewHandler(level Level, formatter *Formatter) *Handler {
	return &Handler{
		level:     level,
		formatter: formatter,
	}
}

func (h *Handler) AddWriter(w io.WriteCloser) {
	h.writers = append(h.writers, w)
}

func (h *Handler) Handle(r *Record) {
	if r.Level >= h.level {
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

func (h *Handler) Close() {
	for _, w := range h.writers {
		err := w.Close()
		if err != nil {
			logError(err)
		}
	}
	h.writers = nil
}
