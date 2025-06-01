package internal

import "github.com/facelang/face/internal/tokens"

const (
	operator_begin tokens.Token = iota // 运算符
	ADD                                // +
	SUB                                // -
	MUL                                // *
	QUO                                // /
	REM                                // %

	AND     // &
	OR      // |
	XOR     // ^
	SHL     // <<
	SHR     // >>
	AND_NOT // &^

	ADD_ASSIGN // +=
	SUB_ASSIGN // -=
	MUL_ASSIGN // *=
	QUO_ASSIGN // /=
	REM_ASSIGN // %=

	AND_ASSIGN     // &=
	OR_ASSIGN      // |=
	XOR_ASSIGN     // ^=
	SHL_ASSIGN     // <<=
	SHR_ASSIGN     // >>=
	AND_NOT_ASSIGN // &^=

	LAND  // &&
	LOR   // ||
	ARROW // <-
	INC   // ++
	DEC   // --

	EQL    // ==
	LSS    // <
	GTR    // >
	ASSIGN // =
	NOT    // !

	NEQ      // !=
	LEQ      // <=
	GEQ      // >=
	DEFINE   // :=
	ELLIPSIS // ...

	LPAREN // (
	LBRACK // [
	LBRACE // {
	COMMA  // ,
	PERIOD // .

	RPAREN       // )
	RBRACK       // ]
	RBRACE       // }
	SEMICOLON    // ;
	COLON        // :
	operator_end // 操作符结束

	keyword_beg
	BREAK
	CASE
	CHAN
	CONST
	CONTINUE

	DEFAULT
	DEFER
	ELSE
	FALLTHROUGH
	FOR

	FUNC
	GO
	GOTO
	IF
	IMPORT

	INTERFACE
	MAP
	PACKAGE
	RANGE
	RETURN

	SELECT
	STRUCT
	SWITCH
	TYPE
	VAR
	keyword_end
)

func init() {
	// 基本运算符
	tokens.NameTable[ADD] = "+"
	tokens.NameTable[SUB] = "-"
	tokens.NameTable[MUL] = "*"
	tokens.NameTable[QUO] = "/"
	tokens.NameTable[REM] = "%"

	// 位运算符
	tokens.NameTable[AND] = "&"
	tokens.NameTable[OR] = "|"
	tokens.NameTable[XOR] = "^"
	tokens.NameTable[SHL] = "<<"
	tokens.NameTable[SHR] = ">>"
	tokens.NameTable[AND_NOT] = "&^"

	// 复合赋值运算符
	tokens.NameTable[ADD_ASSIGN] = "+="
	tokens.NameTable[SUB_ASSIGN] = "-="
	tokens.NameTable[MUL_ASSIGN] = "*="
	tokens.NameTable[QUO_ASSIGN] = "/="
	tokens.NameTable[REM_ASSIGN] = "%="
	tokens.NameTable[AND_ASSIGN] = "&="
	tokens.NameTable[OR_ASSIGN] = "|="
	tokens.NameTable[XOR_ASSIGN] = "^="
	tokens.NameTable[SHL_ASSIGN] = "<<="
	tokens.NameTable[SHR_ASSIGN] = ">>="
	tokens.NameTable[AND_NOT_ASSIGN] = "&^="

	// 逻辑运算符
	tokens.NameTable[LAND] = "&&"
	tokens.NameTable[LOR] = "||"
	tokens.NameTable[ARROW] = "<-"
	tokens.NameTable[INC] = "++"
	tokens.NameTable[DEC] = "--"

	// 比较运算符
	tokens.NameTable[EQL] = "=="
	tokens.NameTable[LSS] = "<"
	tokens.NameTable[GTR] = ">"
	tokens.NameTable[ASSIGN] = "="
	tokens.NameTable[NOT] = "!"
	tokens.NameTable[NEQ] = "!="
	tokens.NameTable[LEQ] = "<="
	tokens.NameTable[GEQ] = ">="
	tokens.NameTable[DEFINE] = ":="
	tokens.NameTable[ELLIPSIS] = "..."

	// 分隔符
	tokens.NameTable[LPAREN] = "("
	tokens.NameTable[LBRACK] = "["
	tokens.NameTable[LBRACE] = "{"
	tokens.NameTable[COMMA] = ","
	tokens.NameTable[PERIOD] = "."
	tokens.NameTable[RPAREN] = ")"
	tokens.NameTable[RBRACK] = "]"
	tokens.NameTable[RBRACE] = "}"
	tokens.NameTable[SEMICOLON] = ";"
	tokens.NameTable[COLON] = ":"

	// 关键字映射
	tokens.NameTable[BREAK] = "break"
	tokens.NameTable[CASE] = "case"
	tokens.NameTable[CHAN] = "chan"
	tokens.NameTable[CONST] = "const"
	tokens.NameTable[CONTINUE] = "continue"
	tokens.NameTable[DEFAULT] = "default"
	tokens.NameTable[DEFER] = "defer"
	tokens.NameTable[ELSE] = "else"
	tokens.NameTable[FALLTHROUGH] = "fallthrough"
	tokens.NameTable[FOR] = "for"
	tokens.NameTable[FUNC] = "func"
	tokens.NameTable[GO] = "go"
	tokens.NameTable[GOTO] = "goto"
	tokens.NameTable[IF] = "if"
	tokens.NameTable[IMPORT] = "import"
	tokens.NameTable[INTERFACE] = "interface"
	tokens.NameTable[MAP] = "map"
	tokens.NameTable[PACKAGE] = "package"
	tokens.NameTable[RANGE] = "range"
	tokens.NameTable[RETURN] = "return"
	tokens.NameTable[SELECT] = "select"
	tokens.NameTable[STRUCT] = "struct"
	tokens.NameTable[SWITCH] = "switch"
	tokens.NameTable[TYPE] = "type"
	tokens.NameTable[VAR] = "var"
}

var keywordsList = []string{
	"break",       // BREAK
	"case",        // CASE
	"chan",        // CHAN
	"const",       // CONST
	"continue",    // CONTINUE
	"default",     // DEFAULT
	"defer",       // DEFER
	"else",        // ELSE
	"fallthrough", // FALLTHROUGH
	"for",         // FOR
	"func",        // FUNC
	"go",          // GO
	"goto",        // GOTO
	"if",          // IF
	"import",      // IMPORT
	"interface",   // INTERFACE
	"map",         // MAP
	"package",     // PACKAGE
	"range",       // RANGE
	"return",      // RETURN
	"select",      // SELECT
	"struct",      // STRUCT
	"switch",      // SWITCH
	"type",        // TYPE
	"var",         // VAR
}

var keywordsTable = []tokens.Token{
	BREAK,
	CASE,
	CHAN,
	CONST,
	CONTINUE,
	DEFAULT,
	DEFER,
	ELSE,
	FALLTHROUGH,
	FOR,
	FUNC,
	GO,
	GOTO,
	IF,
	IMPORT,
	INTERFACE,
	MAP,
	PACKAGE,
	RANGE,
	RETURN,
	SELECT,
	STRUCT,
	SWITCH,
	TYPE,
	VAR,
}

func Keywords(ident string) (tokens.Token, bool) {
	for i, k := range keywordsList {
		if k == ident {
			return keywordsTable[i], true
		}
	}
	return tokens.ILLEGAL, false
}

//
//func (tok Token) String() string {
//	s := ""
//	if 0 <= tok && tok < Token(len(tokens)) {
//		s = tokens[tok]
//	}
//	if s == "" {
//		s = "token(" + strconv.Itoa(int(tok)) + ")"
//	}
//	return s
//}
//
//const (
//	LowestPrec  = 0 // non-operators
//	UnaryPrec   = 6
//	HighestPrec = 7
//)
//
//// Precedence returns the operator precedence of the binary
//// operator op. If op is not a binary operator, the result
//// is LowestPrecedence.
//func (tok Token) Precedence() int {
//	switch tok {
//	case LOR:
//		return 1
//	case LAND:
//		return 2
//	case EQL, NEQ, LSS, LEQ, GTR, GEQ:
//		return 3
//	case ADD, SUB, OR, XOR:
//		return 4
//	case MUL, QUO, REM, SHL, SHR, AND, AND_NOT:
//		return 5
//	}
//	return LowestPrec
//}
//
//var keywordsList = []string{"break", "char", "continue", "else", "extern", "if", "in", "int", "out", "return", "string", "void", "while"}
//var keywordsTable = []Token{BREAK, CHAR, CONTINUE, ELSE, EXTERN, IF, IN, INT, OUT, RETURN, STRING, VOID, WHILE}
//
//func Keywords(ident string) (Token, bool) {
//	for i, k := range keywordsList {
//		if k == ident {
//			return keywordsTable[i], true
//		}
//	}
//	return ILLEGAL, false
//}
//
//func (tok Token) IsLiteral() bool { return _literal < tok && tok < _literalEnd }

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
