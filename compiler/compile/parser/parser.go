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
	inRhs   bool        // 是否右值表达式
	nestLev int         // 递归嵌套计数器
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
	d := &ast.GenDecl{Pos: p.expect(require), Token: require}

	d.Names = p.nameList(p.name())
	if p.token != token.EOF && p.token != token.SEMICOLON && p.token != token.RPAREN {
		d.Type = p.tryIdentOrType()
		if p.token == token.ASSIGN {
			p.next()
			d.Values = exprList(p, true)
		}
	}

	return d
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

// TypeDecl = "type" ( TypeSpec | "(" { TypeSpec ";" } ")" ) .
func (p *parser) typeDecl() ast.Decl {
	d := &ast.TypeDecl{Pos: p.expect(token.TYPE), Name: p.name()}

	if p.token == token.ASSIGN {
		d.Assign = p.expect(p.token)
	}

	d.Type = p.parseType()

	return d
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
		case token.TYPE:
			p.next()
			f.DeclList = append(f.DeclList, p.typeDecl())
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

// gotAssign = "=" .
func (p *parser) gotAssign() bool {
	if p.token == token.ASSIGN {
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
	for p.token == token.COMMA {
		p.next()
		list = append(list, internal.expr(p, &vn))
	}
	return list
}

// paramList = ParameterDecl { "," ParameterDecl } .
func (p *parser) paramList() []*ast.ParamDecl {
	var list []*ast.ParamDecl
	list = append(list, p.paramDecl())
	for p.token == token.COMMA {
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

	for p.token != token.RBRACE && p.token != token.EOF {
		s.List = append(s.List, p.stmt())
	}

	p.expect(token.RBRACE)
	return s
}
