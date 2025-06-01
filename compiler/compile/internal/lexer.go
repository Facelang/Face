package internal

import (
	"github.com/facelang/face/compiler/compile"
	"github.com/facelang/face/internal/reader"
	"github.com/facelang/face/internal/tokens"
	"unicode"
)

// Whitespace 对比 map, switch 位掩码 比较效率最高, 忽略 \n
const Whitespace = 1<<'\t' | 1<<'\r' | 1<<' '

type lexer struct {
	reader *reader.Reader
	token  tokens.Token
	ident  string
}

//type lexer struct {
//	buffer            *buffer       // 读取器
//	content           string        // 暂存字符
//	col, line, offset int           // 文件读取指针行列号
//	back              bool          // 回退标识
//	backToken         compile.Token // 回退Token
//}

//func (l *lexer) init(file string, errFunc compile.ErrorFunc) error {
//	defer func() { l.buffer.read() }()
//	return l.buffer.init(file, errFunc)
//}
//
//func (l *lexer) Back(token compile.Token) {
//	l.back = true
//	l.backToken = token
//}

func (lex *lexer) NextToken() tokens.Token {
	ch, chw := lex.reader.ReadRune()
	if chw == 0 {
		return tokens.EOF
	}

	info := lex.reader.GetFile()

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
	case '\n':
		return tokens.RETURN
	case '"':
		ident, _ := reader.String(lex.reader, '"')
		lex.ident = ident
		return tokens.STRING
	case '\'':
		lex.ident = reader.Char(lex.reader)
		return tokens.Char
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

func (l *lexer) NextToken() compile.Token {
	defer func() {
		l.back = false
	}()

	// 如果有回退，先获取回退
	if l.back {
		return l.backToken
	}

	l.content = ""

	if isIndentPrefix(l.buffer.ch) {
		for isIndentContent(l.buffer.ch) {
			if !l.buffer.read() {
				break
			}
		}
		l.content = l.buffer.segment()
		key, ok := compile.Keywords(l.content)
		if ok {
			return key
		}
		return compile.IDENT
		// 检查是否为 关键字
	} else if isNumeric(l.buffer.ch) {
		for isNumeric(l.buffer.ch) {
			if !l.buffer.read() {
				break
			}
		}
		l.content = l.buffer.segment()
		return compile.V_INT
	} else {
		switch l.buffer.ch {
		case '+':
			l.buffer.read()
			return compile.ADD
		case '-':
			l.buffer.read()
			return compile.SUB
		case '*':
			l.buffer.read()
			return compile.MUL
		case '/':
			l.buffer.read()
			if l.buffer.ch == '/' {
				l.buffer.read()
				for l.buffer.ch != '\n' {
					if !l.buffer.read() {
						break
					}
				}
				return compile.COMMENT
			}
			return compile.QUO
		case '>':
			l.buffer.read()
			switch l.buffer.ch {
			case '=':
				l.buffer.read()
				return compile.GEQ
			case '>':
				l.buffer.read()
				return compile.SHR
			default:
				return compile.GTR
			}
		case '<':
			l.buffer.read()
			switch l.buffer.ch {
			case '=':
				l.buffer.read()
				return compile.LEQ
			case '<':
				l.buffer.read()
				return compile.SHL
			default:
				return compile.LSS
			}
		case '=':
			l.buffer.read()
			if l.buffer.ch == '=' {
				l.buffer.read()
				return compile.EQL
			}
			return compile.ASSIGN
		case '!':
			l.buffer.read()
			if l.buffer.ch == '=' {
				l.buffer.read()
				return compile.NEQ
			}
			return compile.ILLEGAL
		case ';':
			l.buffer.read()
			return compile.SEMICOLON
		case ',':
			l.buffer.read()
			return compile.COMMA
		case '"': // 查找字符串，到 " 结束
			next := l.buffer.read()
			for !next || l.buffer.ch != '"' {
				if l.buffer.ch == '\\' {
					l.buffer.read()
				}
				next = l.buffer.read()
			}
			l.buffer.read()
			l.content = l.buffer.segment()
			return compile.V_STRING
		case '\'': // 读一个字符
			l.buffer.read()
			if l.buffer.ch == '\\' {
				l.buffer.read()
			}
			l.content = string(l.buffer.ch)
			l.buffer.read()
			return compile.ILLEGAL
		case '(':
			l.buffer.read()
			return compile.LPAREN
		case ')':
			l.buffer.read()
			return compile.RPAREN
		case '{':
			l.buffer.read()
			return compile.LBRACE
		case '}':
			l.buffer.read()
			return compile.RBRACE
		default:
			l.buffer.read()
			return compile.ILLEGAL
		}
	}
}

func isIndentContent(ch byte) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' &&
		ch <= 'Z' || ch == '_' || ch >= '0' && ch <= '9'
}

func isIndentPrefix(ch byte) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch == '_'
}

func isNumeric(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func CheckIdent(ch rune, i int) bool {
	unicode
	return ch == '.' || ch == '_' || unicode.IsLetter(ch) ||
		unicode.IsLower() || unicode.IsDigit(ch) && i > 0 // 第一个字符必须是字母或下划线
}

func GetDecimal(lex *lexer, ch rune) tokens.Token {
	token, val := reader.Decimal(lex.reader, ch)
	lex.ident = val
	return token
}

func NewLexer(file string) *lexer { // 封装后的读取器
	return &lexer{reader: reader.FileReader(file)}
}
