package golog

import (
	"bytes"
	"fmt"
	"path/filepath"
)

var (
	unknownFile      = []byte("???")
	DefaultFormatter = ParseFormat("[%l %d %t %s] %m")
)

type Formatter struct {
	formatParts []FormatPart
}

func ParseFormat(format string) (formatter *Formatter) {
	if format == "" {
		return
	}
	formatter = &Formatter{}
	formatter.findParts([]byte(format))
	formatter.appendByte('\n')
	return
}

func (f *Formatter) Format(r *Record, buf *bytes.Buffer) {
	for _, part := range f.formatParts {
		part.Format(r, buf)
	}
}

func (f *Formatter) findParts(format []byte) {
	length := len(format)
	index := bytes.IndexByte(format, '%')
	if index == -1 || index == length-1 {
		if length == 0 {
			return
		}
		if length == 1 {
			f.appendByte(format[0])
		} else {
			f.appendBytes(format)
		}
		return
	}

	if index > 1 {
		f.appendBytes(format[:index])
	} else if index == 1 {
		f.appendByte(format[0])
	}
	switch c := format[index+1]; c {
	case '%':
		f.appendByte('%')
	case 'l':
		f.formatParts = append(f.formatParts, &LevelFormatPart{})
	case 't':
		f.formatParts = append(f.formatParts, &TimeFormatPart{})
	case 'd':
		f.formatParts = append(f.formatParts, &DateFormatPart{})
	case 's':
		f.formatParts = append(f.formatParts, &SourceFormatPart{})
	case 'S':
		f.formatParts = append(f.formatParts, &FullSourceFormatPart{})
	case 'm':
		f.formatParts = append(f.formatParts, &MessageFormatPart{})
	default:
		f.appendBytes([]byte{'%', c})
	}
	f.findParts(format[index+2:])
	return
}

type FormatPart interface {
	Format(r *Record, buf *bytes.Buffer)
}

type ByteFormatPart struct {
	byte byte
}

func (p *ByteFormatPart) Format(r *Record, buf *bytes.Buffer) {
	buf.WriteByte(p.byte)
}

func (f *Formatter) appendByte(b byte) {
	parts := f.formatParts
	count := len(parts)
	if count == 0 {
		f.formatParts = append(parts, &ByteFormatPart{byte: b})
	} else {
		var p FormatPart
		lastPart := parts[count-1]
		switch lp := lastPart.(type) {
		case *ByteFormatPart:
			p = &BytesFormatPart{
				bytes: []byte{lp.byte, b},
			}
		case *BytesFormatPart:
			p = &BytesFormatPart{
				bytes: append(lp.bytes, b),
			}
		default:
			p = &ByteFormatPart{byte: b}
		}
		f.formatParts = append(parts, p)
	}
}

type BytesFormatPart struct {
	bytes []byte
}

func (p *BytesFormatPart) Format(r *Record, buf *bytes.Buffer) {
	buf.Write(p.bytes)
}

func (f *Formatter) appendBytes(bs []byte) {
	parts := f.formatParts
	count := len(parts)
	if count == 0 {
		f.formatParts = append(parts, &BytesFormatPart{bytes: bs})
	} else {
		var p FormatPart
		lastPart := parts[count-1]
		switch lp := lastPart.(type) {
		case *ByteFormatPart:
			p = &BytesFormatPart{
				bytes: append([]byte{lp.byte}, bs...),
			}
		case *BytesFormatPart:
			p = &BytesFormatPart{
				bytes: append(lp.bytes, bs...),
			}
		default:
			p = &BytesFormatPart{bytes: bs}
		}
		f.formatParts = append(parts, p)
	}
}

type LevelFormatPart struct{}

func (p *LevelFormatPart) Format(r *Record, buf *bytes.Buffer) {
	buf.WriteByte(levelNames[int(r.level)])
}

type TimeFormatPart struct{}

func (p *TimeFormatPart) Format(r *Record, buf *bytes.Buffer) {
	hour, min, sec := r.time.Clock()
	buf.Write(uint2Bytes2(hour))
	buf.WriteByte(':')
	buf.Write(uint2Bytes2(min))
	buf.WriteByte(':')
	buf.Write(uint2Bytes2(sec))
}

type DateFormatPart struct{}

func (p *DateFormatPart) Format(r *Record, buf *bytes.Buffer) {
	year, mon, day := r.time.Date()
	buf.Write(uint2Bytes4(year))
	buf.WriteByte('-')
	buf.Write(uint2Bytes2(int(mon)))
	buf.WriteByte('-')
	buf.Write(uint2Bytes2(day))
}

type SourceFormatPart struct{}

func (p *SourceFormatPart) Format(r *Record, buf *bytes.Buffer) {
	if r.line > 0 {
		buf.WriteString(filepath.Base(r.file))
		buf.WriteByte(':')
		buf.Write(fastUint2DynamicBytes(r.line))
	} else {
		buf.Write(unknownFile)
	}
}

type FullSourceFormatPart struct{}

func (p *FullSourceFormatPart) Format(r *Record, buf *bytes.Buffer) {
	if r.line > 0 {
		buf.WriteString(r.file)
		buf.WriteByte(':')
		buf.Write(fastUint2DynamicBytes(r.line))
	} else {
		buf.Write(unknownFile)
	}
}

type MessageFormatPart struct{}

func (p *MessageFormatPart) Format(r *Record, buf *bytes.Buffer) {
	msg := ""
	if len(r.args) > 0 {
		if r.message == "" {
			msg = fmt.Sprint(r.args...)
		} else {
			msg = fmt.Sprintf(r.message, r.args...)
		}
	} else {
		msg = r.message
	}
	if msg != "" {
		buf.WriteString(msg)
	}
}
