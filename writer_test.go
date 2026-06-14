package golog

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

const maxRetryCount = 10

func checkFileSize(t *testing.T, path string, size int64) {
	stat, err := os.Stat(path)
	if err != nil {
		t.Error(err)
	}

	if stat.Size() != size {
		t.Fatalf("file size are %d bytes", stat.Size())
	}
}

func checkFileSizeN(t *testing.T, path string, size int64) {
	for i := 0; i < maxRetryCount; i++ {
		time.Sleep(flushDuration)

		stat, err := os.Stat(path)
		if err != nil {
			if i == maxRetryCount-1 {
				t.Error(err)
			} else {
				continue
			}
		}

		if stat.Size() != size {
			if i == maxRetryCount-1 {
				t.Errorf("file size are %d bytes", stat.Size())
			} else {
				continue
			}
		} else {
			break
		}
	}
}

func TestMain(m *testing.M) {
	SetInternalLogger(NewStderrLogger())
	os.Exit(m.Run())
}

func TestConsoleWriterCloseIdempotent(t *testing.T) {
	w := NewConsoleWriter(os.Stdout)
	if err := w.Close(); err != nil {
		t.Error(err)
	}
	if err := w.Close(); err != nil {
		t.Error(err)
	}
	if w.File == nil {
		t.Error("ConsoleWriter.Close() closed its file")
	}
}

func TestDiscardWriterCloseIdempotent(t *testing.T) {
	w := NewDiscardWriter()
	if err := w.Close(); err != nil {
		t.Error(err)
	}
	if err := w.Close(); err != nil {
		t.Error(err)
	}
	if w.Writer == nil {
		t.Error("DiscardWriter.Close() closed its writer")
	}
}

func TestBufferedFileWriter(t *testing.T) {
	const bufferSize = 1024

	path := filepath.Join(os.TempDir(), "test.log")
	os.Remove(path)
	w, err := NewBufferedFileWriter(path, BufferSize(bufferSize))
	if err != nil {
		t.Error(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Error(err)
	}
	stat, err := f.Stat()
	if err != nil {
		t.Error(err)
	}
	if stat.Size() != 0 {
		t.Errorf("file size are %d bytes", stat.Size())
	}

	n, err := w.Write([]byte("test"))
	if err != nil {
		t.Error(err)
	}
	if n != 4 {
		t.Errorf("write %d bytes, expected 4", n)
	}

	buf := make([]byte, bufferSize*2)

	for i := 0; i < maxRetryCount; i++ {
		n, err = f.Read(buf)
		if err != nil {
			if i == maxRetryCount-1 {
				t.Error(err)
			} else {
				time.Sleep(flushDuration)
				continue
			}
		} else {
			break
		}
	}
	if n != 4 {
		t.Errorf("read %d bytes, expected 4", n)
	}
	bs := string(buf[:4])
	if bs != "test" {
		t.Error("read bytes are " + bs)
	}

	for i := 0; i < bufferSize; i++ {
		w.Write([]byte{'1'})
	}
	w.Write([]byte{'2'}) // writes over bufferSize cause flushing
	n, err = f.Read(buf)
	if err != nil {
		t.Error(err)
	}
	if n != bufferSize {
		t.Errorf("read %d bytes", n)
	}
	if buf[bufferSize-1] != '1' {
		t.Errorf("last byte is %d", buf[bufferSize-1])
	}
	if buf[bufferSize] != 0 {
		t.Errorf("next byte is %d", buf[bufferSize-1])
	}

	for i := 0; i < maxRetryCount; i++ {
		n, err = f.Read(buf)
		if err != nil {
			if i == maxRetryCount-1 {
				t.Error(err)
			} else {
				time.Sleep(flushDuration)
				continue
			}
		} else {
			break
		}
	}

	if n != 1 {
		t.Errorf("read %d bytes", n)
	}
	if buf[0] != '2' {
		t.Errorf("first byte is %d", buf[0])
	}
	if buf[1] != '1' {
		t.Errorf("next byte is %d", buf[1])
	}

	f.Close()
	if err := w.Close(); err != nil {
		t.Error(err)
	}
	if err := w.Close(); err != nil {
		t.Error(err)
	}
	if _, err := w.Write([]byte("closed")); !errors.Is(err, os.ErrClosed) {
		t.Errorf("Write() after Close() error is %v, expected %v", err, os.ErrClosed)
	}
}

func TestRotatingFileWriter(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "test")
	path := filepath.Join(dir, "test.log")
	err := os.RemoveAll(dir)
	if err != nil {
		t.Error(err)
	}
	err = os.Mkdir(dir, 0755)
	if err != nil {
		t.Error(err)
	}

	_, err = NewRotatingFileWriter(path, 0, 2)
	if err == nil {
		t.Errorf("NewRotatingFileWriter with maxSize 0 is invalid")
	}

	_, err = NewRotatingFileWriter(path, 128, 0)
	if err == nil {
		t.Errorf("NewRotatingFileWriter with backupCount 0 is invalid")
	}

	w, err := NewRotatingFileWriter(path, 128, 2)
	if err != nil {
		t.Error(err)
	}
	stat, err := os.Stat(path)
	if err != nil {
		t.Error(err)
	}
	if stat.Size() != 0 {
		t.Errorf("file size are %d bytes", stat.Size())
	}

	bs := []byte("0123456789")
	for i := 0; i < 20; i++ {
		w.Write(bs)
	}

	checkFileSize(t, path, 0)

	checkFileSize(t, path+".1", 130)

	_, err = os.Stat(path + ".2")
	if !os.IsNotExist(err) {
		t.Error(err)
	}

	checkFileSizeN(t, path, 70)

	// second write
	for i := 0; i < 20; i++ {
		w.Write(bs)
	}

	checkFileSize(t, path, 0)
	checkFileSize(t, path+".1", 130)
	checkFileSize(t, path+".2", 130)
	checkFileSizeN(t, path, 10)

	bs = make([]byte, 200)
	for i := 0; i < 200; i++ {
		bs[i] = '1'
	}
	w.Write(bs)

	checkFileSize(t, path, 0)
	checkFileSize(t, path+".1", 210)
	checkFileSize(t, path+".2", 130)
	checkFileSizeN(t, path, 0)

	w.Write(bs)

	checkFileSize(t, path, 0)
	checkFileSize(t, path+".1", 200)
	checkFileSize(t, path+".2", 210)
	checkFileSizeN(t, path, 0)

	if err := w.Close(); err != nil {
		t.Error(err)
	}
	if _, err := w.Write([]byte("closed")); !errors.Is(err, os.ErrClosed) {
		t.Errorf("Write() after Close() error is %v, expected %v", err, os.ErrClosed)
	}
}

func TestTimedRotatingFileWriterByDate(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "test")
	pathPrefix := filepath.Join(dir, "test")
	err := os.RemoveAll(dir)
	if err != nil {
		t.Error(err)
	}
	err = os.Mkdir(dir, 0755)
	if err != nil {
		t.Error(err)
	}

	tm := time.Date(2018, 11, 19, 16, 12, 34, 56, time.Local)
	var lock sync.RWMutex
	setNowFunc(func() time.Time {
		lock.RLock()
		now := tm
		lock.RUnlock()
		return now
	})
	var setNow = func(now time.Time) {
		lock.Lock()
		tm = now
		lock.Unlock()
	}

	oldNextRotateDuration := nextRotateDuration
	nextRotateDuration = func(rotateDuration RotateDuration) time.Duration {
		return flushDuration * 3
	}

	_, err = NewTimedRotatingFileWriter(pathPrefix, RotateByDate, 0)
	if err == nil {
		t.Errorf("NewTimedRotatingFileWriter with backupCount 0 is invalid")
	}

	w, err := NewTimedRotatingFileWriter(pathPrefix, RotateByDate, 2)
	if err != nil {
		t.Error(err)
	}
	path := pathPrefix + "-20181119.log"
	checkFileSize(t, path, 0)

	w.Write([]byte("123"))
	checkFileSize(t, path, 0)

	setNow(time.Date(2018, 11, 20, 16, 12, 34, 56, time.Local))
	time.Sleep(flushDuration * 2)
	checkFileSizeN(t, path, 3)

	time.Sleep(flushDuration * 2)
	path = pathPrefix + "-20181120.log"
	checkFileSizeN(t, path, 0)

	w.Write([]byte("4567"))
	setNow(time.Date(2018, 11, 21, 16, 12, 34, 56, time.Local))

	time.Sleep(flushDuration * 2)
	checkFileSizeN(t, path, 4)

	time.Sleep(flushDuration * 3)
	checkFileSizeN(t, path, 4)
	checkFileSizeN(t, pathPrefix+"-20181121.log", 0)

	setNow(time.Date(2018, 11, 22, 16, 12, 34, 56, time.Local))
	time.Sleep(flushDuration * 3)
	checkFileSizeN(t, pathPrefix+"-20181121.log", 0)
	_, err = os.Stat(pathPrefix + "-20181119.log")
	if !os.IsNotExist(err) {
		t.Error(err)
	}

	if err := w.Close(); err != nil {
		t.Error(err)
	}
	if _, err := w.Write([]byte("closed")); !errors.Is(err, os.ErrClosed) {
		t.Errorf("Write() after Close() error is %v, expected %v", err, os.ErrClosed)
	}
	setNowFunc(time.Now)
	nextRotateDuration = oldNextRotateDuration
}

func TestTimedRotatingFileWriterByHour(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "test")
	pathPrefix := filepath.Join(dir, "test")
	err := os.RemoveAll(dir)
	if err != nil {
		t.Error(err)
	}
	err = os.Mkdir(dir, 0755)
	if err != nil {
		t.Error(err)
	}

	tm := time.Date(2018, 11, 19, 16, 12, 34, 56, time.Local)
	var lock sync.RWMutex
	setNowFunc(func() time.Time {
		lock.RLock()
		now := tm
		lock.RUnlock()
		return now
	})
	var setNow = func(now time.Time) {
		lock.Lock()
		tm = now
		lock.Unlock()
	}

	oldNextRotateDuration := nextRotateDuration
	nextRotateDuration = func(rotateDuration RotateDuration) time.Duration {
		return flushDuration * 3
	}

	w, err := NewTimedRotatingFileWriter(pathPrefix, RotateByHour, 2)
	if err != nil {
		t.Error(err)
	}
	path := pathPrefix + "-2018111916.log"
	checkFileSize(t, path, 0)

	w.Write([]byte("123"))
	checkFileSize(t, path, 0)

	setNow(time.Date(2018, 11, 19, 17, 12, 34, 56, time.Local))
	time.Sleep(flushDuration * 3)
	checkFileSizeN(t, path, 3)

	time.Sleep(flushDuration * 3)
	path = pathPrefix + "-2018111917.log"
	checkFileSizeN(t, path, 0)

	w.Write([]byte("4567"))
	setNow(time.Date(2018, 11, 19, 18, 12, 34, 56, time.Local))
	time.Sleep(flushDuration * 3)
	checkFileSizeN(t, path, 4)
	checkFileSizeN(t, pathPrefix+"-2018111918.log", 0)

	setNow(time.Date(2018, 11, 22, 16, 12, 34, 56, time.Local))
	time.Sleep(flushDuration * 3)
	checkFileSizeN(t, pathPrefix+"-2018112216.log", 0)
	_, err = os.Stat(pathPrefix + "-2018111916.log")
	if !os.IsNotExist(err) {
		t.Error(err)
	}

	w.Close()
	setNowFunc(time.Now)
	nextRotateDuration = oldNextRotateDuration
}

type badWriter struct{}

func (w *badWriter) Write(p []byte) (n int, err error) {
	return 0, io.ErrShortWrite
}

func (w *badWriter) Close() error {
	return nil
}

func TestBadWriter(t *testing.T) {
	path := filepath.Join(os.TempDir(), "error.log")
	os.Remove(path)
	w, err := NewBufferedFileWriter(path)
	if err != nil {
		t.Error(err)
	}

	newLogger := NewLoggerWithWriter(w)
	oldLogger := internalLogger
	SetInternalLogger(newLogger)
	defer SetInternalLogger(oldLogger)

	l := NewLoggerWithWriter(&badWriter{})
	l.Log(InfoLevel, "", 0, "test")
	l.Close()

	time.Sleep(flushDuration * 2)

	content, err := os.ReadFile(path)
	if err != nil {
		t.Error(err)
	}
	size := len(content)
	if size == 0 {
		t.Error("log is empty")
		return
	}

	if !strings.Contains(string(content), io.ErrShortWrite.Error()) {
		t.Error("bad writer raised no error")
		return
	}
}

func TestTimedRotatingFileWriterPurgeKeepsUnrelatedFiles(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "test-purge")
	pathPrefix := filepath.Join(dir, "test")
	if err := os.RemoveAll(dir); err != nil {
		t.Error(err)
	}
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Error(err)
	}

	files := []string{
		pathPrefix + "-20181119.log",
		pathPrefix + "-20181120.log",
		pathPrefix + "-20181121.log",
		pathPrefix + "-20181122.log",
		pathPrefix + "-keep.txt",
		pathPrefix + "-201811.log",
		pathPrefix + "-201811221.log",
	}
	for _, file := range files {
		if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
			t.Error(err)
		}
	}

	w := &TimedRotatingFileWriter{
		pathPrefix:     pathPrefix,
		rotateDuration: RotateByDate,
		backupCount:    1,
	}
	w.purge()

	for _, file := range files[4:] {
		if _, err := os.Stat(file); err != nil {
			t.Errorf("%s should not be purged: %v", file, err)
		}
	}
}

func TestTimedRotatingFileWriterPurgeKeepsUnrelatedHourlyFiles(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "test-purge-hour")
	pathPrefix := filepath.Join(dir, "test")
	if err := os.RemoveAll(dir); err != nil {
		t.Error(err)
	}
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Error(err)
	}

	files := []string{
		pathPrefix + "-2018111916.log",
		pathPrefix + "-2018111917.log",
		pathPrefix + "-2018111918.log",
		pathPrefix + "-2018112216.log",
		pathPrefix + "-keep.txt",
		pathPrefix + "-20181122.log",
		pathPrefix + "-201811221.log",
		pathPrefix + "-201811221600.log",
	}
	for _, file := range files {
		if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
			t.Error(err)
		}
	}

	w := &TimedRotatingFileWriter{
		pathPrefix:     pathPrefix,
		rotateDuration: RotateByHour,
		backupCount:    1,
	}
	w.purge()

	for _, file := range files[4:] {
		if _, err := os.Stat(file); err != nil {
			t.Errorf("%s should not be purged: %v", file, err)
		}
	}
}

func TestNextRotateDuration(t *testing.T) {
	if nextRotateDuration(RotateByDate) > time.Hour*24 {
		t.Errorf("nextRotateDuration(RotateByDate) longer than 1 day")
	}
	if nextRotateDuration(RotateByHour) > time.Hour {
		t.Errorf("nextRotateDuration(RotateByHour) longer than 1 hour")
	}
}

func TestConcurrentFileWriter(t *testing.T) {
	path := filepath.Join(os.TempDir(), "test.log")
	os.Remove(path)
	w, err := NewConcurrentFileWriter(path, BufferSize(1024*1024))
	if err != nil {
		t.Error(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Error(err)
	}
	stat, err := f.Stat()
	if err != nil {
		t.Error(err)
	}
	if stat.Size() != 0 {
		t.Errorf("file size are %d bytes", stat.Size())
	}

	n, err := w.Write([]byte("test"))
	if err != nil {
		t.Error(err)
	}
	if n != 4 {
		t.Errorf("write %d bytes, expected 4", n)
	}

	buf := make([]byte, defaultBufferSize)
	for i := 0; i < maxRetryCount; i++ {
		n, err = f.Read(buf)
		if err != nil {
			if i == maxRetryCount-1 {
				t.Error(err)
			} else {
				time.Sleep(flushDuration)
				continue
			}
		} else {
			break
		}
	}
	if n != 4 {
		t.Errorf("read %d bytes, expected 4", n)
	}
	bs := string(buf[:4])
	if bs != "test" {
		t.Error("read bytes are " + bs)
	}

	var count = w.cpuCount
	if count < 4 {
		count = 4
	} else if count > 10 {
		count = 10
	}

	wg := sync.WaitGroup{}
	wg.Add(count)
	const writeCount = 10000
	var dataSize int
	for i := 0; i < count; i++ {
		data := []byte("test" + strconv.Itoa(i) + "\n")
		dataSize = len(data)
		go func() {
			for j := 0; j < writeCount; j++ {
				w.Write(data)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	if err := w.Close(); err != nil {
		t.Error(err)
	}

	for i := 0; i < maxRetryCount; i++ {
		time.Sleep(flushDuration)
		n, err = f.Read(buf)
		if err != nil {
			if i == maxRetryCount-1 {
				t.Error(err)
			} else {
				continue
			}
		} else {
			break
		}
	}
	if n != count*dataSize*writeCount {
		t.Fatalf("read %d bytes, expected %d bytes", n, count*dataSize*writeCount)
	}

	lines := bytes.Split(buf[:n], []byte{'\n'})
	if len(lines) != count*writeCount+1 {
		t.Fatalf("read %d lines, expected %d lines", len(lines), count*writeCount+1)
	}
	if len(lines[count*writeCount]) != 0 {
		t.Error("last part is not empty")
	}
	lines = lines[:count*writeCount]
	for i, line := range lines {
		if len(line) != dataSize-1 {
			t.Errorf("length of line %d is %d, expected %d", i, len(line), dataSize-1)
		}
	}

	f.Close()
	if err := w.Close(); err != nil {
		t.Error(err)
	}
	if err := w.Close(); err != nil {
		t.Error(err)
	}
	if _, err := w.Write([]byte("closed")); !errors.Is(err, os.ErrClosed) {
		t.Errorf("Write() after Close() error is %v, expected %v", err, os.ErrClosed)
	}
}
