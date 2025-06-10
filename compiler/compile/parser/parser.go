package parser

import (
	"fmt"
	"github.com/facelang/face/compiler/compile/ast"
	"github.com/facelang/face/compiler/compile/token"
)

type parser struct {
	*lexer              // 符号读取器
	token   token.Token // 符号
	literal string      // 字面量
	exprLev int         // 表达式层级
	inRhs   bool        // 是否右值表达式
	nestLev int         // 递归嵌套计数器
}

func (p *parser) next() {
	for {
		p.token = p.NextToken()
		p.literal += p.identifier
		if p.token == token.COMMENT {
			continue
		}
		if p.token == token.NEWLINE {
			continue
		}
		break
	}
}

func (p *parser) got(token token.Token) bool {
	if p.token == token {
		p.next()
		return true
	}
	return false
}

func (p *parser) error(pos token.Pos, msg string) {
	//if p.errors.Len() > 10 {
	//	panic(p.errors)
	//}
	//p.errors.Add(pos, msg)
	panic(fmt.Errorf("%s:%d:%d: %s", pos, p.literal, pos, msg))
}

func (p *parser) errorf(format string, args ...interface{}) {
	//p.errors.Add(p.pos, fmt.Sprintf(format, args...))
}

func (p *parser) expect(token token.Token) token.Pos {
	pos := p.pos
	if p.token != token {
		p.unexpect(token.String())
	}

	p.next()
	return pos
}

func (p *parser) unexpect(except string) {
	found := token.TokenLabel(p.token, p.identifier)
	p.errorf("except %s, found %s", except, found)
}

// ----------------------------------------------------------------------------
// Identifiers

// name = identifier .
func (p *parser) name() *ast.Name {
	if p.token != token.IDENT {
		p.unexpect("identifier")
	}

	n := new(ast.Name)
	n.Pos = p.pos
	n.Name = p.literal

	p.next()
	return n
}

// nameList = name { "," name } .
func (p *parser) nameList(name *ast.Name) []*ast.Name {
	list := []*ast.Name{name}
	for p.token == token.COMMA {
		p.next()
		list = append(list, p.name())
	}
	return list
}

// 参考 ES6 import {} from "" 语法
// 暂不支持解包，只支持两种语法：
// import name from ""
// import ""
func (p *parser) pkg() *ast.Package {
	d := &ast.Package{Pos: p.expect(token.IMPORT)}

	if p.token == token.IDENT {
		d.Name = p.literal
		p.expect(token.FROM)
	}

	d.Path = p.literal
	return d
}

// const name1, name2, ... type = val1, val2, ...
// let name1, name2, ... type = val1, val2, ...
func (p *parser) genDecl(require token.Token) ast.Decl {
	pos := p.expect(require)

	names := p.nameList(p.name())
	var typ ast.Expr
	var values []ast.Expr
	if p.token != token.EOF && p.token != token.SEMICOLON && p.token != token.RPAREN {
		typ = p.tryIdentOrType()
		if p.token == token.ASSIGN {
			p.next()
			values = exprList(p, true)
		}
	}

	return &ast.GenDecl{
		Pos:    pos,
		Token:  require,
		Names:  names,
		Type:   typ,
		Values: values,
	}
}

func (p *parser) funcDecl() ast.Decl {
	pos := p.expect(token.FUNC)
	name := p.name()

	// 参数列表，包括泛型参数
	_, params := p.parseParameters(true)

	results := p.parseResult() // (...) 返回结果

	var body *ast.BlockStmt
	switch p.token {
	case token.LBRACE: // {}
		body = p.parseBody()
	case token.ASSIGN:
	// todo 单行表达式
	default:
		// 第二种情况： func func2(a, b int) [int] = a + b
		// 第三种情况： const func3 = (a, b) => a + b
		// 				const func4 = func() {}
		//				const func5 = func4 别名
		panic("函数声明 func name(){} 或者 func name() = express")
	}

	return &ast.FuncDecl{
		Pos:  pos,
		Name: name,
		Type: &ast.FuncType{
			Params:  params,
			Results: results,
		},
		Body: body,
	}
}

// SourceFile = { ImportDecl ";" } { TopLevelDecl ";" } .
func (p *parser) parseFile() *ast.File {
	f := new(ast.File)

	prev := token.EOF
	for p.token != token.EOF {
		prev = p.token

		switch p.token {
		case token.IMPORT:
			if prev != token.IMPORT {
				p.error(p.pos, "import 语法只能出现在文件头部！")
			}
			f.Imports = append(f.Imports, p.pkg())
		case token.CONST, token.LET:
			f.DeclList = append(f.DeclList, p.genDecl(p.token))
		case token.FUNC:
			p.next()
			f.DeclList = append(f.DeclList, p.funcDecl())
		default:
			p.error(p.pos, "顶层语法仅支持 const, let, type, func 关键字定义！")
		}
	}

	return f
}

func (p *parser) parseBody() *ast.BlockStmt {
	lbrace := p.expect(token.LBRACE) // {
	list := p.parseStmtList()
	rbrace := p.expect(token.RBRACE) // }

	return &ast.BlockStmt{Lbrace: lbrace, List: list, Rbrace: rbrace}
}

func (p *parser) parseBlockStmt() *ast.BlockStmt {
	return p.parseBody()
}

// gotAssign = "=" .
func (p *parser) gotAssign() bool {
	if p.token == token.ASSIGN {
		p.next()
		return true
	}
	return false
}

// block{}, case:, select case 会调用
func (p *parser) parseStmtList() (list []ast.Stmt) {
	for p.token != token.CASE && p.token != token.DEFAULT && p.token != token.RBRACE && p.token != token.EOF {
		list = append(list, p.parseStmt())
	}

	return
}

func (p *parser) parseStmt() (s ast.Stmt) {
	defer decNestLev(incNestLev(p))

	switch p.token {
	case token.CONST, token.LET:
		s = &ast.DeclStmt{Decl: p.genDecl(p.token)}
	case
		token.IDENT, token.INT, token.FLOAT, token.IMAG, token.CHAR, token.STRING, token.FUNC, token.LPAREN, // operands
		token.LBRACK, token.STRUCT, token.MAP, token.CHAN, token.INTERFACE, // composite types
		token.ADD, token.SUB, token.MUL, token.AND, token.XOR, token.ARROW, token.NOT: // unary operators
		s, _ = p.parseSimpleStmt(labelOk)
		// because of the required look-ahead, labeled statements are
		// parsed by parseSimpleStmt - don't expect a semicolon after
		// them
		if _, isLabeledStmt := s.(*ast.LabeledStmt); !isLabeledStmt {
			p.expectSemi()
		}
	case token.RETURN:
		s = p.parseReturnStmt()
	case token.BREAK, token.CONTINUE, token.GOTO, token.FALLTHROUGH:
		s = p.parseBranchStmt(p.token)
		// todo 存在块代码嵌套需要处理{ {} }
	case token.IF:
		s = p.parseIfStmt()
	case token.SWITCH:
		s = p.parseSwitchStmt()
	case token.FOR:
		s = p.parseForStmt()
	case token.SEMICOLON:
		// Is it ever possible to have an implicit semicolon
		// producing an empty statement in a valid program?
		// (handle correctly anyway)
		s = &ast.EmptyStmt{Semicolon: p.pos, Implicit: p.lit == "\n"}
		p.next()
	case token.RBRACE:
		// a semicolon may be omitted before a closing "}"
		s = &ast.EmptyStmt{Semicolon: p.pos, Implicit: true}
	default:
		// no statement found
		pos := p.pos
		p.errorExpected(pos, "statement")
		p.advance(stmtStart)
		s = &ast.BadStmt{From: pos, To: p.pos}
	}

	return
}

// ----------------------------------------------------------------------------
// Statements

// Parsing modes for parseSimpleStmt.
const (
	basic = iota
	labelOk
	rangeOk
)

// parseSimpleStmt returns true as 2nd result if it parsed the assignment
// of a range clause (with mode == rangeOk). The returned statement is an
// assignment with a right-hand side that is a single unary expression of
// the form "range x". No guarantees are given for the left-hand side.
func (p *parser) parseSimpleStmt(mode int) (ast.Stmt, bool) {
	if p.trace {
		defer un(trace(p, "SimpleStmt"))
	}

	x := p.parseList(false)

	switch p.tok {
	case
		token.DEFINE, token.ASSIGN, token.ADD_ASSIGN,
		token.SUB_ASSIGN, token.MUL_ASSIGN, token.QUO_ASSIGN,
		token.REM_ASSIGN, token.AND_ASSIGN, token.OR_ASSIGN,
		token.XOR_ASSIGN, token.SHL_ASSIGN, token.SHR_ASSIGN, token.AND_NOT_ASSIGN:
		// assignment statement, possibly part of a range clause
		pos, tok := p.pos, p.tok
		p.next()
		var y []ast.Expr
		isRange := false
		if mode == rangeOk && p.tok == token.RANGE && (tok == token.DEFINE || tok == token.ASSIGN) {
			pos := p.pos
			p.next()
			y = []ast.Expr{&ast.UnaryExpr{OpPos: pos, Op: token.RANGE, X: p.parseRhs()}}
			isRange = true
		} else {
			y = p.parseList(true)
		}
		return &ast.AssignStmt{Lhs: x, TokPos: pos, Tok: tok, Rhs: y}, isRange
	}

	if len(x) > 1 {
		p.errorExpected(x[0].Pos(), "1 expression")
		// continue with first expression
	}

	switch p.tok {
	case token.COLON:
		// labeled statement
		colon := p.pos
		p.next()
		if label, isIdent := x[0].(*ast.Ident); mode == labelOk && isIdent {
			// Go spec: The scope of a label is the body of the function
			// in which it is declared and excludes the body of any nested
			// function.
			stmt := &ast.LabeledStmt{Label: label, Colon: colon, Stmt: p.parseStmt()}
			return stmt, false
		}
		// The label declaration typically starts at x[0].Pos(), but the label
		// declaration may be erroneous due to a token after that position (and
		// before the ':'). If SpuriousErrors is not set, the (only) error
		// reported for the line is the illegal label error instead of the token
		// before the ':' that caused the problem. Thus, use the (latest) colon
		// position for error reporting.
		p.error(colon, "illegal label declaration")
		return &ast.BadStmt{From: x[0].Pos(), To: colon + 1}, false

	case token.ARROW:
		// send statement
		arrow := p.pos
		p.next()
		y := p.parseRhs()
		return &ast.SendStmt{Chan: x[0], Arrow: arrow, Value: y}, false

	case token.INC, token.DEC:
		// increment or decrement
		s := &ast.IncDecStmt{X: x[0], TokPos: p.pos, Tok: p.token}
		p.next()
		return s, false
	}

	// expression
	return &ast.ExprStmt{X: x[0]}, false
}

func (p *parser) parseReturnStmt() *ast.ReturnStmt {
	pos := p.pos
	p.expect(token.RETURN)
	var x []ast.Expr
	if p.token != token.SEMICOLON && p.token != token.RBRACE {
		x = exprList(p, true)
	}
	p.expectSemi()

	return &ast.ReturnStmt{Return: pos, Results: x}
}

func (p *parser) parseBranchStmt(tok token.Token) *ast.BranchStmt {
	pos := p.expect(tok)
	var label *ast.Name
	if tok != token.FALLTHROUGH && p.token == token.IDENT {
		label = p.name()
	}
	p.expectSemi()

	return &ast.BranchStmt{TokPos: pos, Tok: tok, Label: label}
}

func (p *parser) makeExpr(s ast.Stmt, want string) ast.Expr {
	if s == nil {
		return nil
	}
	if es, isExpr := s.(*ast.ExprStmt); isExpr {
		return es.X
	}
	found := "simple statement"
	if _, isAss := s.(*ast.AssignStmt); isAss {
		found = "assignment"
	}
	p.error(s.Position(), fmt.Sprintf("expected %s, found %s (missing parentheses around composite literal?)", want, found))
	return &ast.BadExpr{From: s.Position(), To: p.safePos(s.End())}
}

func (p *parser) parseIfHeader() (init ast.Stmt, cond ast.Expr) {
	if p.token == token.LBRACE {
		p.error(p.pos, "missing condition in if statement")
		cond = &ast.BadExpr{From: p.pos, To: p.pos}
		return
	}
	// p.tok != token.LBRACE

	prevLev := p.exprLev // 记录层级
	p.exprLev = -1

	if p.token != token.SEMICOLON { // 初始化语句
		// accept potential variable declaration but complain
		if p.token == token.LET {
			p.next()
			p.error(p.pos, "var declaration not allowed in if initializer")
		}
		init, _ = p.parseSimpleStmt(basic)
	}

	var condStmt ast.Stmt // 条件语句
	var semi struct {
		pos token.Pos
		lit string // ";" or "\n"; valid if pos.IsValid()
	}
	if p.token != token.LBRACE { // {}
		if p.token == token.SEMICOLON { // ;
			semi.pos = p.pos
			semi.lit = p.identifier
			p.next()
		} else {
			p.expect(token.SEMICOLON)
		}
		if p.token != token.LBRACE { // 条件语句, 可能是 if ; {}
			condStmt, _ = p.parseSimpleStmt(basic)
		}
	} else {
		condStmt = init
		init = nil
	}

	if condStmt != nil {
		cond = p.makeExpr(condStmt, "boolean expression")
	} else if semi.pos.IsValid() {
		if semi.lit == "\n" {
			p.error(semi.pos, "unexpected newline, expecting { after if clause")
		} else {
			p.error(semi.pos, "missing condition in if statement")
		}
	}

	// make sure we have a valid AST
	if cond == nil {
		cond = &ast.BadExpr{From: p.pos, To: p.pos}
	}

	p.exprLev = prevLev
	return
}

func (p *parser) parseIfStmt() *ast.IfStmt {
	defer decNestLev(incNestLev(p))

	pos := p.expect(token.IF)

	init, cond := p.parseIfHeader()
	body := p.parseBody() // parseBlockStmt

	var else_ ast.Stmt
	if p.token == token.ELSE {
		p.next()
		switch p.token {
		case token.IF:
			else_ = p.parseIfStmt()
		case token.LBRACE:
			else_ = p.parseBlockStmt()
			p.expectSemi()
		default:
			p.errorExpected(p.pos, "if statement or block")
			else_ = &ast.BadStmt{From: p.pos, To: p.pos}
		}
	} else {
		p.expectSemi()
	}

	return &ast.IfStmt{If: pos, Init: init, Cond: cond, Body: body, Else: else_}
}
