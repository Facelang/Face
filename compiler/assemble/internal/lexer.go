package internal

import (
	"github.com/facelang/face/internal/reader"
	"github.com/facelang/face/internal/tokens"
	"unicode"
)

// Whitespace 对比 map, switch 位掩码 比较效率最高, 忽略 \n
const Whitespace = 1<<'\t' | 1<<'\n' | 1<<'\r' | 1<<' '

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
	case '"':
		ident, _ := reader.String(lex.reader, '"')
		lex.ident = ident
		return tokens.STRING
	case '\'':
		lex.ident = reader.Char(lex.reader)
		return tokens.CHAR
	case '`':
		lex.ident = reader.RawString(lex.reader)
		return tokens.STRING
	case ';': // todo at&t 语法使用 # 作为注解
		lex.ident = reader.Comment(lex.reader)
		return tokens.COMMENT
	default:
		return tokens.Token(ch)
	}
}

func CheckIdent(ch rune, i int) bool {
	return ch == '.' || ch == '_' || unicode.IsLetter(ch) ||
		unicode.IsDigit(ch) && i > 0 // 第一个字符必须是字母或下划线
}

func GetDecimal(lex *lexer, ch rune) tokens.Token {
	token, val := reader.Decimal(lex.reader, ch)
	lex.ident = val
	return token
}

func NewLexer(file string) *lexer { // 封装后的读取器
	return &lexer{reader: reader.FileReader(file)}
}
