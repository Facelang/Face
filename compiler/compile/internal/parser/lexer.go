package parser

import (
	"github.com/facelang/face/compiler/compile/internal/api"
	"github.com/facelang/face/internal/prog"
	"github.com/facelang/face/internal/reader"
	"github.com/facelang/face/internal/tokens"
	"unicode"
	"unicode/utf8"
)

// Whitespace 对比 map, switch 位掩码 比较效率最高, 忽略 \n
const Whitespace = 1<<'\t' | 1<<'\r' | 1<<' '

type lexer struct {
	*reader.Reader              // 读取器
	FilePos        prog.FilePos // 位置信息
	identifier     string       // 标识符
}

//type lexer struct {
//	buffer            *buffer       // 读取器
//	content           string        // 暂存字符
//	col, line, offset int           // 文件读取指针行列号
//	back              bool          // 回退标识
//	backToken         Token // 回退Token
//}

//func (l *lexer) init(file string, errFunc ErrorFunc) error {
//	defer func() { next, _ := lex.ReadByte() }()
//	return l.buffer.init(file, errFunc)
//}
//
//func (l *lexer) Back(token Token) {
//	l.back = true
//	l.backToken = token
//}

// NextToken todo 需要处理分号，和换行符， 还需要处理：分支语句中，必须是分号，其它情况可以是换行符或者分号
func (lex *lexer) NextToken() tokens.Token {
	//defer func() {
	//	l.back = false
	//}()
	//
	//// 如果有回退，先获取回退
	//if l.back {
	//	return l.backToken
	//}

	ch, chw := lex.ReadRune()
	if chw == 0 {
		return tokens.EOF
	}

	lex.FilePos = lex.Pos()

	// skip white space
	for Whitespace&(1<<ch) != 0 {
		ch, chw = lex.ReadRune()
	}

	if chw == 0 {
		return tokens.EOF
	}

	lex.identifier = ""

	// start collecting token text
	lex.TextReady()

	if '0' <= ch && ch <= '9' { // 数字
		return GetDecimal(lex, ch)
	}

	if CheckIdent(ch, 0) { // 符号
		for i := 1; CheckIdent(ch, i); i++ {
			ch, chw = lex.ReadRune()
		}
		lex.identifier = lex.ReadText()
		if key, ok := api.Keywords(lex.identifier); ok {
			return key
		}
		return tokens.IDENT
	}

	switch ch {
	case '\n':
		return tokens.NEWLINE
	case '+':
		return api.ADD
	case '-':
		return api.SUB
	case '*':
		return api.MUL
	case '/':
		next, _ := lex.ReadByte()
		if next == '/' {
			lex.identifier = reader.Comment(lex.Reader)
			return tokens.COMMENT
		}
		lex.GoBack()
		return api.QUO
	case '>':
		next, _ := lex.ReadByte()
		if next == '=' {
			return api.GEQ
		} else if next == '>' {
			return api.SHR
		} else {
			lex.GoBack()
			return api.GTR
		}
	case '<':
		next, _ := lex.ReadByte()
		if next == '=' {
			return api.LEQ
		} else if next == '>' {
			return api.SHL
		} else {
			lex.GoBack()
			return api.LSS
		}
	case '=':
		next, _ := lex.ReadByte()
		if next == '=' {
			return api.EQL
		}
		lex.GoBack()
		return api.ASSIGN
	case '!':
		next, _ := lex.ReadByte()
		if next == '=' {
			return api.NEQ
		}
		lex.GoBack()
		return api.NOT
	case ';':
		return api.SEMICOLON
	case ',':
		return api.COMMA
	case '"': // 查找字符串，到 " 结束, 最后一个字符是 ", 所以不需要回退
		ident, _ := reader.String(lex.Reader, '"')
		lex.identifier = ident
		return tokens.STRING
	case '\'': // 读一个字符, 字符串读， \' 结尾， 不需要回退
		lex.identifier = reader.Char(lex.Reader)
		return tokens.CHAR
	case '`': // todo 多行文本，需要进一步处理为一般字符串
		lex.identifier = reader.RawString(lex.Reader)
		return tokens.STRING
	case '(':
		return api.LPAREN
	case ')':
		return api.RPAREN
	case '{':
		return api.LBRACE
	case '}':
		return api.RBRACE
	default:
		return tokens.ILLEGAL
	}
}

func CheckIdent(ch rune, i int) bool {
	return ch == '.' || ch == '_' || unicode.IsLetter(ch) ||
		ch > utf8.RuneSelf || unicode.IsDigit(ch) && i > 0 // 第一个字符必须是字母或下划线
}

func GetDecimal(lex *lexer, ch rune) tokens.Token {
	token, val := reader.Decimal(lex.Reader, ch)
	lex.identifier = val
	return token
}

func NewLexer(file string) *lexer { // 封装后的读取器
	return &lexer{Reader: reader.FileReader(file)}
}
