package parser

import (
	"fmt"
	"github.com/facelang/face/compiler/compile/ast"
	"github.com/facelang/face/compiler/compile/tokens"
)

type parser struct {
	*lexer                // 符号读取器
	token   tokens.Token  // 符号
	literal string        // 字面量
	errors  ast.ErrorList // 异常列表
	inRhs   bool          // if set, the parser is parsing a rhs expression
}

//func (p *parser) nextToken() tokens.Token {
//	token := p.lexer.NextToken()
//	for token == compile.COMMENT {
//		token = p.lexer.NextToken()
//	}
//	return token
//}
//
//func (p *parser) NextToken() (tokens.Token, string, int, int) {
//	token := p.nextToken()
//	c := p.lexer.content
//	return token, c, p.lexer.line, p.lexer.col
//}

func (p *parser) next() {
	for {
		p.token = p.NextToken()
		p.literal += p.identifier
		if p.token == tokens.COMMENT {
			continue
		}
		if p.token == tokens.NEWLINE {
			continue
		}
		break
	}
}

func (p *parser) got(token tokens.Token) bool {
	if p.token == token {
		p.next()
		return true
	}
	return false
}

func (p *parser) error(pos tokens.Pos, msg string) {
	if p.errors.Len() > 10 {
		panic(p.errors)
	}
	p.errors.Add(pos, msg)
}

func (p *parser) errorf(format string, args ...interface{}) {
	p.errors.Add(p.pos, fmt.Sprintf(format, args...))
}

func (p *parser) expect(token tokens.Token) tokens.Pos {
	pos := p.pos
	if p.token != token {
		p.unexpect(token.String())
	}

	p.next()
	return pos
}

func (p *parser) unexpect(except string) {
	found := tokens.TokenLabel(p.token, p.identifier)
	p.errorf("except %s, found %s", except, found)
}

// expectClosing is like expect but provides a better error message
// for the common case of a missing comma before a newline.
func (p *parser) expectClosing(token tokens.Token, context string) tokens.Pos {
	//if p.token != token && p.token == SEMICOLON && p.literal == "\n" {
	if p.token != token && p.token == tokens.NEWLINE {
		p.error(p.pos, "missing ',' before newline in "+context)
		p.next()
	}
	return p.expect(token)
}

// expectSemi consumes a semicolon and returns the applicable line comment.
func (p *parser) expectSemi() (comment *ast.CommentGroup) {
	// semicolon is optional before a closing ')' or '}'
	if p.token != RPAREN && p.token != RBRACE {
		switch p.token {
		case COMMA:
			// permit a ',' instead of a ';' but complain
			p.errorExpected(p.Pos, "';'")
			fallthrough
		case SEMICOLON:
			p.next()
			return comment
		default:
			p.errorExpected(p.Pos, "';'")
			p.advance(stmtStart)
		}
	}
	return nil
}

// gotAssign is like got(_Assign) but it also accepts ":="
// (and reports an error) for better parser error recovery.
func (p *parser) gotAssign() bool {
	switch p.tok {
	case _Define:
		p.syntaxError("expected =")
		fallthrough
	case _Assign:
		p.next()
		return true
	}
	return false
}

// ----------------------------------------------------------------------------
// Identifiers

// name = identifier .
func (p *parser) name() *ast.Name {
	if p.token != tokens.IDENT {
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
	for p.token == tokens.COMMA {
		p.next()
		list = append(list, p.name())
	}
	return list
}

// 参考 ES6 import {} from "" 语法
// 暂不支持解包，只支持两种语法：
// import name from ""
// import ""
func (p *parser) importDecl() ast.Decl {
	d := new(ast.ImportDecl)
	d.SetPos(p.Pos())

	if p.token == tokens.IDENT {
		d.Alias = p.literal
		p.expect(FROM)
	}

	d.Path = p.literal
	p.expect(tokens.STRING)

	return d
}

// ConstSpec = IdentifierList [ [ Type ] "=" ExpressionList ] .
func (p *parser) constDecl() ast.Decl {
	d := new(ast.ConstDecl)
	d.SetPos(p.Pos())

	d.NameList = p.nameList(p.name())
	if p.token != tokens.EOF && p.token != SEMICOLON && p.token != _Rparen {
		d.Type = p.typeOrNil()
		if p.gotAssign() {
			d.Values = p.exprList()
		}
	}

	return d
}

// VarSpec = IdentifierList ( Type [ "=" ExpressionList ] | "=" ExpressionList ) .
func (p *parser) letDecl() ast.Decl {
	d := new(ast.VarDecl)
	d.pos = p.pos()

	d.NameList = p.nameList(p.name())
	if p.gotAssign() { // 没有类型
		d.Values = p.exprList()
	} else { // 有类型
		d.Type = RequireType(p)
		if p.gotAssign() {
			d.Values = p.exprList()
		}
	}

	return d
}

// TypeDecl = "type" ( TypeSpec | "(" { TypeSpec ";" } ")" ) .
func (p *parser) typeDecl() ast.Decl {
	d := new(ast.TypeDecl)
	d.SetPos(p.Pos())

	// 解析类型名称
	d.Name = p.name()

	// 解析类型定义
	if p.token == tokens.ASSIGN {
		p.next()
		d.Type = p.typeOrNil()
	} else {
		p.error(p.Pos, "类型声明需要指定类型定义")
	}

	return d
}

// FuncDecl = "func" FunctionName Signature [ FunctionBody ] .
// FunctionName = identifier .
// Signature = Parameters [ Result ] .
// Result = Parameters | Type .
func (p *parser) funcDecl() ast.Decl {
	d := new(ast.FuncDecl)
	d.SetPos(p.Pos())

	// 解析函数名
	d.Name = p.name()

	// 解析参数列表
	if p.token == tokens.LPAREN {
		p.next()
		d.Params = p.paramList()
		p.expect(tokens.RPAREN)
	}

	// 解析返回值类型
	if p.token != tokens.LBRACE {
		d.Result = p.typeOrNil()
	}

	// 解析函数体
	if p.token == tokens.LBRACE {
		p.next()
		d.Body = p.blockStmt()
	}

	return d
}

// SourceFile = { ImportDecl ";" } { TopLevelDecl ";" } .
func (p *parser) parseFile() *ast.File {

	f := new(ast.File)

	// import decls
	for p.token == IMPORT {
		p.next()
		f.DeclList = append(f.DeclList, p.importDecl())
	}

	for p.token != tokens.EOF {
		if p.token == IMPORT {
			p.error(p.Pos, "import 语法只能出现在文件头部！")
		}

		switch p.token {
		case CONST:
			p.next()
			f.DeclList = append(f.DeclList, p.constDecl())
		case LET:
			p.next()
			f.DeclList = append(f.DeclList, p.letDecl())
		case TYPE:
			p.next()
			f.DeclList = append(f.DeclList, p.typeDecl())
		case FUNC:
			p.next()
			f.DeclList = append(f.DeclList, p.funcDecl())
		default:
			p.error(p.Pos, "顶层语法仅支持 const, let, type, func 关键字定义！")
		}
	}

	return f
}

// gotAssign = "=" .
func (p *parser) gotAssign() bool {
	if p.token == tokens.ASSIGN {
		p.next()
		return true
	}
	return false
}

// exprList = Expression { "," Expression } .
func (p *parser) exprList() []*internal.ProgDec {
	var list []*internal.ProgDec
	var vn int
	list = append(list, internal.expr(p, &vn))
	for p.token == tokens.COMMA {
		p.next()
		list = append(list, internal.expr(p, &vn))
	}
	return list
}

// paramList = ParameterDecl { "," ParameterDecl } .
func (p *parser) paramList() []*ast.ParamDecl {
	var list []*ast.ParamDecl
	list = append(list, p.paramDecl())
	for p.token == tokens.COMMA {
		p.next()
		list = append(list, p.paramDecl())
	}
	return list
}

// ParameterDecl = IdentifierList Type .
func (p *parser) paramDecl() *ast.ParamDecl {
	d := new(ast.ParamDecl)
	d.SetPos(p.Pos())

	d.NameList = p.nameList(p.name())
	d.Type = p.typeOrNil()

	return d
}

// blockStmt = "{" StatementList "}" .
func (p *parser) blockStmt() *ast.BlockStmt {
	s := new(ast.BlockStmt)
	s.SetPos(p.Pos())

	for p.token != tokens.RBRACE && p.token != tokens.EOF {
		s.List = append(s.List, p.stmt())
	}

	p.expect(tokens.RBRACE)
	return s
}
