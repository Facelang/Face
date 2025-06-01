package reader

import (
	"fmt"
	"github.com/facelang/face/internal/tokens"
	"os"
	"unicode/utf8"
)

type Reader struct {
	filename       string // 文件名称
	buff           []byte // 缓存池
	ch             byte   // 主要记录换行符更新
	chw            int    // 缓存字符宽度, 下一次读更新上一个字符宽度
	b, r, e        int    // 读取器游标
	line, col, off int    // 文件指针
}

func (r *Reader) errorf(format string, args ...any) {
	panic(fmt.Errorf("Reader Error: %s\n\t->[%d, %d] %s",
		fmt.Sprintf(format, args...), r.line+1, r.col+1, r.filename))
}

func (r *Reader) Pos() tokens.FilePos {
	return tokens.FilePos{
		Filename: r.filename,
		Col:      r.col + 1,
		Line:     r.line + 1,
		Offset:   r.off,
	}
}

// GoBack 回退一个字符
func (r *Reader) GoBack() {
	r.ch = 0
	r.chw = 0
}

// ReadByte 返回值是否为 eof
func (r *Reader) ReadByte() (byte, bool) {
	if r.chw > 0 { // 文件位置信息记录更新， 下一个字符开始 = 上一个字符结束 + 上一个字符宽度
		r.r += r.chw
		r.off += r.chw

		if r.ch == '\n' {
			r.col = 0
			r.line += 1
		} else {
			r.col += 1 // utf8 字符占一列
		}

		r.chw = 0
	}

	// eof
	if r.r == r.e {
		r.ch = 0
		r.chw = 0
		return 0, true
	}

	r.ch = r.buff[r.r]
	r.chw = 1
	return r.ch, false
}

func (r *Reader) ReadRune() (rune, int) {
redo:
	c, eof := r.ReadByte()
	if eof {
		return 0, 0
	}

	if c < utf8.RuneSelf {
		return rune(c), 1
	}

	// 解码 UTF-8 字符
	ch, chw := utf8.DecodeRune(r.buff[r.r:r.e])

	r.chw = chw

	// 检查解码错误
	if ch == utf8.RuneError && chw == 1 { // 无效的 UTF-8 编码
		r.errorf("invalid UTF-8 encoding at position %d", r.off-1)
	}

	const BOM = 0xfeff
	if ch == BOM {
		if r.off > 0 {
			r.errorf("invalid BOM in the middle of the file")
		}
		goto redo // 忽略 BOM 字符
	}

	return ch, chw
}

// TextReady 文本读取器准备就绪
func (r *Reader) TextReady() {
	r.b = r.r
}

// ReadText 读取一段本文
func (r *Reader) ReadText() string {
	defer func() {
		r.b = -1 // 重置游标
	}()

	return string(r.buff[r.b : r.r+r.chw])
}

// FileReader todo 蔚来可能扩展支持 多种数据源读取模式，比如数据流
func FileReader(file string) *Reader {
	r := &Reader{filename: file}

	buff, err := os.ReadFile(file)
	if err != nil {
		r.errorf("failed to read file: %s", err)
	}

	r.buff = buff
	r.e = len(r.buff)
	return r
}

func BytesReader(input []byte) *Reader {
	r := &Reader{filename: "#Bytes"}
	r.buff = input
	r.e = len(input)
	return r
}
