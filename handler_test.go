package golog

import (
	"bytes"
	"strings"
	"testing"
)

// captureWriter records everything written to it so a test can assert the exact
// bytes a handler produced.
type captureWriter struct {
	bytes.Buffer
}

func (w *captureWriter) Close() error { return nil }

func TestHandle(t *testing.T) {
	h := NewHandler(InfoLevel, DefaultFormatter)
	r := &Record{tm: now()}
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

func TestHandleWithNilFormatter(t *testing.T) {
	h := NewHandler(InfoLevel, nil)
	r := &Record{level: InfoLevel, tm: now()}
	if !h.Handle(r) {
		t.Error("info handler ignored info record")
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
	if w.Writer == nil {
		t.Error("close handler closed discard writer")
	}
	h.Close()
}

// TestHandleOversizedBufferNotPooled drives a record whose formatted output
// exceeds maxPooledBufSize. The handler must still write it correctly while
// taking the drop branch (buf.Cap() > maxPooledBufSize) instead of returning the
// oversized buffer to bufPool, so that capacity cannot be pinned for the process
// lifetime. Handling a normal record afterwards confirms the pool still works.
func TestHandleOversizedBufferNotPooled(t *testing.T) {
	w := &captureWriter{}
	h := NewHandler(InfoLevel, ParseFormat("%m"))
	h.AddWriter(w)

	huge := strings.Repeat("a", maxPooledBufSize+1)
	if !h.Handle(&Record{level: InfoLevel, message: huge}) {
		t.Fatal("handler ignored info record")
	}
	if got := w.String(); got != huge+"\n" {
		t.Fatalf("oversized output length is %d, expected %d", len(got), len(huge)+1)
	}

	w.Reset()
	if !h.Handle(&Record{level: InfoLevel, message: "small"}) {
		t.Fatal("handler ignored info record")
	}
	if got := w.String(); got != "small\n" {
		t.Fatalf("output after oversized record is %q", got)
	}
}
