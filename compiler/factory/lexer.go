package factory

import (
	"go/token"
)

type lexer struct {
	buf *buffer // 读取器
}

func (l *lexer) init(file string, src interface{}, errFunc ErrorFunc) {
	l.buf.init(file, src, errFunc)
}

func (l *lexer) scan() uint32 {
	ch, eof := l.buf.next()
	for ch <= ' ' {
		if eof {
			return 0
		}
		ch, eof = l.buf.next()
	}

	line, col := l.buf.line+1, l.buf.col+1
	l.buf.start()
	switch ch {
	case '"':
		return ReadString(s, s.file)
	case '`':
		return ReadRawText(s, s.file)
	case '\'':
		return ReadRune(s, s.file)
	case '(':
		return format(token.LPAREN, s.file)
	case '[':
		return format(token.LBRACK, s.file)
	case '{':
		return format(token.LBRACE, s.file)
	case ',':
		return format(token.COMMA, s.file)
	case ';':
		return format(token.SEMICOLON, s.file)
	case ')':
		return format(token.RPAREN, s.file)
	case ']':
		return format(token.RBRACK, s.file)
	case '}':
		return format(token.RBRACE, s.file)
	case ':':
		return switch2(s, s.file, token.COLON, token.DEFINE)
	case '.':
		if s.buf.test("..") {
			s.buf.skip(DoubleByteLen)
			return format(token.ELLIPSIS, s.file)
		}
		if isDecimal(rune(s.buf.peek())) {
			return ReadNumber(s, s.file, true)
		}
		return format(token.PERIOD, s.file)
	case '+': // +=, ++, +
		return switch3(s, s.file, token.ADD, token.ADD_ASSIGN, '+', token.INC)
	case '-': // -=, --, -
		return switch3(s, s.file, token.SUB, token.SUB_ASSIGN, '-', token.DEC)
	case '*': // *=, *
		return switch2(s, s.file, token.MUL, token.MUL_ASSIGN)
	case '/': // //, /*, /=, /
		if s.buf.peek() == '/' {
			// TODO 取注释
		}
		return switch2(s, s.file, token.QUO, token.QUO_ASSIGN)
	case '%':
		return switch2(s, s.file, token.REM, token.REM_ASSIGN)
	case '&':
		if s.buf.peek() == '^' {
			s.buf.skip(SingleByteLen)
			return switch2(s, s.file, token.AND_NOT, token.AND_NOT_ASSIGN)
		}
		return switch3(s, s.file, token.AND, token.AND_ASSIGN, '&', token.LAND)
	case '|':
		return switch3(s, s.file, token.OR, token.OR_ASSIGN, '|', token.LOR)
	case '^':
		return switch2(s, s.file, token.XOR, token.XOR_ASSIGN)
	case '<': // <<, <=, <<=, <
		return s.switch4(s, s.file, token.LSS, token.LEQ, '<', token.SHL, token.SHL_ASSIGN)
	case '>':
		return s.switch4(s, s.file, token.GTR, token.GEQ, '>', token.SHR, token.SHR_ASSIGN)
	case '=':
		return switch2(s, s.file, token.ASSIGN, token.EQL)
	case '!':
		return switch2(s, s.file, token.NOT, token.NEQ)
	//case '~':
	//	s.nextch()
	//	s.op, s.prec = Tilde, 0
	//	s.tok = _Operator
	default:
		//if isLetter(s.buf.ch) || s.buf.ch >= utf8.RuneSelf && s.atIdentChar(true) {
		//	s.nextch()
		//	s.ident()
		//	return
		//}
		if isLetter(s.buf.ch) {
			ReadIdent(s, s.file)
		}

		s.errorf("invalid character %#U", s.buf.ch)
		return token.ILLEGAL, token.Position{}, string(s.buf.ch)
	}
}
