package parser

func ExceptError(p *parser, except string) {
	p.errorf("except %s, found %s", except, p.token.Label(p.identifier))
}
