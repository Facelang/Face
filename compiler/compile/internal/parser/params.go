package parser

import (
	"github.com/facelang/face/compiler/compile/internal/api"
	"go/ast"
	"go/token"
)

func ParamList(p *parser, acceptTParams bool) (tparams, params *ast.FieldList) {
	if acceptTParams && p.token == api.LBRACK {
		opening := p.FilePos
		p.next()
		// [T any](params) syntax
		list := p.parseParameterList(nil, nil, token.RBRACK)
		rbrack := p.expect(token.RBRACK)
		tparams = &ast.FieldList{Opening: opening, List: list, Closing: rbrack}
		// Type parameter lists must not be empty.
		if tparams.NumFields() == 0 {
			p.error(tparams.Closing, "empty type parameter list")
			tparams = nil // avoid follow-on errors
		}
	}

	opening := p.expect(token.LPAREN)

	var fields []*ast.Field
	if p.tok != token.RPAREN {
		fields = p.parseParameterList(nil, nil, token.RPAREN)
	}

	rparen := p.expect(token.RPAREN)
	params = &ast.FieldList{Opening: opening, List: fields, Closing: rparen}

	return
}
