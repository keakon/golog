package golog

import (
	"io"
	"os"
	"testing"
	"time"
)

func TestBufferedFileWriter(t *testing.T) {
	oldBufferSize := bufferSize
	bufferSize = 1024
	path := "/tmp/test.log"
	os.Remove(path)
	w, err := NewBufferdFileWriter(path)
	if err != nil {
		t.Error()
	}

	f, err := os.Open(path)
	if err != nil {
		t.Error()
	}
	info, err := f.Stat()
	if err != nil {
		t.Error()
	}
	if info.Size() != 0 {
		t.Error()
	}

	n, err := w.Write([]byte("test"))
	if err != nil || n != 4 {
		t.Error()
	}

	buf := make([]byte, bufferSize*2)
	n, err = f.Read(buf)
	if err != io.EOF || n != 0 {
		t.Error()
	}

	time.Sleep(flushDuration * 2)
	n, err = f.Read(buf)
	if err != nil || n != 4 || string(buf[:4]) != "test" {
		t.Error()
	}

	for i := 0; i < bufferSize; i++ {
		w.Write([]byte{'1'})
	}
	w.Write([]byte{'2'}) // writes over bufferSize cause flushing
	n, err = f.Read(buf)
	if err != nil || n != bufferSize || buf[bufferSize-1] != '1' || buf[bufferSize] != 0 {
		t.Error()
	}

	time.Sleep(flushDuration * 2)
	n, err = f.Read(buf)
	if err != nil || n != 1 || buf[0] != '2' || buf[1] != '1' {
		t.Error()
	}

	f.Close()
	w.Close()
	bufferSize = oldBufferSize
}

func TestRotatingFileWriter(t *testing.T) {
	dir := "/tmp/test/"
	path := dir + "test.log"
	err := os.RemoveAll(dir)
	if err != nil {
		t.Error()
	}
	err = os.Mkdir(dir, 0755)
	if err != nil {
		t.Error()
	}

	w, err := NewRotatingFileWriter(path, 128, 2)
	if err != nil {
		t.Error()
	} else {
		defer w.Close()
	}
	stat, err := os.Stat(path)
	if err != nil {
		t.Error()
	}
	if stat.Size() != 0 {
		t.Error()
	}

	bs := []byte("0123456789")
	for i := 0; i < 20; i++ {
		w.Write(bs)
	}

	stat, err = os.Stat(path)
	if err != nil {
		t.Error()
	}
	if stat.Size() != 0 {
		t.Error()
	}

	stat, err = os.Stat(path + ".1")
	if err != nil {
		t.Error()
	}
	if stat.Size() != 120 {
		t.Error()
	}

	_, err = os.Stat(path + ".2")
	if !os.IsNotExist(err) {
		t.Error()
	}

	time.Sleep(flushDuration * 2)
	stat, err = os.Stat(path)
	if err != nil {
		t.Error()
	}
	if stat.Size() != 80 {
		t.Error()
	}

	// second write
	for i := 0; i < 20; i++ {
		w.Write(bs)
	}

	stat, err = os.Stat(path)
	if err != nil {
		t.Error()
	}
	if stat.Size() != 0 {
		t.Error()
	}

	stat, err = os.Stat(path + ".1")
	if err != nil {
		t.Error()
	}
	if stat.Size() != 120 {
		t.Error()
	}

	stat, err = os.Stat(path + ".2")
	if stat.Size() != 120 {
		t.Error()
	}

	time.Sleep(flushDuration * 2)
	stat, err = os.Stat(path)
	if err != nil {
		t.Error()
	}
	if stat.Size() != 40 {
		t.Error()
	}
}

func TestTimedRotatingFileWriterByDate(t *testing.T) {
	dir := "/tmp/test/"
	pathPrefix := dir + "test"
	err := os.RemoveAll(dir)
	if err != nil {
		t.Error()
	}
	err = os.Mkdir(dir, 0755)
	if err != nil {
		t.Error()
	}

	tm := time.Date(2018, 11, 19, 16, 12, 34, 56, time.Local)
	setNowFunc(func() time.Time {
		return tm
	})

	oldNextRotateDuration := nextRotateDuration
	nextRotateDuration = func(rotateDuration RotateDuration) time.Duration {
		return flushDuration * 3
	}

	w, err := NewTimedRotatingFileWriter(pathPrefix, RotateByDate, 2)
	if err != nil {
		t.Error()
	}
	path := pathPrefix + "-20181119.log"
	stat, err := os.Stat(path)
	if err != nil {
		t.Error()
	}
	if stat.Size() != 0 {
		t.Error()
	}

	w.Write([]byte("123"))
	stat, err = os.Stat(path)
	if err != nil {
		t.Error()
	}
	if stat.Size() != 0 {
		t.Error()
	}

	tm = time.Date(2018, 11, 20, 16, 12, 34, 56, time.Local)
	time.Sleep(flushDuration * 2)
	stat, err = os.Stat(path)
	if err != nil {
		t.Error()
	}
	if stat.Size() != 3 {
		t.Error()
	}

	time.Sleep(flushDuration * 2)
	path = pathPrefix + "-20181120.log"
	stat, err = os.Stat(path)
	if err != nil {
		t.Error()
	}
	if stat.Size() != 0 {
		t.Error()
	}

	w.Write([]byte("4567"))
	tm = time.Date(2018, 11, 21, 16, 12, 34, 56, time.Local)
	time.Sleep(flushDuration * 4)
	stat, err = os.Stat(path)
	if err != nil {
		t.Error()
	}
	if stat.Size() != 4 {
		t.Error()
	}
	stat, err = os.Stat(pathPrefix + "-20181121.log")
	if err != nil {
		t.Error()
	}
	if stat.Size() != 0 {
		t.Error()
	}

	tm = time.Date(2018, 11, 22, 16, 12, 34, 56, time.Local)
	time.Sleep(flushDuration * 4)
	stat, err = os.Stat(pathPrefix + "-20181122.log")
	if err != nil {
		t.Error()
	}
	if stat.Size() != 0 {
		t.Error()
	}
	_, err = os.Stat(pathPrefix + "-20181119.log")
	if !os.IsNotExist(err) {
		t.Error()
	}

	w.Close()
	setNowFunc(time.Now)
	nextRotateDuration = oldNextRotateDuration
}

func TestTimedRotatingFileWriterByHour(t *testing.T) {
	dir := "/tmp/test/"
	pathPrefix := dir + "test"
	err := os.RemoveAll(dir)
	if err != nil {
		t.Error()
	}
	err = os.Mkdir(dir, 0755)
	if err != nil {
		t.Error()
	}

	tm := time.Date(2018, 11, 19, 16, 12, 34, 56, time.Local)
	setNowFunc(func() time.Time {
		return tm
	})

	oldNextRotateDuration := nextRotateDuration
	nextRotateDuration = func(rotateDuration RotateDuration) time.Duration {
		return flushDuration * 3
	}

	w, err := NewTimedRotatingFileWriter(pathPrefix, RotateByHour, 2)
	if err != nil {
		t.Error()
	}
	path := pathPrefix + "-2018111916.log"
	stat, err := os.Stat(path)
	if err != nil {
		t.Error(err)
	}
	if stat.Size() != 0 {
		t.Error()
	}

	w.Write([]byte("123"))
	stat, err = os.Stat(path)
	if err != nil {
		t.Error()
	}
	if stat.Size() != 0 {
		t.Error()
	}

	tm = time.Date(2018, 11, 19, 17, 12, 34, 56, time.Local)
	time.Sleep(flushDuration * 2)
	stat, err = os.Stat(path)
	if err != nil {
		t.Error()
	}
	if stat.Size() != 3 {
		t.Error()
	}

	time.Sleep(flushDuration * 2)
	path = pathPrefix + "-2018111917.log"
	stat, err = os.Stat(path)
	if err != nil {
		t.Error()
	}
	if stat.Size() != 0 {
		t.Error()
	}

	w.Write([]byte("4567"))
	tm = time.Date(2018, 11, 19, 18, 12, 34, 56, time.Local)
	time.Sleep(flushDuration * 4)
	stat, err = os.Stat(path)
	if err != nil {
		t.Error()
	}
	if stat.Size() != 4 {
		t.Error()
	}
	stat, err = os.Stat(pathPrefix + "-2018111918.log")
	if err != nil {
		t.Error()
	}
	if stat.Size() != 0 {
		t.Error()
	}

	tm = time.Date(2018, 11, 22, 16, 12, 34, 56, time.Local)
	time.Sleep(flushDuration * 4)
	stat, err = os.Stat(pathPrefix + "-2018112216.log")
	if err != nil {
		t.Error()
	}
	if stat.Size() != 0 {
		t.Error()
	}
	_, err = os.Stat(pathPrefix + "-2018111916.log")
	if !os.IsNotExist(err) {
		t.Error()
	}

	w.Close()
	setNowFunc(time.Now)
	nextRotateDuration = oldNextRotateDuration
}
