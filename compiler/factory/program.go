package factory

func Program(file string) {
	p := Parser(file)
	_, err := p.Parse()
	if err != nil {
		panic(err)
	}
}
