package prog

import "go/token"

// ----------------------------------------------------------------------------
// Nodes

type Node interface {
	Pos() FilePos
	SetPos(FilePos)
	node()
}

type node struct {
	pos FilePos
}

func (n *node) Pos() FilePos       { return n.pos }
func (n *node) SetPos(pos FilePos) { n.pos = pos }
func (*node) node()                {}

// ----------------------------------------------------------------------------
// Files

// package PkgName; DeclList[0], DeclList[1], ...
type File struct {
	PkgName  *Name
	DeclList []Decl
	EOF      FilePos
	Version  string
	node
}

// ----------------------------------------------------------------------------
// Declarations

type (
	Decl interface {
		Node
		aDecl()
	}

	// GenDecl 通用语句: 常量、变量（包括函数形式， 内部结构体）的定义
	GenDecl struct { // 这是一个数组
		Pos   FilePos     // position of Tok
		Token token.Token // IMPORT, CONST, TYPE, or VAR
		Specs []Spec
	}

	// FuncDecl 函数定义
	FuncDecl struct {
		Recv       *Field // nil means regular function
		Name       *Name
		TParamList []*Field // nil means no type parameters
		Type       *FuncType
		Body       *BlockStmt // nil means no body (forward declaration)
		decl
	}

	// ImportDecl 包导入
	ImportDecl struct {
		Alias string // 别名
		Path  string // 路径
		decl
	}

	// TypeDecl 数据类型定义， 结构体， 接口
	TypeDecl struct {
		Name       *Name
		TParamList []*Field // nil means no type parameters
		Alias      bool
		Type       Expr
		decl
	}

	// NameList
	// NameList      = Values
	// NameList Type = Values
	ValueDecl struct {
		NameList []*Name
		Type     Expr // nil means no type
		Values   Expr // nil means no values
		decl
	}
)

type decl struct{ node }

func (*decl) aDecl() {}

// ----------------------------------------------------------------------------
// Expressions

func NewName(pos FilePos, value string) *Name {
	n := new(Name)
	n.pos = pos
	n.Value = value
	return n
}

type inode struct {
	Pos FilePos
}

type ErrorExpr struct {
	inode
}

type (
	Expr interface {
		Node
		typeInfo
		aExpr()
	}

	// Placeholder for an expression that failed to parse
	// correctly and where we can't provide a better node.
	BadExpr struct {
		expr
	}

	// Value
	Name struct {
		Value string
		expr
	}

	// Value
	BasicLit struct {
		Value string
		Kind  LitKind
		Bad   bool // true means the literal Value has syntax errors
		expr
	}

	// Type { ElemList[0], ElemList[1], ... }
	CompositeLit struct {
		Type     Expr // nil means no literal type
		ElemList []Expr
		NKeys    int // number of elements with keys
		Rbrace   Pos
		expr
	}

	// Key: Value
	KeyValueExpr struct {
		Key, Value Expr
		expr
	}

	// func Type { Body }
	FuncLit struct {
		Type *FuncType
		Body *BlockStmt
		expr
	}

	// (X)
	ParenExpr struct {
		X Expr
		expr
	}

	// X.Sel
	SelectorExpr struct {
		X   Expr
		Sel *Name
		expr
	}

	// X[Index]
	// X[T1, T2, ...] (with Ti = Index.(*ListExpr).ElemList[i])
	IndexExpr struct {
		X     Expr
		Index Expr
		expr
	}

	// X[Index[0] : Index[1] : Index[2]]
	SliceExpr struct {
		X     Expr
		Index [3]Expr
		// Full indicates whether this is a simple or full slice expression.
		// In a valid AST, this is equivalent to Index[2] != nil.
		// TODO(mdempsky): This is only needed to report the "3-index
		// slice of string" error when Index[2] is missing.
		Full bool
		expr
	}

	// X.(Type)
	AssertExpr struct {
		X    Expr
		Type Expr
		expr
	}

	// X.(type)
	// Lhs := X.(type)
	TypeSwitchGuard struct {
		Lhs *Name // nil means no Lhs :=
		X   Expr  // X.(type)
		expr
	}

	Operation struct {
		Op   Operator
		X, Y Expr // Y == nil means unary expression
		expr
	}

	// Fun(ArgList[0], ArgList[1], ...)
	CallExpr struct {
		Fun     Expr
		ArgList []Expr // nil means no arguments
		HasDots bool   // last argument is followed by ...
		expr
	}

	// ElemList[0], ElemList[1], ...
	ListExpr struct {
		ElemList []Expr
		expr
	}

	// [Len]Elem
	ArrayType struct {
		// TODO(gri) consider using Name{"..."} instead of nil (permits attaching of comments)
		Len  Expr // nil means Len is ...
		Elem Expr
		expr
	}

	// []Elem
	SliceType struct {
		Elem Expr
		expr
	}

	// ...Elem
	DotsType struct {
		Elem Expr
		expr
	}

	StructType struct {
		FieldList []*Field
		TagList   []*BasicLit // i >= len(TagList) || TagList[i] == nil means no tag for field i
		expr
	}

	Field struct {
		Name *Name // nil means anonymous field/parameter (structs/parameters), or embedded element (interfaces)
		Type Expr  // field names declared in a list share the same Type (identical pointers)
		node
	}

	InterfaceType struct {
		MethodList []*Field
		expr
	}

	FuncType struct {
		ParamList  []*Field
		ResultList []*Field
		expr
	}

	// map[Key]Value
	MapType struct {
		Key, Value Expr
		expr
	}

	//   chan Elem
	// <-chan Elem
	// chan<- Elem
	ChanType struct {
		Dir  ChanDir // 0 means no direction
		Elem Expr
		expr
	}

	BadExpr struct {
		From, To token.Pos // position range of bad expression
	}

	// An Ident node represents an identifier.
	Ident struct {
		NamePos token.Pos // identifier position
		Name    string    // identifier name
		Obj     *Object   // denoted object, or nil. Deprecated: see Object.
	}

	// An Ellipsis node stands for the "..." type in a
	// parameter list or the "..." length in an array type.
	//
	Ellipsis struct {
		Ellipsis token.Pos // position of "..."
		Elt      Expr      // ellipsis element type (parameter lists only); or nil
	}

	// A BasicLit node represents a literal of basic type.
	//
	// Note that for the CHAR and STRING kinds, the literal is stored
	// with its quotes. For example, for a double-quoted STRING, the
	// first and the last rune in the Value field will be ". The
	// [strconv.Unquote] and [strconv.UnquoteChar] functions can be
	// used to unquote STRING and CHAR values, respectively.
	//
	// For raw string literals (Kind == token.STRING && Value[0] == '`'),
	// the Value field contains the string text without carriage returns (\r) that
	// may have been present in the source. Because the end position is
	// computed using len(Value), the position reported by [BasicLit.End] does not match the
	// true source end position for raw string literals containing carriage returns.
	BasicLit struct {
		ValuePos token.Pos   // literal position
		Kind     token.Token // token.INT, token.FLOAT, token.IMAG, token.CHAR, or token.STRING
		Value    string      // literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f', "foo" or `\m\n\o`
	}

	// A FuncLit node represents a function literal.
	FuncLit struct {
		Type *FuncType  // function type
		Body *BlockStmt // function body
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
		X   Expr   // expression
		Sel *Ident // field selector
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
		Ellipsis token.Pos // position of "..." (token.NoPos if there is no "...")
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

type expr struct {
	node
	typeAndValue // After typechecking, contains the results of typechecking this expression.
}

func (*expr) aExpr() {}

type ChanDir uint

const (
	_ ChanDir = iota
	SendOnly
	RecvOnly
)

// ----------------------------------------------------------------------------
// Statements

type (
	Stmt interface {
		Node
		aStmt()
	}

	SimpleStmt interface {
		Stmt
		aSimpleStmt()
	}

	EmptyStmt struct {
		simpleStmt
	}

	LabeledStmt struct {
		Label *Name
		Stmt  Stmt
		stmt
	}

	BlockStmt struct {
		List   []Stmt
		Rbrace Pos
		stmt
	}

	ExprStmt struct {
		X Expr
		simpleStmt
	}

	SendStmt struct {
		Chan, Value Expr // Chan <- Value
		simpleStmt
	}

	DeclStmt struct {
		DeclList []Decl
		stmt
	}

	AssignStmt struct {
		Op       Operator // 0 means no operation
		Lhs, Rhs Expr     // Rhs == nil means Lhs++ (Op == Add) or Lhs-- (Op == Sub)
		simpleStmt
	}

	BranchStmt struct {
		Tok   token // Break, Continue, Fallthrough, or Goto
		Label *Name
		// Target is the continuation of the control flow after executing
		// the branch; it is computed by the parser if CheckBranches is set.
		// Target is a *LabeledStmt for gotos, and a *SwitchStmt, *SelectStmt,
		// or *ForStmt for breaks and continues, depending on the context of
		// the branch. Target is not set for fallthroughs.
		Target Stmt
		stmt
	}

	CallStmt struct {
		Tok     token // Go or Defer
		Call    Expr
		DeferAt Expr // argument to runtime.deferprocat
		stmt
	}

	ReturnStmt struct {
		Results Expr // nil means no explicit return values
		stmt
	}

	IfStmt struct {
		Init SimpleStmt
		Cond Expr
		Then *BlockStmt
		Else Stmt // either nil, *IfStmt, or *BlockStmt
		stmt
	}

	ForStmt struct {
		Init SimpleStmt // incl. *RangeClause
		Cond Expr
		Post SimpleStmt
		Body *BlockStmt
		stmt
	}

	SwitchStmt struct {
		Init   SimpleStmt
		Tag    Expr // incl. *TypeSwitchGuard
		Body   []*CaseClause
		Rbrace Pos
		stmt
	}

	SelectStmt struct {
		Body   []*CommClause
		Rbrace Pos
		stmt
	}
)

type (
	RangeClause struct {
		Lhs Expr // nil means no Lhs = or Lhs :=
		Def bool // means :=
		X   Expr // range X
		simpleStmt
	}

	CaseClause struct {
		Cases Expr // nil means default clause
		Body  []Stmt
		Colon Pos
		node
	}

	CommClause struct {
		Comm  SimpleStmt // send or receive stmt; nil means default clause
		Body  []Stmt
		Colon Pos
		node
	}
)

type stmt struct{ node }

func (stmt) aStmt() {}

type simpleStmt struct {
	stmt
}

func (simpleStmt) aSimpleStmt() {}
