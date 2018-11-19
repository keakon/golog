package golog

import (
	"bytes"
	"runtime"
	"sync"
	"time"
)

const recordBufSize = 128

var (
	bufPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, recordBufSize))
		},
	}

	uintBytes2 = make([][]byte, 61) // 0 - 60
	uintBytes4 = make([][]byte, 69) // 1970 - 2038

	now = time.Now
)

func uint2Bytes(x uint, length int) []byte {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		remainder := x % 10
		result[length-i-1] = toASCIIByte(byte(remainder))
		x /= 10
	}
	return result
}

func toASCIIByte(x byte) byte {
	return x + '0'
}

func init() {
	for i := uint(0); i < 61; i++ { // hour / minute / second is between 0 and 60
		uintBytes2[i] = uint2Bytes(i, 2)
	}
	for i := uint(0); i < 69; i++ { // year is between 1970 and 2038
		uintBytes4[i] = uint2Bytes(1970+i, 4)
	}
}

func uint2Bytes2(x uint) []byte {
	return uintBytes2[x]
}

func uint2Bytes4(x uint) []byte {
	return uintBytes4[x-1970]
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
	_, file, line, _ := runtime.Caller(1)
	internalLogger.Log(ErrorLevel, file, line, err.Error())
}

func setNowFunc(nowFunc func() time.Time) {
	now = nowFunc
}
