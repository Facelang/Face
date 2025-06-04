package parser

import (
	"github.com/facelang/face/compiler/compile/ast"
	"github.com/facelang/face/compiler/compile/tokens"
)

// ----------------------------------------------------------------------------
// Common productions

// inRhs = true 代表右侧表达式，否则为左侧表达式
func (p *parser) parseList(inRhs bool) []ast.Expr {
	old := p.inRhs
	p.inRhs = inRhs
	list := p.parseExprList()
	p.inRhs = old
	return list
}

func (p *parser) parseRhs() ast.Expr {
	old := p.inRhs
	p.inRhs = true
	x := p.parseExpr()
	p.inRhs = old
	return x
}

// If lhs is set, result list elements which are identifiers are not resolved.
func (p *parser) parseExprList() (list []ast.Expr) {
	list = append(list, p.parseExpr())
	for p.token == tokens.COMMA {
		p.next()
		list = append(list, p.parseExpr())
	}

	return
}

// ----------------------------------------------------------------------------
// Expressions

func (p *parser) parseFuncTypeOrLit() ast.Expr {

	typ := p.parseFuncType()
	if p.token != tokens.LBRACE {
		// function type only
		return typ
	}

	p.exprLev++
	body := p.parseBody()
	p.exprLev--

	return &ast.FuncLit{Type: typ, Body: body}
}

// parseOperand may return an expression or a raw type (incl. array
// types of the form [...]T). Callers must verify the result.
func (p *parser) parseOperand() ast.Expr {
	switch p.token {
	case tokens.IDENT: // 变量符号
		x := p.name()
		return x

	case tokens.INT, tokens.FLOAT, tokens.IMAG, tokens.CHAR, tokens.STRING: // 值类型
		x := &ast.BasicLit{Pos: 0, Kind: p.token, Value: p.identifier}
		p.next()
		return x

	case LPAREN: // (...) 多了一层优先级
		lparen := p.FilePos
		p.next()
		//p.exprLev++
		x := p.parseRhs() // types may be parenthesized: (some type)
		//p.exprLev--
		rparen := p.expect(RPAREN)
		return &ast.ParenExpr{Lparen: lparen, X: x, Rparen: rparen}

	case FUNC: // func ...
		return p.parseFuncTypeOrLit()
	}

	if typ := p.tryIdentOrType(); typ != nil { // do not consume trailing type parameters
		// could be type for composite literal or conversion
		_, isIdent := typ.(*ast.Ident)
		assert(!isIdent, "type cannot be identifier")
		return typ
	}

	// we have an error
	pos := p.pos
	p.errorExpected(pos, "operand")
	p.advance(stmtStart)
	return &ast.BadExpr{From: pos, To: p.pos}
}

// 语法解析中 x = nil !
func (p *parser) parsePrimaryExpr(x ast.Expr) ast.Expr {
	if x == nil {
		x = p.parseOperand()
	}
	// We track the nesting here rather than at the entry for the function,
	// since it can iteratively produce a nested output, and we want to
	// limit how deep a structure we generate.
	var n int
	defer func() { p.nestLev -= n }()
	for n = 1; ; n++ {
		incNestLev(p)
		switch p.tok {
		case PERIOD:
			p.next()
			switch p.tok {
			case IDENT:
				x = p.parseSelector(x)
			case LPAREN:
				x = p.parseTypeAssertion(x)
			default:
				pos := p.pos
				p.errorExpected(pos, "selector or type assertion")
				// TODO(rFindley) The check for RBRACE below is a targeted fix
				//                to error recovery sufficient to make the x/tools tests to
				//                pass with the new parsing logic introduced for type
				//                parameters. Remove this once error recovery has been
				//                more generally reconsidered.
				if p.tok != RBRACE {
					p.next() // make progress
				}
				sel := &ast.Ident{NamePos: pos, Name: "_"}
				x = &ast.SelectorExpr{X: x, Sel: sel}
			}
		case LBRACK:
			x = p.parseIndexOrSliceOrInstance(x)
		case LPAREN:
			x = p.parseCallOrConversion(x)
		case LBRACE:
			// operand may have returned a parenthesized complit
			// type; accept it but complain if we have a complit
			t := ast.Unparen(x)
			// determine if '{' belongs to a composite literal or a block statement
			switch t.(type) {
			case *ast.BadExpr, *ast.Ident, *ast.SelectorExpr:
				if p.exprLev < 0 {
					return x
				}
				// x is possibly a composite literal type
			case *ast.IndexExpr, *ast.IndexListExpr:
				if p.exprLev < 0 {
					return x
				}
				// x is possibly a composite literal type
			case *ast.ArrayType, *ast.StructType, *ast.MapType:
				// x is a composite literal type
			default:
				return x
			}
			if t != x {
				p.error(t.Pos(), "cannot parenthesize type in composite literal")
				// already progressed, no need to advance
			}
			x = p.parseLiteralValue(x)
		default:
			return x
		}
	}
}

func (p *parser) parseUnaryExpr() ast.Expr {
	//defer decNestLev(incNestLev(p))

	switch p.token {
	case ADD, SUB, NOT, XOR, AND, TILDE: // +, -, !, ^， ~
		pos, op := p.FilePos, p.token
		p.next()
		x := p.parseUnaryExpr() // 再解析...
		return &ast.UnaryExpr{OpPos: pos, Op: op, X: x}

	case ARROW: // 信号
		// channel type or receive expression
		arrow := p.pos
		p.next()

		// If the next token is CHAN we still don't know if it
		// is a channel type or a receive operation - we only know
		// once we have found the end of the unary expression. There
		// are two cases:
		//
		//   <- type  => (<-type) must be channel type
		//   <- expr  => <-(expr) is a receive from an expression
		//
		// In the first case, the arrow must be re-associated with
		// the channel type parsed already:
		//
		//   <- (chan type)    =>  (<-chan type)
		//   <- (chan<- type)  =>  (<-chan (<-type))

		x := p.parseUnaryExpr()

		// determine which case we have
		if typ, ok := x.(*ast.ChanType); ok {
			// (<-type)

			// re-associate position info and <-
			dir := ast.SEND
			for ok && dir == ast.SEND {
				if typ.Dir == ast.RECV {
					// error: (<-type) is (<-(<-chan T))
					p.errorExpected(typ.Arrow, "'chan'")
				}
				arrow, typ.Begin, typ.Arrow = typ.Arrow, arrow, arrow
				dir, typ.Dir = typ.Dir, ast.RECV
				typ, ok = typ.Value.(*ast.ChanType)
			}
			if dir == ast.SEND {
				p.errorExpected(arrow, "channel type")
			}

			return x
		}

		// <-(expr)
		return &ast.UnaryExpr{OpPos: arrow, Op: ARROW, X: x}

	case MUL: // * ...
		// pointer type or unary "*" expression
		pos := p.pos
		p.next()
		x := p.parseUnaryExpr()
		return &ast.StarExpr{Star: pos, X: x}
	}

	return p.parsePrimaryExpr(nil) // 更低级表达式
}

func (p *parser) tokPrec() (Token, int) {
	tok := p.token
	if p.inRhs && tok == ASSIGN {
		tok = EQL
	}
	return tok, tok.Precedence() // 这个应该是优先级
}

// parseBinaryExpr parses a (possibly) binary expression.
// If x is non-nil, it is used as the left operand.
//
// TODO(rfindley): parseBinaryExpr has become overloaded. Consider refactoring.
func (p *parser) parseBinaryExpr(x ast.Expr, prec1 int) ast.Expr {
	if x == nil { // 第一次调用为空， 一定会执行
		x = p.parseUnaryExpr() // 先取一元表达式
	}
	// ....
	// We track the nesting here rather than at the entry for the function,
	// since it can iteratively produce a nested output, and we want to
	// limit how deep a structure we generate.
	var n int
	defer func() { p.nestLev -= n }()
	for n = 1; ; n++ {
		incNestLev(p)
		op, oprec := p.tokPrec()
		if oprec < prec1 {
			return x
		}
		pos := p.expect(op)
		y := p.parseBinaryExpr(nil, oprec+1)
		x = &ast.BinaryExpr{X: x, OpPos: pos, Op: op, Y: y}
	}
}

// The result may be a type or even a raw type ([...]int).
// expr() -> binaryExpr() -> unaryExpr() -> pexpr() -> operand()
// 从高到低： 二元运算符优先级最高, 其次一元运算符, 其他运算符, 操作数
// 二元运算符 还需要进一步判断优先级
func (p *parser) parseExpr() ast.Expr {
	return p.parseBinaryExpr(nil, LowestPrec+1) // 最低优先级？
}
