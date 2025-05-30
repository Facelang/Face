package internal

import (
	"log"
	"os"
	"unicode"

	"github.com/facelang/face/compiler/common/parser"
	"github.com/facelang/face/compiler/common/tokens"

	"github.com/facelang/face/compiler/common/reader"
)

// Whitespace 对比 map, switch 位掩码 比较效率最高
const Whitespace = 1<<'\t' | 1<<'\n' | 1<<'\r' | 1<<' '

// 数字类型常量
const (
	Int   = iota // 整数
	Float        // 浮点数
)

type lexer struct {
	reader *reader.Reader
	token  tokens.Token
	ident  string
}

func (lex *lexer) NextToken() tokens.Token {
	ch, chw := lex.reader.ReadRune()
	if chw == 0 {
		return tokens.EOF
	}

redo:
	// skip white space
	for Whitespace&(1<<ch) != 0 {
		ch, chw = lex.reader.ReadRune()
	}

	if chw == 0 {
		return tokens.EOF
	}

	lex.ident = ""

	// start collecting token text
	lex.reader.TextReady()

	if '0' <= ch && ch <= '9' { // 数字
		return GetDecimal(lex, ch)
	}

	if CheckIdent(ch, 0) { // 符号
		for i := 1; CheckIdent(ch, i); i++ {
			ch, chw = lex.reader.ReadRune()
		}
		lex.ident = lex.reader.ReadText()
		return tokens.IDENT
	}

	// determine token value
	switch ch {
	case tokens.EOF:
		break
	case '"':
		if s.Mode&ScanStrings != 0 {
			s.scanString('"')
			tok = String
		}
		ch = s.next()
	case '\'':
		if s.Mode&ScanChars != 0 {
			s.scanChar()
			tok = Char
		}
		ch = s.next()
	case '.':
		ch = s.next()
		if isDecimal(ch) && s.Mode&ScanFloats != 0 {
			tok, ch = s.scanNumber(ch, true)
		}
	case '/':
		ch = s.next()
		if (ch == '/' || ch == '*') && s.Mode&ScanComments != 0 {
			if s.Mode&SkipComments != 0 {
				s.tokPos = -1 // don't collect token text
				ch = s.scanComment(ch)
				goto redo
			}
			ch = s.scanComment(ch)
			tok = Comment
		}
	case '`':
		if s.Mode&ScanRawStrings != 0 {
			s.scanRawString()
			tok = RawString
		}
		ch = s.next()
	default:
		ch = s.next()
	}

	// end of token text
	s.tokEnd = s.srcPos - s.lastCharLen

	s.ch = ch
	return tok
}

func CheckIdent(ch rune, i int) bool {
	return ch == '_' || unicode.IsLetter(ch) ||
		unicode.IsDigit(ch) && i > 0 // 第一个字符必须是字母或下划线
}

func GetDecimal(lex *lexer, ch rune) tokens.Token {
	token, val := parser.Decimal(lex.reader, byte(ch))
	lex.ident = val
	return token
}

// NewLexer returns a lexer for the named file and the given link context.
func NewLexer(name string) TokenReader { // 封装后的读取器
	input := NewInput(name)
	fd, err := os.Open(name)
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	input.Push(NewTokenizer(name, fd, fd))
	return input
}
