package parser

import (
	"github.com/facelang/face/compiler/compile/ast"
	"github.com/facelang/face/compiler/compile/tokens"
	"github.com/facelang/face/internal/prog"
	"go/token"
)

// NewIndirect 指针类型 todo, 暂时忽略
func NewIndirect(pos prog.FilePos, typ prog.Expr) prog.Expr {
	o := new(prog.Operation)
	o.pos = pos
	o.Op = Mul
	o.X = typ
	return o
}

// FuncType If context != "", type parameters are not permitted.
func FuncType(p *parser, context string) ([]*prog.Field, *prog.FuncType) {

	typ := new(prog.FuncType)
	typ.pos = p.FilePos

	var tparamList []*prog.Field
	// 目标语法使用 尖括号
	//if p.got(api.LBRACK) { // [] 泛型 func [] name(args)
	//	if context != "" {
	//		// accept but complain
	//		p.syntaxErrorAt(typ.pos, context+" must have no type parameters")
	//	}
	//	if p.tok == _Rbrack {
	//		p.syntaxError("empty type parameter list")
	//		p.next()
	//	} else {
	//		tparamList = p.paramList(nil, nil, _Rbrack, true)
	//	}
	//}

	p.want(LPAREN)
	typ.ParamList = p.paramList(nil, nil, _Rparen, false)
	typ.ResultList = p.funcResult()

	return tparamList, typ
}

// TypeOrNil is like type_ but it returns nil if there was no type
// instead of reporting an error.
//
//	Type     = TypeName | TypeLit | "(" Type ")" .
//	TypeName = identifier | QualifiedIdent .
//	TypeLit  = ArrayType | StructType | PointerType | FunctionType | InterfaceType |
//		      SliceType | MapType | Channel_Type .
func TypeOrNil(p *parser) prog.Expr {
	//defer decNestLev(incNestLev(p)) // 递归统计，避免递归太深
	switch p.token {
	case token.IDENT:
		typ := p.parseTypeName(nil)
		if p.tok == token.LBRACK {
			typ = p.parseTypeInstance(typ)
		}
		return typ
	case LBRACK:
		lbrack := p.expect(LBRACK)
		return p.parseArrayType(lbrack, nil)
	case STRUCT:
		return p.parseStructType()
	case MUL:
		return p.parsePointerType()
	case FUNC:
		return p.parseFuncType()
	case INTERFACE:
		return p.parseInterfaceType()
	case MAP:
		return p.parseMapType()
	case CHAN, ARROW:
		return p.parseChanType()
	case LPAREN:
		lparen := p.pos
		p.next()
		typ := p.parseType()
		rparen := p.expect(RPAREN)
		return &ast.ParenExpr{Lparen: lparen, X: typ, Rparen: rparen}
	}

	// no type found
	return nil
}

func RequireType(p *parser) prog.Expr {
	typ := TypeOrNil(p)
	if typ == nil {
		p.unexpect("type")
	}
	return typ
}

/**
函数类型: let v1 func(string)
函数类型: let v2 (string) => string
数组类型: let v3 array<int>
字典类型: let v4 map<string, string>
基本数据类型: let v5 int [string, float]
其它自定义类型: let v6 http.Http [或其它类型别名]
*/

func (p *parser) parseTypeInstance(typ ast.Expr) ast.Expr {
	opening := p.expect(token.LBRACK) // [
	//p.exprLev++
	var list []ast.Expr
	for p.token != token.RBRACK && p.token != token.EOF {
		list = append(list, p.parseType())
		if p.token != token.COMMA {
			break
		}
		p.next()
	}
	//p.exprLev--

	closing := p.expect(token.RBRACK) // ]

	if len(list) == 0 {
		p.unexpect("type argument list")
		return &ast.IndexExpr{
			X:      typ,
			Lbrack: opening,
			Index:  &ast.BadExpr{From: opening + 1, To: closing},
			Rbrack: closing,
		}
	}

	return packIndexExpr(typ, opening, list, closing)
}

// If the result is an identifier, it is not resolved.
func (p *parser) parseTypeName(ident *ast.Name) ast.Expr {
	if ident == nil {
		ident = p.name()
	}

	if p.token == token.PERIOD {
		p.next()
		sel := p.name()
		return &ast.SelectorExpr{X: ident, Sel: sel}
	}

	return ident
}

// "[" has already been consumed, and lbrack is its position.
// If len != nil it is the already consumed array length.
func (p *parser) parseArrayType(lbrack token.Pos, len ast.Expr) *ast.ArrayType {

	if len == nil { // 没有解析 [x] 中间的参数
		//p.exprLev++
		// always permit ellipsis for more fault-tolerant parsing
		if p.token == token.ELLIPSIS { // [...]
			len = &ast.Ellipsis{Ellipsis: p.pos}
			p.next()
		} else if p.token != token.RBRACK { // [len]
			len = exprRhs(p)
		}
		// len 可能为 nil
		//p.exprLev--
	}
	if p.token == token.COMMA { // , 不应该出现
		// Trailing commas are accepted in type parameter
		// lists but not in array type declarations.
		// Accept for better error handling but complain.
		p.error(p.pos, "unexpected comma; expecting ]")
		p.next()
	}
	p.expect(token.RBRACK) // ] 结束符
	elt := p.parseType()   // 可能是多维数组
	return &ast.ArrayType{Lbrack: lbrack, Len: len, Elt: elt}
}

func (p *parser) parseMapType() *ast.MapType {
	pos := p.expect(token.MAP) // map
	p.expect(token.LBRACK)     // [
	key := p.parseType()       // keyType
	p.expect(token.RBRACK)     // ]
	value := p.parseType()     // valType

	return &ast.MapType{Map: pos, Key: key, Value: value}
}

func (p *parser) parseQualifiedIdent(ident *ast.Name) ast.Expr {

	typ := p.parseTypeName(ident)
	if p.token == token.LBRACK {
		typ = p.parseTypeInstance(typ)
	}

	return typ
}

func (p *parser) parseArrayFieldOrTypeInstance(x *ast.Name) (*ast.Name, ast.Expr) {
	lbrack := p.expect(token.LBRACK)
	trailingComma := token.NoPos // if valid, the position of a trailing comma preceding the ']'
	var args []ast.Expr
	if p.token != token.RBRACK {
		//p.exprLev++
		args = append(args, exprRhs(p))
		for p.token == token.COMMA {
			comma := p.pos
			p.next()
			if p.token == token.RBRACK {
				trailingComma = comma
				break
			}
			args = append(args, exprRhs(p))
		}
		//p.exprLev--
	}
	rbrack := p.expect(token.RBRACK)

	if len(args) == 0 {
		// x []E
		elt := p.parseType()
		return x, &ast.ArrayType{Lbrack: lbrack, Elt: elt}
	}

	// x [P]E or x[P]
	if len(args) == 1 {
		elt := p.tryIdentOrType()
		if elt != nil {
			// x [P]E
			if trailingComma.IsValid() {
				// Trailing commas are invalid in array type fields.
				p.error(trailingComma, "unexpected comma; expecting ]")
			}
			return x, &ast.ArrayType{Lbrack: lbrack, Len: args[0], Elt: elt}
		}
	}

	// x[P], x[P1, P2], ...
	return nil, packIndexExpr(x, lbrack, args, rbrack)
}

// 只在结构体中
func (p *parser) parseFieldDecl() *ast.Field {

	//doc := p.leadComment

	var names []*ast.Name
	var typ ast.Expr
	switch p.token {
	case token.IDENT: // 先解析字段名
		name := p.name()
		if p.token == token.PERIOD || p.token == token.STRING || p.token == token.SEMICOLON || p.token == token.RBRACE {
			// embedded type
			// 继续解析 name.   . "" ; }
			typ = name
			if p.token == token.PERIOD {
				typ = p.parseQualifiedIdent(name)
			}
		} else { // 其它符号
			// name1, name2, ... T
			names = []*ast.Name{name}
			for p.token == token.COMMA { // struct { a, b, c int }
				p.next()
				names = append(names, p.name())
			}
			// Careful dance: We don't know if we have an embedded instantiated
			// type T[P1, P2, ...] or a field T of array type []E or [P]E.
			// { a }
			if len(names) == 1 && p.token == token.LBRACK {
				name, typ = p.parseArrayFieldOrTypeInstance(name) // todo
				if name == nil {
					names = nil
				}
			} else {
				// T P
				typ = p.parseType()
			}
		}
	case token.MUL:
		star := p.pos
		p.next()
		if p.token == token.LPAREN {
			// *(T)
			p.error(p.pos, "cannot parenthesize embedded type")
			p.next()
			typ = p.parseQualifiedIdent(nil)
			// expect closing ')' but no need to complain if missing
			if p.token == token.RPAREN {
				p.next()
			}
		} else {
			// *T
			typ = p.parseQualifiedIdent(nil)
		}
		typ = &ast.StarExpr{Star: star, X: typ}

	case token.LPAREN:
		p.error(p.pos, "cannot parenthesize embedded type")
		p.next()
		if p.token == token.MUL {
			// (*T)
			star := p.pos
			p.next()
			typ = &ast.StarExpr{Star: star, X: p.parseQualifiedIdent(nil)}
		} else {
			// (T)
			typ = p.parseQualifiedIdent(nil)
		}
		// expect closing ')' but no need to complain if missing
		if p.token == token.RPAREN {
			p.next()
		}

	default:
		pos := p.pos
		p.unexpect("field name or embedded type")
		typ = &ast.BadExpr{From: pos, To: p.pos}
	}

	var tag *ast.BasicLit
	if p.token == token.STRING {
		tag = &ast.BasicLit{Pos: p.pos, Kind: p.token, Value: p.identifier}
		p.next()
	}

	field := &ast.Field{Names: names, Type: typ, Tag: tag}
	return field
}

func (p *parser) parseStructType() *ast.StructType {
	pos := p.expect(token.STRUCT) // struct {}
	lbrace := p.expect(token.LBRACE)
	var list []*ast.Field
	for p.token == token.IDENT || p.token == token.MUL || p.token == token.LPAREN {
		// a field declaration cannot start with a '(' but we accept
		// it here for more robust parsing and better error messages
		// (parseFieldDecl will check and complain if necessary)
		list = append(list, p.parseFieldDecl())
	}
	rbrace := p.expect(token.RBRACE)

	return &ast.StructType{
		Struct: pos,
		Fields: &ast.FieldList{
			Opening: lbrace,
			List:    list,
			Closing: rbrace,
		},
	}
}

func (p *parser) parsePointerType() *ast.StarExpr {
	star := p.expect(token.MUL)
	base := p.parseType()

	return &ast.StarExpr{Star: star, X: base}
}

func (p *parser) parseMethodSpec() *ast.Field {
	var idents []*ast.Name
	var typ ast.Expr
	x := p.parseTypeName(nil)
	if ident, _ := x.(*ast.Name); ident != nil {
		switch {
		case p.token == token.LBRACK:
			// generic method or embedded instantiated type
			lbrack := p.pos
			p.next()
			//p.exprLev++
			x := expr(p)
			//p.exprLev--
			if name0, _ := x.(*ast.Name); name0 != nil && p.token != token.COMMA && p.token != token.RBRACK {
				// generic method m[T any]
				//
				// Interface methods do not have type parameters. We parse them for a
				// better error message and improved error recovery.
				_ = p.parseParameterList(name0, nil, token.RBRACK)
				_ = p.expect(token.RBRACK)
				p.error(lbrack, "interface method must have no type parameters")

				// TODO(rfindley) refactor to share code with parseFuncType.
				_, params := p.parseParameters(false)
				results := p.parseResult()
				idents = []*ast.Name{ident}
				typ = &ast.FuncType{
					Func:    token.NoPos,
					Params:  params,
					Results: results,
				}
			} else {
				// embedded instantiated type
				// TODO(rfindley) should resolve all identifiers in x.
				list := []ast.Expr{x}
				if p.token == token.COMMA {
					//p.exprLev++
					p.next()
					for p.token != token.RBRACK && p.token != token.EOF {
						list = append(list, p.parseType())
						if p.token != token.COMMA {
							break
						}
						p.next()
					}
					//p.exprLev--
				}
				rbrack := p.expectClosing(token.RBRACK, "type argument list")
				typ = packIndexExpr(ident, lbrack, list, rbrack)
			}
		case p.token == token.LPAREN:
			// ordinary method
			// TODO(rfindley) refactor to share code with parseFuncType.
			_, params := p.parseParameters(false)
			results := p.parseResult()
			idents = []*ast.Ident{ident}
			typ = &ast.FuncType{Func: token.NoPos, Params: params, Results: results}
		default:
			// embedded type
			typ = x
		}
	} else {
		// embedded, possibly instantiated type
		typ = x
		if p.token == token.LBRACK {
			// embedded instantiated interface
			typ = p.parseTypeInstance(typ)
		}
	}

	return &ast.Field{Names: idents, Type: typ}
}

func (p *parser) embeddedElem(x ast.Expr) ast.Expr {
	if x == nil {
		x = p.embeddedTerm()
	}
	for p.token == token.OR {
		t := new(ast.BinaryExpr)
		t.OpPos = p.pos
		t.Op = token.OR
		p.next()
		t.X = x
		t.Y = p.embeddedTerm()
		x = t
	}
	return x
}

func (p *parser) embeddedTerm() ast.Expr {
	if p.token == token.TILDE {
		t := new(ast.UnaryExpr)
		t.OpPos = p.pos
		t.Op = token.TILDE
		p.next()
		t.X = p.parseType()
		return t
	}

	t := p.tryIdentOrType()
	if t == nil {
		pos := p.pos
		p.unexpect("~ term or type")
		return &ast.BadExpr{From: pos, To: p.pos}
	}

	return t
}

func (p *parser) parseInterfaceType() *ast.InterfaceType {
	pos := p.expect(token.INTERFACE) // interface {}
	lbrace := p.expect(token.LBRACE)

	var list []*ast.Field

parseElements:
	for {
		switch {
		case p.token == token.IDENT: // 只能声明函数
			f := p.parseMethodSpec()
			if f.Names == nil {
				f.Type = p.embeddedElem(f.Type)
			}
			f.Comment = p.expectSemi()
			list = append(list, f)
		case p.token == token.TILDE:
			typ := p.embeddedElem(nil)
			comment := p.expectSemi()
			list = append(list, &ast.Field{Type: typ, Comment: comment})
		default:
			if t := p.tryIdentOrType(); t != nil {
				typ := p.embeddedElem(t)
				comment := p.expectSemi()
				list = append(list, &ast.Field{Type: typ, Comment: comment})
			} else {
				break parseElements
			}
		}
	}

	// TODO(rfindley): the error produced here could be improved, since we could
	// accept an identifier, 'type', or a '}' at this point.
	rbrace := p.expect(token.RBRACE)

	return &ast.InterfaceType{
		Interface: pos,
		Methods: &ast.FieldList{
			Opening: lbrace,
			List:    list,
			Closing: rbrace,
		},
	}
}

func (p *parser) tryIdentOrType() ast.Expr {
	defer decNestLev(incNestLev(p))

	switch p.token {
	case token.IDENT:
		typ := p.parseTypeName(nil)  // 可能是 x.name(包名) 或者 x
		if p.token == token.LBRACK { // x[]
			typ = p.parseTypeInstance(typ) // todo
		}
		return typ
	case token.LBRACK:
		lbrack := p.expect(token.LBRACK) // n[]
		return p.parseArrayType(lbrack, nil)
	case token.STRUCT:
		return p.parseStructType()
	case token.MUL:
		return p.parsePointerType()
	case token.FUNC:
		return p.parseFuncType()
	case token.INTERFACE:
		return p.parseInterfaceType()
	case token.MAP:
		return p.parseMapType()
	//case tokens.CHAN, tokens.ARROW:
	//	return p.parseChanType()
	case token.LPAREN: // (
		lparen := p.pos
		p.next()
		typ := p.parseType()
		rparen := p.expect(token.RPAREN)
		return &ast.ParenExpr{Lparen: lparen, X: typ, Rparen: rparen}
	}

	// no type found
	return nil
}

func (p *parser) parseType() ast.Expr {
	typ := p.tryIdentOrType()

	if typ == nil {
		pos := p.pos
		p.unexpect("type")
		return &ast.BadExpr{From: pos, To: p.pos}
	}

	return typ
}
