package factory

import (
	"strconv"
)

type Token int

const (
	ILLEGAL Token = iota // 无效标记
	EOF                  // 文件结束标记
	COMMENT              // 文档注释符

	_literal    // 字面量开始标记
	IDENT       // main
	INT         // 12345
	FLOAT       // 123.45
	IMAG        // 123.45i
	CHAR        // 'a'
	STRING      // "abc"
	_literalEnd // 字面量结束标记

	_operator // 运算符
	ADD       // +
	SUB       // -
	MUL       // *
	QUO       // /
	REM       // %
	AND       // &
	OR        // |
	XOR       // ^
	SHL       // <<
	SHR       // >>

	AND_NOT        // &^
	ADD_ASSIGN     // +=
	SUB_ASSIGN     // -=
	MUL_ASSIGN     // *=
	QUO_ASSIGN     // /=
	REM_ASSIGN     // %=
	AND_ASSIGN     // &=
	OR_ASSIGN      // |=
	XOR_ASSIGN     // ^=
	SHL_ASSIGN     // <<=
	SHR_ASSIGN     // >>=
	AND_NOT_ASSIGN // &^=

	LAND     // &&
	LOR      // ||
	ARROW    // <-
	INC      // ++
	DEC      // --
	EQL      // ==
	LSS      // <
	GTR      // >
	ASSIGN   // =
	NOT      // !
	NEQ      // !=
	LEQ      // <=
	GEQ      // >=
	DEFINE   // :=
	ELLIPSIS // ... // #

	LPAREN       // (
	LBRACK       // [
	LBRACE       // {
	COMMA        // ,
	PERIOD       // .
	RPAREN       // )
	RBRACK       // ]
	RBRACE       // }
	COLON        // :
	_operatorEnd // 操作符结束

	additional_beg
	NEWLINE   // \n
	POUND     // #
	SEMICOLON // ;
	additional_end
)

const POINTER = MUL
const GENERIC_LEFT = LSS
const GENERIC_RIGHT = GTR

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",

	EOF:     "EOF",
	COMMENT: "COMMENT",

	IDENT:  "IDENT",
	INT:    "INT",
	FLOAT:  "FLOAT",
	IMAG:   "IMAG",
	CHAR:   "CHAR",
	STRING: "STRING",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	QUO: "/",
	REM: "%",

	AND:     "&",
	OR:      "|",
	XOR:     "^",
	SHL:     "<<",
	SHR:     ">>",
	AND_NOT: "&^",

	ADD_ASSIGN: "+=",
	SUB_ASSIGN: "-=",
	MUL_ASSIGN: "*=",
	QUO_ASSIGN: "/=",
	REM_ASSIGN: "%=",

	AND_ASSIGN:     "&=",
	OR_ASSIGN:      "|=",
	XOR_ASSIGN:     "^=",
	SHL_ASSIGN:     "<<=",
	SHR_ASSIGN:     ">>=",
	AND_NOT_ASSIGN: "&^=",

	LAND:  "&&",
	LOR:   "||",
	ARROW: "<-",
	INC:   "++",
	DEC:   "--",

	EQL:    "==",
	LSS:    "<",
	GTR:    ">",
	ASSIGN: "=",
	NOT:    "!",

	NEQ:      "!=",
	LEQ:      "<=",
	GEQ:      ">=",
	DEFINE:   ":=",
	ELLIPSIS: "...",

	LPAREN: "(",
	LBRACK: "[",
	LBRACE: "{",
	COMMA:  ",",
	PERIOD: ".",

	RPAREN:    ")",
	RBRACK:    "]",
	RBRACE:    "}",
	SEMICOLON: ";",
	COLON:     ":",
	//TILDE:     "~",
}

func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

const (
	LowestPrec  = 0 // non-operators
	UnaryPrec   = 6
	HighestPrec = 7
)

// Precedence returns the operator precedence of the binary
// operator op. If op is not a binary operator, the result
// is LowestPrecedence.
func (tok Token) Precedence() int {
	switch tok {
	case LOR:
		return 1
	case LAND:
		return 2
	case EQL, NEQ, LSS, LEQ, GTR, GEQ:
		return 3
	case ADD, SUB, OR, XOR:
		return 4
	case MUL, QUO, REM, SHL, SHR, AND, AND_NOT:
		return 5
	}
	return LowestPrec
}

//// Lookup maps an identifier to its keyword token or [IDENT] (if not a keyword).
//func Lookup(ident string) Token {
//	if tok, isKeyword := keywords[ident]; isKeyword {
//		return tok
//	}
//	return IDENT
//}

func (tok Token) IsLiteral() bool { return _literal < tok && tok < _literalEnd }

//func (tok Token) IsOperator() bool {
//	return (_operator < tok && tok < _operatorEnd) || tok == TILDE
//}
//
//func IsExported(name string) bool {
//	ch, _ := utf8.DecodeRuneInString(name)
//	return unicode.IsUpper(ch)
//}
//
//func IsIdentifier(name string) bool {
//	if name == "" || IsKeyword(name) {
//		return false
//	}
//	for i, c := range name {
//		if !unicode.IsLetter(c) && c != '_' && (i == 0 || !unicode.IsDigit(c)) {
//			return false
//		}
//	}
//	return true
//}
