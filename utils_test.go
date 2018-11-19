package golog

import "testing"

func TestUint2Bytes(t *testing.T) {
	bs := string(uint2Bytes(0, 2))
	if bs != "00" {
		t.Error()
	}

	bs = string(uint2Bytes(60, 2))
	if bs != "60" {
		t.Error()
	}

	bs = string(uint2Bytes(1970, 4))
	if bs != "1970" {
		t.Error()
	}

	bs = string(uint2Bytes(2038, 4))
	if bs != "2038" {
		t.Error()
	}
}

func TestUint2Bytes2(t *testing.T) {
	bs := string(uint2Bytes2(0))
	if bs != "00" {
		t.Error()
	}

	bs = string(uint2Bytes2(60))
	if bs != "60" {
		t.Error()
	}
}

func TestUint2Bytes4(t *testing.T) {
	bs := string(uint2Bytes4(1970))
	if bs != "1970" {
		t.Error()
	}

	bs = string(uint2Bytes4(2038))
	if bs != "2038" {
		t.Error()
	}
}
