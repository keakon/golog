package golog

import (
	"bytes"
	"runtime"
	"sync"
	"time"
)

const recordBufSize = 128

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

	now = time.Now
)

func uint2Bytes(x, size int) []byte {
	// x and size shoule be uint32
	result := make([]byte, size)
	for i := 0; i < size; i++ {
		r := x % 10
		result[size-i-1] = byte(r) + '0'
		x /= 10
	}
	return result
}

func uint2DynamicBytes(x int) []byte {
	// x shoule be uint32
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
	case x < 10000:
		size = 5
	case x < 10000:
		size = 6
	case x < 100000:
		size = 7
	case x < 1000000:
		size = 8
	case x < 10000000:
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

func uint2Bytes2(x int) []byte {
	// x shoule between 0 and 59
	return uintBytes2[x]
}

func uint2Bytes4(x int) []byte {
	// x shoule between 1970 and 2038
	return uintBytes4[x-1970]
}

func fastUint2DynamicBytes(x int) []byte {
	// x shoule be uint32
	size := 0
	switch {
	case x < 2:
		return []byte{byte(x) + '0'}
	case x <= 1000:
		return uintBytes[x-2]
	case x < 10000:
		size = 4
	case x < 10000:
		size = 5
	case x < 10000:
		size = 6
	case x < 100000:
		size = 7
	case x < 1000000:
		size = 8
	case x < 10000000:
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
		_, file, line, _ := runtime.Caller(1)
		internalLogger.Log(ErrorLevel, file, line, err.Error())
	}
}

func setNowFunc(nowFunc func() time.Time) {
	now = nowFunc
}
