package golog

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const (
	fileFlag      = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	fileMode      = 0644
	flushDuration = time.Millisecond * 100

	RotateByDate RotateDuration = iota
	RotateByHour

	RotateByDateFormat = "-20060102.log"   // -YYYYmmdd.log
	RotateByHourFormat = "-2006010215.log" // -YYYYmmddHH.log
)

var bufferSize = 1024 * 1024 * 4

type RotateDuration uint8

type ConsoleWriter struct {
	*os.File // faster than io.Writer
}

func NewConsoleWriter(f *os.File) *ConsoleWriter {
	w := ConsoleWriter{
		File: f,
	}
	return &w
}

func NewStdoutWriter() *ConsoleWriter {
	return NewConsoleWriter(os.Stdout)
}

func NewStderrWriter() *ConsoleWriter {
	return NewConsoleWriter(os.Stderr)
}

func (w *ConsoleWriter) Close() error {
	w.File = nil
	return nil
}

func NewFileWriter(path string) (*os.File, error) {
	return os.OpenFile(path, fileFlag, fileMode)
}

type BufferedFileWriter struct {
	writer     *os.File
	buffer     *bufio.Writer
	locker     sync.Mutex
	updated    bool
	updateChan chan struct{}
	stopChan   chan struct{}
}

func NewBufferedFileWriter(path string) (*BufferedFileWriter, error) {
	f, err := os.OpenFile(path, fileFlag, fileMode)
	if err != nil {
		return nil, err
	}
	w := &BufferedFileWriter{
		writer:     f,
		buffer:     bufio.NewWriterSize(f, bufferSize),
		updateChan: make(chan struct{}, 1),
		stopChan:   make(chan struct{}, 1),
	}
	go w.schedule()
	return w, nil
}

func (w *BufferedFileWriter) schedule() {
	locker := &w.locker
	timer := time.NewTimer(0)
	bw := w.buffer
	for {
		select {
		case <-w.updateChan:
			stopTimer(timer)
			timer.Reset(flushDuration)
		case <-w.stopChan:
			stopTimer(timer)
			return
		}

		select {
		case <-timer.C:
			locker.Lock()
			w.updated = false
			err := bw.Flush()
			locker.Unlock()
			if err != nil {
				logError(err)
			}
		case <-w.stopChan:
			stopTimer(timer)
			return
		}
	}
}

func (w *BufferedFileWriter) Write(p []byte) (n int, err error) {
	w.locker.Lock()
	n, err = w.buffer.Write(p)
	if !w.updated && n > 0 {
		w.updated = true
		w.updateChan <- struct{}{}
	}
	w.locker.Unlock()
	return
}

func (w *BufferedFileWriter) Close() error {
	w.locker.Lock()
	err := w.buffer.Flush()
	w.buffer = nil
	if err == nil {
		err = w.writer.Close()
	} else {
		e := w.writer.Close()
		if e != nil {
			logError(e)
		}
	}
	w.writer = nil
	w.locker.Unlock()
	w.stopChan <- struct{}{}
	return err
}

type RotatingFileWriter struct {
	BufferedFileWriter
	path        string
	pos         uint64
	maxSize     uint64
	backupCount uint8
}

func NewRotatingFileWriter(path string, maxSize uint64, backupCount uint8) (*RotatingFileWriter, error) {
	if maxSize == 0 {
		return nil, errors.New("maxSize cannot be 0")
	}

	if backupCount == 0 {
		return nil, errors.New("backupCount cannot be 0")
	}

	f, err := os.OpenFile(path, fileFlag, fileMode)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		e := f.Close()
		if e != nil {
			logError(e)
		}
		return nil, err
	}

	w := RotatingFileWriter{
		BufferedFileWriter: BufferedFileWriter{
			writer:     f,
			buffer:     bufio.NewWriterSize(f, bufferSize),
			updateChan: make(chan struct{}, 1),
			stopChan:   make(chan struct{}, 1),
		},
		path:        path,
		pos:         uint64(stat.Size()),
		maxSize:     maxSize,
		backupCount: backupCount,
	}

	go w.schedule()
	return &w, nil
}

func (w *RotatingFileWriter) Write(p []byte) (n int, err error) {
	length := uint64(len(p))
	w.locker.Lock()
	defer w.locker.Unlock()

	if length >= w.maxSize {
		err = w.rotate()
		if err != nil {
			return
		}

		n, err = w.buffer.Write(p)
		if err != nil {
			w.pos = uint64(n)
			return
		}

		err = w.rotate()
	} else {
		pos := w.pos + length
		if pos > w.maxSize {
			err = w.rotate()
			if err != nil {
				return
			}
		}

		n, err = w.buffer.Write(p)
		if !w.updated && w.buffer.Buffered() > 0 {
			w.updated = true
			w.updateChan <- struct{}{}
		}
		w.pos += uint64(n)
	}

	return
}

func (w *RotatingFileWriter) rotate() error { // should be called within lock
	err := w.buffer.Flush()
	if err != nil {
		return err
	}

	err = w.writer.Close()
	w.pos = 0
	if err != nil {
		w.writer = nil
		w.buffer = nil
		return err
	}

	for i := w.backupCount; i > 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", w.path, i-1)
		newPath := fmt.Sprintf("%s.%d", w.path, i)
		os.Rename(oldPath, newPath) // ignore error
	}

	err = os.Rename(w.path, w.path+".1")
	if err != nil {
		w.writer = nil
		w.buffer = nil
		return err
	}

	f, err := os.OpenFile(w.path, fileFlag, fileMode)
	if err != nil {
		w.writer = nil
		w.buffer = nil
		return err
	}

	w.writer = f
	w.buffer.Reset(f)
	return nil
}

func openTimedRotatingFile(path string, rotateDuration RotateDuration) (*os.File, error) {
	var pathSuffix string
	t := now()
	switch rotateDuration {
	case RotateByDate:
		pathSuffix = t.Format(RotateByDateFormat)
	case RotateByHour:
		pathSuffix = t.Format(RotateByHourFormat)
	default:
		return nil, errors.New("invalid rotateDuration")
	}

	return os.OpenFile(path+pathSuffix, fileFlag, fileMode)
}

var nextRotateDuration = func(rotateDuration RotateDuration) time.Duration {
	now := now()
	var nextTime time.Time
	if rotateDuration == RotateByDate {
		nextTime = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	} else {
		nextTime = time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location())
	}
	return nextTime.Sub(now)
}

type TimedRotatingFileWriter struct {
	BufferedFileWriter
	pathPrefix     string
	rotateDuration RotateDuration
	backupCount    uint8
}

func NewTimedRotatingFileWriter(pathPrefix string, rotateDuration RotateDuration, backupCount uint8) (*TimedRotatingFileWriter, error) {
	if backupCount == 0 {
		return nil, errors.New("backupCount cannot be 0")
	}

	f, err := openTimedRotatingFile(pathPrefix, rotateDuration)
	if err != nil {
		return nil, err
	}

	w := TimedRotatingFileWriter{
		BufferedFileWriter: BufferedFileWriter{
			writer:     f,
			buffer:     bufio.NewWriterSize(f, bufferSize),
			updateChan: make(chan struct{}, 1),
			stopChan:   make(chan struct{}, 1),
		},
		pathPrefix:     pathPrefix,
		rotateDuration: rotateDuration,
		backupCount:    backupCount,
	}

	go w.schedule()
	return &w, nil
}

func (w *TimedRotatingFileWriter) schedule() {
	locker := &w.locker
	flushTimer := time.NewTimer(0)
	duration := nextRotateDuration(w.rotateDuration)
	rotateTimer := time.NewTimer(duration)
	bw := w.buffer

	for {
	updateLoop:
		for {
			select {
			case <-w.updateChan:
				stopTimer(flushTimer)
				flushTimer.Reset(flushDuration)
				break updateLoop
			case <-rotateTimer.C:
				err := w.rotate(rotateTimer)
				if err != nil {
					logError(err)
				}
			case <-w.stopChan:
				stopTimer(flushTimer)
				stopTimer(rotateTimer)
				return
			}
		}

	flushLoop:
		for {
			select {
			case <-flushTimer.C:
				locker.Lock()
				w.updated = false
				err := bw.Flush()
				locker.Unlock()
				if err != nil {
					logError(err)
				}
				break flushLoop
			case <-rotateTimer.C:
				err := w.rotate(rotateTimer)
				if err != nil {
					logError(err)
				}
			case <-w.stopChan:
				stopTimer(flushTimer)
				stopTimer(rotateTimer)
				return
			}
		}
	}
}

func (w *TimedRotatingFileWriter) rotate(timer *time.Timer) error {
	w.locker.Lock()
	err := w.buffer.Flush()
	if err != nil {
		w.locker.Unlock()
		return err
	}

	err = w.writer.Close()
	if err != nil {
		w.locker.Unlock()
		return err
	}

	f, err := openTimedRotatingFile(w.pathPrefix, w.rotateDuration)
	if err != nil {
		w.buffer = nil
		w.writer = nil
		w.locker.Unlock()
		return err
	}

	w.writer = f
	w.buffer.Reset(f)
	w.locker.Unlock()

	go w.purge()

	duration := nextRotateDuration(w.rotateDuration)
	timer.Reset(duration)
	return nil
}

func (w *TimedRotatingFileWriter) purge() {
	pathes, err := filepath.Glob(w.pathPrefix + "*")
	if err != nil {
		logError(err)
		return
	}

	count := len(pathes) - int(w.backupCount) - 1
	if count > 0 {
		name := w.writer.Name()
		sort.Strings(pathes)
		for i := 0; i < count; i++ {
			path := pathes[i]
			if path != name {
				err = os.Remove(path)
				if err != nil {
					logError(err)
				}
			}
		}
	}
}
