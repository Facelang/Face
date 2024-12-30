package factory

type parser struct {
	lex   *lexer // 读取器
	table *ProgTable
}

func (p *parser) init(file string) error {
	return p.lex.init(file, func(file string, line, col, off int, msg string) {
		return
	})
}

func (p *parser) parse() {
	p.program()
}

// .：语法结束符，表示规则的终结（EOF）。
// {} 表示 零个或多个 顶层声明，因此顶层声明也是可选的。
// program = decl { program } .
func (p *parser) program() {
	//for {
	//	token := p.lex.scan()
	//	if token == EOF {
	//		println("文件解析结束！")
	//		return
	//	}
	//	fmt.Printf("[%d,%d,%d] %s %s \n",
	//		p.lex.line, p.lex.line, p.lex.offset,
	//		token.String(), p.lex.content,
	//	)
	//}
	token := p.lex.NextToken()
	if token == EOF {
		return
	}
	p.decl(token)
	p.program()
}

// decl = <type>ident<dectail>|semicon|rsv_extern<type>ident semicon
func (p *parser) decl(token Token) {
	if token == SEMICOLON {
		return
	} else if token == IDENT && p.lex.content == "extern" {
		kind := p.kind()
		token = p.lex.NextToken()
		if token != IDENT {
			panic("符号解析不正确！")
		}
		name := p.lex.content
		v := VarRecord{
			kind:   kind,
			name:   name,
			extern: 1,
		}
		if v.kind == "string" {
			v.strValId = -2
		}
		p.table.addVar(&v)
		// 读取分号
	} else {
		kind := p.kind()
		token = p.lex.NextToken()
		if token != IDENT {
			panic("符号解析不正确！")
		}
		name := p.lex.content

	}

}

// Type -> char | int | void
func (p *parser) kind() string {
	token := p.lex.NextToken()
	if token != IDENT {
		return ""
	}
	switch p.lex.content {
	case "char":
		return "char"
	case "int":
		return "int"
	case "string":
		return "string"
	case "void":
		return "void"
	default:
		return ""
	}
}

// dectail -> semicon|<varlist>semicon|lparen<para>rparen<block>
func (p *parser) dectail(kind, name string) {
	switch token := p.lex.NextToken(); token {
	case SEMICOLON:
		v := VarRecord{
			kind: kind,
			name: name,
		}
		if v.kind == "string" {
			v.strValId = -2
		}
		p.table.addVar(&v)
	case LPAREN:
		fn := FnRecord{
			kind: kind,
			name: name,
		}
		p.para(&fn)
		p.funtail(&fn)
	default:
		v := VarRecord{
			kind: kind,
			name: name,
		}
		if v.kind == "string" {
			v.strValId = -2
		}
		p.table.addVar(&v)
		p.varlist(token, &v) //可能是^,会向下取符号，不需要重复取
	}
}

// funtail	-> <block>|semicon TODO
// 到这里已经有参数了 void fn()...
//
//	如果是分号，说明是函数定义
//	如果是大括号，说明函数有实现
//	其他情况
func (p *parser) funtail(fn *FnRecord) {
	level := 0 // 复合语句的层次
	token := p.lex.NextToken()
	if token == SEMICOLON {
		p.table.addFn(fn)
		return
	} else if token == LBRACE { // 函数定义
		fn.defined = 1 // 标记函数定义属性
		p.table.addFn(fn)
		// 解析代码块
		//wait=1;
		p.block(0, &level,0,0);
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

// <para>		->	<type>ident<paralist>|^
// 解析参数
func (p *parser) para(fn *FnRecord) {
	token := p.lex.NextToken()
	if token == RPAREN {
		return
	}

	kind := p.kind()
	token = p.lex.NextToken()
	if token != IDENT {
		panic("类型不正确！")
	}
	name := p.lex.content
	if p.table.exist(name) {
		panic("参数名称重复！")
	}
	v := VarRecord{
		kind: kind,
		name: name,
	}
	fn.addArg(&v)
	p.table.addVar(&v)
	p.paralist(fn)
}

// <paralist> -> comma<type>ident<paralist>|^
func (p *parser) paralist(fn *FnRecord) {
	token := p.lex.NextToken()
	if token == COMMA {
		kind := p.kind()
		token = p.lex.NextToken()
		if token != IDENT {
			panic("类型不正确！")
		}
		name := p.lex.content
		if p.table.exist(name) {
			panic("参数名称重复！")
		}
		v := VarRecord{
			kind: kind,
			name: name,
		}
		fn.addArg(&v)
		p.table.addVar(&v)
		p.paralist(fn)
	} else if token == RPAREN { // ) 结束， 正常关闭
		return
	} else {
		// { ;
		// 定义或者声明时候缺少)
		panic("语法不正确！")
	}

}

// <varlist>	->	comma ident<varlist>|^
func (p *parser) varlist(token Token, v *VarRecord) {
	switch token {
	case COMMA:
		token = p.lex.NextToken()
		if token != IDENT {
			panic("类型不正确！")
		}
		name := p.lex.content
		nv := &VarRecord{
			kind: v.kind,
			name: name,
		}
		if v.kind == "string" {
			nv.strValId = -2
		}
		p.table.addVar(nv)
		token = p.lex.NextToken()
		p.varlist(token, v)
		return
	case SEMICOLON:
		return
	default:
		panic("语法不正确")
	}

}



//<block>		->	lbrac<childprogram>rbrac
func (p *parser) block(initvar_num int,level *int,lopId int, blockAddr int) {
	token := p.lex.NextToken() //


}
void block(int initvar_num,int& level,int lopId,int blockAddr)
{
nextToken();
if(!match(lbrac))//丢失{
{
synterror(lbraclost,-1);
BACK
}
int var_num=initvar_num;//复合语句里变量的个数,一般是0，但是在if和while语句就不一定了
level++;//每次进入时加1

childprogram(var_num,level,lopId,blockAddr);

level--;
//match(rbrac);
//要清除局部变量名字表
tfun.poplocalvars(var_num);
}