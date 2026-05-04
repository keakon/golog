package golog

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

func TestParseFormat(t *testing.T) {
	if ParseFormat("") != nil {
		t.Error("ParseFormat empty string is not nil")
	}

	formatter := ParseFormat("%")
	if len(formatter.formatParts) != 1 {
		t.Error("ParseFormat % failed")
	}
	p, ok := formatter.formatParts[0].(*BytesFormatPart)
	if !ok {
		t.Error("ParseFormat % failed")
	}
	if string(p.bytes) != "%\n" {
		t.Error("ParseFormat % failed")
	}

	formatter = ParseFormat("%%")
	if len(formatter.formatParts) != 1 {
		t.Error("ParseFormat % failed")
	}
	p, ok = formatter.formatParts[0].(*BytesFormatPart)
	if !ok {
		t.Error("ParseFormat % failed")
	}
	if string(p.bytes) != "%\n" {
		t.Error("ParseFormat % failed")
	}

	formatter = ParseFormat("%a")
	if len(formatter.formatParts) != 1 {
		t.Error("ParseFormat %a failed")
	}
	p, ok = formatter.formatParts[0].(*BytesFormatPart)
	if !ok {
		t.Error("ParseFormat %a failed")
	}
	if string(p.bytes) != "%a\n" {
		t.Error("ParseFormat %a failed")
	}

	formatter = ParseFormat("% %")
	if len(formatter.formatParts) != 1 {
		t.Error("ParseFormat % % failed")
	}
	p, ok = formatter.formatParts[0].(*BytesFormatPart)
	if !ok {
		t.Error("ParseFormat % % failed")
	}
	if string(p.bytes) != "% %\n" {
		t.Error("ParseFormat % % failed")
	}

	formatter = ParseFormat("% %a")
	if len(formatter.formatParts) != 1 {
		t.Error("ParseFormat % %a failed")
	}
	p, ok = formatter.formatParts[0].(*BytesFormatPart)
	if !ok {
		t.Error("ParseFormat % %a failed")
	}
	if string(p.bytes) != "% %a\n" {
		t.Error("ParseFormat % %a failed")
	}

	formatter = ParseFormat("abc")
	if len(formatter.formatParts) != 1 {
		t.Error("ParseFormat abc failed")
	}
	p, ok = formatter.formatParts[0].(*BytesFormatPart)
	if !ok {
		t.Error("ParseFormat abc failed")
	}
	if string(p.bytes) != "abc\n" {
		t.Error("ParseFormat abc failed")
	}

	formatter = ParseFormat("%S")
	if len(formatter.formatParts) != 2 {
		t.Error("ParseFormat abc failed")
	}
	if _, ok = formatter.formatParts[0].(*FullSourceFormatPart); !ok {
		t.Error("ParseFormat %S failed")
	}
	if _, ok = formatter.formatParts[1].(*ByteFormatPart); !ok {
		t.Error("ParseFormat %S failed")
	}

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

// TestAppendBytesMergesPrecedingByteFormatPart guards against regressing the
// merge of a *ByteFormatPart with a following appendBytes call. The format
// "%lx%da" exercises this path: findParts emits LevelFormatPart, then
// appendByte('x') (single-char literal prefix), then appendBytes("%d") for the
// unknown directive 'd'. The expected result is a single BytesFormatPart
// containing "x%da\n", not a ByteFormatPart followed by a separate
// BytesFormatPart.
func TestAppendBytesMergesPrecedingByteFormatPart(t *testing.T) {
	formatter := ParseFormat("%lx%da")
	if formatter == nil {
		t.Fatal("ParseFormat returned nil")
	}
	if len(formatter.formatParts) != 2 {
		var got []string
		for _, p := range formatter.formatParts {
			got = append(got, reflect.TypeOf(p).String())
		}
		t.Fatalf("expected 2 parts, got %d: %v", len(formatter.formatParts), got)
	}
	if _, ok := formatter.formatParts[0].(*LevelFormatPart); !ok {
		t.Errorf("part0 is %s, expected *LevelFormatPart", reflect.TypeOf(formatter.formatParts[0]).String())
	}
	bp, ok := formatter.formatParts[1].(*BytesFormatPart)
	if !ok {
		t.Fatalf("part1 is %s, expected *BytesFormatPart", reflect.TypeOf(formatter.formatParts[1]).String())
	}
	if string(bp.bytes) != "x%da\n" {
		t.Errorf("part1 bytes = %q, expected %q", string(bp.bytes), "x%da\n")
	}

	// Sanity check: the formatted output is unchanged.
	r := &Record{level: InfoLevel}
	buf := &bytes.Buffer{}
	formatter.Format(r, buf)
	if buf.String() != "Ix%da\n" {
		t.Errorf("Format output = %q, expected %q", buf.String(), "Ix%da\n")
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

	r.level = Level(99)
	buf.Reset()
	part.Format(r, buf)
	bs = buf.String()
	if bs != "?" {
		t.Error()
	}
}

func TestTimeFormatPart(t *testing.T) {
	r := &Record{
		time: "16:12:34",
		date: "2018-11-19",
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
	r := &Record{
		time: "16:12:34",
		date: "2018-11-19",
	}
	buf := &bytes.Buffer{}
	part := DateFormatPart{}
	part.Format(r, buf)
	bs := buf.String()
	if bs != "2018-11-19" {
		t.Error()
	}

	r.date = ""
	r.tm = time.Date(2039, 1, 2, 0, 0, 0, 0, time.Local)
	buf.Reset()
	part.Format(r, buf)
	bs = buf.String()
	if bs != "2039-01-02" {
		t.Error(bs)
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
