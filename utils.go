package golog

import (
	"bytes"
	"runtime"
	"sync"
	"time"
	_ "unsafe"
)

const (
	recordBufSize   = 128
	dateTimeBufSize = 10 // length of date string
)

var (
	recordPool = sync.Pool{
		New: func() interface{} {
			return &Record{}
		},
	}

	bufPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, recordBufSize))
		},
	}

	uintBytes2 [60][]byte  // 0 - 59
	uintBytes4 [69][]byte  // 1970 - 2038
	uintBytes  [999][]byte // 2 - 1000

	frameCache sync.Map

	now = time.Now

	fastTimer = FastTimer{}
)

func init() {
	for i := 0; i < 60; i++ { // hour / minute / second is between 0 and 59
		uintBytes2[i] = uint2Bytes(i, 2)
	}
	for i := 0; i < 69; i++ { // year is between 1970 and 2038
		uintBytes4[i] = uint2Bytes(1970+i, 4)
	}
	for i := 0; i < 999; i++ { // source code line number is usually between 2 and 1000
		uintBytes[i] = uint2DynamicBytes(i + 2)
	}
}

func uint2Bytes(x, size int) []byte {
	// x and size should be uint32
	result := make([]byte, size)
	for i := 0; i < size; i++ {
		r := x % 10
		result[size-i-1] = byte(r) + '0'
		x /= 10
	}
	return result
}

func uint2DynamicBytes(x int) []byte {
	// x should be uint32
	size := 0
	switch {
	case x < 10:
		return []byte{byte(x) + '0'}
	case x < 100:
		size = 2
	case x < 1000:
		size = 3
	case x < 10000:
		size = 4
	case x < 100000:
		size = 5
	case x < 1000000:
		size = 6
	case x < 10000000:
		size = 7
	case x < 100000000:
		size = 8
	case x < 1000000000:
		size = 9
	default:
		size = 10
	}
	result := make([]byte, size)
	for i := 0; i < size; i++ {
		r := x % 10
		result[size-i-1] = byte(r) + '0'
		x /= 10
	}
	return result
}

func uint2Bytes2(x int) []byte {
	// x should between 0 and 59
	return uintBytes2[x]
}

func uint2Bytes4(x int) []byte {
	if x >= 1970 && x < 1970+len(uintBytes4) {
		return uintBytes4[x-1970]
	}
	return uint2Bytes(x, 4)
}

func fastUint2DynamicBytes(x int) []byte {
	// x should be uint32
	size := 0
	switch {
	case x < 2:
		return []byte{byte(x) + '0'}
	case x <= 1000:
		return uintBytes[x-2]
	case x < 10000:
		size = 4
	case x < 100000:
		size = 5
	case x < 1000000:
		size = 6
	case x < 10000000:
		size = 7
	case x < 100000000:
		size = 8
	case x < 1000000000:
		size = 9
	default:
		size = 10
	}
	result := make([]byte, size)
	for i := 0; i < size; i++ {
		r := x % 10
		result[size-i-1] = byte(r) + '0'
		x /= 10
	}
	return result
}

func stopTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}

func logError(err error) {
	if internalLogger != nil {
		file, line := Caller(1)
		internalLogger.Log(ErrorLevel, file, line, err.Error())
	}
}

func setNowFunc(nowFunc func() time.Time) {
	now = nowFunc
}

//go:noescape
//go:linkname callers runtime.callers
func callers(skip int, pcbuf []uintptr) int

// Caller caches the result for runtime.Caller().
// Inspired by https://zhuanlan.zhihu.com/p/403417640
func Caller(skip int) (file string, line int) {
	rpc := [1]uintptr{}
	n := callers(skip+1, rpc[:])
	if n < 1 {
		return
	}

	var frame runtime.Frame
	pc := rpc[0]
	if f, ok := frameCache.Load(pc); ok {
		frame = f.(runtime.Frame)
	} else {
		frame, _ = runtime.CallersFrames([]uintptr{pc}).Next()
		frameCache.Store(pc, frame)
	}
	return frame.File, frame.Line
}

// FastTimer is not thread-safe for performance reason, but all the threads will notice its changes in a few milliseconds.
type FastTimer struct {
	date      string
	time      string
	stopChan  chan struct{}
	isRunning bool
	controlMu sync.Mutex
}

func (t *FastTimer) update(tm time.Time, buf *bytes.Buffer) {
	buf.Reset()
	year, mon, day := tm.Date()
	buf.Write(uint2Bytes4(year))
	buf.WriteByte('-')
	buf.Write(uint2Bytes2(int(mon)))
	buf.WriteByte('-')
	buf.Write(uint2Bytes2(day))
	t.date = buf.String()

	buf.Reset()
	hour, min, sec := tm.Clock()
	buf.Write(uint2Bytes2(hour))
	buf.WriteByte(':')
	buf.Write(uint2Bytes2(min))
	buf.WriteByte(':')
	buf.Write(uint2Bytes2(sec))
	t.time = buf.String()
}

func (t *FastTimer) start() {
	t.controlMu.Lock()
	if t.isRunning {
		t.controlMu.Unlock()
		return
	}

	buf := bytes.NewBuffer(make([]byte, 0, dateTimeBufSize))
	t.update(now(), buf)
	t.isRunning = true
	t.stopChan = make(chan struct{})
	stopChan := t.stopChan
	t.controlMu.Unlock()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case tm := <-ticker.C:
				t.update(tm, buf)
			case <-stopChan:
				return
			}
		}
	}()
}

func (t *FastTimer) stop() {
	t.controlMu.Lock()
	if !t.isRunning {
		t.controlMu.Unlock()
		return
	}
	close(t.stopChan)
	t.isRunning = false
	t.controlMu.Unlock()
}

// StartFastTimer starts the fastTimer.
func StartFastTimer() {
	fastTimer.start()
}

// StopFastTimer stops the fastTimer.
func StopFastTimer() {
	fastTimer.stop()
}

//go:noescape
//go:linkname runtime_procPin runtime.procPin
func runtime_procPin() int

//go:noescape
//go:linkname runtime_procUnpin runtime.procUnpin
func runtime_procUnpin()
