package factory

func Program(file string) {
	p := parser{
		lex: &lexer{
			buf: &buffer{},
		},
	}
	p.init(file)
	p.parse()
}
