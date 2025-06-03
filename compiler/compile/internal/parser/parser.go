package parser

import (
	"github.com/facelang/face/compiler/compile/internal"
	"github.com/facelang/face/compiler/compile/internal/api"
	"github.com/facelang/face/internal/ast"
	"github.com/facelang/face/internal/prog"
	"github.com/facelang/face/internal/tokens"
	"go/token"
	"os"
)

type ParserFactory interface {
	Parse() (interface{}, error)
	NextToken() (tokens.Token, string, int, int)
}

func OpenFile(filepath string) (*os.File, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filepath); err == nil {
		// 文件已存在，先删除
		err = os.Remove(filepath)
		if err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		// 如果是其他错误，直接返回
		return nil, err
	}

	// 创建文件
	file, err := os.Create(filepath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

type parser struct {
	*lexer                   // 符号读取器
	token   tokens.Token     // 符号
	literal string           // 字面量
	errors  tokens.ErrorList // 异常列表
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

func (p *parser) error(pos prog.FilePos, msg string) {
	if p.errors.Len() > 10 {
		panic(p.errors)
	}
	p.errors.Add(pos, msg)
}

func (p *parser) errorExpected(pos prog.FilePos, msg string) {
	msg = "expected " + msg
	switch {
	case p.token == tokens.NEWLINE:
		msg += ", found newline"
	case p.token.IsLiteral():
		msg += ", found " + p.identifier
	default:
		msg += ", found '" + p.token.String() + "'"
	}
	p.error(pos, msg)
}

func (p *parser) expect(token tokens.Token) prog.FilePos {
	pos := p.FilePos
	if p.token != token {
		p.errorExpected(pos, "'"+token.String()+"'")
	}
	p.next() // make progress
	return pos
}

// expect2 is like expect, but it returns an invalid position
// if the expected token is not found.
func (p *parser) expect2(token tokens.Token) (pos prog.FilePos) {
	if p.token == token {
		pos = p.FilePos
	} else {
		p.errorExpected(p.FilePos, "'"+token.String()+"'")
	}
	p.next() // make progress
	return
}

// expectClosing is like expect but provides a better error message
// for the common case of a missing comma before a newline.
func (p *parser) expectClosing(token tokens.Token, context string) prog.FilePos {
	//if p.token != token && p.token == SEMICOLON && p.literal == "\n" {
	if p.token != token && p.token == tokens.NEWLINE {
		p.error(p.FilePos, "missing ',' before newline in "+context)
		p.next()
	}
	return p.expect(token)
}

// expectSemi consumes a semicolon and returns the applicable line comment.
func (p *parser) expectSemi() (comment *ast.CommentGroup) {
	// semicolon is optional before a closing ')' or '}'
	if p.token != api.RPAREN && p.token != api.RBRACE {
		switch p.token {
		case api.COMMA:
			// permit a ',' instead of a ';' but complain
			p.errorExpected(p.FilePos, "';'")
			fallthrough
		case api.SEMICOLON:
			p.next()
			return comment
		default:
			p.errorExpected(p.FilePos, "';'")
			p.advance(stmtStart)
		}
	}
	return nil
}

func (p *parser) got(token tokens.Token) bool {
	if p.token == token {
		p.next()
		return true
	}
	return false
}

func (p *parser) want(token tokens.Token) {
	if !p.got(token) {
		p.syntaxError("expected " + tokstring(token))
		p.advance()
	}
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

func (p *parser) parseIdent() *ast.Ident {
	pos := p.pos
	name := "_"
	if p.tok == token.IDENT {
		name = p.lit
		p.next()
	} else {
		p.expect(token.IDENT) // use expect() error handling
	}
	return &ast.Ident{NamePos: pos, Name: name}
}

func (p *parser) parseIdentList() (list []*ast.Ident) {
	if p.trace {
		defer un(trace(p, "IdentList"))
	}

	list = append(list, p.parseIdent())
	for p.tok == token.COMMA {
		p.next()
		list = append(list, p.parseIdent())
	}

	return
}

// name = identifier .
func (p *parser) name() *prog.Name {
	if p.token != tokens.IDENT {
		name := p.literal
		p.next()
		return prog.NewName(p.FilePos, name)
	}
	p.errorExpected(p.FilePos, "identifier")
	return prog.NewName(p.FilePos, "_")
}

// nameList = name { "," name } .
func (p *parser) nameList(name *prog.Name) []*prog.Name {
	list := []*prog.Name{name}
	for p.token == api.COMMA {
		p.next()
		list = append(list, p.name())
	}
	return list
}

// 参考 ES6 import {} from "" 语法
// 暂不支持解包，只支持两种语法：
// import name from ""
// import ""
func (p *parser) importDecl() prog.Decl {
	d := new(prog.ImportDecl)
	d.SetPos(p.Pos())

	if p.token == tokens.IDENT {
		d.Alias = p.literal
		p.expect(api.FROM)
	}

	d.Path = p.literal
	p.expect(tokens.STRING)

	return d
}

// ConstSpec = IdentifierList [ [ Type ] "=" ExpressionList ] .
func (p *parser) constDecl() prog.Decl {
	d := new(prog.ConstDecl)
	d.SetPos(p.Pos())

	d.NameList = p.nameList(p.name())
	if p.token != tokens.EOF && p.token != api.SEMICOLON && p.token != _Rparen {
		d.Type = p.typeOrNil()
		if p.gotAssign() {
			d.Values = p.exprList()
		}
	}

	return d
}

// LetDecl = "let" IdentifierList [ Type ] [ "=" ExpressionList ] .
func (p *parser) letDecl() prog.Decl {
	d := new(prog.LetDecl)
	d.SetPos(p.Pos())

	d.NameList = p.nameList(p.name())
	if p.token != tokens.EOF && p.token != tokens.SEMICOLON && p.token != tokens.RPAREN {
		d.Type = p.typeOrNil()
		if p.gotAssign() {
			d.Values = p.exprList()
		}
	}

	return d
}

// TypeDecl = "type" ( TypeSpec | "(" { TypeSpec ";" } ")" ) .
func (p *parser) typeDecl() prog.Decl {
	d := new(prog.TypeDecl)
	d.SetPos(p.Pos())

	// 解析类型名称
	d.Name = p.name()

	// 解析类型定义
	if p.token == tokens.ASSIGN {
		p.next()
		d.Type = p.typeOrNil()
	} else {
		p.error(p.FilePos, "类型声明需要指定类型定义")
	}

	return d
}

// FuncDecl = "func" FunctionName Signature [ FunctionBody ] .
// FunctionName = identifier .
// Signature = Parameters [ Result ] .
// Result = Parameters | Type .
func (p *parser) funcDecl() prog.Decl {
	d := new(prog.FuncDecl)
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
func (p *parser) parseFile() *prog.File {

	f := new(prog.File)

	// import decls
	for p.token == api.IMPORT {
		p.next()
		f.DeclList = append(f.DeclList, p.importDecl())
	}

	for p.token != tokens.EOF {
		if p.token == api.IMPORT {
			p.error(p.FilePos, "import 语法只能出现在文件头部！")
		}

		switch p.token {
		case api.CONST:
			p.next()
			f.DeclList = append(f.DeclList, p.constDecl())
		case api.LET:
			p.next()
			f.DeclList = append(f.DeclList, p.letDecl())
		case api.TYPE:
			p.next()
			f.DeclList = append(f.DeclList, p.typeDecl())
		case api.FUNC:
			p.next()
			f.DeclList = append(f.DeclList, p.funcDecl())
		default:
			p.error(p.FilePos, "顶层语法仅支持 const, let, type, func 关键字定义！")
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
func (p *parser) paramList() []*prog.ParamDecl {
	var list []*prog.ParamDecl
	list = append(list, p.paramDecl())
	for p.token == tokens.COMMA {
		p.next()
		list = append(list, p.paramDecl())
	}
	return list
}

// ParameterDecl = IdentifierList Type .
func (p *parser) paramDecl() *prog.ParamDecl {
	d := new(prog.ParamDecl)
	d.SetPos(p.Pos())

	d.NameList = p.nameList(p.name())
	d.Type = p.typeOrNil()

	return d
}

// blockStmt = "{" StatementList "}" .
func (p *parser) blockStmt() *prog.BlockStmt {
	s := new(prog.BlockStmt)
	s.SetPos(p.Pos())

	for p.token != tokens.RBRACE && p.token != tokens.EOF {
		s.List = append(s.List, p.stmt())
	}

	p.expect(tokens.RBRACE)
	return s
}
