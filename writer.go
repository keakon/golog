package golog

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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

	rotateByDateFormat = "-20060102.log"   // -YYYYmmdd.log
	rotateByHourFormat = "-2006010215.log" // -YYYYmmddHH.log
)

const (
	// RotateByDate set the log file to be rotated each day.
	RotateByDate RotateDuration = iota
	// RotateByHour set the log file to be rotated each hour.
	RotateByHour
)

var bufferSize = 1024 * 1024 * 4

// RotateDuration specifies rotate duration type, should be either RotateByDate or RotateByHour.
type RotateDuration uint8

// DiscardWriter is a WriteCloser which write everything to devNull
type DiscardWriter struct {
	io.Writer
}

// NewDiscardWriter creates a new ConsoleWriter.
func NewDiscardWriter() *DiscardWriter {
	return &DiscardWriter{Writer: ioutil.Discard}
}

// Close sets its Writer to nil.
func (w *DiscardWriter) Close() error {
	w.Writer = nil
	return nil
}

// A ConsoleWriter is a writer which should not be acturelly closed.
type ConsoleWriter struct {
	*os.File // faster than io.Writer
}

// NewConsoleWriter creates a new ConsoleWriter.
func NewConsoleWriter(f *os.File) *ConsoleWriter {
	return &ConsoleWriter{File: f}
}

// NewStdoutWriter creates a new stdout writer.
func NewStdoutWriter() *ConsoleWriter {
	return NewConsoleWriter(os.Stdout)
}

// NewStderrWriter creates a new stderr writer.
func NewStderrWriter() *ConsoleWriter {
	return NewConsoleWriter(os.Stderr)
}

// Close sets its File to nil.
func (w *ConsoleWriter) Close() error {
	w.File = nil
	return nil
}

// NewFileWriter creates a FileWriter by its path.
func NewFileWriter(path string) (*os.File, error) {
	return os.OpenFile(path, fileFlag, fileMode)
}

// A BufferedFileWriter is a buffered file writer.
// The written bytes will be flushed to the log file every 0.1 second,
// or when reaching the buffer capacity (4 MB).
type BufferedFileWriter struct {
	writer     *os.File
	buffer     *bufio.Writer
	locker     sync.Mutex
	updateChan chan struct{}
	stopChan   chan struct{}
	updated    bool
}

// NewBufferedFileWriter creates a new BufferedFileWriter.
func NewBufferedFileWriter(path string) (*BufferedFileWriter, error) {
	f, err := os.OpenFile(path, fileFlag, fileMode)
	if err != nil {
		return nil, err
	}
	w := &BufferedFileWriter{
		writer:     f,
		buffer:     bufio.NewWriterSize(f, bufferSize),
		updateChan: make(chan struct{}, 1),
		stopChan:   make(chan struct{}),
	}
	go w.schedule()
	return w, nil
}

func (w *BufferedFileWriter) schedule() {
	timer := time.NewTimer(0)
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
			w.locker.Lock()
			var err error
			if w.writer != nil { // not closed
				w.updated = false
				err = w.buffer.Flush()
			}
			w.locker.Unlock()
			if err != nil {
				logError(err)
			}
		case <-w.stopChan:
			stopTimer(timer)
			return
		}
	}
}

// Write writes a byte slice to the buffer.
func (w *BufferedFileWriter) Write(p []byte) (n int, err error) {
	w.locker.Lock()
	n, err = w.buffer.Write(p)
	if !w.updated && n > 0 && w.buffer.Buffered() > 0 {
		w.updated = true
		w.updateChan <- struct{}{}
	}
	w.locker.Unlock()
	return
}

// Close flushes the buffer, then closes the file writer.
func (w *BufferedFileWriter) Close() error {
	close(w.stopChan)
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
	return err
}

// A RotatingFileWriter is a buffered file writer which will rotate before reaching its maxSize.
// An exception is when a record is larger than maxSize, it won't be separated into 2 files.
// It keeps at most backupCount backups.
type RotatingFileWriter struct {
	BufferedFileWriter
	path        string
	pos         uint64
	maxSize     uint64
	backupCount uint8
}

// NewRotatingFileWriter creates a new RotatingFileWriter.
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
			stopChan:   make(chan struct{}),
		},
		path:        path,
		pos:         uint64(stat.Size()),
		maxSize:     maxSize,
		backupCount: backupCount,
	}

	go w.schedule()
	return &w, nil
}

// Write writes a byte slice to the buffer and rotates if reaching its maxSize.
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
		if n > 0 {
			w.pos += uint64(n)
			if !w.updated && w.buffer.Buffered() > 0 {
				w.updated = true
				w.updateChan <- struct{}{}
			}
		}
	}

	return
}

// rotate rotates the log file. It should be called within a lock block.
func (w *RotatingFileWriter) rotate() error {
	if w.writer == nil { // was closed
		return os.ErrClosed
	}

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

// A TimedRotatingFileWriter is a buffered file writer which will rotate by time.
// Its rotateDuration can be either RotateByDate or RotateByHour.
// It keeps at most backupCount backups.
type TimedRotatingFileWriter struct {
	BufferedFileWriter
	pathPrefix     string
	rotateDuration RotateDuration
	backupCount    uint8
}

// NewTimedRotatingFileWriter creates a new TimedRotatingFileWriter.
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
			stopChan:   make(chan struct{}),
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
				var err error
				if w.writer != nil { // not closed
					w.updated = false
					err = w.buffer.Flush()
				}
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

// rotate rotates the log file.
func (w *TimedRotatingFileWriter) rotate(timer *time.Timer) error {
	w.locker.Lock()
	if w.writer == nil { // was closed
		w.locker.Unlock()
		return os.ErrClosed
	}

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

	duration := nextRotateDuration(w.rotateDuration)
	timer.Reset(duration)
	w.locker.Unlock()

	go w.purge()
	return nil
}

// purge removes the outdated backups.
func (w *TimedRotatingFileWriter) purge() {
	pathes, err := filepath.Glob(w.pathPrefix + "*")
	if err != nil {
		logError(err)
		return
	}

	count := len(pathes) - int(w.backupCount) - 1
	if count > 0 {
		var name string
		w.locker.Lock()
		if w.writer != nil { // not closed
			name = w.writer.Name()
		}
		w.locker.Unlock()
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

// openTimedRotatingFile opens a log file for TimedRotatingFileWriter
func openTimedRotatingFile(path string, rotateDuration RotateDuration) (*os.File, error) {
	var pathSuffix string
	t := now()
	switch rotateDuration {
	case RotateByDate:
		pathSuffix = t.Format(rotateByDateFormat)
	case RotateByHour:
		pathSuffix = t.Format(rotateByHourFormat)
	default:
		return nil, errors.New("invalid rotateDuration")
	}

	return os.OpenFile(path+pathSuffix, fileFlag, fileMode)
}

// nextRotateDuration returns the next rotate duration for the rotateTimer.
// It is defined as a variable in order to mock it in the unit testing.
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
