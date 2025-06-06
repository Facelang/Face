package parser

import (
	"github.com/facelang/face/compiler/compile/token"
	"github.com/facelang/face/internal/reader"
	"unicode"
	"unicode/utf8"
)

// Whitespace 对比 map, switch 位掩码 比较效率最高, 忽略 \n
const Whitespace = 1<<'\t' | 1<<'\r' | 1<<' '

type lexer struct {
	*reader.Reader           // 读取器
	pos            token.Pos // 位置信息
	identifier     string    // 标识符
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
func (lex *lexer) NextToken() token.Token {
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
		return token.EOF
	}

	lex.pos = token.Pos(lex.Pos())

	// skip white space
	for Whitespace&(1<<ch) != 0 {
		ch, chw = lex.ReadRune()
	}

	if chw == 0 {
		return token.EOF
	}

	lex.identifier = ""

	// start collecting token text
	lex.TextReady()

	if '0' <= ch && ch <= '9' { // 数字
		return Number(lex, ch)
	}

	if CheckIdent(ch, 0) { // 符号
		for i := 1; CheckIdent(ch, i); i++ {
			ch, chw = lex.ReadRune()
		}
		lex.identifier = lex.ReadText()
		return token.Lookup(lex.identifier)
	}

	switch ch {
	case '\n':
		return token.NEWLINE
	case '+':
		return token.ADD
	case '-':
		return token.SUB
	case '*':
		return token.MUL
	case '/':
		next, _ := lex.ReadByte()
		if next == '/' {
			lex.identifier = reader.Comment(lex.Reader)
			return token.COMMENT
		}
		lex.GoBack()
		return token.QUO
	case '>':
		next, _ := lex.ReadByte()
		if next == '=' {
			return token.GEQ
		} else if next == '>' {
			return token.SHR
		} else {
			lex.GoBack()
			return token.GTR
		}
	case '<':
		next, _ := lex.ReadByte()
		if next == '=' {
			return token.LEQ
		} else if next == '>' {
			return token.SHL
		} else {
			lex.GoBack()
			return token.LSS
		}
	case '=':
		next, _ := lex.ReadByte()
		if next == '=' {
			return token.EQL
		}
		lex.GoBack()
		return token.ASSIGN
	case '!':
		next, _ := lex.ReadByte()
		if next == '=' {
			return token.NEQ
		}
		lex.GoBack()
		return token.NOT
	case ';':
		return token.SEMICOLON
	case ',':
		return token.COMMA
	case '"': // 查找字符串，到 " 结束, 最后一个字符是 ", 所以不需要回退
		ident, _ := reader.String(lex.Reader, '"')
		lex.identifier = ident
		return token.STRING
	case '\'': // 读一个字符, 字符串读， \' 结尾， 不需要回退
		lex.identifier = reader.Char(lex.Reader)
		return token.CHAR
	case '`': // todo 多行文本，需要进一步处理为一般字符串
		lex.identifier = reader.RawString(lex.Reader)
		return token.STRING
	case '(':
		return token.LPAREN
	case ')':
		return token.RPAREN
	case '{':
		return token.LBRACE
	case '}':
		return token.RBRACE
	default:
		return token.ILLEGAL
	}
}

func CheckIdent(ch rune, i int) bool {
	return ch == '.' || ch == '_' || unicode.IsLetter(ch) ||
		ch > utf8.RuneSelf || unicode.IsDigit(ch) && i > 0 // 第一个字符必须是字母或下划线
}

func Number(lex *lexer, ch rune) token.Token {
	typ, val := reader.Number(lex.Reader, ch)
	lex.identifier = val

	if typ == reader.INT_TYPE {
		return token.INT
	}
	return token.FLOAT
}

func NewLexer(file string) *lexer { // 封装后的读取器
	return &lexer{Reader: reader.FileReader(file)}
}
