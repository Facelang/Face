package factory

import "io"

type codegen struct {
	out io.Writer
}

func genAss(dst, src *ProgDec, vn *int) *ProgDec {

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
