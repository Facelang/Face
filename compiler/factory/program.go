package factory

func Program(file string) {
	p := parser{
		lex: &lexer{
			buffer: &buffer{},
		},
		table: &ProgTable{},
	}
	err := p.init(file)
	if err != nil {
		panic(err)
	}
	p.parse()
}
