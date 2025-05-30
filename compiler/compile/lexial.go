package compile

type Lexer interface {
	NextToken() Token
}

type lexer struct {
	buffer            *buffer // 读取器
	content           string  // 暂存字符
	col, line, offset int     // 文件读取指针行列号
	back              bool    // 回退标识
	backToken         Token   // 回退Token
}

func (l *lexer) init(file string, errFunc ErrorFunc) error {
	defer func() { l.buffer.read() }()
	return l.buffer.init(file, errFunc)
}

func (l *lexer) Back(token Token) {
	l.back = true
	l.backToken = token
}

func (l *lexer) NextToken() Token {
	defer func() {
		l.back = false
	}()

	// 如果有回退，先获取回退
	if l.back {
		return l.backToken
	}

	l.content = ""

	for l.buffer.ch <= ' ' { // 空格符号之前的符号全部忽略
		if !l.buffer.read() {
			return EOF
		}
	}

	l.buffer.start()

	l.col = l.buffer.col
	l.line = l.buffer.line
	l.offset = l.buffer.offset

	if isIndentPrefix(l.buffer.ch) {
		for isIndentContent(l.buffer.ch) {
			if !l.buffer.read() {
				break
			}
		}
		l.content = l.buffer.segment()
		key, ok := Keywords(l.content)
		if ok {
			return key
		}
		return IDENT
		// 检查是否为 关键字
	} else if isNumeric(l.buffer.ch) {
		for isNumeric(l.buffer.ch) {
			if !l.buffer.read() {
				break
			}
		}
		l.content = l.buffer.segment()
		return V_INT
	} else {
		switch l.buffer.ch {
		case '+':
			l.buffer.read()
			return ADD
		case '-':
			l.buffer.read()
			return SUB
		case '*':
			l.buffer.read()
			return MUL
		case '/':
			l.buffer.read()
			if l.buffer.ch == '/' {
				l.buffer.read()
				for l.buffer.ch != '\n' {
					if !l.buffer.read() {
						break
					}
				}
				return COMMENT
			}
			return QUO
		case '>':
			l.buffer.read()
			switch l.buffer.ch {
			case '=':
				l.buffer.read()
				return GEQ
			case '>':
				l.buffer.read()
				return SHR
			default:
				return GTR
			}
		case '<':
			l.buffer.read()
			switch l.buffer.ch {
			case '=':
				l.buffer.read()
				return LEQ
			case '<':
				l.buffer.read()
				return SHL
			default:
				return LSS
			}
		case '=':
			l.buffer.read()
			if l.buffer.ch == '=' {
				l.buffer.read()
				return EQL
			}
			return ASSIGN
		case '!':
			l.buffer.read()
			if l.buffer.ch == '=' {
				l.buffer.read()
				return NEQ
			}
			return ILLEGAL
		case ';':
			l.buffer.read()
			return SEMICOLON
		case ',':
			l.buffer.read()
			return COMMA
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
			return V_STRING
		case '\'': // 读一个字符
			l.buffer.read()
			if l.buffer.ch == '\\' {
				l.buffer.read()
			}
			l.content = string(l.buffer.ch)
			l.buffer.read()
			return ILLEGAL
		case '(':
			l.buffer.read()
			return LPAREN
		case ')':
			l.buffer.read()
			return RPAREN
		case '{':
			l.buffer.read()
			return LBRACE
		case '}':
			l.buffer.read()
			return RBRACE
		default:
			l.buffer.read()
			return ILLEGAL
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
