//go:build !race
// +build !race

// golog.FastTimer is not thread-safe.

package log

import (
	"bytes"
	"errors"
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

func errorFunc(args ...interface{}) {
	if len(args) == 1 {
		arg := args[0]
		if _, ok := arg.(error); ok {
			// skip
			return
		}
	}
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.ErrorLevel, file, line, "", args...)
}

func errorfFunc(msg string, args ...interface{}) {
	if len(args) == 1 {
		arg := args[0]
		if _, ok := arg.(error); ok {
			// skip
			return
		}
	}
	file, line := golog.Caller(1)
	defaultLogger.Log(golog.ErrorLevel, file, line, msg, args...)
}

func TestSetLogFunc(t *testing.T) {
	golog.StartFastTimer()
	defer golog.StopFastTimer()

	w := &memoryWriter{}
	h := golog.NewHandler(golog.InfoLevel, golog.DefaultFormatter)
	h.AddWriter(w)
	l := golog.NewLogger(golog.InfoLevel)
	l.AddHandler(h)

	SetDefaultLogger(l)
	SetLogFunc(errorFunc, golog.ErrorLevel)
	SetLogfFunc(errorfFunc, golog.ErrorLevel)

	Debug("test")
	if w.Buffer.Len() != 0 {
		t.Error("memoryWriter is not empty")
	}
	w.Buffer.Reset()

	Error(errors.New("test"))
	if w.Buffer.Len() != 0 {
		t.Error("memoryWriter is not empty")
	}

	Error("test")
	if w.Buffer.Len() == 0 {
		t.Error("memoryWriter is empty")
	}
	w.Buffer.Reset()

	Errorf("error: %v", errors.New("test"))
	if w.Buffer.Len() != 0 {
		t.Error("memoryWriter is not empty")
	}

	Errorf("error: %s", "test")
	if w.Buffer.Len() == 0 {
		t.Error("memoryWriter is empty")
	}
	w.Buffer.Reset()

	l.Close()
}

func TestLogFuncs(t *testing.T) {
	golog.StartFastTimer()
	defer golog.StopFastTimer()

	w := &memoryWriter{}
	h := golog.NewHandler(golog.InfoLevel, golog.DefaultFormatter)
	h.AddWriter(w)
	l := golog.NewLogger(golog.InfoLevel)
	l.AddHandler(h)
	SetDefaultLogger(l)

	Debug("test")
	if w.Buffer.Len() != 0 {
		t.Error("memoryWriter is not empty")
	}
	Debugf("test")
	if w.Buffer.Len() != 0 {
		t.Error("memoryWriter is not empty")
	}

	Info("test")
	if w.Buffer.Len() == 0 {
		t.Error("memoryWriter is empty")
	}
	w.Buffer.Reset()

	Infof("test")
	if w.Buffer.Len() == 0 {
		t.Error("memoryWriter is empty")
	}
	w.Buffer.Reset()

	Error("test")
	if w.Buffer.Len() == 0 {
		t.Error("memoryWriter is empty")
	}
	w.Buffer.Reset()

	Errorf("test")
	if w.Buffer.Len() == 0 {
		t.Error("memoryWriter is empty")
	}
	l.Close()

	h = golog.NewHandler(golog.ErrorLevel, golog.DefaultFormatter)
	h.AddWriter(w)
	l = golog.NewLogger(golog.ErrorLevel)
	l.AddHandler(h)
	SetDefaultLogger(l)

	Info("test")
	if w.Buffer.Len() != 0 {
		t.Error("memoryWriter is not empty")
	}
	w.Buffer.Reset()

	Error("test")
	if w.Buffer.Len() == 0 {
		t.Error("memoryWriter is empty")
	}
	l.Close()
}

func BenchmarkDiscardLogger(b *testing.B) {
	golog.StartFastTimer()
	defer golog.StopFastTimer()

	w := golog.NewDiscardWriter()
	h := golog.NewHandler(golog.InfoLevel, golog.DefaultFormatter)
	h.AddWriter(w)
	l := golog.NewLogger(golog.InfoLevel)
	l.AddHandler(h)
	SetDefaultLogger(l)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Infof("test")
	}
	l.Close()
}

func BenchmarkDiscardLoggerParallel(b *testing.B) {
	golog.StartFastTimer()
	defer golog.StopFastTimer()

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

func BenchmarkDiscardLoggerWithoutTimer(b *testing.B) {
	w := golog.NewDiscardWriter()
	h := golog.NewHandler(golog.InfoLevel, golog.DefaultFormatter)
	h.AddWriter(w)
	l := golog.NewLogger(golog.InfoLevel)
	l.AddHandler(h)
	SetDefaultLogger(l)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Infof("test")
	}
	l.Close()
}

func BenchmarkDiscardLoggerWithoutTimerParallel(b *testing.B) {
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
	golog.StartFastTimer()
	defer golog.StopFastTimer()

	w := golog.NewDiscardWriter()
	h := golog.NewHandler(golog.InfoLevel, golog.DefaultFormatter)
	h.AddWriter(w)
	l := golog.NewLogger(golog.InfoLevel)
	l.AddHandler(h)
	SetDefaultLogger(l)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Infof("test")
	}
	l.Close()
}

func BenchmarkNopLogParallel(b *testing.B) {
	golog.StartFastTimer()
	defer golog.StopFastTimer()

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

func BenchmarkMultiLevels(b *testing.B) {
	golog.StartFastTimer()
	defer golog.StopFastTimer()

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

	for i := 0; i < b.N; i++ {
		Debugf("test")
		Infof("test")
		Warnf("test")
		Errorf("test")
		Critf("test")
	}
	l.Close()
}

func BenchmarkMultiLevelsParallel(b *testing.B) {
	golog.StartFastTimer()
	defer golog.StopFastTimer()

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

func BenchmarkBufferedFileLogger(b *testing.B) {
	golog.StartFastTimer()
	defer golog.StopFastTimer()

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

	for i := 0; i < b.N; i++ {
		Infof("test")
	}
	l.Close()
}

func BenchmarkBufferedFileLoggerParallel(b *testing.B) {
	golog.StartFastTimer()
	defer golog.StopFastTimer()

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
