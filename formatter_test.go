package golog

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

func TestParseFormat(t *testing.T) {
	if len(DefaultFormatter.formatParts) != 11 {
		t.Errorf("formatParts are %d", len(DefaultFormatter.formatParts))
	}

	part0, ok := DefaultFormatter.formatParts[0].(*ByteFormatPart)
	if !ok {
		t.Errorf("part0 is " + reflect.TypeOf(DefaultFormatter.formatParts[0]).String())
	}
	if part0.byte != '[' {
		t.Errorf("byte of part0 is %d", part0.byte)
	}

	_, ok = DefaultFormatter.formatParts[1].(*LevelFormatPart)
	if !ok {
		t.Errorf("part1 is " + reflect.TypeOf(DefaultFormatter.formatParts[1]).String())
	}

	part2, ok := DefaultFormatter.formatParts[2].(*ByteFormatPart)
	if !ok {
		t.Errorf("part2 is " + reflect.TypeOf(DefaultFormatter.formatParts[2]).String())
	}
	if part2.byte != ' ' {
		t.Errorf("byte of part2 is %d", part2.byte)
	}

	_, ok = DefaultFormatter.formatParts[3].(*DateFormatPart)
	if !ok {
		t.Errorf("part3 is " + reflect.TypeOf(DefaultFormatter.formatParts[3]).String())
	}

	part4, ok := DefaultFormatter.formatParts[4].(*ByteFormatPart)
	if !ok {
		t.Errorf("part4 is " + reflect.TypeOf(DefaultFormatter.formatParts[4]).String())
	}
	if part4.byte != ' ' {
		t.Errorf("byte of part4 is %d", part4.byte)
	}

	_, ok = DefaultFormatter.formatParts[5].(*TimeFormatPart)
	if !ok {
		t.Errorf("part5 is " + reflect.TypeOf(DefaultFormatter.formatParts[5]).String())
	}

	part6, ok := DefaultFormatter.formatParts[6].(*ByteFormatPart)
	if !ok {
		t.Errorf("part6 is " + reflect.TypeOf(DefaultFormatter.formatParts[6]).String())
	}
	if part6.byte != ' ' {
		t.Errorf("byte of part6 is %d", part6.byte)
	}

	_, ok = DefaultFormatter.formatParts[7].(*SourceFormatPart)
	if !ok {
		t.Errorf("part7 is " + reflect.TypeOf(DefaultFormatter.formatParts[7]).String())
	}

	part8, ok := DefaultFormatter.formatParts[8].(*BytesFormatPart)
	if !ok {
		t.Errorf("part8 is " + reflect.TypeOf(DefaultFormatter.formatParts[8]).String())
	}
	bs := part8.bytes
	if len(bs) != 2 || bs[0] != ']' || bs[1] != ' ' {
		t.Errorf("bytes of part8 is " + string(part8.bytes))
	}

	_, ok = DefaultFormatter.formatParts[9].(*MessageFormatPart)
	if !ok {
		t.Errorf("part9 is " + reflect.TypeOf(DefaultFormatter.formatParts[9]).String())
	}

	part10, ok := DefaultFormatter.formatParts[10].(*ByteFormatPart)
	if !ok {
		t.Errorf("part10 is " + reflect.TypeOf(DefaultFormatter.formatParts[10]).String())
	}
	if part10.byte != '\n' {
		t.Errorf("byte of part6 is %d", part6.byte)
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

	r.level = InfoLevel
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
		time: tm,
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
		time: tm,
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

	r.file = "/test/test.go"
	r.line = 10
	buf.Reset()
	part.Format(r, buf)
	bs = buf.String()
	if bs != "test:10" {
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

	r.file = "/test/test.go"
	r.line = 10
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

	r.message = "abc"
	buf.Reset()
	part.Format(r, buf)
	bs = buf.String()
	if bs != "abc" {
		t.Error()
	}

	r.message = "abc %d %d"
	r.args = []interface{}{1, 2}
	buf.Reset()
	part.Format(r, buf)
	bs = buf.String()
	if bs != "abc 1 2" {
		t.Error()
	}

	r.message = ""
	r.args = []interface{}{1, 2}
	buf.Reset()
	part.Format(r, buf)
	bs = buf.String()
	if bs != "1 2" {
		t.Error()
	}
}
