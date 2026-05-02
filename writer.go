package golog

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultBufferSize = 1024 * 1024 * 4

	fileFlag      = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	fileMode      = 0644
	flushDuration = time.Millisecond * 100

	rotateByDateFormat = "-20060102.log"   // -YYYYmmdd.log
	rotateByHourFormat = "-2006010215.log" // -YYYYmmddHH.log
)

// RotateDuration specifies rotate duration type, should be either RotateByDate or RotateByHour.
type RotateDuration uint8

const (
	// RotateByDate set the log file to be rotated each day.
	RotateByDate RotateDuration = iota
	// RotateByHour set the log file to be rotated each hour.
	RotateByHour
)

// DiscardWriter is a WriteCloser which write everything to devNull
type DiscardWriter struct {
	io.Writer
}

// NewDiscardWriter creates a new DiscardWriter.
func NewDiscardWriter() *DiscardWriter {
	return &DiscardWriter{Writer: io.Discard}
}

// Close does nothing.
func (w *DiscardWriter) Close() error {
	return nil
}

// A ConsoleWriter is a writer which should not be actually closed.
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

// Close does nothing.
func (w *ConsoleWriter) Close() error {
	return nil
}

// NewFileWriter creates a FileWriter by its path.
func NewFileWriter(path string) (*os.File, error) {
	return os.OpenFile(path, fileFlag, fileMode)
}

type bufferedFileWriter struct {
	file       *os.File
	buffer     *bufio.Writer
	bufferSize uint32
}

type BufferedFileWriterOption func(*bufferedFileWriter)

// BufferSize sets the buffer size.
func BufferSize(size uint32) BufferedFileWriterOption {
	return func(w *bufferedFileWriter) {
		if size >= 1024 {
			w.bufferSize = size
		}
	}
}

// A BufferedFileWriter is a buffered file writer.
// The written bytes will be flushed to the log file every 0.1 second,
// or when reaching the buffer capacity (4 MB).
type BufferedFileWriter struct {
	bufferedFileWriter
	lock       sync.Mutex
	stopChan   chan struct{}
	updateChan chan struct{}
	updated    bool
}

// NewBufferedFileWriter creates a new BufferedFileWriter.
func NewBufferedFileWriter(path string, options ...BufferedFileWriterOption) (*BufferedFileWriter, error) {
	f, err := os.OpenFile(path, fileFlag, fileMode)
	if err != nil {
		return nil, err
	}
	w := &BufferedFileWriter{
		bufferedFileWriter: bufferedFileWriter{
			file:       f,
			bufferSize: defaultBufferSize,
		},
		updateChan: make(chan struct{}, 1),
		stopChan:   make(chan struct{}),
	}

	for _, option := range options {
		option(&w.bufferedFileWriter)
	}
	w.buffer = bufio.NewWriterSize(f, int(w.bufferSize))

	go w.schedule()
	return w, nil
}

func (w *BufferedFileWriter) schedule() {
	timer := time.NewTimer(0)
	for {
		select {
		case <-w.updateChan:
			// something has been written to the buffer, it can be flushed to the file later
			stopTimer(timer)
			timer.Reset(flushDuration)
		case <-w.stopChan:
			stopTimer(timer)
			return
		}

		select {
		case <-timer.C:
			var err error
			w.lock.Lock()
			if w.file != nil { // not closed
				w.updated = false
				err = w.buffer.Flush()
			}
			w.lock.Unlock()
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
	w.lock.Lock()
	if w.file == nil {
		w.lock.Unlock()
		return 0, os.ErrClosed
	}
	n, err = w.buffer.Write(p)
	if !w.updated && n > 0 && w.buffer.Buffered() > 0 { // checks w.updated to prevent notifying w.updateChan twice
		w.updated = true
		w.lock.Unlock()

		select { // ignores if blocked
		case w.updateChan <- struct{}{}:
		default:
		}
	} else {
		w.lock.Unlock()
	}
	return
}

// Close flushes the buffer, then closes the file writer.
func (w *BufferedFileWriter) Close() error {
	w.lock.Lock()
	if w.file == nil {
		w.lock.Unlock()
		return nil
	}

	close(w.stopChan)
	err := w.buffer.Flush()
	w.buffer = nil
	if err == nil {
		err = w.file.Close()
	} else {
		e := w.file.Close()
		if e != nil {
			logError(e)
		}
	}
	w.file = nil
	w.lock.Unlock()
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
func NewRotatingFileWriter(path string, maxSize uint64, backupCount uint8, options ...BufferedFileWriterOption) (*RotatingFileWriter, error) {
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
			bufferedFileWriter: bufferedFileWriter{
				file:       f,
				bufferSize: defaultBufferSize,
			},
			updateChan: make(chan struct{}, 1),
			stopChan:   make(chan struct{}),
		},
		path:        path,
		pos:         uint64(stat.Size()),
		maxSize:     maxSize,
		backupCount: backupCount,
	}

	for _, option := range options {
		option(&w.bufferedFileWriter)
	}
	w.buffer = bufio.NewWriterSize(f, int(w.bufferSize))

	go w.schedule()
	return &w, nil
}

// Write writes a byte slice to the buffer and rotates if reaching its maxSize.
func (w *RotatingFileWriter) Write(p []byte) (n int, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.file == nil {
		return 0, os.ErrClosed
	}

	n, err = w.buffer.Write(p)
	if n > 0 {
		w.pos += uint64(n)

		if w.pos >= w.maxSize {
			e := w.rotate()
			if e != nil {
				logError(e)
				if err == nil { // don't shadow Write() error
					err = e
				}
			}
			return // w.rotate() also calls w.buffer.Flush(), no need to notify w.updateChan
		}

		if !w.updated && w.buffer.Buffered() > 0 {
			w.updated = true

			select { // ignores if blocked
			case w.updateChan <- struct{}{}:
			default:
			}
		}
	}

	return
}

// rotate rotates the log file. It should be called within a lock block.
func (w *RotatingFileWriter) rotate() error {
	if w.file == nil { // was closed
		return os.ErrClosed
	}

	err := w.buffer.Flush()
	if err != nil {
		return err
	}

	err = w.file.Close()
	w.pos = 0
	if err != nil {
		w.file = nil
		w.buffer = nil
		return err
	}

	for i := w.backupCount; i > 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", w.path, i-1)
		newPath := fmt.Sprintf("%s.%d", w.path, i)
		e := os.Rename(oldPath, newPath)
		if e != nil && !os.IsNotExist(e) {
			logError(e)
		}
	}

	err = os.Rename(w.path, w.path+".1")
	if err != nil {
		w.file = nil
		w.buffer = nil
		return err
	}

	f, err := os.OpenFile(w.path, fileFlag, fileMode)
	if err != nil {
		w.file = nil
		w.buffer = nil
		return err
	}

	w.file = f
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
func NewTimedRotatingFileWriter(pathPrefix string, rotateDuration RotateDuration, backupCount uint8, options ...BufferedFileWriterOption) (*TimedRotatingFileWriter, error) {
	if backupCount == 0 {
		return nil, errors.New("backupCount cannot be 0")
	}

	f, err := openTimedRotatingFile(pathPrefix, rotateDuration)
	if err != nil {
		return nil, err
	}

	w := TimedRotatingFileWriter{
		BufferedFileWriter: BufferedFileWriter{
			bufferedFileWriter: bufferedFileWriter{
				file:       f,
				bufferSize: defaultBufferSize,
			},
			updateChan: make(chan struct{}, 1),
			stopChan:   make(chan struct{}),
		},
		pathPrefix:     pathPrefix,
		rotateDuration: rotateDuration,
		backupCount:    backupCount,
	}

	for _, option := range options {
		option(&w.bufferedFileWriter)
	}
	w.buffer = bufio.NewWriterSize(f, int(w.bufferSize))

	go w.schedule()
	return &w, nil
}

func (w *TimedRotatingFileWriter) schedule() {
	lock := &w.lock
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
				lock.Lock()
				var err error
				if w.file != nil { // not closed
					w.updated = false
					err = w.buffer.Flush()
				}
				lock.Unlock()
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
	w.lock.Lock()
	if w.file == nil { // was closed
		w.lock.Unlock()
		return nil // usually happens when program exits, should be ignored
	}

	err := w.buffer.Flush()
	if err != nil {
		w.lock.Unlock()
		return err
	}

	err = w.file.Close()
	if err != nil {
		w.lock.Unlock()
		return err
	}

	f, err := openTimedRotatingFile(w.pathPrefix, w.rotateDuration)
	if err != nil {
		w.buffer = nil
		w.file = nil
		w.lock.Unlock()
		return err
	}

	w.file = f
	w.buffer.Reset(f)

	duration := nextRotateDuration(w.rotateDuration)
	timer.Reset(duration)
	w.lock.Unlock()

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

	pathes = filterTimedRotatingFiles(pathes, w.pathPrefix, w.rotateDuration)
	count := len(pathes) - int(w.backupCount) - 1
	if count > 0 {
		var name string
		w.lock.Lock()
		if w.file != nil { // not closed
			name = w.file.Name()
		}
		w.lock.Unlock()
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

func filterTimedRotatingFiles(pathes []string, pathPrefix string, rotateDuration RotateDuration) []string {
	filtered := pathes[:0]
	for _, path := range pathes {
		if isTimedRotatingFile(path, pathPrefix, rotateDuration) {
			filtered = append(filtered, path)
		}
	}
	return filtered
}

func isTimedRotatingFile(path string, pathPrefix string, rotateDuration RotateDuration) bool {
	if !strings.HasPrefix(path, pathPrefix) {
		return false
	}

	suffix := path[len(pathPrefix):]
	switch rotateDuration {
	case RotateByDate:
		if len(suffix) != len(rotateByDateFormat) {
			return false
		}
		for i, c := range suffix {
			switch i {
			case 0:
				if c != '-' {
					return false
				}
			case 9:
				if c != '.' {
					return false
				}
			case 10:
				if c != 'l' {
					return false
				}
			case 11:
				if c != 'o' {
					return false
				}
			case 12:
				if c != 'g' {
					return false
				}
			default:
				if c < '0' || c > '9' {
					return false
				}
			}
		}
		return true
	case RotateByHour:
		if len(suffix) != len(rotateByHourFormat) {
			return false
		}
		for i, c := range suffix {
			switch i {
			case 0:
				if c != '-' {
					return false
				}
			case 11:
				if c != '.' {
					return false
				}
			case 12:
				if c != 'l' {
					return false
				}
			case 13:
				if c != 'o' {
					return false
				}
			case 14:
				if c != 'g' {
					return false
				}
			default:
				if c < '0' || c > '9' {
					return false
				}
			}
		}
		return true
	}
	return false
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

type ConcurrentFileWriter struct {
	bufferedFileWriter
	cpuCount    int
	locks       []sync.Mutex
	buffers     []*bytes.Buffer
	stopChan    chan struct{}
	stoppedChan chan struct{}
	closeOnce   sync.Once
	closeErr    error
	closed      uint32
}

// NewConcurrentFileWriter creates a new ConcurrentFileWriter.
func NewConcurrentFileWriter(path string, options ...BufferedFileWriterOption) (*ConcurrentFileWriter, error) {
	f, err := os.OpenFile(path, fileFlag, fileMode)
	if err != nil {
		return nil, err
	}

	cpuCount := runtime.GOMAXPROCS(0)

	w := &ConcurrentFileWriter{
		bufferedFileWriter: bufferedFileWriter{
			file:       f,
			bufferSize: defaultBufferSize,
		},
		cpuCount:    cpuCount,
		locks:       make([]sync.Mutex, cpuCount),
		buffers:     make([]*bytes.Buffer, cpuCount),
		stopChan:    make(chan struct{}),
		stoppedChan: make(chan struct{}, 1),
	}

	for _, option := range options {
		option(&w.bufferedFileWriter)
	}

	w.buffer = bufio.NewWriterSize(f, int(w.bufferSize))
	for i := 0; i < cpuCount; i++ {
		w.buffers[i] = bytes.NewBuffer(make([]byte, 0, w.bufferSize))
	}

	go w.schedule()
	return w, nil
}

func (w *ConcurrentFileWriter) schedule() {
	timer := time.NewTimer(flushDuration)
	for {
		select {
		case <-timer.C:
			for shard := 0; shard < w.cpuCount; shard++ {
				w.locks[shard].Lock()
				buffer := w.buffers[shard]
				if buffer.Len() > 0 {
					w.buffer.Write(buffer.Bytes())
					buffer.Reset()
				}
				w.locks[shard].Unlock()
			}

			if w.buffer.Buffered() > 0 {
				err := w.buffer.Flush()
				if err != nil {
					logError(err)
				}
			}

			timer.Reset(flushDuration)
		case <-w.stopChan:
			stopTimer(timer)
			w.stoppedChan <- struct{}{}
			return
		}
	}
}

// Write writes a byte slice to the buffer.
func (w *ConcurrentFileWriter) Write(p []byte) (n int, err error) {
	if atomic.LoadUint32(&w.closed) != 0 {
		return 0, os.ErrClosed
	}

	shard := runtime_procPin()
	runtime_procUnpin() // can't hold the lock for long

	w.locks[shard].Lock()
	defer w.locks[shard].Unlock()

	if atomic.LoadUint32(&w.closed) != 0 {
		return 0, os.ErrClosed
	}
	return w.buffers[shard].Write(p)
}

// Close flushes the buffer, then closes the file writer.
func (w *ConcurrentFileWriter) Close() error {
	w.closeOnce.Do(func() {
		atomic.StoreUint32(&w.closed, 1)
		close(w.stopChan) // stops schedule()
		<-w.stoppedChan   // waits for schedule() to finish, so the rest code can run without its flush loop

		for shard := 0; shard < w.cpuCount; shard++ {
			w.locks[shard].Lock()
			buffer := w.buffers[shard]
			if buffer.Len() > 0 {
				w.buffer.Write(buffer.Bytes())
				buffer.Reset()
			}
			w.locks[shard].Unlock()
		}

		if w.buffer.Buffered() > 0 {
			w.closeErr = w.buffer.Flush()
		}
		if w.closeErr == nil {
			w.closeErr = w.file.Close()
		} else {
			e := w.file.Close()
			if e != nil {
				logError(e)
			}
		}
		w.file = nil
	})
	return w.closeErr
}
