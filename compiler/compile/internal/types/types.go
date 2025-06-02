package types

import (
	"github.com/facelang/face/compiler/compile/internal"
	"github.com/facelang/face/internal/prog"
	"github.com/facelang/face/internal/tokens"
)

// NewIndirect 指针类型 todo, 暂时忽略
func NewIndirect(pos prog.FilePos, typ prog.Expr) prog.Expr {
	o := new(prog.Operation)
	o.pos = pos
	o.Op = Mul
	o.X = typ
	return o
}

// typeOrNil is like type_ but it returns nil if there was no type
// instead of reporting an error.
//
//	Type     = TypeName | TypeLit | "(" Type ")" .
//	TypeName = identifier | QualifiedIdent .
//	TypeLit  = ArrayType | StructType | PointerType | FunctionType | InterfaceType |
//		      SliceType | MapType | Channel_Type .
func (p *internal.parser) typeOrNil() string {
	switch p.token {
	case '*':
		p.next()
		return "*"
	case tokens.IDENT:
	case internal.FUNC:
		p.next()
		_, t := p.funcType("function type")
		return t
	case internal.LBRACK: // []
	case internal.MAP: // map[_]_
	case internal.STRUCT:
	case internal.INTERFACE:
	case NAME:
	case internal.LPAREN:

	}
	if p.token == tokens.IDENT {
		return p.name()
	}
	return ""
}
