package factory

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

const BOMTag = 0xfeff

const BufferSize = 4 << 10    // 4K: minimum buffer size
const BufferSizeMax = 4 << 24 // 64M: maximum buffer size

type buffer struct {
	in             io.Reader // 读取器
	file           string    // 文件名 / 路径
	buf            []byte    // 缓存池
	ch             byte      // 当前字符
	b, r, e        int       // 读取器游标
	col, line, off int       // 文件读取指针行列号
	err            error     // 错误信息
	errFunc        ErrorFunc // 错误处理函数
}

func (b *buffer) error(msg string) {
	b.errFunc(b.file, b.line+1, b.col+1, b.off+1, msg)
}

func (b *buffer) errorf(format string, args ...interface{}) {
	b.error(fmt.Sprintf(format, args...))
}

func (b *buffer) init(file string, src interface{}, errFunc ErrorFunc) {
	b.in = nil
	b.file = file
	b.col = 0
	b.line = 0
	b.off = 0
	b.b = 0
	b.r = 0
	b.e = 0
	b.err = io.EOF
	b.errFunc = errFunc

	if b.buf == nil {
		b.buf = make([]byte, BufferSize)
	}

	if src != nil {
		b.in, b.err = os.Open(file)
		if b.err != nil {
			panic(b.err)
		}
		return
	}

	switch s := src.(type) {
	case string:
		b.buf = []byte(s)
	case []byte:
		b.buf = s
	case *bytes.Buffer:
		b.buf = s.Bytes()
	case io.Reader:
		b.in = s
		b.err = b.fill()
	}
}

func (b *buffer) fill() error {
	n, err := b.in.Read(b.buf[b.e:len(b.buf)])
	if n > 0 {
		b.e += n
	}

	return err
}

func (b *buffer) fillNext() error {
	if b.err != nil {
		return b.err
	}

	i := b.r
	if b.b >= 0 {
		i = b.b
		b.b = 0
	}

	content := b.buf[i:b.e]

	if len(b.buf) < BufferSizeMax {
		size := b.e - i
		if size<<1 >= BufferSizeMax {
			b.buf = make([]byte, BufferSizeMax)
		} else {
			b.buf = make([]byte, size<<1)
		}
	}

	if i > 0 {
		copy(b.buf, content)
	}

	b.r -= i
	b.e -= i

	if b.e >= len(b.buf) {
		return overflow(b)
	}

	return b.fill()
}

//func (b *buffer) read() rune {
//	if b.chw > 0 {
//		return b.ch
//	}
//
//	//if b.err == nil || b.isFull() {
//	//	if b.err == io.EOF {
//	//		return -1
//	//	}
//	//	_ = b.fill()
//	//}
//
//	// EOF
//	if b.r == b.e {
//		if b.err != io.EOF {
//			b.error("I/O error: " + b.err.Error())
//		}
//		return -1
//	}
//
//	if b.buf[b.r] < utf8.RuneSelf {
//		b.ch = rune(b.buf[b.r])
//		b.chw = 1
//		return b.ch
//	}
//
//	b.ch, b.chw = utf8.DecodeRune(b.buf[b.r:b.e])
//	if b.ch == utf8.RuneError && b.chw == 1 {
//		b.error("invalid UTF-8 encoding")
//		return -1
//	}
//
//	if b.ch == BOMTag {
//		if b.line > 0 || b.col > 0 {
//			b.error("invalid BOM in the middle of the file")
//		}
//		return -1
//	}
//
//	return b.ch
//}

func (b *buffer) next() (byte, bool) {
	if b.r == b.e {
		b.err = b.fillNext()
	}
	if b.r == b.e {
		return 0, true
	} else {
		b.ch = b.buf[b.r]
		b.r += 1
		b.off += 1
	}

	if b.ch == '\n' {
		b.col = 0
		b.line += 1
	} else {
		b.col += 1
	}
	return b.ch, false
}

//func (b *buffer) next() rune {
//	b.move()
//	return b.read()
//}

//	func (b *buffer) isFull() bool {
//		return b.e == b.r || b.buf[b.r] > utf8.RuneSelf && !utf8.FullRune(b.buf)
//	}
func (b *buffer) start()          { b.b = b.r - 1 }
func (b *buffer) stop()           { b.b = -1 }
func (b *buffer) segment() []byte { return b.buf[b.b : b.r-1] }

//func (b *buffer) offset(s uint) (rune, uint) {
//	return GetRune(b.buf[s:], b.errorBy)
//}

//func (b *buffer) test(next string) bool {
//	offset := uint(0)
//	for _, ch := range []rune(next) {
//		rv, rl := b.offset(offset)
//		offset += rl
//		if rv != ch {
//			return false
//		}
//	}
//	return true
//}
//
//func (b *buffer) Scan(offset int, check func(rune) bool) []byte {
//	b.start()
//	for {
//		ch, _ := b.rune()
//		if ch < 0 || !check(ch) {
//			return b.segment()
//		}
//	}
//}
//
//func (b *buffer) Peek() (rune, uint) {
//	return b.rune()
//}
//
//func (b *buffer) Match(next string) bool {
//	return b.test(next)
//}
//
//func (b *buffer) ReadNext() (rune, uint) {
//	return b.next()
//}
//
//func (b *buffer) FilePos() (uint, uint) {
//	return b.pos()
//}
//
//func (b *buffer) FileName() string { return b.file }
