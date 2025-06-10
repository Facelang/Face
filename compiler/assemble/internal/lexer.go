package internal

import (
	"github.com/facelang/face/internal/reader"
	"unicode"
	"unicode/utf8"
)

// Whitespace 对比 map, switch 位掩码 比较效率最高, 忽略 \n
const Whitespace = 1<<'\t' | 1<<'\n' | 1<<'\r' | 1<<' '

//type lexer struct {
//	reader *reader.Reader
//	token  tokens2.Token
//	ident  string
//}
//
//func (lex *lexer) NextToken() tokens2.Token {
//	ch, chw := lex.reader.ReadRune()
//	if chw == 0 {
//		return tokens2.EOF
//	}
//
//	// skip white space
//	for Whitespace&(1<<ch) != 0 {
//		ch, chw = lex.reader.ReadRune()
//	}
//
//	if chw == 0 {
//		return tokens2.EOF
//	}
//
//	lex.ident = ""
//
//	// start collecting token text
//	lex.reader.TextReady()
//
//	if '0' <= ch && ch <= '9' { // 数字
//		return GetDecimal(lex, ch)
//	}
//
//	if CheckIdent(ch, 0) { // 符号
//		for i := 1; CheckIdent(ch, i); i++ {
//			ch, chw = lex.reader.ReadRune()
//		}
//		lex.ident = lex.reader.ReadText()
//		return tokens2.IDENT
//	}
//
//	// determine token value
//	switch ch {
//	case '"':
//		ident, _ := reader.String(lex.reader, '"')
//		lex.ident = ident
//		return tokens2.STRING
//	case '\'':
//		lex.ident = reader.Char(lex.reader)
//		return tokens.CHAR
//	case '`':
//		lex.ident = reader.RawString(lex.reader)
//		return tokens2.STRING
//	case ';': // todo at&t 语法使用 # 作为注解
//		lex.ident = reader.Comment(lex.reader)
//		return tokens2.COMMENT
//	default:
//		return tokens2.Token(ch)
//	}
//}
//
//func CheckIdent(ch rune, i int) bool {
//	return ch == '.' || ch == '_' || unicode.IsLetter(ch) ||
//		unicode.IsDigit(ch) && i > 0 // 第一个字符必须是字母或下划线
//}
//
//func GetDecimal(lex *lexer, ch rune) tokens2.Token {
//	token, val := reader.Decimal(lex.reader, ch)
//	lex.ident = val
//	return token
//}

type lexer struct {
	*reader.Reader        // 读取器
	id             string // 暂存字符
	pos            int    // 文件读取指针行列号
	back           bool   // 回退标识
	backToken      Token  // 回退Token
}

func (lex *lexer) Back(token Token) {
	lex.back = true
	lex.backToken = token
}

func (lex *lexer) NextToken() Token {
	defer func() {
		lex.back = false
	}()

	// 如果有回退，先获取回退
	if lex.back {
		return lex.backToken
	}

	ch, chw := lex.ReadRune()
	if chw == 0 {
		return EOF
	}

	lex.pos = lex.Pos()

	// skip white space
	for Whitespace&(1<<ch) != 0 {
		ch, chw = lex.ReadRune()
	}

	if chw == 0 {
		return EOF
	}

	lex.id = ""

	lex.TextReady()

	if '0' <= ch && ch <= '9' { // 数字
		return Number(lex, ch)
	}

	if CheckIdent(ch, 0) { // 符号
		for i := 1; CheckIdent(ch, i); i++ {
			ch, chw = lex.ReadRune()
		}
		lex.id = lex.ReadText()
		return Lookup(lex.id)
	}

	switch ch {
	case '+':
		return ADD
	case '-':
		return SUB
	case ':':
		return COLON
	case ',':
		return COMMA
	case ';':
		lex.id = reader.Comment(lex.Reader)
		return COMMENT
	case '"': // 查找字符串，到 " 结束
		lex.id, _ = reader.String(lex.Reader, '"')
		return STRING
	case '[':
		return LBRACK
	case ']':
		return RBRACK
	default:
		return ILLEGAL
	}
}

func CheckIdent(ch rune, i int) bool {
	return ch == '.' || ch == '_' || ch == '@' || unicode.IsLetter(ch) ||
		ch > utf8.RuneSelf || unicode.IsDigit(ch) && i > 0 // 第一个字符必须是字母或下划线
}

func Number(lex *lexer, ch rune) Token {
	typ, val := reader.Number(lex.Reader, ch)
	lex.id = val

	if typ == reader.INT_TYPE {
		return INT
	}
	return FLOAT
}

func NewLexer(file string) *lexer { // 封装后的读取器
	return &lexer{Reader: reader.FileReader(file)}
}
