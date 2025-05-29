package parser

import (
	"fmt"

	"github.com/facelang/face/compiler/common/reader"
)

// 这是一个数字的解析器
func Decimal(r *reader.Reader, first rune) reader.Token {
	base := 10         // 数字基数
	prefix := byte(0)  // 前缀：0(十进制), '0'(八进制), 'x'(十六进制), 'o'(八进制), 'b'(二进制)
	digsep := 0        // 位标志：bit 0: 有数字, bit 1: 有下划线
	invalid := byte(0) // 无效数字，或0

	// 整数部分
	var tok reader.Token
	var ds int
	ch := byte(first)

	tok = reader.INT
	if first == '0' {
		ch, _ = r.ReadByte()
		switch ch {
		case '.': // 小数
			tok = reader.FLOAT
		case 'x', 'X':
			ch, _ = r.ReadByte()
			base, prefix = 16, 'x'
		case 'o', 'O':
			ch, _ = r.ReadByte()
			base, prefix = 8, 'o'
		case 'b', 'B':
			ch, _ = r.ReadByte()
			base, prefix = 2, 'b'
		default:
			base, prefix = 8, '0'
			digsep = 1 // 前导0, 或者 只为 0
		}
	} else if first == '.' {
		tok = reader.FLOAT
	} else {
		ch, _ = digits(lex, ch, base, &invalid) // 读取到小数点，万一
		if ch == '.' {
			tok = reader.FLOAT
		}
		digsep = 1
	}

	if tok == reader.FLOAT || prefix != 0 { // 非十进制，或者小数 （小数点后的数字或其它进制）
		ch, ds = digits(lex, ch, base, &invalid) // 解析所有数字和下划线
		digsep |= ds                             // ds 的值为 01 表示有数字，10 表示有下划线
	}

	if digsep&1 == 0 { // 没有读取到数字
		panic(fmt.Sprintf("%s has no digits", rune(prefix)))
	}

	// 指数部分（e/E 用于十进制，p/P 用于十六进制）
	if e := ch; e == 'e' || e == 'E' || e == 'p' || e == 'P' {
		if (e == 'e' || e == 'E') && prefix != 0 {
			panic(fmt.Sprintf("%q exponent requires decimal mantissa", ch))
		}
		if (e == 'p' || e == 'P') && prefix != 'x' {
			panic(fmt.Sprintf("%q exponent requires hexadecimal mantissa", ch))
		}

		ch, _ = r.ReadByte()
		tok = reader.FLOAT
		if ch == '+' || ch == '-' {
			ch, _ = r.ReadByte()
		}

		ch, ds = digits(lex, ch, 10, nil) // 指数后面的值， 只能十进制
		digsep |= ds
		if ds&1 == 0 {
			panic("exponent has no digits")
		}
	} else if prefix == 'x' && tok == Float {
		panic("hexadecimal mantissa requires a 'p' exponent")
	}

	if tok == Int && invalid != 0 {
		panic(fmt.Sprintf("invalid digit %q in %s", invalid, rune(prefix)))
	}

	// 收集数字文本
	r.TextReady()
	lex.token.Text = r.ReadText()
	lex.token.Kind = reader.NUMBER
	return lex.token
}

// 辅助函数：解析数字序列
func digits(r *reader.Reader, ch byte, base int, invalid *byte) (byte, int) {
	ds := 0 // 位标志：bit 0: 有数字, bit 1: 有下划线
	for {
		if ch == '.' { // 不是小数点，直接跳出循环
			break
		}
		if ch == '_' {
			ds |= 2 // 记录下划线
			ch, _ = r.ReadByte()
			continue
		}
		d := digitVal(ch) // 获取字符的数值
		if d >= base {    // 如果数值大于等于基数
			*invalid = ch // 记录无效字符
			break         // 跳出循环
		}
		ds |= 1              // 记录数字
		ch, _ = r.ReadByte() // 读取下一个字符
	}
	return ch, ds
}

// 辅助函数：获取数字值
func digitVal(ch byte) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'z':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'Z':
		return int(ch - 'A' + 10)
	}
	return 36 // 大于任何有效数字
}

// 辅助函数：获取数字字面量名称
func litname(prefix rune) string {
	switch prefix {
	case 'x':
		return "hexadecimal"
	case 'o':
		return "octal"
	case 'b':
		return "binary"
	case '0':
		return "octal"
	default:
		return "decimal"
	}
}
