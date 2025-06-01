package internal

import (
	"fmt"
	"github.com/facelang/face/compiler/compile"
	"github.com/facelang/face/internal/ast"
	"github.com/facelang/face/internal/tokens"
	"go/build/constraint"
	"go/token"
	"os"
	"strings"
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

func (p *parser) error(pos tokens.FilePos, msg string) {
	if p.errors.Len() > 10 {
		panic(p.errors)
	}
	p.errors.Add(pos, msg)
}

func (p *parser) errorExpected(pos tokens.FilePos, msg string) {
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

func (p *parser) expect(token tokens.Token) tokens.FilePos {
	pos := p.FilePos
	if p.token != token {
		p.errorExpected(pos, "'"+token.String()+"'")
	}
	p.next() // make progress
	return pos
}

// expect2 is like expect, but it returns an invalid position
// if the expected token is not found.
func (p *parser) expect2(token tokens.Token) (pos tokens.FilePos) {
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
func (p *parser) expectClosing(token tokens.Token, context string) tokens.FilePos {
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
	if p.token != RPAREN && p.token != RBRACE {
		switch p.token {
		case COMMA:
			// permit a ',' instead of a ';' but complain
			p.errorExpected(p.FilePos, "';'")
			fallthrough
		case SEMICOLON:
			if p.lit == ";" {
				// explicit semicolon
				p.next()
				comment = p.lineComment // use following comments
			} else {
				// artificial semicolon
				comment = p.lineComment // use preceding comments
				p.next()
			}
			return comment
		default:
			p.errorExpected(p.FilePos, "';'")
			p.advance(stmtStart)
		}
	}
	return nil
}

type parseSpecFunction func(doc *ast.CommentGroup, keyword token.Token, iota int) ast.Spec

func (p *parser) parseImportSpec(doc *ast.CommentGroup, _ token.Token, _ int) ast.Spec {
	var ident *ast.Ident
	switch p.token {
	case tokens.IDENT:
		ident = p.parseIdent()
	case PERIOD:
		ident = &ast.Ident{NamePos: p.pos, Name: "."}
		p.next()
	}

	pos := p.pos
	var path string
	if p.token == tokens.STRING {
		path = p.lit
		p.next()
	} else if p.token.IsLiteral() {
		p.error(pos, "import path must be a string")
		p.next()
	} else {
		p.error(pos, "missing import path")
		p.advance(exprEnd)
	}
	comment := p.expectSemi()

	// collect imports
	spec := &ast.ImportSpec{
		Doc:     doc,
		Name:    ident,
		Path:    &ast.BasicLit{ValuePos: pos, Kind: token.STRING, Value: path},
		Comment: comment,
	}
	p.imports = append(p.imports, spec)

	return spec
}

func (p *parser) parseGenDecl(keyword tokens.Token, f parseSpecFunction) *ast.GenDecl {
	pos := p.expect(keyword)
	var lparen, rparen token.Pos
	var list []ast.Spec
	if p.tok == token.LPAREN {
		lparen = p.pos
		p.next()
		for iota := 0; p.tok != token.RPAREN && p.tok != token.EOF; iota++ {
			list = append(list, f(p.leadComment, keyword, iota))
		}
		rparen = p.expect(token.RPAREN)
		p.expectSemi()
	} else {
		list = append(list, f(nil, keyword, 0))
	}

	return &ast.GenDecl{
		Doc:    doc,
		TokPos: pos,
		Tok:    keyword,
		Lparen: lparen,
		Specs:  list,
		Rparen: rparen,
	}
}

func ParseFile(lexer *lexer) (interface{}, error) {
	var decls []ast.Decl

	// import decls
	for p.tok == token.IMPORT {
		decls = append(decls, p.parseGenDecl(token.IMPORT, p.parseImportSpec))
	}

	if p.mode&ImportsOnly == 0 {
		// rest of package body
		prev := token.IMPORT
		for p.tok != token.EOF {
			// Continue to accept import declarations for error tolerance, but complain.
			if p.tok == token.IMPORT && prev != token.IMPORT {
				p.error(p.pos, "imports must appear before other declarations")
			}
			prev = p.tok

			decls = append(decls, p.parseDecl(declStart))
		}
	}

	f := &ast.File{
		Doc:     doc,
		Package: pos,
		Name:    ident,
		Decls:   decls,
		// File{Start,End} are set by the defer in the caller.
		Imports:   p.imports,
		Comments:  p.comments,
		GoVersion: p.goVersion,
	}
	var declErr func(token.Pos, string)
	if p.mode&DeclarationErrors != 0 {
		declErr = p.error
	}
	if p.mode&SkipObjectResolution == 0 {
		resolveFile(f, p.file, declErr)
	}
}

// Type -> char | int | void
func (p *parser) kind(token tokens.Token) string {
	switch token {
	case compile.CHAR:
		return "char"
	case compile.INT:
		return "int"
	case compile.STRING:
		return "string"
	case compile.VOID:
		return "void"
	default:
		p.err = fmt.Errorf("不支持的数据类型[%d,%d]: %s, %s",
			p.lexer.line, p.lexer.col, token.String(), p.lexer.content)
		panic(p.err)
	}
}

func (p *parser) ident() string {
	token := p.lexer.NextToken()
	if token != compile.IDENT {
		p.err = fmt.Errorf("[%d,%d]: %s, %s，需要一个ID类型！",
			p.lexer.line, p.lexer.col, token.String(), p.lexer.content)
		panic(p.err)
	}
	return p.lexer.content
}

func (p *parser) require(next tokens.Token) string {
	token := p.lexer.NextToken()
	if token != next {
		p.err = fmt.Errorf("[%d,%d]: %s, %s，需要一个 %s 类型！",
			p.lexer.line, p.lexer.col, token.String(), p.lexer.content, next)
		panic(p.err)
	}
	return p.lexer.content
}

// .：语法结束符，表示规则的终结（EOF）。
// {} 表示 零个或多个 顶层声明，因此顶层声明也是可选的。
// program = decl { program } .
func (p *parser) program() (interface{}, error) {
	if p.err != nil {
		return nil, p.err
	}
	//for {
	//	token := p.lexer.scan()
	//	if token == EOF {
	//		println("文件解析结束！")
	//		return
	//	}
	//	fmt.Printf("[%d,%d,%d] %s %s \n",
	//		p.lexer.line, p.lexer.line, p.lexer.offset,
	//		token.String(), p.lexer.content,
	//	)
	//}
	token := p.lexer.NextToken()
	if token == compile.EOF {
		return nil, nil
	}
	p.dec(token)
	return p.program()
}

// SourceFile = { ImportDecl ";" } { TopLevelDecl ";" } .
func (p *parser) parseFile() *ast.File {
	var decls []ast.Decl

	// import decls
	for p.token == IMPORT {
		decls = append(decls, p.parseGenDecl(token.IMPORT, p.parseImportSpec))
	}

	if p.mode&ImportsOnly == 0 {
		// rest of package body
		prev := token.IMPORT
		for p.tok != token.EOF {
			// Continue to accept import declarations for error tolerance, but complain.
			if p.tok == token.IMPORT && prev != token.IMPORT {
				p.error(p.pos, "imports must appear before other declarations")
			}
			prev = p.tok

			decls = append(decls, p.parseDecl(declStart))
		}
	}

	f := &ast.File{
		Doc:     doc,
		Package: pos,
		Name:    ident,
		Decls:   decls,
		// File{Start,End} are set by the defer in the caller.
		Imports:   p.imports,
		Comments:  p.comments,
		GoVersion: p.goVersion,
	}
	var declErr func(token.Pos, string)
	if p.mode&DeclarationErrors != 0 {
		declErr = p.error
	}
	if p.mode&SkipObjectResolution == 0 {
		resolveFile(f, p.file, declErr)
	}

	return f
}
func (p *parser) fileOrNil() *ast.File {

	f := new(ast.File)
	f.Version = "" // todo

	// Accept import declarations anywhere for error tolerance, but complain.
	// { ( ImportDecl | TopLevelDecl ) ";" }

	prev := _Import
	for p.tok != _EOF {
		if p.tok == _Import && prev != _Import {
			p.syntaxError("imports must appear before other declarations")
		}
		prev = p.tok

		switch p.tok {
		case _Import:
			p.next()
			f.DeclList = p.appendGroup(f.DeclList, p.importDecl)

		case _Const:
			p.next()
			f.DeclList = p.appendGroup(f.DeclList, p.constDecl)

		case _Type:
			p.next()
			f.DeclList = p.appendGroup(f.DeclList, p.typeDecl)

		case _Var:
			p.next()
			f.DeclList = p.appendGroup(f.DeclList, p.varDecl)

		case _Func:
			p.next()
			if d := p.funcDeclOrNil(); d != nil {
				f.DeclList = append(f.DeclList, d)
			}

		default:
			if p.tok == _Lbrace && len(f.DeclList) > 0 && isEmptyFuncDecl(f.DeclList[len(f.DeclList)-1]) {
				// opening { of function declaration on next line
				p.syntaxError("unexpected semicolon or newline before {")
			} else {
				p.syntaxError("non-declaration statement outside function body")
			}
			p.advance(_Import, _Const, _Type, _Var, _Func)
			continue
		}

		// Reset p.pragma BEFORE advancing to the next token (consuming ';')
		// since comments before may set pragmas for the next function decl.
		p.clearPragma()

		if p.tok != _EOF && !p.got(_Semi) {
			p.syntaxError("after top level declaration")
			p.advance(_Import, _Const, _Type, _Var, _Func)
		}
	}
	// p.tok == _EOF

	p.clearPragma()
	f.EOF = p.pos()

	return f
}
