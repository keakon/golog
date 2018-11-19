package golog

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strconv"
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
	Byte byte
}

func (p *ByteFormatPart) Format(r *Record, buf *bytes.Buffer) {
	buf.WriteByte(p.Byte)
}

func (f *Formatter) appendByte(b byte) {
	parts := f.formatParts
	count := len(parts)
	if count == 0 {
		f.formatParts = append(parts, &ByteFormatPart{Byte: b})
	} else {
		var p FormatPart
		lastPart := parts[count-1]
		switch lp := lastPart.(type) {
		case *ByteFormatPart:
			p = &BytesFormatPart{
				Bytes: []byte{lp.Byte, b},
			}
		case *BytesFormatPart:
			p = &BytesFormatPart{
				Bytes: append(lp.Bytes, b),
			}
		default:
			p = &ByteFormatPart{Byte: b}
		}
		f.formatParts = append(parts, p)
	}
}

type BytesFormatPart struct {
	Bytes []byte
}

func (p *BytesFormatPart) Format(r *Record, buf *bytes.Buffer) {
	buf.Write(p.Bytes)
}

func (f *Formatter) appendBytes(bs []byte) {
	parts := f.formatParts
	count := len(parts)
	if count == 0 {
		f.formatParts = append(parts, &BytesFormatPart{Bytes: bs})
	} else {
		var p FormatPart
		lastPart := parts[count-1]
		switch lp := lastPart.(type) {
		case *ByteFormatPart:
			p = &BytesFormatPart{
				Bytes: append([]byte{lp.Byte}, bs...),
			}
		case *BytesFormatPart:
			p = &BytesFormatPart{
				Bytes: append(lp.Bytes, bs...),
			}
		default:
			p = &BytesFormatPart{Bytes: bs}
		}
		f.formatParts = append(parts, p)
	}
}

type LevelFormatPart struct{}

func (p *LevelFormatPart) Format(r *Record, buf *bytes.Buffer) {
	buf.WriteByte(levelNames[int(r.Level)])
}

type TimeFormatPart struct{}

func (p *TimeFormatPart) Format(r *Record, buf *bytes.Buffer) {
	hour, min, sec := r.Time.Clock()
	buf.Write(uint2Bytes2(uint(hour)))
	buf.WriteByte(':')
	buf.Write(uint2Bytes2(uint(min)))
	buf.WriteByte(':')
	buf.Write(uint2Bytes2(uint(sec)))
}

type DateFormatPart struct{}

func (p *DateFormatPart) Format(r *Record, buf *bytes.Buffer) {
	year, mon, day := r.Time.Date()
	buf.Write(uint2Bytes4(uint(year)))
	buf.WriteByte('-')
	buf.Write(uint2Bytes2(uint(mon)))
	buf.WriteByte('-')
	buf.Write(uint2Bytes2(uint(day)))
}

type SourceFormatPart struct{}

func (p *SourceFormatPart) Format(r *Record, buf *bytes.Buffer) {
	if r.Line > 0 {
		buf.WriteString(filepath.Base(r.File))
		buf.WriteByte(':')
		buf.WriteString(strconv.Itoa(r.Line))
	} else {
		buf.Write(unknownFile)
	}
}

type FullSourceFormatPart struct{}

func (p *FullSourceFormatPart) Format(r *Record, buf *bytes.Buffer) {
	if r.Line > 0 {
		buf.WriteString(r.File)
		buf.WriteByte(':')
		buf.WriteString(strconv.Itoa(r.Line))
	} else {
		buf.Write(unknownFile)
	}
}

type MessageFormatPart struct{}

func (p *MessageFormatPart) Format(r *Record, buf *bytes.Buffer) {
	msg := ""
	if len(r.Args) > 0 {
		if r.Message == "" {
			msg = fmt.Sprint(r.Args...)
		} else {
			msg = fmt.Sprintf(r.Message, r.Args...)
		}
	} else {
		msg = r.Message
	}
	if msg != "" {
		buf.WriteString(msg)
	}
}
