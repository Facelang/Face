package parser

import (
	"github.com/facelang/face/compiler/compile/tokens"
	"github.com/facelang/face/internal/prog"
	"go/ast"
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
	case tokens.IDENT:
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

var name func(string) = func(s string) {

}
