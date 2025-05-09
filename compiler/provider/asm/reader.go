package asm

import (
	"fmt"
	"os"
)

type reader struct {
	buf               []byte    // 缓存池
	ch                byte      // 缓存字符
	file              string    // 文件名 / 路径
	b, r, e           int       // 读取器游标
	col, line, offset int       // 文件读取指针行列号
	err               error     // 错误信息
	errFunc           ErrorFunc // 错误处理函数
}

func (r *reader) error(msg string) {
	r.errFunc(r.file, r.line+1, r.col+1, r.offset+1, msg)
}

func (r *reader) errorf(format string, args ...interface{}) {
	r.error(fmt.Sprintf(format, args...))
}

func (r *reader) init(file string, errFunc ErrorFunc) error {
	r.file = file
	r.errFunc = errFunc
	return r.reset()
}

func (r *reader) reset() error {
	r.b, r.r = 0, 0
	r.col, r.line, r.offset = 0, 0, 0

	r.buf, r.err = os.ReadFile(r.file)
	r.e = len(r.buf)
	return r.err
}

// 返回值是否为 eof
func (r *reader) read() bool {
	if r.r == r.e {
		r.ch = 0
		return false
	}

	r.ch = r.buf[r.r]
	r.r += 1
	r.offset += 1

	if r.ch == '\n' {
		r.col = 0
		r.line += 1
	} else {
		r.col += 1
	}

	return true
}

//func (r *reader) next() rune {
//	r.move()
//	return r.read()
//}

//	func (r *reader) isFull() bool {
//		return r.e == r.r || r.buf[r.r] > utf8.RuneSelf && !utf8.FullRune(r.buf)
//	}
func (r *reader) start()          { r.b = r.r - 1 }
func (r *reader) stop()           { r.b = -1 }
func (r *reader) segment() string { return string(r.buf[r.b : r.r-1]) }

//func (r *reader) offset(s uint) (rune, uint) {
//	return GetRune(r.buf[s:], r.errorBy)
//}

//func (r *reader) test(next string) bool {
//	offset := uint(0)
//	for _, ch := range []rune(next) {
//		rv, rl := r.offset(offset)
//		offset += rl
//		if rv != ch {
//			return false
//		}
//	}
//	return true
//}
//
//func (r *reader) Scan(offset int, check func(rune) bool) []byte {
//	r.start()
//	for {
//		ch, _ := r.rune()
//		if ch < 0 || !check(ch) {
//			return r.segment()
//		}
//	}
//}
//
//func (r *reader) Peek() (rune, uint) {
//	return r.rune()
//}
//
//func (r *reader) Match(next string) bool {
//	return r.test(next)
//}
//
//func (r *reader) ReadNext() (rune, uint) {
//	return r.next()
//}
//
//func (r *reader) FilePos() (uint, uint) {
//	return r.pos()
//}
//
//func (r *reader) FileName() string { return r.file }
