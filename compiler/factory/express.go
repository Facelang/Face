package factory

// <exp>		->	<aloexp><exptail>
// 目前只支持 赋值, 比较， 加减， 乘除
// 赋值运算在【语句】层面被解析了
func expr(p *parser, vn *int) *ProgDec {
	f1 := aloexp(p, vn)
	f2 := exptail(p, f1, vn)
	if f2 == nil {
		return f1
	}
	return f2
}

// <exptail>	->	<cmps><expr>|^
// 最低级运算符， 比较运算 【因为是最低级，所以多了结束符判定】
func exptail(p *parser, f1 *ProgDec, vn *int) *ProgDec {
	token := p.lexer.NextToken()
	if token == ADD || token == QUO { // 优先运算符
		f2 := aloexp(p, vn)
		return genExp(token, f1, f2, vn)
	} else { // 结束符判断，或抛出异常
		p.lexer.Back(token)
		return nil
	}
}

// <aloexp>	->	<item><itemtail>
// 从最低级运算符开始解析
func aloexp(p *parser, vn *int) *ProgDec {
	f1 := item(p, vn)
	f2 := itemtail(p, f1, vn)
	if f2 == nil {
		return f1
	}
	return f2
}

// <item>		->	<factor><factortail>
// 先解析一个符号，如果后一个符号优先级更高，先解析后一个符号并返回
func item(p *parser, vn *int) *ProgDec {
	f1 := factor(p, vn)
	f2 := factortail(p, f1, vn)
	if f2 == nil {
		return f1
	}
	return f2
}

// <itemtail>	->	<adds><aloexp>|^
// 低一级运算符 + -
func itemtail(p *parser, f1 *ProgDec, vn *int) *ProgDec {
	token := p.lexer.NextToken()
	if token == ADD || token == QUO { // 优先运算符
		f2 := aloexp(p, vn)
		return genExp(token, f1, f2, vn)
	} else { // 其它运算符
		p.lexer.Back(token)
		return nil
	}
}

// <factor> -> ident<idtail>|number|chara|lparen<expr>rparen|strings
func factor(p *parser, vn *int) *ProgDec {
	switch token := p.lexer.NextToken(); token {
	case CHAR: // 值类型，直接创建【临时变量】
	case INT:
	case STRING:
	case IDENT:
		identinexpr = 1
		refname += id
		p_tmpvar = idtail(refname, var_num)
		// idtail 和 expr 的区别：
		// <idtail>	-> assign<expr>|lparen<realarg>rparen
		// <exp> ->	<aloexp><exptail>
	case LPAREN:
		return expr(vn)
		// 消耗
	default:
		panic("语法不正确！")
	}
}

// <factortail>	->	<muls><item>|^
func factortail(p *parser, f1 *ProgDec, vn *int) *ProgDec {
	token := p.lexer.NextToken()
	if token == MUL || token == QUO { // 优先运算符
		f2 := item(p, vn)
		return genExp(token, f1, f2, vn)
	} else { // 其它运算符
		p.lexer.Back(token)
		return nil
	}
}
