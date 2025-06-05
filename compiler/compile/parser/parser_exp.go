package parser

import (
	"github.com/facelang/face/compiler/compile/ast"
	"github.com/facelang/face/compiler/compile/tokens"
	"go/token"
)

// maxNestLev is the deepest we're willing to recurse during parsing
const maxNestLev int = 1e5

func incNestLev(p *parser) *parser {
	p.nestLev++
	if p.nestLev > maxNestLev {
		p.error(p.pos, "exceeded max nesting depth")
	}
	return p
}

// decNestLev is used to track nesting depth during parsing to prevent stack exhaustion.
// It is used along with incNestLev in a similar fashion to how un and trace are used.
func decNestLev(p *parser) {
	p.nestLev--
}

// ----------------------------------------------------------------------------
// Common productions

// inRhs = true 代表右侧表达式，否则为左侧表达式
func exprList(p *parser, inRhs bool) []ast.Expr {
	old := p.inRhs
	p.inRhs = inRhs

	list := []ast.Expr{expr(p)}
	for p.token == tokens.COMMA {
		p.next()
		list = append(list, expr(p))
	}

	p.inRhs = old
	return list
}

// ----------------------------------------------------------------------------
// Expressions

//func (p *parser) parseFuncTypeOrLit() ast.Expr {
//
//	typ := p.parseFuncType()
//	if p.token != tokens.LBRACE {
//		// function type only
//		return typ
//	}
//
//	p.exprLev++
//	body := p.parseBody()
//	p.exprLev--
//
//	return &ast.FuncLit{Type: typ, Body: body}
//}

// operand may return an expression or a raw type (incl. array
// types of the form [...]T). Callers must verify the result.
func operand(p *parser) ast.Expr {
	switch p.token {
	case tokens.IDENT: // 变量符号
		x := p.name()
		return x

	case tokens.INT, tokens.FLOAT, tokens.IMAG, tokens.CHAR, tokens.STRING: // 值类型
		x := &ast.BasicLit{Pos: 0, Kind: p.token, Value: p.identifier}
		p.next()
		return x

	case tokens.LPAREN: // (...) 多了一层优先级
		lparen := p.pos
		p.next()
		//p.exprLev++
		x := exprRhs(p) // types may be parenthesized: (some type)
		//p.exprLev--
		rparen := p.expect(tokens.RPAREN)
		return &ast.ParenExpr{Lparen: lparen, X: x, Rparen: rparen}

		//case tokens.FUNC: // func ...
		//	return p.parseFuncTypeOrLit() // todo 暂时忽略
	}

	// 上面都是具体值类型
	// 下面是数据类型、关键字一类

	// 类型转换 int(123), []string{"a", "b", "c"}
	if typ := p.tryIdentOrType(); typ != nil { // do not consume trailing type parameters
		// could be type for composite literal or conversion
		if _, isIdent := typ.(*ast.Name); !isIdent {
			p.error(p.pos, "type cannot be identifier")
		}
		return typ
	}

	// we have an error
	pos := p.pos
	p.unexpect("operand")
	return &ast.BadExpr{From: pos, To: p.pos}
}

// 只在 parseElement 被调用
func (p *parser) parseValue() ast.Expr {
	if p.token == tokens.LBRACE {
		return p.parseLiteralValue(nil)
	}

	return expr(p)
}

// 只在 parseElementList 被调用
func (p *parser) parseElement() ast.Expr {
	x := p.parseValue()
	if p.token == tokens.COLON {
		colon := p.pos
		p.next()
		x = &ast.KeyValueExpr{Key: x, Colon: colon, Value: p.parseValue()}
	}

	return x
}

func (p *parser) parseElementList() (list []ast.Expr) {
	for p.token != tokens.RBRACE && p.token != tokens.EOF {
		list = append(list, p.parseElement())
		if p.token != tokens.COMMA {
			break
		}
		p.next()
	}

	return
}

// 解析复合字面量， {1, 2, 3} {key: value} 类型
func (p *parser) parseLiteralValue(typ ast.Expr) ast.Expr {
	defer decNestLev(incNestLev(p))

	lbrace := p.expect(tokens.LBRACE)
	var elts []ast.Expr
	//p.exprLev++
	if p.token != tokens.RBRACE {
		elts = p.parseElementList()
	}
	//p.exprLev--
	rbrace := p.expect(tokens.RBRACE)
	return &ast.CompositeLit{Type: typ, Lbrace: lbrace, Elts: elts, Rbrace: rbrace}
}

// packIndexExpr returns an IndexExpr x[expr0] or IndexListExpr x[expr0, ...].
func packIndexExpr(x ast.Expr, lbrack tokens.Pos, exprs []ast.Expr, rbrack tokens.Pos) ast.Expr {
	switch len(exprs) {
	case 0:
		panic("internal error: packIndexExpr with empty expr slice")
	case 1:
		return &ast.IndexExpr{
			X:      x,
			Lbrack: lbrack,
			Index:  exprs[0],
			Rbrack: rbrack,
		}
	default:
		return &ast.IndexListExpr{
			X:       x,
			Lbrack:  lbrack,
			Indices: exprs,
			Rbrack:  rbrack,
		}
	}
}

func (p *parser) parseIndexOrSliceOrInstance(x ast.Expr) ast.Expr {
	lbrack := p.expect(tokens.LBRACK)
	if p.token == tokens.RBRACK { // 直接结束， 抛异常
		p.unexpect("[operand is empty]")
		rbrack := p.pos
		p.next()
		return &ast.IndexExpr{
			X:      x,
			Lbrack: lbrack,
			Index:  &ast.BadExpr{From: rbrack, To: rbrack},
			Rbrack: rbrack,
		}
	}
	//p.exprLev++

	const N = 3         // [index] [:] [::]
	var args []ast.Expr // 值类型 [1, 2, 3]
	var index [N]ast.Expr
	var colons [N - 1]tokens.Pos
	if p.token != tokens.COLON {
		index[0] = exprRhs(p)
	}
	ncolons := 0
	switch p.token {
	case tokens.COLON:
		// slice expression
		for p.token == tokens.COLON && ncolons < len(colons) {
			colons[ncolons] = p.pos
			ncolons++
			p.next()
			if p.token != tokens.COLON && p.token != tokens.RBRACK && p.token != tokens.EOF {
				index[ncolons] = exprRhs(p)
			}
		}
	case tokens.COMMA: // ,
		// instance expression
		args = append(args, index[0])
		for p.token == tokens.COMMA {
			p.next()
			if p.token != tokens.RBRACK && p.token != tokens.EOF {
				args = append(args, p.parseType())
			}
		}
	}

	// p.exprLev--
	rbrack := p.expect(tokens.RBRACK)

	if ncolons > 0 { // 切片类型
		// slice expression
		slice3 := false
		if ncolons == 2 {
			slice3 = true
			// Check presence of middle and final index here rather than during type-checking
			// to prevent erroneous programs from passing through gofmt (was go.dev/issue/7305).
			if index[1] == nil {
				p.error(colons[0], "middle index required in 3-index slice")
				index[1] = &ast.BadExpr{From: colons[0] + 1, To: colons[1]}
			}
			if index[2] == nil {
				p.error(colons[1], "final index required in 3-index slice")
				index[2] = &ast.BadExpr{From: colons[1] + 1, To: rbrack}
			}
		}
		return &ast.SliceExpr{X: x, Lbrack: lbrack, Low: index[0], High: index[1], Max: index[2], Slice3: slice3, Rbrack: rbrack}
	}

	if len(args) == 0 {
		// index expression
		return &ast.IndexExpr{X: x, Lbrack: lbrack, Index: index[0], Rbrack: rbrack}
	}

	// instance expression
	return packIndexExpr(x, lbrack, args, rbrack)
}

// 函数调用或类型转换，类型转换本身就是一种函数调用
func (p *parser) funcCall(fun ast.Expr) *ast.CallExpr {
	lparen := p.expect(tokens.LPAREN) // 开始
	//p.exprLev++
	var list []ast.Expr
	var ellipsis tokens.Pos
	for p.token != tokens.RPAREN && p.token != tokens.EOF && !ellipsis.IsValid() {
		list = append(list, exprRhs(p)) // builtins may expect a type: make(some type, ...)
		if p.token == tokens.ELLIPSIS {
			ellipsis = p.pos
			p.next()
		}

		// 逗号，继续解析下一个参数， 否则结束
		if p.token != tokens.COMMA {
			break
		}
		p.next()
	}
	//p.exprLev--
	rparen := p.expect(tokens.RPAREN) // 关闭

	return &ast.CallExpr{Fun: fun, Lparen: lparen, Args: list, Ellipsis: ellipsis, Rparen: rparen}
}

// 处理后缀表达式， 比如： x.name, x[123]
func primaryExpr(p *parser, x ast.Expr) ast.Expr {
	if x == nil {
		x = operand(p)
	}

	var n int
	//defer func() { p.nestLev -= n }()
	for n = 1; ; n++ { // 持续++
		//incNestLev(p)
		switch p.token {
		case tokens.PERIOD: // x. 只能接 ident
			p.next()
			x = &ast.SelectorExpr{X: x, Sel: p.name()}
		case tokens.LBRACK: // x[...], x[1], x[:]
			x = p.parseIndexOrSliceOrInstance(x) // todo
		case tokens.LPAREN: // x(...), 函数调用或类型转换
			x = p.funcCall(x)
		case tokens.LBRACE: // todo {} 什么意思？
			// operand may have returned a parenthesized complit
			// type; accept it but complain if we have a complit
			t := ast.Unparen(x) // 解括号 (), 获取 x 真实类型
			// determine if '{' belongs to a composite literal or a block statement
			switch t.(type) { // 一些特殊情况直接返回 x, 其它情况，需要继续解析
			case *ast.BadExpr, *ast.Name, *ast.SelectorExpr: // 有条件解析
				//if p.exprLev < 0 { // 有一些解析过程会将 exprLev = -1
				//	return x
				//}
				// x is possibly a composite literal type
			case *ast.IndexExpr, *ast.IndexListExpr: // 有条件解析
				//if p.exprLev < 0 {
				//	return x
				//}
				// x is possibly a composite literal type
			case *ast.ArrayType, *ast.StructType, *ast.MapType:
				// x is a composite literal type
				// 数组，结构体， 字典，直接解析
			default:
				return x
			}
			if t != x {
				p.error(t.Offset(), "cannot parenthesize type in composite literal")
				// already progressed, no need to advance
			}
			x = p.parseLiteralValue(x) // todo 已实现，可能不需要
		default:
			return x
		}
	}
}

// 一元运算符， go 支持 <- 和 *, 目前仅支持 +-!&|
func unaryExpr(p *parser) ast.Expr {
	defer decNestLev(incNestLev(p))

	switch p.token {
	case tokens.ADD, tokens.SUB, tokens.NOT, tokens.XOR, tokens.AND, tokens.TILDE: // +, -, !, ^， ~
		pos, op := p.pos, p.token
		p.next()
		x := unaryExpr(p) // 再解析...
		return &ast.UnaryExpr{OpPos: pos, Op: op, X: x}
	}

	return primaryExpr(p, nil) // 更低级表达式
}

// 获得 token 和 优先级； 特例：将右值表达式中的 赋值符号 视为 ==
func precedence(p *parser) (tokens.Token, int) {
	tok := p.token
	if p.inRhs && tok == tokens.ASSIGN {
		tok = tokens.EQL
	}
	return tok, tok.Precedence() // 这个应该是优先级
}

// 二元表达式
func binaryExpr(p *parser, x ast.Expr, prec1 int) ast.Expr {
	if x == nil { // 第一次调用为空， 一定会执行
		x = unaryExpr(p) // 先取一元表达式
	}

	var n int
	defer func() { p.nestLev -= n }()
	for n = 1; ; n++ {
		incNestLev(p)
		// 判断优先级
		op, oprec := precedence(p)
		if oprec < prec1 { // 传入优先级 会 +1, 所以相同优先级会终止
			return x
		}
		pos := p.expect(op)
		y := binaryExpr(p, nil, oprec+1) // 优先级 +1, 同优先级，直接返回
		x = &ast.BinaryExpr{X: x, OpPos: pos, Op: op, Y: y}
	}
}

func exprRhs(p *parser) ast.Expr {
	old := p.inRhs
	p.inRhs = true
	x := expr(p)
	p.inRhs = old
	return x
}

// The result may be a type or even a raw type ([...]int).
// expr() -> binaryExpr() -> unaryExpr() -> pexpr() -> operand()
// 从高到低： 二元运算符优先级最高, 其次一元运算符, 其他运算符, 操作数
// 二元运算符 还需要进一步判断优先级
func expr(p *parser) ast.Expr {
	return binaryExpr(p, nil, tokens.LowestPrec+1) // 最低优先级？
}

type field struct {
	name *ast.Name
	typ  ast.Expr
}

func (p *parser) parseDotsType() *ast.Ellipsis {
	pos := p.expect(tokens.ELLIPSIS)
	elt := p.parseType()

	return &ast.Ellipsis{Ellipsis: pos, Elt: elt}
}

// 解析单条参数， name 一般为空（大部分时间）， typesetsok 一般为 false
func (p *parser) parseParamDecl(name *ast.Name, typeSetsOK bool) (f field) {

	ptok := p.token
	if name != nil { // 有参数名， 强制 tokens.IDENT
		p.token = tokens.IDENT // force tokens.IDENT case in switch below
	} else if typeSetsOK && p.token == tokens.TILDE {
		// "~" ...
		return field{nil, p.embeddedElem(nil)}
	}

	switch p.token { // 判断符号类型
	case tokens.IDENT:
		// name
		if name != nil {
			f.name = name
			p.token = ptok // 暂存， 恢复后尝试解析类型
		} else {
			f.name = p.name() // 解析参数名
		}
		switch p.token { // 再次判断符号
		case tokens.IDENT, tokens.MUL, tokens.ARROW, tokens.FUNC, tokens.CHAN, tokens.MAP, tokens.STRUCT, tokens.INTERFACE, tokens.LPAREN:
			// name type
			f.typ = p.parseType() // 解析符号

		case tokens.LBRACK: // [] 数组类型
			// name "[" type1, ..., typeN "]" or name "[" n "]" type
			f.name, f.typ = p.parseArrayFieldOrTypeInstance(f.name)

		case tokens.ELLIPSIS: // ... 可变参数
			// name "..." type
			f.typ = p.parseDotsType()
			return // don't allow ...type "|" ...

		case tokens.PERIOD: // . 选择器 name.xxx, 这种一定判定为 类型， 而不是参数名
			// name "." ...
			f.typ = p.parseQualifiedIdent(f.name)
			f.name = nil

		case tokens.TILDE: // ~ 类型约束
			if typeSetsOK {
				f.typ = p.embeddedElem(nil)
				return
			}

		case tokens.OR: // | 类型约束
			if typeSetsOK {
				// name "|" typeset
				f.typ = p.embeddedElem(f.name)
				f.name = nil
				return
			}
		}

	case tokens.MUL, tokens.ARROW, tokens.FUNC, tokens.LBRACK, tokens.CHAN, tokens.MAP, tokens.STRUCT, tokens.INTERFACE, tokens.LPAREN:
		// type
		f.typ = p.parseType()

	case tokens.ELLIPSIS:
		// "..." type
		// (always accepted)
		f.typ = p.parseDotsType()
		return // don't allow ...type "|" ...

	default:
		// TODO(rfindley): this is incorrect in the case of type parameter lists
		//                 (should be "']'" in that case)
		p.unexpect("')'")
	}

	// [name] type "|"
	if typeSetsOK && p.token == tokens.OR && f.typ != nil {
		f.typ = p.embeddedElem(f.typ)
	}

	return
}

// 多处调用， 默认调用 name0, type0 = nil ] or )
// parseMethodSpec中 name0 != nil, typ0 = nil ]
// parseGenericType中 name0, typ0 != nil ]
func (p *parser) parseParameterList(name0 *ast.Name, typ0 ast.Expr, closing tokens.Token) (params []*ast.Field) {
	// Type parameters are the only parameter list closed by ']'.
	tparams := closing == tokens.RBRACK // 是否是泛型参数

	pos0 := p.pos
	if name0 != nil {
		pos0 = name0.Offset()
	} else if typ0 != nil {
		pos0 = typ0.Offset()
	}

	// Note: The code below matches the corresponding code in the syntax
	//       parser closely. Changes must be reflected in either parser.
	//       For the code to match, we use the local []field list that
	//       corresponds to []syntax.Field. At the end, the list must be
	//       converted into an []*ast.Field.

	var list []field
	var named int // number of parameters that have an explicit name and type
	var typed int // number of parameters that have an explicit type

	// todo 第一个参数不为空，或者不是结束符，则继续解析
	//       p.tok != closing, 就会一直循环
	for name0 != nil || p.token != closing && p.token != tokens.EOF {
		var par field
		if typ0 != nil { // todo 有泛型参数的情况
			if tparams {
				typ0 = p.embeddedElem(typ0)
			}
			par = field{name0, typ0}
		} else { // 主要解析过程， 解析单条参数
			par = p.parseParamDecl(name0, tparams) // name0 可能为空
		}
		name0 = nil                            // 1st name was consumed if present // 第一次使用后删除
		typ0 = nil                             // 1st typ was consumed if present // 第一次使用后删除
		if par.name != nil || par.typ != nil { // 解析到参数，添加到list， 并统计（参数数量和类型数量）
			list = append(list, par)
			if par.name != nil && par.typ != nil {
				named++
			}
			if par.typ != nil { // 参数名可以为空？
				typed++
			}
			// todo 实际解析， 单类型参数，会被解析为 par.name && par.typ = nil
		}
		if p.token != tokens.COMMA {
			break
		}
		p.next() // 取下一个符号，继续解析
	}

	if len(list) == 0 {
		return // not uncommon
	}

	// distribute parameter types (len(list) > 0)
	if named == 0 { // 处理未命名参数， 声明段，可以不命名参数
		// all unnamed => found names are type names
		for i := 0; i < len(list); i++ { // 类似 func(int, string) 这样的会被解析为 只有 name, 需要转为 仅 type
			par := &list[i]
			if typ := par.name; typ != nil {
				par.typ = typ
				par.name = nil
			}
		}
		if tparams { // 一般为 false, 处理单泛型类型（没有类型约束）Class[T, B, C]， 直接抛出异常？？？
			// This is the same error handling as below, adjusted for type parameters only.
			// See comment below for details. (go.dev/issue/64534)
			var errPos tokens.Pos
			var msg string
			if named == typed /* same as typed == 0 */ {
				errPos = p.pos // position error at closing ]
				msg = "missing type constraint"
			} else {
				errPos = pos0 // position at opening [ or first name
				msg = "missing type parameter name"
				if len(list) == 1 {
					msg += " or invalid array length"
				}
			}
			p.error(errPos, msg)
		}
	} else if named != len(list) { // 类似 ？？ func (a, b, c int)
		// some named or we're in a type parameter list => all must be named
		var errPos tokens.Pos                 // left-most error position (or invalid)
		var typ ast.Expr                      // current type (from right to left)
		for i := len(list) - 1; i >= 0; i-- { // 从右向左扫描参数列表
			if par := &list[i]; par.typ != nil { // par.typ != nil 记录类型，向前
				typ = par.typ
				if par.name == nil { // 参数名为空？
					errPos = typ.Offset() // 记录一个异常
					n := &ast.Name{Pos: errPos, Name: "_"}
					par.name = n       // 记录一个 _ 下划线变量
				}
			} else if typ != nil { // par.typ == nil && typ != nil
				par.typ = typ
			} else {
				// par.typ == nil && typ == nil => we only have a par.name
				errPos = par.name.Offset()
				par.typ = &ast.BadExpr{From: errPos, To: p.pos}
			}
		}
		if errPos.IsValid() { // par.name == nil || typ == nil && par.typ == nil
			// Not all parameters are named because named != len(list).
			// If named == typed, there must be parameters that have no types.
			// They must be at the end of the parameter list, otherwise types
			// would have been filled in by the right-to-left sweep above and
			// there would be no error.
			// If tparams is set, the parameter list is a type parameter list.
			var msg string
			if named == typed {
				errPos = p.pos // position error at closing token ) or ]
				if tparams {
					msg = "missing type constraint"
				} else {
					msg = "missing parameter type"
				}
			} else {
				if tparams {
					msg = "missing type parameter name"
					// go.dev/issue/60812
					if len(list) == 1 {
						msg += " or invalid array length"
					}
				} else {
					msg = "missing parameter name"
				}
			}
			p.error(errPos, msg)
		}
	}

	// Convert list to []*ast.Field.
	// If list contains types only, each type gets its own ast.Field.
	if named == 0 {
		// parameter list consists of types only
		for _, par := range list { // 再一次过滤空异常
			assert(par.typ != nil, "nil type in unnamed parameter list")
			params = append(params, &ast.Field{Type: par.typ})
		}
		return
	}

	// If the parameter list consists of named parameters with types,
	// collect all names with the same types into a single ast.Field.
	var names []*ast.Ident
	var typ ast.Expr
	addParams := func() {
		assert(typ != nil, "nil type in named parameter list")
		field := &ast.Field{Names: names, Type: typ}
		params = append(params, field)
		names = nil
	}
	for _, par := range list {
		if par.typ != typ {
			// 将参数分组，相同类型的参数，添加到一个字段
			if len(names) > 0 { // 第一次为0
				addParams() // 添加一次， 清空一次 names
			}
			typ = par.typ // 记录
		}
		names = append(names, par.name)
	}
	// 最后调用一次，避免循环结束漏掉了
	if len(names) > 0 {
		addParams()
	}
	return
}

type name interface {
	~int | float32
	Main(string) string
}

type name0 int | string