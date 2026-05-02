package golog

import (
	"bytes"
	"strconv"
	"testing"
)

func TestUint2Bytes(t *testing.T) {
	bs := string(uint2Bytes(0, 2))
	if bs != "00" {
		t.Errorf("result is " + bs)
	}

	bs = string(uint2Bytes(59, 2))
	if bs != "59" {
		t.Errorf("result is " + bs)
	}

	bs = string(uint2Bytes(1970, 4))
	if bs != "1970" {
		t.Errorf("result is " + bs)
	}

	bs = string(uint2Bytes(2038, 4))
	if bs != "2038" {
		t.Errorf("result is " + bs)
	}
}

func TestUint2Bytes2(t *testing.T) {
	bs := string(uint2Bytes2(0))
	if bs != "00" {
		t.Errorf("result is " + bs)
	}

	bs = string(uint2Bytes2(59))
	if bs != "59" {
		t.Errorf("result is " + bs)
	}
}

func TestUint2Bytes4(t *testing.T) {
	bs := string(uint2Bytes4(1970))
	if bs != "1970" {
		t.Errorf("result is " + bs)
	}

	bs = string(uint2Bytes4(2038))
	if bs != "2038" {
		t.Errorf("result is " + bs)
	}
}

func TestFastTimerStartStopIdempotent(t *testing.T) {
	StopFastTimer()
	StartFastTimer()
	StartFastTimer()
	StopFastTimer()
	StopFastTimer()
}

func TestUint2DynamicBytes(t *testing.T) {
	bs := string(uint2DynamicBytes(0))
	if bs != "0" {
		t.Errorf("result is " + bs)
	}

	bs = string(uint2DynamicBytes(59))
	if bs != "59" {
		t.Errorf("result is " + bs)
	}

	bs = string(uint2DynamicBytes(999))
	if bs != "999" {
		t.Errorf("result is " + bs)
	}

	bs = string(uint2DynamicBytes(1000))
	if bs != "1000" {
		t.Errorf("result is " + bs)
	}

	bs = string(uint2DynamicBytes(1970))
	if bs != "1970" {
		t.Errorf("result is " + bs)
	}

	bs = string(uint2DynamicBytes(2038))
	if bs != "2038" {
		t.Errorf("result is " + bs)
	}

	bs = string(uint2DynamicBytes(12345))
	if bs != "12345" {
		t.Errorf("result is " + bs)
	}

	bs = string(uint2DynamicBytes(123456))
	if bs != "123456" {
		t.Errorf("result is " + bs)
	}

	bs = string(uint2DynamicBytes(1234567))
	if bs != "1234567" {
		t.Errorf("result is " + bs)
	}

	bs = string(uint2DynamicBytes(12345678))
	if bs != "12345678" {
		t.Errorf("result is " + bs)
	}

	bs = string(uint2DynamicBytes(123456789))
	if bs != "123456789" {
		t.Errorf("result is " + bs)
	}

	bs = string(uint2DynamicBytes(1234567890))
	if bs != "1234567890" {
		t.Errorf("result is " + bs)
	}

	bs = string(uint2DynamicBytes(2<<31 - 1))
	if bs != "4294967295" {
		t.Errorf("result is " + bs)
	}
}

func TestWriteUintToBuf(t *testing.T) {
	tests := []struct {
		input  int
		expect string
	}{
		{0, "0"},
		{1, "1"},
		{2, "2"},
		{9, "9"},
		{10, "10"},
		{59, "59"},
		{99, "99"},
		{100, "100"},
		{999, "999"},
		{1000, "1000"},
		{1970, "1970"},
		{2038, "2038"},
		{12345, "12345"},
		{123456, "123456"},
		{1234567, "1234567"},
		{12345678, "12345678"},
		{123456789, "123456789"},
		{1234567890, "1234567890"},
		{2<<31 - 1, "4294967295"},
	}
	maxInt := int(^uint(0) >> 1)
	tests = append(tests, struct {
		input  int
		expect string
	}{maxInt, strconv.FormatInt(int64(maxInt), 10)})
	for _, tt := range tests {
		buf := &bytes.Buffer{}
		writeUintToBuf(buf, tt.input)
		if buf.String() != tt.expect {
			t.Errorf("writeUintToBuf(%d) = %s, want %s", tt.input, buf.String(), tt.expect)
		}
	}
}

func BenchmarkCaller(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Caller(1)
	}
}
