package internal

import (
	tokens2 "github.com/facelang/face/compiler/compile/internal/tokens"
	"github.com/facelang/face/internal/reader"
	"github.com/facelang/face/internal/tokens"
	"unicode"
)

// Whitespace 对比 map, switch 位掩码 比较效率最高, 忽略 \n
const Whitespace = 1<<'\t' | 1<<'\n' | 1<<'\r' | 1<<' '

type lexer struct {
	reader *reader.Reader
	token  tokens2.Token
	ident  string
}

func (lex *lexer) NextToken() tokens2.Token {
	ch, chw := lex.reader.ReadRune()
	if chw == 0 {
		return tokens2.EOF
	}

	// skip white space
	for Whitespace&(1<<ch) != 0 {
		ch, chw = lex.reader.ReadRune()
	}

	if chw == 0 {
		return tokens2.EOF
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
		return tokens2.IDENT
	}

	// determine token value
	switch ch {
	case '"':
		ident, _ := reader.String(lex.reader, '"')
		lex.ident = ident
		return tokens2.STRING
	case '\'':
		lex.ident = reader.Char(lex.reader)
		return tokens.CHAR
	case '`':
		lex.ident = reader.RawString(lex.reader)
		return tokens2.STRING
	case ';': // todo at&t 语法使用 # 作为注解
		lex.ident = reader.Comment(lex.reader)
		return tokens2.COMMENT
	default:
		return tokens2.Token(ch)
	}
}

func CheckIdent(ch rune, i int) bool {
	return ch == '.' || ch == '_' || unicode.IsLetter(ch) ||
		unicode.IsDigit(ch) && i > 0 // 第一个字符必须是字母或下划线
}

func GetDecimal(lex *lexer, ch rune) tokens2.Token {
	token, val := reader.Decimal(lex.reader, ch)
	lex.ident = val
	return token
}

func NewLexer(file string) *lexer { // 封装后的读取器
	return &lexer{reader: reader.FileReader(file)}
}
