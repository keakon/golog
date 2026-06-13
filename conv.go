package golog

import "bytes"

const (
	uintBytes2Count = 60   // 0..59 (hour/minute/second components)
	uintBytes4Count = 130  // years uintBytes4Base..uintBytes4Base+129
	uintBytesCount  = 1001 // 0..1000 (covers virtually all source line numbers)
	uintBytes4Base  = 1970

	// uintBytesBackingSize is the total width of the minimal-width decimal forms
	// of 0..1000: 10*1 + 90*2 + 900*3 + 1*4 = 2894 bytes.
	uintBytesBackingSize = 10*1 + 90*2 + 900*3 + 1*4
)

var (
	// uintBytes2 caches 2-digit decimal forms of values 0..59 (hour/minute/second components).
	uintBytes2 [uintBytes2Count][]byte
	// uintBytes4 caches 4-digit decimal forms of years 1970..2099 (covers all reasonable
	// log timestamps; queries outside this range fall back to uint2Bytes).
	uintBytes4 [uintBytes4Count][]byte
	// uintBytes caches decimal forms of values 0..1000, indexed directly by value.
	// Covers virtually all source line numbers in real code.
	uintBytes [uintBytesCount][]byte

	// The lookup tables above point into these contiguous backing arrays instead
	// of one heap allocation per entry. Packing the digits together improves cache
	// locality and removes ~1200 separately GC-scanned slices.
	uintBytes2Backing [uintBytes2Count * 2]byte
	uintBytes4Backing [uintBytes4Count * 4]byte
	uintBytesBacking  [uintBytesBackingSize]byte
)

func init() {
	for i := 0; i < uintBytes2Count; i++ {
		b := uintBytes2Backing[i*2 : i*2+2 : i*2+2]
		writeFixedUint(b, i)
		uintBytes2[i] = b
	}
	for i := 0; i < uintBytes4Count; i++ {
		b := uintBytes4Backing[i*4 : i*4+4 : i*4+4]
		writeFixedUint(b, uintBytes4Base+i)
		uintBytes4[i] = b
	}
	offset := 0
	for i := 0; i < uintBytesCount; i++ {
		w := decimalWidth(i)
		b := uintBytesBacking[offset : offset+w : offset+w]
		writeFixedUint(b, i)
		uintBytes[i] = b
		offset += w
	}
}

// writeFixedUint writes the decimal digits of x right-aligned into b, where
// len(b) is the fixed width. Caller must ensure len(b) is large enough to hold x.
func writeFixedUint(b []byte, x int) {
	for i := len(b) - 1; i >= 0; i-- {
		b[i] = byte(x%10) + '0'
		x /= 10
	}
}

// decimalWidth returns the number of decimal digits of x for 0 <= x <= 1000.
func decimalWidth(x int) int {
	switch {
	case x < 10:
		return 1
	case x < 100:
		return 2
	case x < 1000:
		return 3
	default:
		return 4
	}
}

// uint2Bytes encodes x as a fixed-width decimal byte slice of the given size.
// Caller must ensure size is large enough to hold x.
func uint2Bytes(x, size int) []byte {
	result := make([]byte, size)
	writeFixedUint(result, x)
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
