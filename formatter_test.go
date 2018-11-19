package golog

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func TestParseFormat(t *testing.T) {
	fmt.Println("")
	if len(DefaultFormatter.formatParts) != 11 {
		t.Error()
	}

	part0, ok := DefaultFormatter.formatParts[0].(*ByteFormatPart)
	if !ok {
		t.Error()
	}
	if part0.Byte != '[' {
		t.Error()
	}

	_, ok = DefaultFormatter.formatParts[1].(*LevelFormatPart)
	if !ok {
		t.Error()
	}

	part2, ok := DefaultFormatter.formatParts[2].(*ByteFormatPart)
	if !ok {
		t.Error()
	}
	if part2.Byte != ' ' {
		t.Error()
	}

	_, ok = DefaultFormatter.formatParts[3].(*DateFormatPart)
	if !ok {
		t.Error()
	}

	part4, ok := DefaultFormatter.formatParts[4].(*ByteFormatPart)
	if !ok {
		t.Error()
	}
	if part4.Byte != ' ' {
		t.Error()
	}

	_, ok = DefaultFormatter.formatParts[5].(*TimeFormatPart)
	if !ok {
		t.Error()
	}

	part6, ok := DefaultFormatter.formatParts[6].(*ByteFormatPart)
	if !ok {
		t.Error()
	}
	if part6.Byte != ' ' {
		t.Error()
	}

	_, ok = DefaultFormatter.formatParts[7].(*SourceFormatPart)
	if !ok {
		t.Error()
	}

	part8, ok := DefaultFormatter.formatParts[8].(*BytesFormatPart)
	if !ok {
		t.Error()
	}
	bs := part8.Bytes
	if len(bs) != 2 || bs[0] != ']' || bs[1] != ' ' {
		t.Error()
	}

	_, ok = DefaultFormatter.formatParts[9].(*MessageFormatPart)
	if !ok {
		t.Error()
	}

	part10, ok := DefaultFormatter.formatParts[10].(*ByteFormatPart)
	if !ok {
		t.Error()
	}
	if part10.Byte != '\n' {
		t.Error()
	}
}

func TestByteFormatPart(t *testing.T) {
	buf := &bytes.Buffer{}
	part := ByteFormatPart{'a'}
	part.Format(nil, buf)
	bs := buf.String()
	if bs != "a" {
		t.Error()
	}
}

func TestBytesFormatPart(t *testing.T) {
	buf := &bytes.Buffer{}
	part := BytesFormatPart{[]byte("abc")}
	part.Format(nil, buf)
	bs := buf.String()
	if bs != "abc" {
		t.Error()
	}
}

func TestLevelFormatPart(t *testing.T) {
	r := &Record{}
	buf := &bytes.Buffer{}
	part := LevelFormatPart{}
	part.Format(r, buf)
	bs := buf.String()
	if bs != "D" {
		t.Error()
	}

	r.Level = InfoLevel
	buf.Reset()
	part.Format(r, buf)
	bs = buf.String()
	if bs != "I" {
		t.Error()
	}
}

func TestTimeFormatPart(t *testing.T) {
	tm := time.Date(2018, 11, 19, 16, 12, 34, 56, time.Local)
	r := &Record{
		Time: tm,
	}
	buf := &bytes.Buffer{}
	part := TimeFormatPart{}
	part.Format(r, buf)
	bs := buf.String()
	if bs != "16:12:34" {
		t.Error()
	}
}

func TestDateFormatPart(t *testing.T) {
	tm := time.Date(2018, 11, 19, 16, 12, 34, 56, time.Local)
	r := &Record{
		Time: tm,
	}
	buf := &bytes.Buffer{}
	part := DateFormatPart{}
	part.Format(r, buf)
	bs := buf.String()
	if bs != "2018-11-19" {
		t.Error()
	}
}

func TestSourceFormatPart(t *testing.T) {
	r := &Record{}
	buf := &bytes.Buffer{}
	part := SourceFormatPart{}
	part.Format(r, buf)
	bs := buf.String()
	if bs != string(unknownFile) {
		t.Error()
	}

	r.File = "/test/test.go"
	r.Line = 10
	buf.Reset()
	part.Format(r, buf)
	bs = buf.String()
	if bs != "test.go:10" {
		t.Error()
	}
}

func TestFullSourceFormatPart(t *testing.T) {
	r := &Record{}
	buf := &bytes.Buffer{}
	part := FullSourceFormatPart{}
	part.Format(r, buf)
	bs := buf.String()
	if bs != string(unknownFile) {
		t.Error()
	}

	r.File = "/test/test.go"
	r.Line = 10
	buf.Reset()
	part.Format(r, buf)
	bs = buf.String()
	if bs != "/test/test.go:10" {
		t.Error()
	}
}

func TestMessageFormatPart(t *testing.T) {
	r := &Record{}
	buf := &bytes.Buffer{}
	part := MessageFormatPart{}
	part.Format(r, buf)
	bs := buf.String()
	if bs != "" {
		t.Error()
	}

	r.Message = "abc"
	buf.Reset()
	part.Format(r, buf)
	bs = buf.String()
	if bs != "abc" {
		t.Error()
	}

	r.Message = "abc %d %d"
	r.Args = []interface{}{1, 2}
	buf.Reset()
	part.Format(r, buf)
	bs = buf.String()
	if bs != "abc 1 2" {
		t.Error()
	}

	r.Message = ""
	r.Args = []interface{}{1, 2}
	buf.Reset()
	part.Format(r, buf)
	bs = buf.String()
	if bs != "1 2" {
		t.Error()
	}
}
