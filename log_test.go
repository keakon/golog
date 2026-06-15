package golog

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	fastTimer.start()
	defer fastTimer.stop()

	dir := t.TempDir()
	infoPath := filepath.Join(dir, "test_info.log")
	debugPath := filepath.Join(dir, "test_debug.log")

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
	defer l.Close()

	l.Debugf("test %d", 1)

	stat, err := os.Stat(infoPath)
	if err != nil {
		t.Error(err)
	}
	if stat.Size() != 0 {
		t.Errorf("file size are %d", stat.Size())
	}

	debugContent, err := os.ReadFile(debugPath)
	if err != nil {
		t.Error(err)
	}
	size1 := len(debugContent)
	if size1 == 0 {
		t.Error("debug log is empty")
	}

	l.Infof("test %d", 2)

	infoContent, err := os.ReadFile(infoPath)
	if err != nil {
		t.Error(err)
	}
	size2 := len(infoContent)
	if size2 != size1 {
		t.Error("the sizes of debug and info logs are not equal")
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

	debugContent, err = os.ReadFile(debugPath)
	if err != nil {
		t.Error(err)
	}
	size3 := len(debugContent)
	if size3 == size1*2 {
		if !bytes.Equal(debugContent[size1:], infoContent) {
			t.Error("log contents are not equal")
		}
	} else {
		t.Errorf("debug log size are %d bytes", size2)
	}

	if !bytes.Equal(debugContent[size1:], infoContent) {
		t.Error("log contents are not equal")
	}

	l.Debug(1)
	l.Info(1)
	l.Warn(1)
	l.Error(1)
	l.Crit(1)
	l.Warnf("1")
	l.Errorf("1")
	l.Critf("1")

	infoContent, err = os.ReadFile(infoPath)
	if err != nil {
		t.Error(err)
	}
	size4 := len(infoContent)
	if size4 <= size2 {
		t.Error("info log size not changed")
	}

	debugContent, err = os.ReadFile(debugPath)
	if err != nil {
		t.Error(err)
	}
	size5 := len(debugContent)
	if size5 <= size3 {
		t.Error("debug log size not changed")
	}
	if size5 <= size4 {
		t.Error("info log size is larger than debug log size")
	}
}

func TestGetMinLevel(t *testing.T) {
	l := NewLogger(InfoLevel)
	defer l.Close()
	if l.GetMinLevel() != disabledLevel {
		t.Errorf("GetMinLevel failed")
	}

	errorHandler := NewHandler(ErrorLevel, DefaultFormatter)
	l.AddHandler(errorHandler)
	if l.GetMinLevel() != ErrorLevel {
		t.Errorf("GetMinLevel failed")
	}

	debugHandler := NewHandler(DebugLevel, DefaultFormatter)
	l.AddHandler(debugHandler)
	if l.GetMinLevel() != InfoLevel {
		t.Errorf("GetMinLevel failed")
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
	if count != 5 {
		t.Errorf("the logger has %d handlers", count)
	}

	for i := 0; i < count-1; i++ {
		if l.handlers[i].level > l.handlers[i+1].level {
			t.Errorf("handlers[%d].level > handlers[%d].level", i, i+1)
		}
	}
}

func TestCloseLogger(t *testing.T) {
	l := &Logger{}
	l.Close()
	l.Close()

	l = NewStdoutLogger()
	h := l.handlers[0]
	w := h.writers[0].(*ConsoleWriter)
	l.Close()
	if len(l.handlers) > 0 {
		t.Error("closed logger is not empty")
	}
	if len(h.writers) > 0 {
		t.Error("closed handler is not empty")
	}
	if w.File == nil {
		t.Error("close logger closed console writer")
	}
	l.Close()
}

func TestNeedsCaller(t *testing.T) {
	// Formatter level: only %s and %S require the caller.
	if !DefaultFormatter.NeedsCaller() {
		t.Error("DefaultFormatter should need the caller")
	}
	if !ParseFormat("[%l %S] %m").NeedsCaller() {
		t.Error("a full-source format should need the caller")
	}
	if ParseFormat("[%l %D %T] %m").NeedsCaller() {
		t.Error("a format without a source directive should not need the caller")
	}

	// Logger level: needsCaller is the OR over its handlers' formatters.
	noSource := NewLogger(InfoLevel)
	noSource.AddHandler(NewHandler(InfoLevel, ParseFormat("[%l %D %T] %m")))
	if noSource.NeedsCaller() {
		t.Error("logger with only a no-source handler should not need the caller")
	}
	// Adding a source handler flips it on and it stays on.
	noSource.AddHandler(NewHandler(InfoLevel, DefaultFormatter))
	if !noSource.NeedsCaller() {
		t.Error("logger with a source handler should need the caller")
	}

	// A nil formatter is treated conservatively as needing the caller.
	nilFmt := NewLogger(InfoLevel)
	nilFmt.AddHandler(&Handler{level: InfoLevel})
	if !nilFmt.NeedsCaller() {
		t.Error("logger with a nil-formatter handler should need the caller")
	}

	// Functionally, a no-source logger still logs the message correctly; the
	// source token is simply absent.
	path := filepath.Join(t.TempDir(), "test_needscaller.log")
	w, err := NewFileWriter(path)
	if err != nil {
		t.Fatal(err)
	}
	h := NewHandler(InfoLevel, ParseFormat("[%l %D %T] %m"))
	h.AddWriter(w)
	l := NewLogger(InfoLevel)
	l.AddHandler(h)
	l.Infof("test")
	l.Close()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Fields(string(content))
	// "[I", date, time, "test" -> 4 fields, no "file:line".
	if len(parts) != 4 {
		t.Errorf("parts length are %d: %q", len(parts), string(content))
	}
	if parts[len(parts)-1] != "test" {
		t.Errorf("last field is %q", parts[len(parts)-1])
	}
	if strings.Contains(string(content), "log_test:") {
		t.Errorf("no-source format should not contain a source token: %q", string(content))
	}
}

func TestNewStdoutLogger(t *testing.T) {
	l := NewStdoutLogger()
	if l.IsEnabledFor(DebugLevel) {
		t.Error("stdout logger is enabled for debug level")
	}
	if !l.IsEnabledFor(InfoLevel) {
		t.Error("stdout logger is not enabled for info level")
	}
	if !l.IsEnabledFor(ErrorLevel) {
		t.Error("stdout logger is not enabled for error level")
	}
}
