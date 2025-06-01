package tokens

type Token rune

const (
	EOF     Token = (1 << 30) - iota // 结束
	ILLEGAL                          // error
	IDENT                            // label
	INT                              // 123456
	FLOAT                            // 123.456
	Char                             // ''
	STRING                           // "", ``
	COMMENT                          //
	RETURN                           // \n
)
