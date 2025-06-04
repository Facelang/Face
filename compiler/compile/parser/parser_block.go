package parser

import (
	"github.com/facelang/face/compiler/compile/internal/tokens"
	tokens2 "github.com/facelang/face/compiler/compile/tokens"
	"go/ast"
	"go/token"
)

// ----------------------------------------------------------------------------
// Blocks

// 与函数 parseBlockStmt 完全等价！
func (p *parser) parseBody() *ast.BlockStmt {
	lbrace := p.expect(LBRACE) // {
	list := p.parseStmtList()
	rbrace := p.expect(RBRACE) // }

	return &ast.BlockStmt{Lbrace: lbrace, List: list, Rbrace: rbrace}
}

// ----------------------------------------------------------------------------
// Statements

// parseSimpleStmt returns true as 2nd result if it parsed the assignment
// of a range clause (with mode == rangeOk). The returned statement is an
// assignment with a right-hand side that is a single unary expression of
// the form "range x". No guarantees are given for the left-hand side.
func (p *parser) parseSimpleStmt(mode int) (ast.Stmt, bool) {
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
		s := &ast.IncDecStmt{X: x[0], TokPos: p.pos, Tok: p.tok}
		p.next()
		return s, false
	}

	// expression
	return &ast.ExprStmt{X: x[0]}, false
}

func (p *parser) parseStmt() (s ast.Stmt) {
	//defer decNestLev(incNestLev(p))

	switch p.token {
	case CONST, TYPE, LET:
		s = &ast.DeclStmt{Decl: p.parseDecl()} // 声明语句
	case
		// tokens that may start an expression
		tokens2.IDENT, tokens2.INT, tokens2.FLOAT, tokens2.IMAG, tokens.BYTE, tokens2.STRING,
		FUNC, LPAREN, // operands
		LBRACK, STRUCT, MAP, CHAN, INTERFACE, // composite types
		ADD, SUB, MUL, AND, XOR, ARROW, NOT: // unary operators
		s, _ = p.parseSimpleStmt(labelOk)
		// because of the required look-ahead, labeled statements are
		// parsed by parseSimpleStmt - don't expect a semicolon after
		// them
		if _, isLabeledStmt := s.(*ast.LabeledStmt); !isLabeledStmt {
			p.expectSemi()
		}
	case GO:
		s = p.parseGoStmt()
	case DEFER:
		s = p.parseDeferStmt()
	case RETURN:
		s = p.parseReturnStmt()
	case BREAK, CONTINUE, GOTO, FALLTHROUGH:
		s = p.parseBranchStmt(p.tok)
	case LBRACE:
		s = p.parseBlockStmt()
		p.expectSemi()
	case IF:
		s = p.parseIfStmt()
	case SWITCH:
		s = p.parseSwitchStmt()
	case SELECT:
		s = p.parseSelectStmt()
	case FOR:
		s = p.parseForStmt()
	case SEMICOLON:
		// Is it ever possible to have an implicit semicolon
		// producing an empty statement in a valid program?
		// (handle correctly anyway)
		s = &ast.EmptyStmt{Semicolon: p.pos, Implicit: p.lit == "\n"}
		p.next()
	case RBRACE:
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

// block{}, case:, select case 会调用
func (p *parser) parseStmtList() (list []ast.Stmt) {
	for p.token != CASE && p.token != DEFAULT && p.token != RBRACE && p.token != tokens2.EOF {
		list = append(list, p.parseStmt())
	}

	return
}
