package internal

import (
	"fmt"
	"github.com/facelang/face/compiler/compile"
	"github.com/facelang/face/internal/ast"
	"os"
)

type ParserFactory interface {
	Parse() (interface{}, error)
	NextToken() (compile.Token, string, int, int)
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

func Parser(file string) ParserFactory {
	lex := &compile.lexer{buffer: &buffer{}}
	err := lex.init(file, func(file string, line, col, off int, msg string) {
		return
	})
	if err != nil {
		return &parser{err: err}
	}
	out, err := OpenFile(file + ".out")
	gen := &compile.codegen{out: out}

	p := &parser{
		file:   file,
		err:    err,
		gen:    gen,
		lexer:  lex,
		progFn: nil,
		progTable: &compile.ProgTable{
			fnRecList:   make(map[string]*compile.ProgFunc),
			varRecList:  make(map[string]*compile.ProgDec),
			stringTable: make([]*string, 0),
			realArgList: make([]*compile.ProgDec, 0),
		},
	}

	gen.parser = p

	return p
}

type parser struct {
	file      string           // 解析文件
	err       error            // 解析异常
	gen       *compile.codegen // 代码生成器
	lexer     *compile.lexer   // 读取器
	progFn    *compile.ProgFunc
	progTable *compile.ProgTable
}

func (p *parser) nextToken() compile.Token {
	token := p.lexer.NextToken()
	for token == compile.COMMENT {
		token = p.lexer.NextToken()
	}
	return token
}

func (p *parser) NextToken() (compile.Token, string, int, int) {
	token := p.nextToken()
	c := p.lexer.content
	return token, c, p.lexer.line, p.lexer.col
}

func (p *parser) Parse() (interface{}, error) {
	if p.err != nil {
		return nil, p.err
	}
	//for {
	//	token := p.lexer.NextToken()
	//	if token == EOF {
	//		return nil, nil
	//	}
	//
	//	fmt.Printf("Token: %v, %s", token, p.lexer.content)
	//	if rand.Intn(10) > 7 {
	//		p.lexer.Back(token)
	//		fmt.Printf(" BACK\n")
	//	} else {
	//		fmt.Printf(" \n")
	//	}
	//}
	return p.program()
}

// Type -> char | int | void
func (p *parser) kind(token compile.Token) string {
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

func (p *parser) require(next compile.Token) string {
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
