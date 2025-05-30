package parser

import (
	"fmt"
	"strings"

	"github.com/facelang/face/compiler/common/tokens"

	"github.com/facelang/face/compiler/common/reader"
)

// Decimal 这是一个数字的解析器
func Decimal(r *reader.Reader, first byte) (tokens.Token, string) {
	defer func() {
		r.GoBack() // 最后一个符号需要回退
	}()

	base := 10        // 数字基数
	prefix := byte(0) // 前缀：0(十进制), '0'(八进制), 'x'(十六进制), 'o'(八进制), 'b'(二进制)
	flags := byte(0)  // 位标志：bit 0: 有数字, bit 1: 有下划线, bit 2 符号异常

	// 整数部分
	ds := byte(0)
	ch := first
	tok := tokens.INT

	if first == '0' {
		ch, _ = r.ReadByte()
		switch ch {
		case '.': // 小数
			tok = tokens.FLOAT
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
			flags = 1 // 前导0, 或者 只为 0
		}
	} else if first == '.' {
		tok = tokens.FLOAT
	} else {
		flags = 1 // 前导数
	}

	// 整数和16进制支持小数表达， 先读取整数部分
	// 123.456 和 0x1.2p3 都是合法的
	if tok == tokens.INT || prefix != 'x' {
		ch, ds = digits(r, ch, base) // 解析所有数字和下划线
		flags |= ds                  // ds 的值为 01 表示有数字，10 表示有下划线
		if ch == '.' {
			if flags&1 == 0 { // 0x. 是非法的
				panic(fmt.Errorf("%s has no digits", decimalName(prefix)))
			}
			tok = tokens.FLOAT
		}
	}

	// 非十进制，或者小数 （小数点后的数字或其它进制）
	if tok == tokens.FLOAT || prefix != 0 {
		ch, ds = digits(r, ch, base) // 解析所有数字和下划线
		flags |= ds                  // ds 的值为 01 表示有数字，10 表示有下划线
		if flags&1 == 0 {            // 没有读取到数字
			panic(fmt.Errorf("%s has no digits", decimalName(prefix)))
		}
	}

	// 指数部分（e/E 用于十进制，p/P 用于十六进制）
	if e := ch; e == 'e' || e == 'E' || e == 'p' || e == 'P' {
		if (e == 'e' || e == 'E') && prefix != 0 {
			panic(fmt.Errorf("%q exponent requires decimal mantissa", ch))
		}
		if (e == 'p' || e == 'P') && prefix != 'x' {
			panic(fmt.Errorf("%q exponent requires hexadecimal mantissa", ch))
		}

		ch, _ = r.ReadByte()
		tok = tokens.FLOAT
		if ch == '+' || ch == '-' {
			ch, _ = r.ReadByte()
		}

		_, ds = digits(r, ch, 10) // 指数后面的值， 只能十进制
		flags |= ds

		if ds&1 == 0 { // 指数后面没有数字
			panic(fmt.Errorf("exponent has no digits"))
		}
	}

	if flags&2 == 0 {
		return tok, r.ReadText()
	}

	// 数字中有 _ 需要踢掉
	return tok, strings.ReplaceAll(r.ReadText(), "_", "")
}

// 辅助函数：解析数字序列
func digits(r *reader.Reader, ch byte, base int) (byte, byte) {
	ds := byte(0) // 位标志：bit 0: 有数字, bit 1: 有下划线 bit 3: 异常
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
			ds |= 4 // 记录异常
			break   // 跳出循环
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
func decimalName(prefix byte) string {
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
