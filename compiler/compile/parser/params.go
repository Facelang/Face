package parser

import (
	"github.com/facelang/face/compiler/compile/tokens"
	"github.com/facelang/face/internal/prog"
)

// const a int = 1 // byte, int8, int16, int32, int64, uint8, uint16, uint32, uint64, bool, string,
// let b array<int> = [1,2,3]
// let b map<int, string> = {1: "a", 2: "b", 3: "c"}
// let c classA = {a: 1, b: 2, c: 3}

// ParamOrNil = [ IdentifierList ] [ "..." ] Type . 只在参数列表中调用 follow=close
func ParamOrNil(p *parser, name *prog.Name, follow tokens.Token) *prog.Field {

	pos := p.FilePos
	if name != nil {
		pos = name.Pos()
	}

	f := new(prog.Field)
	f.pos = pos

	if p.token == tokens.IDENT || name != nil {
		// name
		if name == nil {
			name = p.name()
		}

		if p.tok == _Dot { // name.***
			// name "." ...
			f.Type = p.qualifiedName(name)
			if typeSetsOk && p.tok == _Operator && p.op == Or {
				// name "." name "|" ...
				f = p.embeddedElem(f)
			}
			return f
		}

		if typeSetsOk && p.tok == _Operator && p.op == Or {
			// name "|" ...
			f.Type = name
			return p.embeddedElem(f)
		}

		f.Name = name
	}

	if p.token == prog.DotsType {
		// [name] "..." ...
		t := new(DotsType)
		t.pos = p.pos()
		p.next()
		t.Elem = p.typeOrNil()
		if t.Elem == nil {
			t.Elem = p.badExpr()
			p.syntaxError("... is missing type")
		}
		f.Type = t
		return f
	}

	if typeSetsOk && p.tok == _Operator && p.op == Tilde {
		// [name] "~" ...
		f.Type = p.embeddedElem(nil).Type
		return f
	}

	f.Type = p.typeOrNil()
	if typeSetsOk && p.tok == _Operator && p.op == Or && f.Type != nil {
		// [name] type "|"
		f = p.embeddedElem(f)
	}
	if f.Name != nil || f.Type != nil {
		return f
	}

	p.syntaxError("expected " + tokstring(follow))
	p.advance(_Comma, follow)
	return nil
}

// Parameters    = "(" [ ParameterList [ "," ] ] ")" .
// ParameterList = ParameterDecl { "," ParameterDecl } .
// "(" or "[" has already been consumed.
// If name != nil, it is the first name after "(" or "[".
// If typ != nil, name must be != nil, and (name, typ) is the first field in the list.
// In the result list, either all fields have a name, or no field has a name.

// p.paramList(nil, nil, _Rbrack, true)
func paramList(p *parser, close tokens.Token, requireNames bool) (list []*Field) {

	var named int // number of parameters that have an explicit name and type
	var typed int // number of parameters that have an explicit type
	end := p.list("parameter list", COMMA, close, func() bool {
		var par *prog.Field
		f := ParamOrNil(p)

		name = nil // 1st name was consumed if present
		typ = nil  // 1st type was consumed if present
		if par != nil {
			if debug && par.Name == nil && par.Type == nil {
				panic("parameter without name or type")
			}
			if par.Name != nil && par.Type != nil {
				named++
			}
			if par.Type != nil {
				typed++
			}
			list = append(list, par)
		}
		return false
	})

	if len(list) == 0 {
		return
	}

	// distribute parameter types (len(list) > 0)
	if named == 0 && !requireNames {
		// all unnamed and we're not in a type parameter list => found names are named types
		for _, par := range list {
			if typ := par.Name; typ != nil {
				par.Type = typ
				par.Name = nil
			}
		}
	} else if named != len(list) {
		// some named or we're in a type parameter list => all must be named
		var errPos Pos // left-most error position (or unknown)
		var typ Expr   // current type (from right to left)
		for i := len(list) - 1; i >= 0; i-- {
			par := list[i]
			if par.Type != nil {
				typ = par.Type
				if par.Name == nil {
					errPos = StartPos(typ)
					par.Name = NewName(errPos, "_")
				}
			} else if typ != nil {
				par.Type = typ
			} else {
				// par.Type == nil && typ == nil => we only have a par.Name
				errPos = par.Name.Pos()
				t := p.badExpr()
				t.pos = errPos // correct position
				par.Type = t
			}
		}
		if errPos.IsKnown() {
			// Not all parameters are named because named != len(list).
			// If named == typed, there must be parameters that have no types.
			// They must be at the end of the parameter list, otherwise types
			// would have been filled in by the right-to-left sweep above and
			// there would be no error.
			// If requireNames is set, the parameter list is a type parameter
			// list.
			var msg string
			if named == typed {
				errPos = end // position error at closing token ) or ]
				if requireNames {
					msg = "missing type constraint"
				} else {
					msg = "missing parameter type"
				}
			} else {
				if requireNames {
					msg = "missing type parameter name"
					// go.dev/issue/60812
					if len(list) == 1 {
						msg += " or invalid array length"
					}
				} else {
					msg = "missing parameter name"
				}
			}
			p.syntaxErrorAt(errPos, msg)
		}
	}

	return
}

func (p *parser) list(context string, sep, close tokens.Token, f func() bool) prog.FilePos {
	done := false

	for p.token != tokens.EOF && p.token != close && !done {
		done = f()

		if !p.got(sep) && p.token != close {
			p.errorf("list for %s; missing %s or %s", context, sep, close)
			return p.FilePos
		}
	}

	pos := p.FilePos
	p.expect(close)
	return pos
}
