package factory

import (
	"fmt"
	"io"
)

type codegen struct {
	out io.Writer
}

func genAss(dst, src *ProgDec, vn *int) *ProgDec {

}

// 产生函数头的代码，因为是在函数解析的过程中进行代码生成，所有默认是用的是tfun的信息opp!=addi
func (g *codegen) funhead(f *ProgFunc) {
	_, _ = fmt.Fprintf(g.out, "%s:\n", f.name)                                           //函数头
	_, _ = fmt.Fprintf(g.out, "\tpush ebp\n\tmov ebp,esp\n")                             //enter
	_, _ = fmt.Fprintf(g.out, "\tmov ebx,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,ebx\n") //esp<=>[@s_esp]
	_, _ = fmt.Fprintf(g.out, "\tmov ebx,[@s_ebp]\n\tpush ebx\n\tmov [@s_ebp],esp\n")    //s_enter
	_, _ = fmt.Fprintf(g.out, "\tmov ebx,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,ebx\n") //esp<=>[@s_esp]
	_, _ = fmt.Fprintf(g.out, "\t;函数头\n")
}

func (g *codegen) funtail(f *ProgFunc) {
	if f.hadret != 0 { // todo 不知道含义
		return
	}
	_, _ = fmt.Fprintf(g.out, "\t;函数尾\n")
	_, _ = fmt.Fprintf(g.out, "\tmov ebx,[@s_ebp]\n\tmov [@s_esp],ebx\n")                //s_leave
	_, _ = fmt.Fprintf(g.out, "\tmov ebx,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,ebx\n") //esp<=>[@s_esp]
	_, _ = fmt.Fprintf(g.out, "\tpop ebx\n\tmov [@s_ebp],ebx\n")                         //s_ebp
	_, _ = fmt.Fprintf(g.out, "\tmov ebx,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,ebx\n") //esp<=>[@s_esp]
	_, _ = fmt.Fprintf(g.out, "\tmov esp,ebp\n\tpop ebp\n\tret\n")                       //leave
}

// 为局部变量开辟新的空间，包括临时变量，但不包含参数变量，参数变量的空间一般在调用函数值前申请入栈的
func (g *codegen) genLocvar(val int) {
	_, _ = fmt.Fprintf(g.out, "\tpush %d\n", val)
}

/**
 * 产生表达式代码
 * 运算规则，字符可以当作整数进行运算，string只能加法运算
 * 有string类型的操作数结果是string，否则就是int
 * 参数说明：
 * 	p_factor1——左操作数
 * 	token——运算符
 * 	p_factor2——右操作数
 * 	var_num——复合语句里的变量个数
 */
func genExp(op Token, f1, f2 *ProgDec, vn *int) *ProgDec {
	if f1 == nil || f2 == nil {
		return nil
	}

	rslType := INT // 返回值类型默认 int
	// 字符可以当作整数进行运算，string只能加法运算
	// 有string类型的操作数结果是string，否则就是int
	if f1.kind == "string" || f2.kind == "string" {
		if op == ADD {
			rslType = STRING
		} else {
			panic("字符串不能运用于除了加法以外的运算。\n")
		}
	} else {
		// >,>=,<,<=,==,!= , 比较运算，返回值为一个字符
		if op == GTR || op == GEQ || op == LSS || op == LEQ || op == EQL || op == NEQ {
			rslType = CHAR
		}
	}
	// todo 生成代码
	// 先创建临时变量
	switch rslType {
	case STRING:

	}
}
