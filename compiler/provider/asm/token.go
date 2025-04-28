package asm

import (
	"strconv"
)

type Token int

const (
	ILLEGAL Token = iota // 无效标记
	EOF                  // 文件结束标记
	COMMENT              // 文档注释符

	_literal // 字面量开始标记
	IDENT    // main
	NUMBER
	STRING
	_literalEnd // 字面量结束标记

	_operator    // 运算符
	ADD          // +
	SUB          // -
	LBRACK       // [
	COMMA        // ,
	RBRACK       // ]
	COLON        // :
	_operatorEnd // 操作符结束

	_keywords
	BREAK    // 1
	CHAR     // 2
	CONTINUE // 3

	ELSE   // 4
	EXTERN // 5

	IF     // 5
	IN     // 6
	INT    // 7
	OUT    // 8
	RETURN // 9
	VOID   // 11
	WHILE  // 12
	_keywordsEnd

	// 寄存器
	BR_AL
	BR_CL
	BR_DL
	BR_BL
	BR_AH
	BR_CH
	BR_DH
	BR_BH
	DR_EAX
	DR_ECX
	DR_EDX
	DR_EBX
	DR_ESP
	DR_EBP
	DR_ESI
	DR_EDI
	// 双操作数指令
	I_MOV
	I_CMP
	I_SUB
	I_ADD
	I_LEA
	// 单操作数指令
	I_CALL
	I_INT
	I_IMUL
	I_IDIV
	I_NEG
	I_INC
	I_DEC
	I_JMP
	I_JE
	I_JG
	I_JL
	I_JGE
	I_JLE
	I_JNE
	I_JNA
	I_PUSH
	I_POP
	// 零操作数指令
	I_RET
	// 汇编指令
	A_SEC
	A_GLB
	A_EQU
	A_TIMES
	A_DB
	A_DW
	A_DD
)

var tokens = [...]string{
	ILLEGAL:  "ILLEGAL",
	EOF:      "EOF",
	COMMENT:  "COMMENT",
	IDENT:    "IDENT",
	NUMBER:   "NUMBER",
	STRING:   "STRING",
	ADD:      "+",
	SUB:      "-",
	LBRACK:   "[",
	COMMA:    ",",
	RBRACK:   "]",
	COLON:    ":",
	BREAK:    "break",
	CHAR:     "char",
	CONTINUE: "continue",
	ELSE:     "else",
	EXTERN:   "extern",
	IF:       "if",
	IN:       "in",
	INT:      "int",
	OUT:      "out",
	RETURN:   "return",
	VOID:     "void",
	WHILE:    "while",

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

var keywordsList = []string{
	"al", "cl", "dl", "bl", "ah", "ch", "dh", "bh",
	"eax", "ecx", "edx", "ebx", "esp", "ebp", "esi", "edi",
	"mov", "cmp", "sub", "add", "lea",
	"call", "int", "imul", "idiv", "neg", "inc", "dec", "jmp", "je", "jg", "jl", "jge", "jle", "jne", "jna", "push", "pop",
	"ret",
	"section", "global", "equ", "times", "db", "dw", "dd",
	"text", "data", "bss", // 添加段名
}
var keywordsTable = []Token{
	BR_AL, BR_CL, BR_DL, BR_BL, BR_AH, BR_CH, BR_DH, BR_BH,
	DR_EAX, DR_ECX, DR_EDX, DR_EBX, DR_ESP, DR_EBP, DR_ESI, DR_EDI,
	I_MOV, I_CMP, I_SUB, I_ADD, I_LEA,
	I_CALL, I_INT, I_IMUL, I_IDIV, I_NEG, I_INC, I_DEC, I_JMP, I_JE, I_JG, I_JL, I_JGE, I_JLE, I_JNE, I_JNA, I_PUSH, I_POP,
	I_RET,
	A_SEC, A_GLB, A_EQU, A_TIMES, A_DB, A_DW, A_DD,
	IDENT, IDENT, IDENT, // 段名作为标识符处理
}

func Keywords(ident string) (Token, bool) {
	for i, k := range keywordsList {
		if k == ident {
			return keywordsTable[i], true
		}
	}
	return ILLEGAL, false
}

func (tok Token) IsLiteral() bool { return _literal < tok && tok < _literalEnd }
