// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ast declares the types used to represent syntax trees for Go
// packages.
package ast

import (
	"github.com/facelang/face/compiler/compile/tokens"
)

// ----------------------------------------------------------------------------
// Interfaces
//
// There are 3 main classes of nodes: Expressions and type nodes,
// statement nodes, and declaration nodes. The node names usually
// match the corresponding Go spec production names to which they
// correspond. The node fields correspond to the individual parts
// of the respective productions.
//
// All nodes contain position information marking the beginning of
// the corresponding source text segment; it is accessible via the
// Pos accessor method. Nodes may contain additional position info
// for language constructs where comments may be found between parts
// of the construct (typically any larger, parenthesized subpart).
// That position information is needed to properly position comments
// when printing the construct.

// All node types implement the Node interface.
type Node interface {
	Offset() tokens.Pos // position of first character belonging to the node
}

// All expression nodes implement the Expr interface.
type Expr interface {
	Node
	expr()
}

// All statement nodes implement the Stmt interface.
type Stmt interface {
	Node
	stmtNode()
}

// All declaration nodes implement the Decl interface.
type Decl interface {
	Node
	declNode()
}

// ----------------------------------------------------------------------------
// Comments

// A Comment node represents a single //-style or /*-style comment.
//
// The Text field contains the comment text without carriage returns (\r) that
// may have been present in the source. Because a comment's end position is
// computed using len(Text), the position reported by [Comment.End] does not match the
// true source end position for comments containing carriage returns.
type Comment struct {
	Slash tokens.Pos // position of "/" starting the comment
	Text  string     // comment text (excluding '\n' for //-style comments)
}

func (c *Comment) Offset() tokens.Pos { return c.Slash }
func (c *Comment) End() tokens.Pos    { return tokens.Pos(int(c.Slash) + len(c.Text)) }

// ----------------------------------------------------------------------------
// Expressions and types

// A Field represents a Field declaration list in a struct type,
// a method list in an interface type, or a parameter/result declaration
// in a signature.
// [Field.Names] is nil for unnamed parameters (parameter lists which only contain types)
// and embedded struct fields. In the latter case, the field name is the type name.
type Field struct {
	Names []*Name   // field/method/(type) parameter names; or nil
	Type  Expr      // field/method/parameter type; or nil
	Tag   *BasicLit // field tag; or nil
}

func (f *Field) Offset() tokens.Pos {
	if len(f.Names) > 0 {
		return f.Names[0].Offset()
	}
	if f.Type != nil {
		return f.Type.Offset()
	}
	return tokens.NoPos
}

// A FieldList represents a list of Fields, enclosed by parentheses,
// curly braces, or square brackets.
type FieldList struct {
	Opening tokens.Pos // position of opening parenthesis/brace/bracket, if any
	List    []*Field   // field list; or nil
	Closing tokens.Pos // position of closing parenthesis/brace/bracket, if any
}

func (f *FieldList) Offset() tokens.Pos {
	if f.Opening.IsValid() {
		return f.Opening
	}
	// the list should not be empty in this case;
	// be conservative and guard against bad ASTs
	if len(f.List) > 0 {
		return f.List[0].Offset()
	}
	return tokens.NoPos
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
		From, To tokens.Pos // position range of bad expression
	}

	// An Name node represents an identifier.
	Name struct {
		Pos  tokens.Pos // identifier position
		Name string     // identifier name
		Obj  *Object    // denoted object, or nil. Deprecated: see Object.
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
		Pos   tokens.Pos   // literal position
		Kind  tokens.Token // tokens.INT, tokens.FLOAT, tokens.IMAG, tokens.BYTE, or tokens.STRING
		Value string       // literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f', "foo" or `\m\n\o`
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
		Ellipsis tokens.Pos // position of "..."
		Elt      Expr       // ellipsis element type (parameter lists only); or nil
	}

	// A CompositeLit node represents a composite literal.
	CompositeLit struct {
		Type       Expr       // literal type; or nil
		Lbrace     tokens.Pos // position of "{"
		Elts       []Expr     // list of composite elements; or nil
		Rbrace     tokens.Pos // position of "}"
		Incomplete bool       // true if (source) expressions are missing in the Elts list
	}

	// A ParenExpr node represents a parenthesized expression.
	ParenExpr struct {
		Lparen tokens.Pos // position of "("
		X      Expr       // parenthesized expression
		Rparen tokens.Pos // position of ")"
	}

	// A SelectorExpr node represents an expression followed by a selector.
	SelectorExpr struct {
		X   Expr  // expression
		Sel *Name // field selector
	}

	// An IndexExpr node represents an expression followed by an index.
	IndexExpr struct {
		X      Expr       // expression
		Lbrack tokens.Pos // position of "["
		Index  Expr       // index expression
		Rbrack tokens.Pos // position of "]"
	}

	// An IndexListExpr node represents an expression followed by multiple
	// indices.
	IndexListExpr struct {
		X       Expr       // expression
		Lbrack  tokens.Pos // position of "["
		Indices []Expr     // index expressions
		Rbrack  tokens.Pos // position of "]"
	}

	// A SliceExpr node represents an expression followed by slice indices.
	SliceExpr struct {
		X      Expr       // expression
		Lbrack tokens.Pos // position of "["
		Low    Expr       // begin of slice range; or nil
		High   Expr       // end of slice range; or nil
		Max    Expr       // maximum capacity of slice; or nil
		Slice3 bool       // true if 3-index slice (2 colons present)
		Rbrack tokens.Pos // position of "]"
	}

	// A TypeAssertExpr node represents an expression followed by a
	// type assertion.
	//
	TypeAssertExpr struct {
		X      Expr       // expression
		Lparen tokens.Pos // position of "("
		Type   Expr       // asserted type; nil means type switch X.(type)
		Rparen tokens.Pos // position of ")"
	}

	// A CallExpr node represents an expression followed by an argument list.
	CallExpr struct {
		Fun      Expr       // function expression
		Lparen   tokens.Pos // position of "("
		Args     []Expr     // function arguments; or nil
		Ellipsis tokens.Pos // position of "..." (tokens.NoPos if there is no "...")
		Rparen   tokens.Pos // position of ")"
	}

	// A StarExpr node represents an expression of the form "*" Expression.
	// Semantically it could be a unary "*" expression, or a pointer type.
	//
	StarExpr struct {
		Star tokens.Pos // position of "*"
		X    Expr       // operand
	}

	// A UnaryExpr node represents a unary expression.
	// Unary "*" expressions are represented via StarExpr nodes.
	//
	UnaryExpr struct {
		OpPos tokens.Pos   // position of Op
		Op    tokens.Token // operator
		X     Expr         // operand
	}

	// A BinaryExpr node represents a binary expression.
	BinaryExpr struct {
		X     Expr         // left operand
		OpPos tokens.Pos   // position of Op
		Op    tokens.Token // operator
		Y     Expr         // right operand
	}

	// A KeyValueExpr node represents (key : value) pairs
	// in composite literals.
	//
	KeyValueExpr struct {
		Key   Expr
		Colon tokens.Pos // position of ":"
		Value Expr
	}
)

// A type is represented by a tree consisting of one
// or more of the following type-specific expression
// nodes.
type (
	// An ArrayType node represents an array or slice type.
	ArrayType struct {
		Lbrack tokens.Pos // position of "["
		Len    Expr       // Ellipsis node for [...]T array types, nil for slice types
		Elt    Expr       // element type
	}

	// A StructType node represents a struct type.
	StructType struct {
		Struct     tokens.Pos // position of "struct" keyword
		Fields     *FieldList // list of field declarations
		Incomplete bool       // true if (source) fields are missing in the Fields list
	}

	// Pointer types are represented via StarExpr nodes.

	// A FuncType node represents a function type.
	FuncType struct {
		Func       tokens.Pos // position of "func" keyword (tokens.NoPos if there is no "func")
		TypeParams *FieldList // type parameters; or nil
		Params     *FieldList // (incoming) parameters; non-nil
		Results    *FieldList // (outgoing) results; or nil
	}

	// An InterfaceType node represents an interface type.
	InterfaceType struct {
		Interface  tokens.Pos // position of "interface" keyword
		Methods    *FieldList // list of embedded interfaces, methods, or types
		Incomplete bool       // true if (source) methods or types are missing in the Methods list
	}

	// A MapType node represents a map type.
	MapType struct {
		Map   tokens.Pos // position of "map" keyword
		Key   Expr
		Value Expr
	}
)

// Pos and End implementations for expression/type nodes.

func (x *BadExpr) Offset() tokens.Pos  { return x.From }
func (x *Name) Offset() tokens.Pos     { return x.Pos }
func (x *BasicLit) Offset() tokens.Pos { return x.Pos }
func (x *FuncLit) Offset() tokens.Pos  { return x.Type.Offset() }
func (x *Ellipsis) Offset() tokens.Pos { return x.Ellipsis }
func (x *CompositeLit) Offset() tokens.Pos {
	if x.Type != nil {
		return x.Type.Offset()
	}
	return x.Lbrace
}
func (x *ParenExpr) Offset() tokens.Pos      { return x.Lparen }
func (x *SelectorExpr) Offset() tokens.Pos   { return x.X.Offset() }
func (x *IndexExpr) Offset() tokens.Pos      { return x.X.Offset() }
func (x *IndexListExpr) Offset() tokens.Pos  { return x.X.Offset() }
func (x *SliceExpr) Offset() tokens.Pos      { return x.X.Offset() }
func (x *TypeAssertExpr) Offset() tokens.Pos { return x.X.Offset() }
func (x *CallExpr) Offset() tokens.Pos       { return x.Fun.Offset() }
func (x *StarExpr) Offset() tokens.Pos       { return x.Star }
func (x *UnaryExpr) Offset() tokens.Pos      { return x.OpPos }
func (x *BinaryExpr) Offset() tokens.Pos     { return x.X.Offset() }
func (x *KeyValueExpr) Offset() tokens.Pos   { return x.Key.Offset() }
func (x *ArrayType) Offset() tokens.Pos      { return x.Lbrack }
func (x *StructType) Offset() tokens.Pos     { return x.Struct }
func (x *FuncType) Offset() tokens.Pos {
	if x.Func.IsValid() || x.Params == nil { // see issue 3870
		return x.Func
	}
	return x.Params.Offset() // interface method declarations have no "func" keyword
}
func (x *InterfaceType) Offset() tokens.Pos { return x.Interface }
func (x *MapType) Offset() tokens.Pos       { return x.Map }

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
func IsExported(name string) bool { return tokens.IsExported(name) }

// IsExported reports whether id starts with an upper-case letter.
func (id *Name) IsExported() bool { return tokens.IsExported(id.Name) }

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
		From, To tokens.Pos // position range of bad statement
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
		Semicolon tokens.Pos // position of following ";"
		Implicit  bool       // if set, ";" was omitted in the source
	}

	// A LabeledStmt node represents a labeled statement.
	LabeledStmt struct {
		Label *Name
		Colon tokens.Pos // position of ":"
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
		Arrow tokens.Pos // position of "<-"
		Value Expr
	}

	// An IncDecStmt node represents an increment or decrement statement.
	IncDecStmt struct {
		X      Expr
		TokPos tokens.Pos   // position of Tok
		Tok    tokens.Token // INC or DEC
	}

	// An AssignStmt node represents an assignment or
	// a short variable declaration.
	//
	AssignStmt struct {
		Lhs    []Expr
		TokPos tokens.Pos   // position of Tok
		Tok    tokens.Token // assignment token, DEFINE
		Rhs    []Expr
	}

	// A GoStmt node represents a go statement.
	GoStmt struct {
		Go   tokens.Pos // position of "go" keyword
		Call *CallExpr
	}

	// A DeferStmt node represents a defer statement.
	DeferStmt struct {
		Defer tokens.Pos // position of "defer" keyword
		Call  *CallExpr
	}

	// A ReturnStmt node represents a return statement.
	ReturnStmt struct {
		Return  tokens.Pos // position of "return" keyword
		Results []Expr     // result expressions; or nil
	}

	// A BranchStmt node represents a break, continue, goto,
	// or fallthrough statement.
	//
	BranchStmt struct {
		TokPos tokens.Pos   // position of Tok
		Tok    tokens.Token // keyword token (BREAK, CONTINUE, GOTO, FALLTHROUGH)
		Label  *Name        // label name; or nil
	}

	// A BlockStmt node represents a braced statement list.
	BlockStmt struct {
		Lbrace tokens.Pos // position of "{"
		List   []Stmt
		Rbrace tokens.Pos // position of "}", if any (may be absent due to syntax error)
	}

	// An IfStmt node represents an if statement.
	IfStmt struct {
		If   tokens.Pos // position of "if" keyword
		Init Stmt       // initialization statement; or nil
		Cond Expr       // condition
		Body *BlockStmt
		Else Stmt // else branch; or nil
	}

	// A CaseClause represents a case of an expression or type switch statement.
	CaseClause struct {
		Case  tokens.Pos // position of "case" or "default" keyword
		List  []Expr     // list of expressions or types; nil means default case
		Colon tokens.Pos // position of ":"
		Body  []Stmt     // statement list; or nil
	}

	// A SwitchStmt node represents an expression switch statement.
	SwitchStmt struct {
		Switch tokens.Pos // position of "switch" keyword
		Init   Stmt       // initialization statement; or nil
		Tag    Expr       // tag expression; or nil
		Body   *BlockStmt // CaseClauses only
	}

	// A TypeSwitchStmt node represents a type switch statement.
	TypeSwitchStmt struct {
		Switch tokens.Pos // position of "switch" keyword
		Init   Stmt       // initialization statement; or nil
		Assign Stmt       // x := y.(type) or y.(type)
		Body   *BlockStmt // CaseClauses only
	}

	// A CommClause node represents a case of a select statement.
	CommClause struct {
		Case  tokens.Pos // position of "case" or "default" keyword
		Comm  Stmt       // send or receive statement; nil means default case
		Colon tokens.Pos // position of ":"
		Body  []Stmt     // statement list; or nil
	}

	// A SelectStmt node represents a select statement.
	SelectStmt struct {
		Select tokens.Pos // position of "select" keyword
		Body   *BlockStmt // CommClauses only
	}

	// A ForStmt represents a for statement.
	ForStmt struct {
		For  tokens.Pos // position of "for" keyword
		Init Stmt       // initialization statement; or nil
		Cond Expr       // condition; or nil
		Post Stmt       // post iteration statement; or nil
		Body *BlockStmt
	}

	// A RangeStmt represents a for statement with a range clause.
	RangeStmt struct {
		For        tokens.Pos   // position of "for" keyword
		Key, Value Expr         // Key, Value may be nil
		TokPos     tokens.Pos   // position of Tok; invalid if Key == nil
		Tok        tokens.Token // ILLEGAL if Key == nil, ASSIGN, DEFINE
		Range      tokens.Pos   // position of "range" keyword
		X          Expr         // value to range over
		Body       *BlockStmt
	}
)

// Pos and End implementations for statement nodes.

func (s *BadStmt) Offset() tokens.Pos        { return s.From }
func (s *DeclStmt) Offset() tokens.Pos       { return s.Decl.Offset() }
func (s *EmptyStmt) Offset() tokens.Pos      { return s.Semicolon }
func (s *LabeledStmt) Offset() tokens.Pos    { return s.Label.Offset() }
func (s *ExprStmt) Offset() tokens.Pos       { return s.X.Offset() }
func (s *SendStmt) Offset() tokens.Pos       { return s.Chan.Offset() }
func (s *IncDecStmt) Offset() tokens.Pos     { return s.X.Offset() }
func (s *AssignStmt) Offset() tokens.Pos     { return s.Lhs[0].Offset() }
func (s *GoStmt) Offset() tokens.Pos         { return s.Go }
func (s *DeferStmt) Offset() tokens.Pos      { return s.Defer }
func (s *ReturnStmt) Offset() tokens.Pos     { return s.Return }
func (s *BranchStmt) Offset() tokens.Pos     { return s.TokPos }
func (s *BlockStmt) Offset() tokens.Pos      { return s.Lbrace }
func (s *IfStmt) Offset() tokens.Pos         { return s.If }
func (s *CaseClause) Offset() tokens.Pos     { return s.Case }
func (s *SwitchStmt) Offset() tokens.Pos     { return s.Switch }
func (s *TypeSwitchStmt) Offset() tokens.Pos { return s.Switch }
func (s *CommClause) Offset() tokens.Pos     { return s.Case }
func (s *SelectStmt) Offset() tokens.Pos     { return s.Select }
func (s *ForStmt) Offset() tokens.Pos        { return s.For }
func (s *RangeStmt) Offset() tokens.Pos      { return s.For }

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

// A Spec node represents a single (non-parenthesized) import,
// constant, type, or variable declaration.
type (
	// The Spec type stands for any of *ImportSpec, *ValueSpec, and *TypeSpec.
	//Spec interface {
	//	Node
	//	specNode()
	//}

	// A ValueSpec node represents a constant or variable declaration
	// (ConstSpec or VarSpec production).
	//
	ValueSpec struct {
		Names  []*Name // value names (len(Names) > 0)
		Type   Expr    // value type; or nil
		Values []Expr  // initial values; or nil
	}
)

// Pos and End implementations for spec nodes.

func (s *ValueSpec) Offset() tokens.Pos { return s.Names[0].Offset() }

// specNode() ensures that only spec nodes can be
// assigned to a Spec.
func (*ValueSpec) specNode() {}

// A declaration is represented by one of the following declaration nodes.
type (
	// A BadDecl node is a placeholder for a declaration containing
	// syntax errors for which a correct declaration node cannot be
	// created.
	//
	BadDecl struct {
		From, To tokens.Pos // position range of bad declaration
	}

	GenDecl struct {
		Pos    tokens.Pos   // position of Tok
		Token  tokens.Token // const or Let
		Names  []*Name      // value names (len(Names) > 0)
		Type   Expr         // value type; or nil
		Values []Expr       // initial values; or nil
	}

	// A FuncDecl node represents a function declaration.
	FuncDecl struct {
		Recv *FieldList // receiver (methods); or nil (functions)
		Name *Name      // function/method name
		Type *FuncType  // function signature: type and value parameters, results, and position of "func" keyword
		Body *BlockStmt // function body; or nil for external (non-Go) function
	}

	// TypeDecl type 语句：
	// 		type name1 int 新的类型
	// 		type name2 = int 类型别名
	//		type name3 struct {} 结构体
	// 		type name5 interface{} 接口
	TypeDecl struct {
		Pos        tokens.Pos // position of Tok
		Name       *Name      // type name
		Lparen     tokens.Pos // position of '(', if any
		TypeParams *FieldList // type parameters; or nil
		Assign     tokens.Pos // position of '=', if any
		Rparen     tokens.Pos // position of ')', if any
		Type       Expr       // *Ident, *ParenExpr, *SelectorExpr, *StarExpr, or any of the *XxxTypes
	}
)

// Pos and End implementations for declaration nodes.

func (d *BadDecl) Offset() tokens.Pos  { return d.From }
func (d *GenDecl) Offset() tokens.Pos  { return d.Pos }
func (d *FuncDecl) Offset() tokens.Pos { return d.Type.Offset() }

// declNode() ensures that only declaration nodes can be
// assigned to a Decl.
func (*BadDecl) declNode()  {}
func (*GenDecl) declNode()  {}
func (*FuncDecl) declNode() {}

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
	Pos  tokens.Pos // 位置信息
	Name string     // 别名
	Kind string     // 类型
}

// import ""
// import default from ""
// import {} // todo 暂时不支持解包

type Package struct {
	Pos  tokens.Pos // 位置信息
	Name string     //
	Path string     // import path
}

// End returns the end of the last declaration in the file.
// It may be invalid, for example in an empty file.
//
// (Use FileEnd for the end of the entire file. It is always valid.)
//func (f *File) End() tokens.Pos {
//	if n := len(f.Decls); n > 0 {
//		return f.Decls[n-1].End()
//	}
//	return f.Name.End()
//}

// A Package node represents a set of source files
// collectively building a Go package.
//
// Deprecated: use the type checker [go/types] instead; see [Object].
//type Package struct {
//	Name    string             // package name
//	Scope   *Scope             // package scope across all files
//	Imports map[string]*Object // map of package id -> package object
//	Files   map[string]*File   // Go source files by filename
//}
//
//func (p *Package) Offset() tokens.Pos { return tokens.NoPos }
//func (p *Package) End() tokens.Pos    { return tokens.NoPos }

// Unparen returns the expression with any enclosing parentheses removed.
func Unparen(e Expr) Expr {
	for {
		paren, ok := e.(*ParenExpr)
		if !ok {
			return e
		}
		e = paren.X
	}
}
