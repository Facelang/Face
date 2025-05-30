package compile

import (
	"fmt"
	"os"
)

type ParserFactory interface {
	Parse() (interface{}, error)
	NextToken() (Token, string, int, int)
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
	lex := &lexer{buffer: &buffer{}}
	err := lex.init(file, func(file string, line, col, off int, msg string) {
		return
	})
	if err != nil {
		return &parser{err: err}
	}
	out, err := OpenFile(file + ".out")
	gen := &codegen{out: out}

	p := &parser{
		file:   file,
		err:    err,
		gen:    gen,
		lexer:  lex,
		progFn: nil,
		progTable: &ProgTable{
			fnRecList:   make(map[string]*ProgFunc),
			varRecList:  make(map[string]*ProgDec),
			stringTable: make([]*string, 0),
			realArgList: make([]*ProgDec, 0),
		},
	}

	gen.parser = p

	return p
}

type parser struct {
	file      string   // 解析文件
	err       error    // 解析异常
	gen       *codegen // 代码生成器
	lexer     *lexer   // 读取器
	progFn    *ProgFunc
	progTable *ProgTable
}

func (p *parser) nextToken() Token {
	token := p.lexer.NextToken()
	for token == COMMENT {
		token = p.lexer.NextToken()
	}
	return token
}

func (p *parser) NextToken() (Token, string, int, int) {
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
func (p *parser) kind(token Token) string {
	switch token {
	case CHAR:
		return "char"
	case INT:
		return "int"
	case STRING:
		return "string"
	case VOID:
		return "void"
	default:
		p.err = fmt.Errorf("不支持的数据类型[%d,%d]: %s, %s",
			p.lexer.line, p.lexer.col, token.String(), p.lexer.content)
		panic(p.err)
	}
}

func (p *parser) ident() string {
	token := p.lexer.NextToken()
	if token != IDENT {
		p.err = fmt.Errorf("[%d,%d]: %s, %s，需要一个ID类型！",
			p.lexer.line, p.lexer.col, token.String(), p.lexer.content)
		panic(p.err)
	}
	return p.lexer.content
}

func (p *parser) require(next Token) string {
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
	if token == EOF {
		return nil, nil
	}
	p.dec(token)
	return p.program()
}

// dec = <type>ident<dectail>|semicon|rsv_extern<type>ident semicon
func (p *parser) dec(token Token) {
	if token == SEMICOLON {
		return
	} else if token == EXTERN {
		token = p.lexer.NextToken()
		v := ProgDec{
			kind:   p.kind(token),
			name:   p.ident(),
			extern: 1,
		}
		if v.kind == "string" {
			v.strValId = -2 //全局的string
		}
		p.progTable.addVar(&v)
		p.require(SEMICOLON)
	} else {
		p.dectail(p.kind(token), p.ident())
	}

}

// dectail -> semicon|<varlist>semicon|lparen<para>rparen<block>
func (p *parser) dectail(kind, name string) {
	token := p.lexer.NextToken()
	if token == LPAREN { // type name(...); 函数声明
		p.progFn = &ProgFunc{
			kind:      kind,
			name:      name,
			args:      make([]string, 0),
			localVars: make([]*ProgDec, 0),
		}
		p.params()
		p.funtail()
	} else { // 其它情况，结束，或列表定义
		v := ProgDec{kind: kind, name: name}
		if v.kind == "string" {
			v.strValId = -2
		}
		p.progTable.addVar(&v)

		//p.varlist(token, &v) //可能是^,会向下取符号，不需要重复取
		// <varlist>	->	comma ident<varlist>|^

		for {
			if token == COMMA { // type name1, name2, ...
				nv := &ProgDec{
					kind: v.kind,
					name: p.ident(),
				}
				if nv.kind == "string" {
					nv.strValId = -2
				}
				p.progTable.addVar(nv)
				token = p.lexer.NextToken()
			} else if token == SEMICOLON { // ; 结束
				return
			} else {
				panic("语法不正确 dec")
			}
		}
	}
}

// <params>		->	<type>ident<paralist>|^ 函数申明的参数
func (p *parser) params() {
	token := p.lexer.NextToken()
	if token == RPAREN {
		return
	}
	for {
		p.para() // type name 写入参数表

		// <paralist> -> comma<type>ident<paralist>|^
		token = p.lexer.NextToken()
		if token == COMMA {
			p.para()
		} else if token == RPAREN { // ) 结束， 正常关闭
			return
		} else {
			panic("定义函数时，参数解析异常！")
		}
	}
}

func (p *parser) para() {
	token := p.lexer.NextToken()
	kind := p.kind(token)
	id := p.ident()
	if p.progTable.exist(id) { // todo
		panic("参数名称重复！")
	}
	v := ProgDec{
		kind: kind,
		name: id,
	}
	p.progFn.addArg(p.gen, &v)
}

// funtail	-> <block>|semicon TODO
// 到这里已经有参数了 void fn()...
//
//	如果是分号，说明是函数定义
//	如果是大括号，说明函数有实现
//	其他情况
func (p *parser) funtail() {
	level := 0 // 复合语句的层次
	token := p.lexer.NextToken()
	if token == SEMICOLON {
		// defined = 0
		p.progTable.addFn(p.gen, p.progFn) // 添加函数声明记录
		return
	} else if token == LBRACE { // {函数定义}
		p.progFn.defined = 1 // 标记函数定义属性
		p.progTable.addFn(p.gen, p.progFn)
		p.lexer.Back(token)

		p.block(0, &level, 0, 0)
		level = 0                              // 恢复
		p.progFn.poplocalvars(p.progTable, -1) // 清除参数， block 中也会调用同，什么意思
		p.gen.funtail(p.progFn)
		return
	} else if token == IDENT || token == IF || token == WHILE || token == RETURN ||
		token == CONTINUE || token == IN || token == OUT || token == RBRACE {
		// 关键字语句 todo
		p.lexer.Back(token)
		p.block(0, &level, 0, 0)
		level = 0 // 恢复
		return
	} else {
		panic("; 结束符缺失！1")
	}
}

// <block>		->	lbrac<childprogram>rbrac
func (p *parser) block(initVarNum int, level *int, lopId int, blockAddr int) {
	p.require(LBRACE)
	// 判断大括号
	varNum := initVarNum
	*level += 1

	p.childprogram(&varNum, level, lopId, blockAddr)

	*level -= 1
	p.progFn.poplocalvars(p.progTable, varNum) // 要清除局部变量名字表
}

var rbracislost = 0 //}丢失异常，维护恢复,紧急恢复
func (p *parser) childprogram(vn *int, level *int, lopId int, blockAddr int) {
	switch token := p.lexer.NextToken(); token {
	case SEMICOLON, WHILE, IF, RETURN, BREAK, CONTINUE, IN, OUT, IDENT: // 关键字语法 变量||函数， & 结束符
		p.statement(token, vn, level, lopId, blockAddr)
		p.childprogram(vn, level, lopId, blockAddr)
		return
	case VOID, INT, CHAR, STRING: // 数据类型， 局部变量定义
		p.localdec(token, vn, level)
		//if rbracislost == 1 { // 哪种情况会使用
		//	rbracislost = 0
		//} else { // 为什么调用子程序？
		//	p.childprogram(vn, level, lopId, blockAddr)
		//}
		p.childprogram(vn, level, lopId, blockAddr)
		return
	case RBRACE: // 子程序结束， 开始符号在外部就被使用
		return
	default:
		panic("语法解析错误！")
	}
}

// <localdec>	->	<type>ident<lcoaldectail>semicon
func (p *parser) localdec(token Token, vn *int, level *int) {
	v := ProgDec{
		kind: p.kind(token),
		name: p.ident(),
	}
	if p.progTable.exist(v.name) {
		panic("变量重复定义： localdec 1")
	}
	p.progFn.addLocalVar(p.gen, &v) //添加一个局部变量
	*vn += 1
	// p.localdectail(v.kind, vn, level)
	// <localvartail>	->	comma ident<localvartail>|^
	for {
		token = p.lexer.NextToken()
		if token == COMMA {
			v = ProgDec{
				kind: v.kind,
				name: p.ident(),
			}
			if p.progTable.exist(v.name) {
				panic("变量重复定义： localdec 2")
			}
			p.progFn.addLocalVar(p.gen, &v)
			*vn += 1
		} else if token == SEMICOLON {
			return
		} else {
			panic("; 结束符缺失！2")
		}
	}
}

// <statement>	->	ident<idtail>semicon|<whilestat>|<ifstat>|<retstat>|semicon|rsv_break semicon|rsv_continue semicon
func (p *parser) statement(token Token, vn *int, level *int, lopId int, blockAddr int) {
	switch token {
	case SEMICOLON:
		return
	//case WHILE:
	//	whilestat(var_num, level)
	//	return
	//case IF:
	//	ifstat(var_num, level, lopId, blockAddr)
	//	return
	//case BREAK, CONTINUE:
	//	p.require(SEMICOLON)
	//	if lopId != 0 {
	//		p.gen.block(p.progFn, blockAddr)
	//		if token == BREAK {
	//			p.gen.fprintf("\tjmp @while_%d_exit\n", lopId)
	//		} else {
	//			p.gen.fprintf("\tjmp @while_%d_lop\n", lopId)
	//		}
	//	} else {
	//		panic("break, continue 语句不能出现在while之外。\n")
	//	}
	case RETURN:
		p.retstat(vn, level)
		break
	case IN:
		p.require(SHR) // >>
		id := p.ident()
		v := p.progTable.getVar(id)
		p.gen.input(v, vn)
		p.require(SEMICOLON)
	case OUT:
		p.require(SHL) // <<
		p.gen.output(expr(p, vn), vn)
		p.require(SEMICOLON)
	case IDENT:
		p.idtail(p.lexer.content, vn)
		p.require(SEMICOLON)
	default:
		panic("unhandled default case")
	}
}

// <retstat>	->	rsv_return<expr>semicon
func (p *parser) retstat(vn *int, level *int) {
	//p.returntail(vn, level)
	// <returntail>	->	<expr>|^
	if *level == 1 { // todo 不确定什么作用
		p.progFn.hadret = 1
	}
	token := p.lexer.NextToken()
	if token == IDENT || token == V_INT || token == V_CHAR || token == V_STRING || token == LPAREN {
		p.lexer.Back(token)
		ret := expr(p, vn)
		if ret == nil || ret.kind != p.progFn.kind {
			panic("返回值类型不兼容")
		}
		p.gen.ret(ret, vn)
	} else if token == SEMICOLON {
		p.lexer.Back(token)
		if p.progFn.kind != "void" {
			panic("返回值类型不兼容 void")
		}
		p.gen.ret(nil, vn)
	} else if token == RBRACE {
		p.lexer.Back(token)
		return
	} else {
		panic("语法错误！ return")
	}

	p.require(SEMICOLON)
}

var identinexpr = 0 // 指示标识符是否单独出现在表达式中

// <idtail>	->	assign<expr>|lparen<realarg>rparen
// 赋值语句 或 函数调用
func (p *parser) idtail(refname string, vn *int) *ProgDec {
	token := p.lexer.NextToken()
	if token == ASSIGN { // 赋值语句
		src := expr(p, vn)
		des := p.progTable.getVar(refname)
		return p.gen.assign(des, src, vn)
	} else if token == LPAREN { // 函数调用
		p.realarg(refname, vn) // 调用参数写入符号表
		ret := p.gen.call(p.progTable, refname, vn)
		p.require(RPAREN)
		return ret
	} else if identinexpr == 1 { //表达式中可以单独出现标识符，基于此消除表达式中非a=b类型的错误
		identinexpr = 0
		p.lexer.Back(token)
		return p.progTable.getVar(refname)
	} else {
		panic("语法错误!!")
	}
}

// <realarg>	->	<expr><arglist>|^
// 参数为表达式列表
func (p *parser) realarg(refname string, vn *int) {
	token := p.lexer.NextToken()
	if token == RPAREN || token == SEMICOLON { // ), ; 终结符
		p.lexer.Back(token)
		return
	} else if token == IDENT || token == V_INT || token == V_CHAR || token == V_STRING || token == LPAREN {
		// 可取值类型， 变量， 数字，字符，字符串， （） 使用括号括起来的参数
		p.lexer.Back(token)
		p.progTable.addRealArg(p.gen, expr(p, vn), vn)
		p.arglist(vn)
	} else {
		panic("语法错误！")
	}
}

// <arglist>	->	comma<expr><arglist>|^
func (p *parser) arglist(vn *int) {
	token := p.lexer.NextToken()
	if token == COMMA {
		token = p.lexer.NextToken()
		// 值类型（变量， 数字， 字符， 字符串），结束符
		if token == IDENT || token == V_INT || token == V_CHAR || token == V_STRING || token == LPAREN {
			p.lexer.Back(token)
			p.progTable.addRealArg(p.gen, expr(p, vn), vn)
			p.arglist(vn)
		} else {
			panic("语法错误！")
		}
	} else if token == RPAREN || token == SEMICOLON { // ) ; 结束符
		p.lexer.Back(token)
		return
	} else {
		panic("语法错误！")
	}
}
