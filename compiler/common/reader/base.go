package reader

import (
	"errors"
	"fmt"
	"io"
	"os"
)

type ErrorFunc func(info *FileInfo, msg string)

type FileInfo struct {
	Filename          string
	Col, Line, Offset int
}

func (i *FileInfo) String() string {
	return fmt.Sprintf("行: %d, 列: %d, 文件名：%s", i.Line+1, i.Col+1, i.Filename)
}

const BufferSize = 4 << 10    // 4K: minimum buffer size
const BufferSizeMax = 4 << 24 // 64M: maximum buffer size

type Reader struct {
	*FileInfo           // 文件信息
	buf       []byte    // 缓存池
	ch        byte      // 缓存字符
	b, r, e   int       // 读取器游标
	err       error     // 错误信息
	reader    io.Reader // 读取器
}

func (r *Reader) init(file string) error {
	r.Col, r.Line, r.Offset = 0, 0, 0
	r.Filename = file

	r.b, r.r, r.e = 0, 0, 0

	if r.buf == nil {
		r.buf = make([]byte, BufferSize)
	}

	reader, err := os.Open(file)
	if err != nil {
		return err
	}

	r.err = nil
	r.reader = reader
	return r.fill()
}

func (r *Reader) fill() error {
	n, err := r.reader.Read(r.buf[r.e:len(r.buf)])
	if n > 0 {
		r.e += n
	}
	if err != nil {
		return err
	}

	n, err = r.reader.Read(r.buf[r.e:len(r.buf)])
	if n > 0 {
		r.e += n
	}
	return err
}

func (r *Reader) fillNext() error {
	if r.err != nil {
		return r.err
	}

	i := r.r
	if r.b >= 0 {
		i = r.b
		r.b = 0
	}

	content := r.buf[i:r.e]
	size := r.e - i

	if size > 0 && len(r.buf) < BufferSizeMax {
		if size<<1 >= BufferSizeMax {
			r.buf = make([]byte, BufferSizeMax)
		} else {
			r.buf = make([]byte, size<<1)
		}
	}

	if i > 0 && size > 0 {
		copy(r.buf, content)
		r.r -= i
		r.e -= i
	}

	if r.e >= len(r.buf) {
		panic(errors.New(fmt.Sprintf(
			"解析文件时，文件段落超出缓冲区大小[最大限制64M]\n\t->%s", r.String())))
	}

	return r.fill()
}

//	func (b *Reader) read() rune {
//		if b.chw > 0 {
//			return b.ch
//		}
//
//		//if b.err == nil || b.isFull() {
//		//	if b.err == io.EOF {
//		//		return -1
//		//	}
//		//	_ = b.fill()
//		//}
//
//		// EOF
//		if b.r == b.e {
//			if b.err != io.EOF {
//				b.error("I/O error: " + b.err.Error())
//			}
//			return -1
//		}
//
//		if b.buf[b.r] < utf8.RuneSelf {
//			b.ch = rune(b.buf[b.r])
//			b.chw = 1
//			return b.ch
//		}
//
//		b.ch, b.chw = utf8.DecodeRune(b.buf[b.r:b.e])
//		if b.ch == utf8.RuneError && b.chw == 1 {
//			b.error("invalid UTF-8 encoding")
//			return -1
//		}
//
//		if b.ch == BOMTag {
//			if b.line > 0 || b.col > 0 {
//				b.error("invalid BOM in the middle of the file")
//			}
//			return -1
//		}
//
//		return b.ch
//	}
func (r *Reader) start()          { r.b = r.r - 1 }
func (r *Reader) stop()           { r.b = -1 }
func (r *Reader) segment() string { return string(r.buf[r.b : r.r-1]) }

func (r *Reader) GetFile() *FileInfo {
	return &FileInfo{
		r.Filename,
		r.Col + 1,
		r.Line + 1,
		r.Offset,
	}
}

// ReadByte 返回值是否为 eof
func (r *Reader) ReadByte() (byte, error) {
	if r.r == r.e {
		r.err = r.fillNext()
	}

	if r.r == r.e {
		r.ch = 0
		return 0, io.EOF
	}

	r.ch = r.buf[r.r]
	r.r += 1
	r.Offset += 1

	if r.ch == '\n' {
		r.Col = 0
		r.Line += 1
	} else {
		r.Col += 1
	}

	return r.ch, r.err
}

// FileReader todo 蔚来可能扩展支持 多种数据源读取模式，比如数据流
func FileReader(file string) *Reader {
	r := new(Reader)
	r.err = r.init(file)
	return r
}
