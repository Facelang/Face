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
func (p *parser) dectail(kind, name string) string {
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
		fn := FuncRecord{
			kind: kind,
			name: name,
		}
		para()
		//match(rpren);
		funtail(dec_type, dec_name)
	default:
		v := VarRecord{
			kind: kind,
			name: name,
		}
		if v.kind == "string" {
			v.strValId = -2
		}
		p.table.addVar(&v)
		varlist(dec_type) //可能是^,会向下取符号，不需要重复取
	}
}

// funtail	-> <block>|semicon
func (p *parser) funtail(kind, name string) {
	level := 0 // 复合语句的层次
	token := p.lex.NextToken()
	if token == SEMICOLON {
		p.table.addFn()
		return
	} else if token == LBRACE {
		tfun.defined=1;//标记函数定义属性
		table.addfun();
	}
else if(token==lbrac)//函数定义
{
p("函数定义",2);
tfun.defined=1;//标记函数定义属性
table.addfun();
BACK
block(0,level,0,0);
level=0;//恢复
tfun.poplocalvars(-1);//清除参数
genFuntail();
return;
}
else if(token==ident||token==rsv_if||token==rsv_while||token==rsv_return||token==rsv_break||token==rsv_continue
||token==rsv_in||token==rsv_out||token==rbrac)//必然是函数定义
{
BACK
block(0,level,0,0);
level=0;//恢复
return ;
}
else if(token==rsv_void||token==rsv_int||token==rsv_char||token==rsv_string)//其他的作为声明处理
{
synterror(semiconlost,-1);
BACK
return ;
}
else
{
synterror(semiconwrong,0);
return ;
}
}
