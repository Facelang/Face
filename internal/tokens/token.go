package tokens

type Token rune

const (
	EOF Token = (1 << 30) - iota
	ILLEGAL
	IDENT
	INT
	FLOAT
	Char
	STRING
	COMMENT
	RETURN
)
