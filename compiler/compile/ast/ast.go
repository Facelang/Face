package ast

import (
	"github.com/facelang/face/compiler/compile/token"
)

type Node interface {
	Position() token.Pos
}

type Expr interface {
	Node
	expr()
}

type Stmt interface {
	Node
	stmtNode()
}

type Decl interface {
	Node
	declNode()
}

// ----------------------------------------------------------------------------
// Expressions and types

type Field struct {
	Names []*Name   // field/method/(type) parameter names; or nil
	Type  Expr      // field/method/parameter type; or nil
	Tag   *BasicLit // field tag; or nil
}

func (f *Field) Position() token.Pos {
	if len(f.Names) > 0 {
		return f.Names[0].Position()
	}
	if f.Type != nil {
		return f.Type.Position()
	}
	return token.NoPos
}

type FieldList struct {
	Opening token.Pos // position of opening parenthesis/brace/bracket, if any
	List    []*Field  // field list; or nil
	Closing token.Pos // position of closing parenthesis/brace/bracket, if any
}

func (f *FieldList) Position() token.Pos {
	if f.Opening.IsValid() {
		return f.Opening
	}

	if len(f.List) > 0 {
		return f.List[0].Position()
	}
	return token.NoPos
}

// NumFields returns the number of parameters or struct fields represented by a [FieldList].
func (f *FieldList) NumFields() int {
	n := 0
	if f != nil {
		for _, g := range f.List {
			m := len(g.Names)
			if m == 0 {
				m = 1
			}
			n += m
		}
	}
	return n
}

// An expression is represented by a tree consisting of one
// or more of the following concrete expression nodes.
type (
	// A BadExpr node is a placeholder for an expression containing
	// syntax errors for which a correct expression node cannot be
	// created.
	//
	BadExpr struct {
		From, To token.Pos // position range of bad expression
	}

	// An Name node represents an identifier.
	Name struct {
		Pos  token.Pos // identifier position
		Name string    // identifier name
		Obj  *Object   // denoted object, or nil. Deprecated: see Object.
	}

	// A BasicLit node represents a literal of basic type.
	//
	// Note that for the CHAR and STRING kinds, the literal is stored
	// with its quotes. For example, for a double-quoted STRING, the
	// first and the last rune in the Value field will be ". The
	// [strconv.Unquote] and [strconv.UnquoteChar] functions can be
	// used to unquote STRING and CHAR values, respectively.
	//
	// For raw string literals (Kind == tokens.STRING && Value[0] == '`'),
	// the Value field contains the string text without carriage returns (\r) that
	// may have been present in the source. Because the end position is
	// computed using len(Value), the position reported by [BasicLit.End] does not match the
	// true source end position for raw string literals containing carriage returns.
	BasicLit struct {
		Pos   token.Pos   // literal position
		Kind  token.Token // tokens.INT, tokens.FLOAT, tokens.IMAG, tokens.BYTE, or tokens.STRING
		Value string      // literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f', "foo" or `\m\n\o`
	}

	// A FuncLit node represents a function literal.
	FuncLit struct {
		Type *FuncType  // function type
		Body *BlockStmt // function body
	}

	// An Ellipsis node stands for the "..." type in a
	// parameter list or the "..." length in an array type.
	//
	Ellipsis struct {
		Ellipsis token.Pos // position of "..."
		Elt      Expr      // ellipsis element type (parameter lists only); or nil
	}

	// A CompositeLit node represents a composite literal.
	CompositeLit struct {
		Type       Expr      // literal type; or nil
		Lbrace     token.Pos // position of "{"
		Elts       []Expr    // list of composite elements; or nil
		Rbrace     token.Pos // position of "}"
		Incomplete bool      // true if (source) expressions are missing in the Elts list
	}

	// A ParenExpr node represents a parenthesized expression.
	ParenExpr struct {
		Lparen token.Pos // position of "("
		X      Expr      // parenthesized expression
		Rparen token.Pos // position of ")"
	}

	// A SelectorExpr node represents an expression followed by a selector.
	SelectorExpr struct {
		X   Expr  // expression
		Sel *Name // field selector
	}

	// An IndexExpr node represents an expression followed by an index.
	IndexExpr struct {
		X      Expr      // expression
		Lbrack token.Pos // position of "["
		Index  Expr      // index expression
		Rbrack token.Pos // position of "]"
	}

	// An IndexListExpr node represents an expression followed by multiple
	// indices.
	IndexListExpr struct {
		X       Expr      // expression
		Lbrack  token.Pos // position of "["
		Indices []Expr    // index expressions
		Rbrack  token.Pos // position of "]"
	}

	// A SliceExpr node represents an expression followed by slice indices.
	SliceExpr struct {
		X      Expr      // expression
		Lbrack token.Pos // position of "["
		Low    Expr      // begin of slice range; or nil
		High   Expr      // end of slice range; or nil
		Max    Expr      // maximum capacity of slice; or nil
		Slice3 bool      // true if 3-index slice (2 colons present)
		Rbrack token.Pos // position of "]"
	}

	// A TypeAssertExpr node represents an expression followed by a
	// type assertion.
	//
	TypeAssertExpr struct {
		X      Expr      // expression
		Lparen token.Pos // position of "("
		Type   Expr      // asserted type; nil means type switch X.(type)
		Rparen token.Pos // position of ")"
	}

	// A CallExpr node represents an expression followed by an argument list.
	CallExpr struct {
		Fun      Expr      // function expression
		Lparen   token.Pos // position of "("
		Args     []Expr    // function arguments; or nil
		Ellipsis token.Pos // position of "..." (tokens.NoPos if there is no "...")
		Rparen   token.Pos // position of ")"
	}

	// A StarExpr node represents an expression of the form "*" Expression.
	// Semantically it could be a unary "*" expression, or a pointer type.
	//
	StarExpr struct {
		Star token.Pos // position of "*"
		X    Expr      // operand
	}

	// A UnaryExpr node represents a unary expression.
	// Unary "*" expressions are represented via StarExpr nodes.
	//
	UnaryExpr struct {
		OpPos token.Pos   // position of Op
		Op    token.Token // operator
		X     Expr        // operand
	}

	// A BinaryExpr node represents a binary expression.
	BinaryExpr struct {
		X     Expr        // left operand
		OpPos token.Pos   // position of Op
		Op    token.Token // operator
		Y     Expr        // right operand
	}

	// A KeyValueExpr node represents (key : value) pairs
	// in composite literals.
	//
	KeyValueExpr struct {
		Key   Expr
		Colon token.Pos // position of ":"
		Value Expr
	}
)

// A type is represented by a tree consisting of one
// or more of the following type-specific expression
// nodes.
type (
	// An ArrayType node represents an array or slice type.
	ArrayType struct {
		Lbrack token.Pos // position of "["
		Len    Expr      // Ellipsis node for [...]T array types, nil for slice types
		Elt    Expr      // element type
	}

	// A StructType node represents a struct type.
	StructType struct {
		Struct     token.Pos  // position of "struct" keyword
		Fields     *FieldList // list of field declarations
		Incomplete bool       // true if (source) fields are missing in the Fields list
	}

	// Pointer types are represented via StarExpr nodes.

	// A FuncType node represents a function type.
	FuncType struct {
		Func       token.Pos  // position of "func" keyword (tokens.NoPos if there is no "func")
		TypeParams *FieldList // type parameters; or nil
		Params     *FieldList // (incoming) parameters; non-nil
		Results    *FieldList // (outgoing) results; or nil
	}

	// An InterfaceType node represents an interface type.
	InterfaceType struct {
		Interface  token.Pos  // position of "interface" keyword
		Methods    *FieldList // list of embedded interfaces, methods, or types
		Incomplete bool       // true if (source) methods or types are missing in the Methods list
	}

	// A MapType node represents a map type.
	MapType struct {
		Map   token.Pos // position of "map" keyword
		Key   Expr
		Value Expr
	}
)

// Pos and End implementations for expression/type nodes.

func (x *BadExpr) Position() token.Pos  { return x.From }
func (x *Name) Position() token.Pos     { return x.Pos }
func (x *BasicLit) Position() token.Pos { return x.Pos }
func (x *FuncLit) Position() token.Pos  { return x.Type.Position() }
func (x *Ellipsis) Position() token.Pos { return x.Ellipsis }
func (x *CompositeLit) Position() token.Pos {
	if x.Type != nil {
		return x.Type.Position()
	}
	return x.Lbrace
}
func (x *ParenExpr) Position() token.Pos      { return x.Lparen }
func (x *SelectorExpr) Position() token.Pos   { return x.X.Position() }
func (x *IndexExpr) Position() token.Pos      { return x.X.Position() }
func (x *IndexListExpr) Position() token.Pos  { return x.X.Position() }
func (x *SliceExpr) Position() token.Pos      { return x.X.Position() }
func (x *TypeAssertExpr) Position() token.Pos { return x.X.Position() }
func (x *CallExpr) Position() token.Pos       { return x.Fun.Position() }
func (x *StarExpr) Position() token.Pos       { return x.Star }
func (x *UnaryExpr) Position() token.Pos      { return x.OpPos }
func (x *BinaryExpr) Position() token.Pos     { return x.X.Position() }
func (x *KeyValueExpr) Position() token.Pos   { return x.Key.Position() }
func (x *ArrayType) Position() token.Pos      { return x.Lbrack }
func (x *StructType) Position() token.Pos     { return x.Struct }
func (x *FuncType) Position() token.Pos {
	if x.Func.IsValid() || x.Params == nil { // see issue 3870
		return x.Func
	}
	return x.Params.Position() // interface method declarations have no "func" keyword
}
func (x *InterfaceType) Position() token.Pos { return x.Interface }
func (x *MapType) Position() token.Pos       { return x.Map }

// expr() ensures that only expression/type nodes can be
// assigned to an Expr.
func (*BadExpr) expr()        {}
func (*Name) expr()           {}
func (*Ellipsis) expr()       {}
func (*BasicLit) expr()       {}
func (*FuncLit) expr()        {}
func (*CompositeLit) expr()   {}
func (*ParenExpr) expr()      {}
func (*SelectorExpr) expr()   {}
func (*IndexExpr) expr()      {}
func (*IndexListExpr) expr()  {}
func (*SliceExpr) expr()      {}
func (*TypeAssertExpr) expr() {}
func (*CallExpr) expr()       {}
func (*StarExpr) expr()       {}
func (*UnaryExpr) expr()      {}
func (*BinaryExpr) expr()     {}
func (*KeyValueExpr) expr()   {}

func (*ArrayType) expr()     {}
func (*StructType) expr()    {}
func (*FuncType) expr()      {}
func (*InterfaceType) expr() {}
func (*MapType) expr()       {}

// ----------------------------------------------------------------------------
// Convenience functions for Idents

// NewName creates a new [Ident] without position.
// Useful for ASTs generated by code other than the Go parser.
//func NewName(name string) *Name { return &Name{tokens.NoPos, name, nil} }

// IsExported reports whether name starts with an upper-case letter.
func IsExported(name string) bool { return token.IsExported(name) }

// IsExported reports whether id starts with an upper-case letter.
func (id *Name) IsExported() bool { return token.IsExported(id.Name) }

func (id *Name) String() string {
	if id != nil {
		return id.Name
	}
	return "<nil>"
}

// ----------------------------------------------------------------------------
// Statements

// A statement is represented by a tree consisting of one
// or more of the following concrete statement nodes.
type (
	// A BadStmt node is a placeholder for statements containing
	// syntax errors for which no correct statement nodes can be
	// created.
	//
	BadStmt struct {
		From, To token.Pos // position range of bad statement
	}

	// A DeclStmt node represents a declaration in a statement list.
	DeclStmt struct {
		Decl Decl // *GenDecl with CONST, TYPE, or VAR token
	}

	// An EmptyStmt node represents an empty statement.
	// The "position" of the empty statement is the position
	// of the immediately following (explicit or implicit) semicolon.
	//
	EmptyStmt struct {
		Semicolon token.Pos // position of following ";"
		Implicit  bool      // if set, ";" was omitted in the source
	}

	// A LabeledStmt node represents a labeled statement.
	LabeledStmt struct {
		Label *Name
		Colon token.Pos // position of ":"
		Stmt  Stmt
	}

	// An ExprStmt node represents a (stand-alone) expression
	// in a statement list.
	//
	ExprStmt struct {
		X Expr // expression
	}

	// A SendStmt node represents a send statement.
	SendStmt struct {
		Chan  Expr
		Arrow token.Pos // position of "<-"
		Value Expr
	}

	// An IncDecStmt node represents an increment or decrement statement.
	IncDecStmt struct {
		X      Expr
		TokPos token.Pos   // position of Tok
		Tok    token.Token // INC or DEC
	}

	// An AssignStmt node represents an assignment or
	// a short variable declaration.
	//
	AssignStmt struct {
		Lhs    []Expr
		TokPos token.Pos   // position of Tok
		Tok    token.Token // assignment token, DEFINE
		Rhs    []Expr
	}

	// A GoStmt node represents a go statement.
	GoStmt struct {
		Go   token.Pos // position of "go" keyword
		Call *CallExpr
	}

	// A DeferStmt node represents a defer statement.
	DeferStmt struct {
		Defer token.Pos // position of "defer" keyword
		Call  *CallExpr
	}

	// A ReturnStmt node represents a return statement.
	ReturnStmt struct {
		Return  token.Pos // position of "return" keyword
		Results []Expr    // result expressions; or nil
	}

	// A BranchStmt node represents a break, continue, goto,
	// or fallthrough statement.
	//
	BranchStmt struct {
		TokPos token.Pos   // position of Tok
		Tok    token.Token // keyword token (BREAK, CONTINUE, GOTO, FALLTHROUGH)
		Label  *Name       // label name; or nil
	}

	// A BlockStmt node represents a braced statement list.
	BlockStmt struct {
		Lbrace token.Pos // position of "{"
		List   []Stmt
		Rbrace token.Pos // position of "}", if any (may be absent due to syntax error)
	}

	// An IfStmt node represents an if statement.
	IfStmt struct {
		If   token.Pos // position of "if" keyword
		Init Stmt      // initialization statement; or nil
		Cond Expr      // condition
		Body *BlockStmt
		Else Stmt // else branch; or nil
	}

	// A CaseClause represents a case of an expression or type switch statement.
	CaseClause struct {
		Case  token.Pos // position of "case" or "default" keyword
		List  []Expr    // list of expressions or types; nil means default case
		Colon token.Pos // position of ":"
		Body  []Stmt    // statement list; or nil
	}

	// A SwitchStmt node represents an expression switch statement.
	SwitchStmt struct {
		Switch token.Pos  // position of "switch" keyword
		Init   Stmt       // initialization statement; or nil
		Tag    Expr       // tag expression; or nil
		Body   *BlockStmt // CaseClauses only
	}

	// A TypeSwitchStmt node represents a type switch statement.
	TypeSwitchStmt struct {
		Switch token.Pos  // position of "switch" keyword
		Init   Stmt       // initialization statement; or nil
		Assign Stmt       // x := y.(type) or y.(type)
		Body   *BlockStmt // CaseClauses only
	}

	// A CommClause node represents a case of a select statement.
	CommClause struct {
		Case  token.Pos // position of "case" or "default" keyword
		Comm  Stmt      // send or receive statement; nil means default case
		Colon token.Pos // position of ":"
		Body  []Stmt    // statement list; or nil
	}

	// A SelectStmt node represents a select statement.
	SelectStmt struct {
		Select token.Pos  // position of "select" keyword
		Body   *BlockStmt // CommClauses only
	}

	// A ForStmt represents a for statement.
	ForStmt struct {
		For  token.Pos // position of "for" keyword
		Init Stmt      // initialization statement; or nil
		Cond Expr      // condition; or nil
		Post Stmt      // post iteration statement; or nil
		Body *BlockStmt
	}

	// A RangeStmt represents a for statement with a range clause.
	RangeStmt struct {
		For        token.Pos   // position of "for" keyword
		Key, Value Expr        // Key, Value may be nil
		TokPos     token.Pos   // position of Tok; invalid if Key == nil
		Tok        token.Token // ILLEGAL if Key == nil, ASSIGN, DEFINE
		Range      token.Pos   // position of "range" keyword
		X          Expr        // value to range over
		Body       *BlockStmt
	}
)

// Pos and End implementations for statement nodes.

func (s *BadStmt) Position() token.Pos        { return s.From }
func (s *DeclStmt) Position() token.Pos       { return s.Decl.Position() }
func (s *EmptyStmt) Position() token.Pos      { return s.Semicolon }
func (s *LabeledStmt) Position() token.Pos    { return s.Label.Position() }
func (s *ExprStmt) Position() token.Pos       { return s.X.Position() }
func (s *SendStmt) Position() token.Pos       { return s.Chan.Position() }
func (s *IncDecStmt) Position() token.Pos     { return s.X.Position() }
func (s *AssignStmt) Position() token.Pos     { return s.Lhs[0].Position() }
func (s *GoStmt) Position() token.Pos         { return s.Go }
func (s *DeferStmt) Position() token.Pos      { return s.Defer }
func (s *ReturnStmt) Position() token.Pos     { return s.Return }
func (s *BranchStmt) Position() token.Pos     { return s.TokPos }
func (s *BlockStmt) Position() token.Pos      { return s.Lbrace }
func (s *IfStmt) Position() token.Pos         { return s.If }
func (s *CaseClause) Position() token.Pos     { return s.Case }
func (s *SwitchStmt) Position() token.Pos     { return s.Switch }
func (s *TypeSwitchStmt) Position() token.Pos { return s.Switch }
func (s *CommClause) Position() token.Pos     { return s.Case }
func (s *SelectStmt) Position() token.Pos     { return s.Select }
func (s *ForStmt) Position() token.Pos        { return s.For }
func (s *RangeStmt) Position() token.Pos      { return s.For }

// stmtNode() ensures that only statement nodes can be
// assigned to a Stmt.
func (*BadStmt) stmtNode()        {}
func (*DeclStmt) stmtNode()       {}
func (*EmptyStmt) stmtNode()      {}
func (*LabeledStmt) stmtNode()    {}
func (*ExprStmt) stmtNode()       {}
func (*SendStmt) stmtNode()       {}
func (*IncDecStmt) stmtNode()     {}
func (*AssignStmt) stmtNode()     {}
func (*GoStmt) stmtNode()         {}
func (*DeferStmt) stmtNode()      {}
func (*ReturnStmt) stmtNode()     {}
func (*BranchStmt) stmtNode()     {}
func (*BlockStmt) stmtNode()      {}
func (*IfStmt) stmtNode()         {}
func (*CaseClause) stmtNode()     {}
func (*SwitchStmt) stmtNode()     {}
func (*TypeSwitchStmt) stmtNode() {}
func (*CommClause) stmtNode()     {}
func (*SelectStmt) stmtNode()     {}
func (*ForStmt) stmtNode()        {}
func (*RangeStmt) stmtNode()      {}

// ----------------------------------------------------------------------------
// Declarations

type (
	BadDecl struct {
		From, To token.Pos // position range of bad declaration
	}

	GenDecl struct {
		Pos    token.Pos   // position of Tok
		Token  token.Token // const or Let
		Names  []*Name     // value names (len(Names) > 0)
		Type   Expr        // value type; or nil
		Values []Expr      // initial values; or nil
	}

	// A FuncDecl node represents a function declaration.
	FuncDecl struct {
		Pos token.Pos
		//Recv *FieldList // receiver (methods); or nil (functions)
		Type *FuncType  // function signature: type and value parameters, results, and position of "func" keyword
		Name *Name      // function/method name
		Body *BlockStmt // function body; or nil for external (non-Go) function
	}

	// TypeDecl type 语句：
	// 		type name1 int 新的类型
	// 		type name2 = int 类型别名
	//		type name3 struct {} 结构体
	// 		type name5 interface{} 接口
	TypeDecl struct {
		Pos        token.Pos  // position of Tok
		Name       *Name      // type name
		TypeParams *FieldList // type parameters; or nil
		Assign     token.Pos  // position of '=', if any
		Type       Expr       // *Ident, *ParenExpr, *SelectorExpr, *StarExpr, or any of the *XxxTypes
	}
)

// Pos and End implementations for declaration nodes.

func (d *BadDecl) Position() token.Pos  { return d.From }
func (d *GenDecl) Position() token.Pos  { return d.Pos }
func (d *FuncDecl) Position() token.Pos { return d.Pos }
func (d *TypeDecl) Position() token.Pos { return d.Pos }

// declNode() ensures that only declaration nodes can be
// assigned to a Decl.
func (*BadDecl) declNode()  {}
func (*GenDecl) declNode()  {}
func (*FuncDecl) declNode() {}
func (*TypeDecl) declNode() {}

type File struct {
	DeclList []Decl // top-level declarations; or nil

	//FileStart, FileEnd tokens.Pos      // start and end of entire file
	Scope      *Scope     // package scope (this file only). Deprecated: see Object
	Imports    []*Package // imports in this file
	Unresolved []*Name    // unresolved identifiers in this file. Deprecated: see Object
	//Comments   []*CommentGroup // list of all comments in the source file
	//GoVersion string // minimum Go version required by //go:build or // +build directives
}

type PkgName struct {
	Pos  token.Pos // 位置信息
	Name string    // 别名
	Kind string    // 类型
}

// import ""
// import default from ""
// import {} // todo 暂时不支持解包

type Package struct {
	Pos  token.Pos // 位置信息
	Name string    //
	Path string    // import path
}

// Unparen 扒开括号，查看真实类型
func Unparen(e Expr) Expr {
	for {
		paren, ok := e.(*ParenExpr)
		if !ok {
			return e
		}
		e = paren.X
	}
}
