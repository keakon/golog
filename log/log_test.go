package log

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/keakon/golog"
)

type memoryWriter struct {
	bytes.Buffer
}

func (w *memoryWriter) Close() error {
	w.Buffer.Reset()
	return nil
}

func TestLogFuncs(t *testing.T) {
	w := &memoryWriter{}
	h := golog.NewHandler(golog.InfoLevel, golog.DefaultFormatter)
	h.AddWriter(w)
	l := golog.NewLogger(golog.InfoLevel)
	l.AddHandler(h)
	SetDefaultLogger(l)

	l.Debug("test")
	if w.Buffer.Len() != 0 {
		t.Error("memoryWriter is not empty")
	}
	l.Debugf("test")
	if w.Buffer.Len() != 0 {
		t.Error("memoryWriter is not empty")
	}

	l.Info("test")
	if w.Buffer.Len() == 0 {
		t.Error("memoryWriter is empty")
	}
	w.Buffer.Reset()

	l.Infof("test")
	if w.Buffer.Len() == 0 {
		t.Error("memoryWriter is empty")
	}
	w.Buffer.Reset()

	l.Error("test")
	if w.Buffer.Len() == 0 {
		t.Error("memoryWriter is empty")
	}
	w.Buffer.Reset()

	l.Errorf("test")
	if w.Buffer.Len() == 0 {
		t.Error("memoryWriter is empty")
	}
	l.Close()

	h = golog.NewHandler(golog.ErrorLevel, golog.DefaultFormatter)
	h.AddWriter(w)
	l = golog.NewLogger(golog.ErrorLevel)
	l.AddHandler(h)
	SetDefaultLogger(l)

	l.Info("test")
	if w.Buffer.Len() != 0 {
		t.Error("memoryWriter is not empty")
	}
	w.Buffer.Reset()

	l.Error("test")
	if w.Buffer.Len() == 0 {
		t.Error("memoryWriter is empty")
	}
	l.Close()
}

func BenchmarkBufferedFileLogger(b *testing.B) {
	path := filepath.Join(os.TempDir(), "test.log")
	os.Remove(path)
	w, err := golog.NewBufferedFileWriter(path)
	if err != nil {
		b.Error(err)
	}
	h := golog.NewHandler(golog.InfoLevel, golog.DefaultFormatter)
	h.AddWriter(w)
	l := golog.NewLogger(golog.InfoLevel)
	l.AddHandler(h)
	SetDefaultLogger(l)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Infof("test")
		}
	})
	l.Close()
}

func BenchmarkDiscardLogger(b *testing.B) {
	w := golog.NewDiscardWriter()
	h := golog.NewHandler(golog.InfoLevel, golog.DefaultFormatter)
	h.AddWriter(w)
	l := golog.NewLogger(golog.InfoLevel)
	l.AddHandler(h)
	SetDefaultLogger(l)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Infof("test")
		}
	})
	l.Close()
}

func BenchmarkNopLog(b *testing.B) {
	w := golog.NewDiscardWriter()
	h := golog.NewHandler(golog.InfoLevel, golog.DefaultFormatter)
	h.AddWriter(w)
	l := golog.NewLogger(golog.InfoLevel)
	l.AddHandler(h)
	SetDefaultLogger(l)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Debugf("test")
		}
	})
	l.Close()
}

func BenchmarkMultiLevels(b *testing.B) {
	w := golog.NewDiscardWriter()
	dh := golog.NewHandler(golog.DebugLevel, golog.DefaultFormatter)
	dh.AddWriter(w)
	ih := golog.NewHandler(golog.InfoLevel, golog.DefaultFormatter)
	ih.AddWriter(w)
	wh := golog.NewHandler(golog.WarnLevel, golog.DefaultFormatter)
	wh.AddWriter(w)
	eh := golog.NewHandler(golog.ErrorLevel, golog.DefaultFormatter)
	eh.AddWriter(w)
	ch := golog.NewHandler(golog.CritLevel, golog.DefaultFormatter)
	ch.AddWriter(w)

	l := golog.NewLogger(golog.WarnLevel)
	l.AddHandler(dh)
	l.AddHandler(ih)
	l.AddHandler(wh)
	l.AddHandler(eh)
	l.AddHandler(ch)
	SetDefaultLogger(l)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Debugf("test")
			Infof("test")
			Warnf("test")
			Errorf("test")
			Critf("test")
		}
	})
	l.Close()
}
