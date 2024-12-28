package factory

type Lexer interface {
	NextToken() Token
}

type lexer struct {
	buf            *buffer // 读取器
	content        []byte  // 暂存字符
	col, line, off int     // 文件读取指针行列号
}

func (l *lexer) init(file string, src interface{}, errFunc ErrorFunc) {
	l.buf.init(file, src, errFunc)
}

func (l *lexer) scan() Token {
	ch, eof := l.buf.next()
	for ch <= ' ' { // 空格符号之前的符号全部忽略
		if eof {
			return EOF
		}
		ch, eof = l.buf.next()
	}

	l.col = l.buf.col + 1
	l.line = l.buf.line + 1
	l.off = l.buf.off + 1

	l.buf.start()

	if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch == '_' {
		for ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' ||
			ch == '_' || ch >= '0' && ch <= '9' {
			ch, eof = l.buf.next()
		}
		l.content = l.buf.segment()
		return IDENT
		// 检查是否为 关键字
	} else if ch >= '0' && ch <= '9' {
		for ch >= '0' && ch <= '9' {
			ch, eof = l.buf.next()
		}
		l.content = l.buf.segment()
		return INT
	} else {
		// +, -, :, ;, ,, "
		// ; 分号之后读取换行符
		switch ch {
		case '+':
			return ADD
		case '-':
			return SUB
		case '*':
			return MUL
		case '/':
			if l.buf.look() == '/' {
				for !eof && ch != '\n' {
					ch, eof = l.buf.next()
				}
				return COMMENT
			}
			return QUO
		case '>':
			switch l.buf.look() {
			case '=':
				l.buf.next()
				return GEQ
			case '>':
				l.buf.next()
				return SHR
			default:
				return GTR
			}
		case '<':
			switch l.buf.look() {
			case '=':
				l.buf.next()
				return LEQ
			case '>':
				l.buf.next()
				return SHL
			default:
				return LSS
			}
		case '=':
			if l.buf.look() == '=' {
				l.buf.next()
				return EQL
			}
			return ASSIGN
		case '!':
			if l.buf.look() == '=' {
				l.buf.next()
				return NEQ
			}
			return ILLEGAL
		case ';':
			return SEMICOLON
		case ',':
			return COMMA
		case '"': // 查找字符串，到 " 结束
			ch, eof = l.buf.next()
			for !eof && ch != '"' {
				if ch == '\\' {
					l.buf.next()
				}
				ch, eof = l.buf.next()
			}
			l.content = l.buf.segment()
			return STRING
		case '\'': // 读一个字符
			ch, eof = l.buf.next()
			if l.buf.look() == '\'' {
				ch, eof = l.buf.next()
				l.content = []byte("")
				l.content = append(l.content, ch)
				return CHAR
			}
			return ILLEGAL
		case '(':
			return LPAREN
		case ')':
			return RPAREN
		case '{':
			return LBRACE
		case '}':
			return RBRACE
		default:
			return ILLEGAL
		}
	}
}
