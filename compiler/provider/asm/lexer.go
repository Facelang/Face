package asm

type Lexer interface {
	NextToken() Token
}

type lexer struct {
	reader            *reader // 读取器
	id                string  // 暂存字符
	col, line, offset int     // 文件读取指针行列号
	back              bool    // 回退标识
	backToken         Token   // 回退Token
}

func (l *lexer) init(file string, errFunc ErrorFunc) error {
	defer func() { l.reader.read() }()
	return l.reader.init(file, errFunc)
}

func (l *lexer) reset() error {
	defer func() { l.reader.read() }()
	return l.reader.reset()
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

	l.id = ""

	for l.reader.ch <= ' ' { // 空格符号之前的符号全部忽略
		if !l.reader.read() {
			return EOF
		}
	}

	l.reader.start()

	l.col = l.reader.col
	l.line = l.reader.line
	l.offset = l.reader.offset

	if isIndentPrefix(l.reader.ch) {
		for isIndentContent(l.reader.ch) {
			if !l.reader.read() {
				break
			}
		}
		l.id = l.reader.segment()
		key, ok := Keywords(l.id)
		if ok {
			return key
		}
		return IDENT
		// 检查是否为 关键字
	} else if isNumeric(l.reader.ch) {
		for isNumeric(l.reader.ch) {
			if !l.reader.read() {
				break
			}
		}
		l.id = l.reader.segment()
		return NUMBER
	} else {
		switch l.reader.ch {
		case '+':
			l.reader.read()
			return ADD
		case '-':
			l.reader.read()
			return SUB
		case ':':
			l.reader.read()
			return COLON
		case ',':
			l.reader.read()
			return COMMA
		case ';':
			l.reader.read()
			for l.reader.ch != '\n' {
				if !l.reader.read() {
					break
				}
			}
			return COMMENT
		case '"': // 查找字符串，到 " 结束
			next := l.reader.read()
			for !next || l.reader.ch != '"' {
				if l.reader.ch == '\\' {
					l.reader.read()
				}
				next = l.reader.read()
			}
			l.reader.read()
			l.id = l.reader.segment()
			return STRING
		case '[':
			l.reader.read()
			return LBRACK
		case ']':
			l.reader.read()
			return RBRACK
		default:
			l.reader.read()
			return ILLEGAL
		}
	}
}

func isIndentContent(ch byte) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' &&
		ch <= 'Z' || ch == '_' || ch >= '0' && ch <= '9' ||
		ch == '.' || ch == '@'
}

// 字母、_、@、.开头的标识符或关键字
func isIndentPrefix(ch byte) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' ||
		ch == '_' || ch == '.' || ch == '@'
}

func isNumeric(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func NewLexer() *lexer {
	return &lexer{reader: &reader{}}
}
