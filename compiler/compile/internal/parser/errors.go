package parser

func ExceptError(p *parser, except string) {
	if p.token.IsLiteral() {
		p.errorf("except %s, found %s", except, p.identifier)
	} else {
		p.errorf("except %s, found %s", except, p.token)
	}
}
