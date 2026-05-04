package golog

import "bytes"

var (
	// uintBytes2 caches 2-digit decimal forms of values 0..59 (hour/minute/second components).
	uintBytes2 [60][]byte
	// uintBytes4 caches 4-digit decimal forms of years 1970..2099 (covers all reasonable
	// log timestamps; queries outside this range fall back to uint2Bytes).
	uintBytes4 [130][]byte
	// uintBytes caches decimal forms of values 0..1000, indexed directly by value.
	// Covers virtually all source line numbers in real code.
	uintBytes [1001][]byte
)

const uintBytes4Base = 1970

func init() {
	for i := 0; i < len(uintBytes2); i++ {
		uintBytes2[i] = uint2Bytes(i, 2)
	}
	for i := 0; i < len(uintBytes4); i++ {
		uintBytes4[i] = uint2Bytes(uintBytes4Base+i, 4)
	}
	for i := 0; i < len(uintBytes); i++ {
		uintBytes[i] = uint2DynamicBytes(i)
	}
}

// uint2Bytes encodes x as a fixed-width decimal byte slice of the given size.
// Caller must ensure size is large enough to hold x.
func uint2Bytes(x, size int) []byte {
	result := make([]byte, size)
	for i := 0; i < size; i++ {
		result[size-i-1] = byte(x%10) + '0'
		x /= 10
	}
	return result
}

// uint2DynamicBytes encodes a non-negative integer as a minimal-width decimal byte slice.
// Used only at init time to populate uintBytes; runtime hot paths should use writeUintToBuf.
func uint2DynamicBytes(x int) []byte {
	if x < 10 {
		return []byte{byte(x) + '0'}
	}
	var size int
	switch {
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
		result[size-i-1] = byte(x%10) + '0'
		x /= 10
	}
	return result
}

// uint2Bytes2 returns the 2-digit decimal form of x. Caller must ensure 0 <= x < 60.
func uint2Bytes2(x int) []byte {
	return uintBytes2[x]
}

// uint2Bytes4 returns the 4-digit decimal form of x. Falls back to allocation for years
// outside the cached range.
func uint2Bytes4(x int) []byte {
	if x >= uintBytes4Base && x < uintBytes4Base+len(uintBytes4) {
		return uintBytes4[x-uintBytes4Base]
	}
	return uint2Bytes(x, 4)
}

// writeUintToBuf writes a non-negative integer to buf without heap allocation.
// For values 0..1000 (the vast majority of source line numbers) it uses the
// pre-computed lookup table; larger values are written digit-by-digit through
// a stack-allocated array.
//
// REQUIRES: x >= 0. Negative values fall through and produce no output.
func writeUintToBuf(buf *bytes.Buffer, x int) {
	if x >= 0 && x <= 1000 {
		buf.Write(uintBytes[x])
		return
	}
	// Slow path for x > 1000: compute digits in a stack buffer.
	// 20 bytes is enough for any int64 (max 19 digits) plus safety margin.
	var b [20]byte
	i := len(b)
	for x > 0 {
		i--
		b[i] = byte(x%10) + '0'
		x /= 10
	}
	buf.Write(b[i:])
}
