package tokens

import (
	"strconv"
)

type Token rune

const (
	ILLEGAL Token = (1 << 7) - iota // error
	EOF                             // 结束
	COMMENT                         // 注释
	NEWLINE                         // \n 换行符
	IDENT                           // label
	INT                             // 123456
	FLOAT                           // 123.456
	IMAG                            // 123.1i 复数
	CHAR                            // ''
	STRING                          // "", ``
)

var NameTable = [...]string{
	ILLEGAL: "ILLEGAL",

	EOF:     "EOF",
	COMMENT: "COMMENT",
	NEWLINE: "NEWLINE",

	IDENT:  "IDENT",
	INT:    "INT",
	FLOAT:  "FLOAT",
	IMAG:   "IMAG",
	CHAR:   "CHAR",
	STRING: "STRING",
}

func (token Token) String() (name string) {
	if token >= 0 && token <= ILLEGAL {
		name = NameTable[token]
	}
	if name == "" {
		name = "token(" + strconv.Itoa(int(token)) + ")"
	}
	return name
}

func (token Token) IsLiteral() bool { return IDENT <= token && token <= STRING }
