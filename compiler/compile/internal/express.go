package internal

import (
	"github.com/facelang/face/compiler/compile/internal/api"
	"github.com/facelang/face/compiler/compile/internal/parser"
)

// <exp>		->	<aloexp><exptail>
// 目前只支持 赋值, 比较， 加减， 乘除
// 赋值运算在【语句】层面被解析了
func expr(p *parser.parser, vn *int) *ProgDec {
	f1 := aloexp(p, vn)
	f2 := exptail(p, f1, vn)
	if f2 == nil {
		return f1
	}
	return f2
}

// <exptail>	->	<cmps><expr>|^
// 最低级运算符， 比较运算 【因为是最低级，所以多了结束符判定】
func exptail(p *parser.parser, f1 *ProgDec, vn *int) *ProgDec {
	token := p.lexer.NextToken()
	if token == api.ADD || token == api.QUO { // 优先运算符
		f2 := aloexp(p, vn)
		return p.gen.exp(token, f1, f2, vn)
	} else { // 结束符判断，或抛出异常
		p.lexer.Back(token)
		return nil
	}
}

// <aloexp>	->	<item><itemtail>
// 从最低级运算符开始解析
func aloexp(p *parser.parser, vn *int) *ProgDec {
	f1 := item(p, vn)
	f2 := itemtail(p, f1, vn)
	if f2 == nil {
		return f1
	}
	return f2
}

// <item>		->	<factor><factortail>
// 先解析一个符号，如果后一个符号优先级更高，先解析后一个符号并返回
func item(p *parser.parser, vn *int) *ProgDec {
	f1 := factor(p, vn)
	f2 := factortail(p, f1, vn)
	if f2 == nil {
		return f1
	}
	return f2
}

// <itemtail>	->	<adds><aloexp>|^
// 低一级运算符 + -
func itemtail(p *parser.parser, f1 *ProgDec, vn *int) *ProgDec {
	token := p.lexer.NextToken()
	if token == api.ADD || token == api.QUO { // 优先运算符
		f2 := aloexp(p, vn)
		return p.gen.exp(token, f1, f2, vn)
	} else { // 其它运算符
		p.lexer.Back(token)
		return nil
	}
}

// <factor> -> ident<idtail>|number|chara|lparen<expr>rparen|strings
func factor(p *parser.parser, vn *int) *ProgDec {
	switch token := p.lexer.NextToken(); token {
	case V_CHAR: // 值类型，直接创建【临时变量】
		val := p.lexer.content
		return p.progFn.createTempVar(p, "char", val, true, vn)
	case V_INT:
		val := p.lexer.content
		return p.progFn.createTempVar(p, "int", val, true, vn)
	case V_STRING:
		val := p.lexer.content
		return p.progFn.createTempVar(p, "string", val, true, vn)
	case IDENT:
		identinexpr = 1
		refname := p.lexer.content
		return p.idtail(refname, vn)
		// idtail 和 expr 的区别：
		// <idtail>	-> assign<expr>|lparen<realarg>rparen
		// <exp> ->	<aloexp><exptail>
	case api.LPAREN:
		ret := expr(p, vn)
		p.require(api.RPAREN)
		return ret
		// 消耗
	default:
		panic("语法不正确！")
	}
}

// <factortail>	->	<muls><item>|^
func factortail(p *parser.parser, f1 *ProgDec, vn *int) *ProgDec {
	token := p.lexer.NextToken()
	if token == api.MUL || token == api.QUO { // 优先运算符
		f2 := item(p, vn)
		return p.gen.exp(token, f1, f2, vn)
	} else { // 其它运算符
		p.lexer.Back(token)
		return nil
	}
}
