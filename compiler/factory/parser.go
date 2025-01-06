package factory

import (
	"fmt"
	"os"
)

type ParserFactory interface {
	Parse() (interface{}, error)
}

func Parser(file string) ParserFactory {
	lex := &lexer{buffer: &buffer{}}
	err := lex.init(file, func(file string, line, col, off int, msg string) {
		return
	})
	if err != nil {
		return &parser{err: err}
	}
	out, err := os.Open(file + ".out")
	gen := &codegen{out: out}

	return &parser{
		file:      file,
		err:       err,
		gen:       gen,
		lexer:     lex,
		progFn:    nil,
		progTable: &ProgTable{},
	}
}

type parser struct {
	file      string   // 解析文件
	err       error    // 解析异常
	gen       *codegen // 代码生成器
	lexer     *lexer   // 读取器
	progFn    *ProgFunc
	progTable *ProgTable
}

func (p *parser) Parse() (interface{}, error) {
	if p.err != nil {
		return nil, p.err
	}
	return p.program()
}

// Type -> char | int | void
func (p *parser) kind() string {
	switch token := p.lexer.NextToken(); token {
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

//
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
		v := ProgDec{
			kind:   p.kind(),
			name:   p.ident(),
			extern: 1,
		}
		if v.kind == "string" {
			v.strValId = -2 //全局的string
		}
		p.progTable.addVar(&v)
		p.require(SEMICOLON)
	} else {
		p.dectail(p.kind(), p.ident())
	}

}

// dectail -> semicon|<varlist>semicon|lparen<para>rparen<block>
func (p *parser) dectail(kind, name string) {
	switch token := p.lexer.NextToken(); token {
	case SEMICOLON:
		v := ProgDec{kind: kind, name: name}
		if v.kind == "string" {
			v.strValId = -2
		}
		p.progTable.addVar(&v)
	case LPAREN:
		p.progFn = &ProgFunc{
			kind: kind,
			name: name,
		}
		p.params()
		p.funtail()
	default:
		v := ProgDec{kind: kind, name: name}
		if v.kind == "string" {
			v.strValId = -2
		}
		p.progTable.addVar(&v)
		p.varlist(token, &v) //可能是^,会向下取符号，不需要重复取
	}
}

// <params>		->	<type>ident<paralist>|^ 函数申明的参数
func (p *parser) params() {
	token := p.lexer.NextToken()
	if token == RPAREN {
		return
	}
	for {
		p.para()

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
	kind := p.kind()
	id := p.ident()
	if p.progTable.exist(id) {
		panic("参数名称重复！")
	}
	v := ProgDec{
		kind: kind,
		name: id,
	}

	p.progFn.addArg(&v)
	p.progTable.addVar(&v)
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
		p.progTable.addFn(p.gen, p.progFn)
		return
	} else if token == LBRACE { // {函数定义}
		p.progFn.defined = 1 // 标记函数定义属性
		p.progTable.addFn(p.gen, p.progFn)
		p.lexer.Back(token)

		p.block(0, &level, 0, 0)
		//level=0;//恢复
		//tfun.poplocalvars(-1);//清除参数
		//genFuntail();
		return
	} else if token == IDENT {
		//||token==rsv_if||
		//token==rsv_while||token==rsv_return||
		//token==rsv_break||token==rsv_continue ||
		//token==rsv_in||token==rsv_out||token==rbrac
		// 这里目前不清楚
		//BACK
		//block(0,level,0,0);
		//level=0;//恢复
		//return ;

		//else if token==rsv_void||token==rsv_int||token==rsv_char||token==rsv_string
		//synterror(semiconlost,-1);
		//BACK
		//return ;

		//else
		//synterror(semiconwrong,0);
		//return ;

	}
}

// <varlist>	->	comma ident<varlist>|^
func (p *parser) varlist(token Token, v *ProgDec) {
	switch token {
	case COMMA:
		token = p.lexer.NextToken()
		if token != IDENT {
			panic("类型不正确！")
		}
		name := p.lexer.content
		nv := &ProgDec{
			kind: v.kind,
			name: name,
		}
		if v.kind == "string" {
			nv.strValId = -2
		}
		p.progTable.addVar(nv)
		token = p.lexer.NextToken()
		p.varlist(token, v)
		return
	case SEMICOLON:
		return
	default:
		panic("语法不正确")
	}

}

// <block>		->	lbrac<childprogram>rbrac
func (p *parser) block(initVarNum int, level *int, lopId int, blockAddr int) {
	token := p.lexer.NextToken()
	// 判断大括号
	varNum := initVarNum
	*level += 1

	p.childprogram(&varNum, level, lopId, blockAddr)

	*level -= 1
	p.progFn.poplocalvars(varNum)
}

var rbracislost = 0 //}丢失异常，维护恢复,紧急恢复
func (p *parser) childprogram(fn *ProgFunc, vn *int, level *int, lopId int, blockAddr int) {
	token := p.lexer.NextToken()

	if token == SEMICOLON || p.keywords(token,
		[]string{"while", "if", "return", "break", "continue", "in", "out"}) {
		//statement(vn, level, lopId, blockAddr)
		//childprogram(vn, level, lopId, blockAddr)
	} else if p.keywords(token, []string{"void", "int", "char", "string"}) {
		p.localdec(fn, vn, level)
		if rbracislost == 1 {
			rbracislost = 0
		} else {
			p.childprogram(fn, vn, level, lopId, blockAddr)
		}
	} else if token == IDENT {
		//statement(vn, level, lopId, blockAddr)
		//childprogram(vn, level, lopId, blockAddr)
	} else if token == RBRACE {
		return
	} else {
		panic("语法错误！！！")
	}
}

func (p *parser) keywords(token Token, input []string) bool {
	if token != IDENT {
		return false
	}
	for _, k := range input {
		if k == p.lexer.content {
			return true
		}
	}
	return false
}

// <localdec>	->	<type>ident<lcoaldectail>semicon
func (p *parser) localdec(fn *ProgFunc, vn *int, level *int) {
	kind := p.kind()
	token := p.lexer.NextToken()
	if token != IDENT {
		panic("类型不正确！")
	}
	name := p.lexer.content
	if p.progTable.exist(name) {
		panic("参数名称重复！")
	}
	v := ProgDec{
		kind: kind,
		name: name,
	}
	fn.addLocalVar(&v)
	*vn += 1
	p.localdectail(fn, kind, vn, level)
}

// <localvartail>	->	comma ident<localvartail>|^
func (p *parser) localdectail(fn *ProgFunc, kind string, vn *int, level *int) {
	token := p.lexer.NextToken()
	if token == COMMA {
		token = p.lexer.NextToken()
		// 直接取变量名
		name := p.lexer.content
		if p.progTable.exist(name) {
			panic("参数名称重复！")
		}
		v := ProgDec{
			kind: kind,
			name: name,
		}
		fn.addLocalVar(&v)
		*vn += 1
		p.localdectail(fn, kind, vn, level)
	} else if token == SEMICOLON {
		return
	} else {
		panic("语法错误！！！")
	}
}

// <statement>	->	ident<idtail>semicon|<whilestat>|<ifstat>|<retstat>|semicon|rsv_break semicon|rsv_continue semicon
func (p *parser) statement(fn *ProgFunc, token Token, vn *int, level *int, lopId int, blockAddr int) {
	refname := ""
	switch token {
	case SEMICOLON:
		return
	case IDENT:
		switch p.lexer.content {
		case "while":
			whilestat(var_num, level)
			return
		case "if":
			ifstat(var_num, level, lopId, blockAddr)
			return
		case "break": // 这个还比较复杂
		// TODO
		case "return":
			retstat(var_num, level)
			return
		case "in": // todo
		case "out": //
		default: // 非关键字，申明 或 定义语句
			refname += p.lexer.content // 变量名
			p.idtail(refname, vn)
			nextToken()
			if (!match(semicon)) //赋值语句或者函数调用语句丢失分号，记得回退
			{
				synterror(semiconlost, -1)
				BACK
			}
			break
		}
	}
}

//<idtail>	->	assign<expr>|lparen<realarg>rparen
// 赋值语句 或 函数调用
func (p *parser) idtail(refname string, vn *int) *ProgDec {
	token := p.lexer.NextToken()
	if token == ASSIGN { // 赋值语句
		src := expr(p, vn)
		des := p.progTable.getVar(refname)
		return genAss(des, src, vn)
	} else if token == LPAREN { // 函数调用
		p.realarg(refname, vn) // 调用参数写入符号表
		var_record * var_ret = p.progTable.genCall(refname, var_num)
		nextToken()
		if !match(rparen) {
			synterror(rparenlost, -1)
			BACK
		}
		return var_ret
	} else if identinexpr == 1 { //表达式中可以单独出现标识符，基于此消除表达式中非a=b类型的错误
		identinexpr = 0
		BACK
		return p.progTable.getVar(refname)
	} else {
		panic("语法错误!!")
	}
}

//<realarg>	->	<expr><arglist>|^
// 参数为表达式列表
func (p *parser) realarg(refname string, vn *int) {
	token := p.lexer.NextToken()
	if token == RPAREN || token == SEMICOLON { // ), ; 终结符
		p.lexer.Back(token)
		return
	} else if token == IDENT || token == INT || token == CHAR || token == STRING || token == LPAREN {
		// 可取值类型， 变量， 数字，字符，字符串， （） 使用括号括起来的参数
		p.lexer.Back(token)
		p.progTable.addRealArg(expr(p, vn), vn)
		p.arglist(vn)
	} else {
		panic("语法错误！")
	}
}

//<arglist>	->	comma<expr><arglist>|^
func (p *parser) arglist(vn *int) {
	token := p.lexer.NextToken()
	if token == COMMA {
		token = p.lexer.NextToken()
		if token == IDENT || token == INT || token == CHAR || token == STRING || token == LPAREN {
			p.lexer.Back(token)
			p.progTable.addRealArg(expr(p, vn), vn)
			p.arglist(vn)
		} else {
			panic("语法错误！")
		}
	} else if token == RPAREN || token == SEMICOLON {
		p.lexer.Back(token)
		return
	} else {
		panic("语法错误！")
	}
}
