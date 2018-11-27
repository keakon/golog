package golog

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	infoPath := filepath.Join(os.TempDir(), "test_info.log")
	debugPath := filepath.Join(os.TempDir(), "test_debug.log")
	os.Remove(infoPath)
	os.Remove(debugPath)

	infoWriter, err := NewFileWriter(infoPath)
	if err != nil {
		t.Error(err)
	}
	debugWriter, err := NewFileWriter(debugPath)
	if err != nil {
		t.Error(err)
	}

	infoHandler := NewHandler(InfoLevel, DefaultFormatter)
	infoHandler.AddWriter(infoWriter)

	debugHandler := &Handler{
		formatter: DefaultFormatter,
	}
	debugHandler.AddWriter(debugWriter)

	l := NewLogger(DebugLevel)
	l.AddHandler(infoHandler)
	l.AddHandler(debugHandler)

	l.Debugf("test %d", 1)

	stat, err := os.Stat(infoPath)
	if err != nil {
		t.Error(err)
	}
	if stat.Size() != 0 {
		t.Errorf("file size are %d", stat.Size())
	}

	debugContent, err := ioutil.ReadFile(debugPath)
	if err != nil {
		t.Error(err)
	}
	size1 := len(debugContent)
	if size1 == 0 {
		t.Error("debug log is empty")
		return
	}

	l.Infof("test %d", 2)

	infoContent, err := ioutil.ReadFile(infoPath)
	if err != nil {
		t.Error(err)
	}

	parts := strings.Fields(string(infoContent))
	if len(parts) != 6 {
		t.Errorf("parts length are %d", len(parts))
	}
	if parts[0] != "[I" {
		t.Errorf("parts[0] is " + parts[0])
	}
	if len(parts[1]) != 10 {
		t.Errorf("parts[1] is " + parts[1])
	}
	if len(parts[2]) != 8 {
		t.Errorf("parts[2] is " + parts[2])
	}
	if !strings.HasPrefix(parts[3], "log_test:") {
		t.Errorf("parts[3] is " + parts[3])
	}
	if parts[4] != "test" {
		t.Errorf("parts[4] is " + parts[4])
	}
	if parts[5] != "2" {
		t.Errorf("parts[5] is " + parts[5])
	}

	debugContent, err = ioutil.ReadFile(debugPath)
	if err != nil {
		t.Error(err)
	}
	size2 := len(debugContent)
	if size2 == size1*2 {
		if !bytes.Equal(debugContent[size1:], infoContent) {
			t.Error("log contents are not equal")
		}
	} else {
		t.Errorf("debug log size are %d bytes", size2)
	}

	if !bytes.Equal(debugContent[size1:], infoContent) {
		t.Error("log contents are not equal")
	}
}

func TestAddHandler(t *testing.T) {
	w := NewDiscardWriter()

	dh := NewHandler(DebugLevel, DefaultFormatter)
	dh.AddWriter(w)

	ih := NewHandler(InfoLevel, DefaultFormatter)
	ih.AddWriter(w)

	wh := NewHandler(WarnLevel, DefaultFormatter)
	wh.AddWriter(w)

	eh := NewHandler(ErrorLevel, DefaultFormatter)
	eh.AddWriter(w)

	ch := NewHandler(CritLevel, DefaultFormatter)
	ch.AddWriter(w)

	l := NewLogger(InfoLevel)
	if l.IsEnabledFor(CritLevel) {
		t.Error("an empty logger should not be enabled for any level")
	}

	l.AddHandler(ch)
	if !l.IsEnabledFor(CritLevel) {
		t.Error("the logger is not enable for critical level")
	}
	if l.IsEnabledFor(ErrorLevel) {
		t.Error("the logger is enable for error level")
	}

	l.AddHandler(eh)
	if !l.IsEnabledFor(ErrorLevel) {
		t.Error("the logger is not enable for error level")
	}

	l.AddHandler(wh)
	if !l.IsEnabledFor(WarnLevel) {
		t.Error("the logger is not enable for warning level")
	}

	l.AddHandler(ih)
	if !l.IsEnabledFor(InfoLevel) {
		t.Error("the logger is not enable for info level")
	}

	l.AddHandler(dh)
	if l.IsEnabledFor(DebugLevel) {
		t.Error("info logger should not enable for debug level")
	}

	count := len(l.handlers)
	if count != 4 {
		t.Errorf("the logger has %d handlers", count)
	}

	for i := 0; i < count-1; i++ {
		if l.handlers[i].level > l.handlers[i+1].level {
			t.Errorf("handlers[%d].level > handlers[%d].level", i, i+1)
		}
	}
}

func BenchmarkBufferedFileLogger(b *testing.B) {
	path := filepath.Join(os.TempDir(), "test.log")
	os.Remove(path)
	w, err := NewBufferedFileWriter(path)
	if err != nil {
		b.Error(err)
	}
	h := NewHandler(InfoLevel, DefaultFormatter)
	h.AddWriter(w)
	l := Logger{}
	l.AddHandler(h)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Infof("test")
		}
	})
	l.Close()
}

func BenchmarkDiscardLogger(b *testing.B) {
	w := NewDiscardWriter()
	h := NewHandler(InfoLevel, DefaultFormatter)
	h.AddWriter(w)
	l := NewLogger(InfoLevel)
	l.AddHandler(h)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Infof("test")
		}
	})
	l.Close()
}

func BenchmarkMultiLevel(b *testing.B) {
	l := NewLogger(WarnLevel)
	w := NewDiscardWriter()
	dh := NewHandler(DebugLevel, DefaultFormatter)
	dh.AddWriter(w)
	ih := NewHandler(InfoLevel, DefaultFormatter)
	ih.AddWriter(w)
	wh := NewHandler(WarnLevel, DefaultFormatter)
	wh.AddWriter(w)
	eh := NewHandler(ErrorLevel, DefaultFormatter)
	eh.AddWriter(w)
	ch := NewHandler(CritLevel, DefaultFormatter)
	ch.AddWriter(w)

	l.AddHandler(dh)
	l.AddHandler(ih)
	l.AddHandler(wh)
	l.AddHandler(eh)
	l.AddHandler(ch)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Debugf("test")
			l.Infof("test")
			l.Warnf("test")
			l.Errorf("test")
			l.Critf("test")
		}
	})
	l.Close()
}
