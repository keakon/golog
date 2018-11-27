package golog

import (
	"testing"
)

func TestHandle(t *testing.T) {
	h := NewHandler(InfoLevel, DefaultFormatter)
	r := &Record{time: now()}
	if h.Handle(r) {
		t.Error("info handler handled debug record")
	}

	r.level = InfoLevel
	if !h.Handle(r) {
		t.Error("info handler ignored info record")
	}

	r.level = ErrorLevel
	if !h.Handle(r) {
		t.Error("error handler ignored info record")
	}
}

func TestCloseHandler(t *testing.T) {
	h := NewHandler(InfoLevel, DefaultFormatter)
	h.Close()
	h.Close()

	w := NewDiscardWriter()
	h.AddWriter(w)
	h.Close()
	if len(h.writers) > 0 {
		t.Error("closed handler is not empty")
	}
	if w.Writer != nil {
		t.Error("close handler left its writer opened")
	}
	h.Close()
}
