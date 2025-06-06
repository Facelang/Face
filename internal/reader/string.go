package reader

import "fmt"

func Char(r *Reader) string {
	ident, l := String(r, '\'')
	if l != 1 {
		panic(fmt.Errorf("invalid char literal"))
	}
	return ident
}

func String(r *Reader, quote byte) (string, int) {
	length := 0
	ch, ok := r.ReadByte() // read character after quote
	for ch != quote {
		if ch == '\n' || !ok {
			panic(fmt.Errorf("literal not terminated"))
		}
		if ch == '\\' {
			ch = escape(r, quote)
		} else {
			ch, ok = r.ReadByte()
		}
		length++
	}
	return r.ReadText(), length
}

func RawString(r *Reader) string {
	ch, ok := r.ReadByte() // read character after '`'
	for ch != '`' {
		if !ok {
			panic(fmt.Errorf("literal not terminated"))
		}
		ch, ok = r.ReadByte()
	}
	return r.ReadText()
}

// Comment 单行注释
func Comment(r *Reader) string {
	ch, ok := r.ReadByte() // read character after "//"
	for ok && ch != '\n' {
		ch, ok = r.ReadByte()
	}
	r.GoBack()
	return r.ReadText()
}

// 处理转义字符
func escape(r *Reader, quote byte) byte {
	ch, _ := r.ReadByte() // read character after '/'
	switch ch {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', quote:
		// 常见的转义字符， 只需要读一个字符即可
		ch, _ = r.ReadByte()
	case '0', '1', '2', '3', '4', '5', '6', '7':
		// 处理形如 \123 的八进制转义序列
		// 最多读取 3 位八进制数字
		ch = number(r, ch, 8, 3)
	case 'x':
		ch, _ = r.ReadByte()
		ch = number(r, ch, 16, 2)
	case 'u':
		ch, _ = r.ReadByte()
		ch = number(r, ch, 16, 4)
	case 'U':
		ch, _ = r.ReadByte()
		ch = number(r, ch, 16, 8)
	default:
		panic(fmt.Errorf("invalid char escape"))
	}
	return ch
}

// 处理数字部分
func number(r *Reader, ch byte, base, n int) byte {
	for n > 0 && digitVal(ch) < base {
		ch, _ = r.ReadByte()
		n--
	}
	if n > 0 {
		panic(fmt.Errorf("invalid char escape"))
	}
	return ch
}
