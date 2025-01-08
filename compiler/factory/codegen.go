package factory

import (
	"fmt"
	"io"
	"strconv"
)

type codegen struct {
	out    io.Writer
	parser *parser
}

func fprintf(g *codegen, format string, args ...interface{}) {
	_, _ = fmt.Fprintf(g.out, format, args...)
}

/**
 * 产生目标代码中的唯一的标签名
 * head——标题：tmp,var,str,fun,lab
 * type——类型: rsv_char,rsv_int,rsv_string
 * name——名称
 */

var _id = 0

func (g *codegen) id(head, kind, name string) string {
	_id += 1
	retStr := "@" + head
	if kind != "" {
		retStr += "_" + kind
	}
	if name != "" {
		retStr += "_" + name
	}
	if head != "str" && head != "fun" && head != "var" {
		retStr += "_" + strconv.Itoa(_id)
	}
	return retStr
}

func (g *codegen) assign(dst, src *ProgDec, vn *int) *ProgDec {
	if dst.kind == "void" {
		panic("void类型不能参加赋值运算。\n")
	}
	// 判断数据类型，需要是 string, char , int， 且需要判定类型转换
	if dst.kind == "string" {
		if src.strValId != -1 {
			empty := ProgDec{kind: "string"}
			src = g.exp(ADD, src, &empty, vn)
		}
		if dst.strValId == -2 {
			labLop := g.id("lab", "", "cpy2gstr")
			labExt := g.id("lab", "", "cpy2gstr_exit")
			if src.localAddr < 0 {
				fprintf(g, "\tmov ecx,0\n\tmov esi,[ebp%d]\n\tmov cl,[esi]\n", src.localAddr)
			} else {
				fprintf(g, "\tmov ecx,0\n\tmov esi,[ebp+%d]\n\tmov cl,[esi]\n", src.localAddr)
			}
			fprintf(g, "\tcmp ecx,0\n\tje %s\n", labExt)
			fprintf(g, "\tmov [@str_%s_len],cl\n", dst.name) //先复制长度
			fprintf(g, "\tsub esi,ecx\n")
			fprintf(g, "\tmov edi,@str_%s\n", dst.name)
			fprintf(g, "\tmov edx,0\n")
			fprintf(g, "%s:\n", labLop)
			fprintf(g, "\tmov al,[esi+edx]\n\tmov [edi+edx],al\n")
			fprintf(g, "\tinc edx\n\tcmp edx,ecx\n\tje %s\n\tjmp %s\n", labExt, labLop)
			fprintf(g, "%s:\n", labExt)
		} else {
			dst.strValId = -1
			if src.localAddr < 0 {
				fprintf(g, "\tmov eax,[ebp%d]\n", src.localAddr)
			} else {
				fprintf(g, "\tmov eax,[ebp+%d]\n", src.localAddr)
			}
			if dst.localAddr < 0 {
				fprintf(g, "\tmov [ebp%d],eax\n", dst.localAddr)
			} else {
				fprintf(g, "\tmov [ebp+%d],eax\n", dst.localAddr)
			}
		}
	} else { //int char 默认处理
		if dst.localAddr == 0 { // 全局的
			fprintf(g, "\tmov eax,@var_%s\n", dst.name)
		} else {
			if dst.localAddr < 0 {
				fprintf(g, "\tlea eax,[ebp%d]\n", dst.localAddr)
			} else {
				fprintf(g, "\tlea eax,[ebp+%d]\n", dst.localAddr)
			}
		}
		if dst.localAddr == 0 { // 全局的
			fprintf(g, "\tmov ebx,[@var_%s]\n", src.name)
		} else {
			if dst.localAddr < 0 {
				fprintf(g, "\tmov ebx,[ebp%d]\n", src.localAddr)
			} else {
				fprintf(g, "\tmov ebx,[ebp+%d]\n", src.localAddr)
			}
		}
		fprintf(g, "\tmov [eax],ebx\n")
	}
	return dst
}

func (g *codegen) call(t *ProgTable, name string, vn *int) *ProgDec {
	if fn, ok := t.fnRecList[name]; ok {
		l := len(t.realArgList)
		m := len(fn.args)
		if l < m {
			panic("函数实参的类型不能与函数的形参声明严格匹配。\\n")
		}
		// 产生参数进栈代码
		for i, j := l-1, m-1; j >= 0; i, j = i-1, j-1 {
			kind := t.realArgList[i].kind
			if kind != fn.args[j] {
				panic("函数实参的类型不能与函数的形参声明严格匹配。\\n")
			}
			// 产生参数进栈代码
			ret := t.realArgList[i]
			if ret.kind == "string" {
				// 将副本字符串的地址放在eax中
				fprintf(g, "\tmov eax,[ebp%d]\n", ret.localAddr)
			} else {
				if ret.localAddr == 0 { // 全局的？
					fprintf(g, "\tmov eax,[@var%s]\n", ret.name)
				} else { // 局部的？
					if ret.localAddr < 0 {
						fprintf(g, "\tmov eax,[@ebp%d]\n", ret.localAddr)
					} else {
						fprintf(g, "\tmov eax,[@ebp+%d]\n", ret.localAddr)
					}
				}
			}
			fprintf(g, "\tpush eax\n")
		}
		// 产生函数调用代码
		fprintf(g, "\tcall %s\n", name)
		fprintf(g, "\tadd esp,%d\n", 4*l)

		var rec *ProgDec
		// 产生函数返回代码
		// 非void函数在函数返回的时候将eax的数据放到临时变量中，为调用代码使用
		if fn.kind != "void" {
			// 创建临时变量
			val := g.parser.lexer.content
			rec = g.parser.progFn.createTempVar(g.parser, fn.kind, val, false, vn)
			fprintf(g, "\tmov [ebp%d],eax\n", rec.localAddr)
			if fn.kind == "string" { //返回的是临时string，必须拷贝
				empty := ProgDec{
					kind: "string",
				}
				rec = g.exp(ADD, &empty, rec, vn)
			}
		}

		// 清除实际参数
		for ; m > 0; m-- {
			t.realArgList = t.realArgList[:len(t.realArgList)-1]
		}

		return rec
	}
	panic("变量在使用之前没有合法的声明。\\n")
}

/**
 * 产生block的边界代码
 * 参数isIn：-1-进入block；0..n-退出block
 */
func (g *codegen) block(f *ProgFunc, n int) int {
	if n == -1 {
		return f.getCurAddr()
	} else {
		if n != 0 {
			fprintf(g, "\tlea esp,[ebp%d]\n", n)
		} else {
			fprintf(g, "\tmov esp,ebp\n")
		}
		return -2
	}
}

// 产生函数头的代码，因为是在函数解析的过程中进行代码生成，所有默认是用的是tfun的信息opp!=addi
func (g *codegen) funhead(f *ProgFunc) {
	fprintf(g, "%s:\n", f.name)                                           //函数头
	fprintf(g, "\tpush ebp\n\tmov ebp,esp\n")                             //enter
	fprintf(g, "\tmov ebx,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,ebx\n") //esp<=>[@s_esp]
	fprintf(g, "\tmov ebx,[@s_ebp]\n\tpush ebx\n\tmov [@s_ebp],esp\n")    //s_enter
	fprintf(g, "\tmov ebx,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,ebx\n") //esp<=>[@s_esp]
	fprintf(g, "\t;函数头\n")
}

func (g *codegen) funtail(f *ProgFunc) {
	if f.hadret != 0 { // todo 不知道含义
		return
	}
	fprintf(g, "\t;函数尾\n")
	fprintf(g, "\tmov ebx,[@s_ebp]\n\tmov [@s_esp],ebx\n")                //s_leave
	fprintf(g, "\tmov ebx,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,ebx\n") //esp<=>[@s_esp]
	fprintf(g, "\tpop ebx\n\tmov [@s_ebp],ebx\n")                         //s_ebp
	fprintf(g, "\tmov ebx,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,ebx\n") //esp<=>[@s_esp]
	fprintf(g, "\tmov esp,ebp\n\tpop ebp\n\tret\n")                       //leave
}

// 为局部变量开辟新的空间，包括临时变量，但不包含参数变量，参数变量的空间一般在调用函数值前申请入栈的
func (g *codegen) locvar(val int) {
	fprintf(g, "\tpush %d\n", val)
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
func (g *codegen) exp(op Token, f1, f2 *ProgDec, vn *int) *ProgDec {
	if f1 == nil || f2 == nil {
		return nil
	}
	if f1.kind == "void" || f2.kind == "void" {
		panic("void 类型，不能参数运算")
		return nil
	}

	rslType := "int" // 返回值类型默认 int
	// 字符可以当作整数进行运算，string只能加法运算
	// 有string类型的操作数结果是string，否则就是int
	if f1.kind == "string" || f2.kind == "string" {
		if op == ADD {
			rslType = "string"
		} else {
			panic("字符串不能运用于除了加法以外的运算。\n")
		}
	} else {
		// >,>=,<,<=,==,!= , 比较运算，返回值为一个字符
		if op == GTR || op == GEQ || op == LSS || op == LEQ || op == EQL || op == NEQ {
			rslType = "char"
		}
	}
	// todo 生成代码
	// 先创建临时变量
	val := g.parser.lexer.content
	rec := g.parser.progFn.createTempVar(g.parser, rslType, val, false, vn)
	switch rslType {
	case "string": // 字符串连接运算
		// cout<<"字符串链接运算"<<endl;
		if f2.kind == "string" {
			labLop := g.id("lab", "", "cpystr2")
			labExt := g.id("lab", "", "cpystr2_exit")
			if f2.strValId == -1 { // 动态 string
				fprintf(g, ";----------生成动态string%s的代码----------\n", f2.name)
				fprintf(g, "\tmov eax,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,eax\n") //esp<=>[@s_esp]
				if f2.localAddr < 0 {
					fprintf(g, "\tmov ebx,[ebp%d]\n\tmov eax,0\n\tmov al,[ebx]\n", f2.localAddr)
				} else {
					fprintf(g, "\tmov ebx,[ebp+%d]\n\tmov eax,0\n\tmov al,[ebx]\n", f2.localAddr)
				}
				fprintf(g, "\tsub esp,1\n\tmov [esp],al;长度压入后再压入数据栈\n")
				fprintf(g, "\tmov [ebp%d],esp\n", rec.localAddr) //存入数据指针

				fprintf(g, "\tcmp eax,0\n")
				fprintf(g, "\tje %s\n", labExt)
				fprintf(g, "\tmov ecx,0\n")
				fprintf(g, "\tmov esi,ebx\n\tsub esi,1\n")
				fprintf(g, "\tneg eax\n")
				fprintf(g, "%s:\n", labLop)
				fprintf(g, "\tcmp ecx,eax\n")
				fprintf(g, "\tje %s\n", labExt)
				fprintf(g, "\tmov dl,[esi+ecx]\n")
				fprintf(g, "\tsub esp,1\n\tmov [esp],dl\n")
				fprintf(g, "\tdec ecx\n")
				fprintf(g, "\tjmp %s\n", labLop)
				fprintf(g, "%s:\n", labExt)
				fprintf(g, "\tmov eax,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,eax\n") //esp<=>[@s_esp]
			} else if f2.strValId > 0 { // 常量string
				fprintf(g, ";----------生成常量string%s的代码----------\n", f2.name)
				fprintf(g, "\tmov eax,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,eax\n") //esp<=>[@s_esp]
				fprintf(g, "\tmov eax,@str_%d_len\n\tsub esp,1\n\tmov [esp],al;长度压入后再压入数据栈\n", f2.strValId)
				fprintf(g, "\tmov [ebp%d],esp\n", rec.localAddr) //存入数据指针

				fprintf(g, "\tcmp eax,0\n") //测试长度是否是0
				fprintf(g, "\tje %s\n", labExt)
				fprintf(g, "\tmov ecx,@str_%d_len\n\tdec ecx\n", f2.strValId) //
				fprintf(g, "\tmov esi,@str_%d\n", f2.strValId)                //取得首地址
				fprintf(g, "%s:\n", labLop)
				fprintf(g, "\tcmp ecx,-1\n")
				fprintf(g, "\tje %s\n", labExt)
				fprintf(g, "\tmov al,[esi+ecx]\n")
				fprintf(g, "\tsub esp,1\n\tmov [esp],al\n")
				fprintf(g, "\tdec ecx\n")
				fprintf(g, "\tjmp %s\n", labLop)
				fprintf(g, "%s:\n", labExt)
				fprintf(g, "\tmov eax,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,eax\n") //esp<=>[@s_esp]
			} else if f2.strValId == -2 {
				fprintf(g, ";----------生成全局string%s的代码----------\n", f2.name)
				fprintf(g, "\tmov eax,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,eax\n") //esp<=>[@s_esp]
				if convert_buffer == 0 {
					fprintf(g, "\tmov eax,0\n\tmov al,[@str_%s_len]\n\tsub esp,1\n\tmov [esp],al;长度压入后再压入数据栈\n", f2.name)
				} else {
					fprintf(g, "\tmov eax,0\n\tmov al,[%s_len]\n\tsub esp,1\n\tmov [esp],al;长度压入后再压入数据栈\n", f2.name)
				}
				fprintf(g, "\tmov [ebp%d],esp\n", rec.localAddr) //存入数据指针

				fprintf(g, "\tcmp eax,0\n") //测试长度是否是0
				fprintf(g, "\tje %s\n", labExt)
				fprintf(g, "\tsub eax,1\n\tmov ecx,eax\n")
				if convert_buffer == 0 {
					fprintf(g, "\tmov esi,@str_%s\n", f2.name) //取得首地址
				} else {
					fprintf(g, "\tmov esi,%s\n", f2.name) //取得首地址
				}
				fprintf(g, "%s:\n", labLop)
				fprintf(g, "\tcmp ecx,-1\n")
				fprintf(g, "\tje %s\n", labExt)
				fprintf(g, "\tmov al,[esi+ecx]\n")
				fprintf(g, "\tsub esp,1\n\tmov [esp],al\n")
				fprintf(g, "\tdec ecx\n")
				fprintf(g, "\tjmp %s\n", labLop)
				fprintf(g, "%s:\n", labExt)
				fprintf(g, "\tmov eax,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,eax\n") //esp<=>[@s_esp]
			} else if f2.strValId == 0 {
				fprintf(g, "\tmov eax,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,eax\n") //esp<=>[@s_esp]
				fprintf(g, "\tmov eax,0\n\tsub esp,1\n\tmov [esp],al;长度压入后再压入数据栈\n")
				fprintf(g, "\tmov [ebp%d],esp\n", rec.localAddr)                      //存入数据指针
				fprintf(g, "\tmov eax,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,eax\n") //esp<=>[@s_esp]
			}
		} else if f2.kind == "int" {
			labLop := g.id("lab", "", "numtostr2")
			labExt := g.id("lab", "", "numtostr2_exit")
			labNumSign := g.id("lab", "", "numsign2")
			labNumSignExt := g.id("lab", "", "numsign2_exit")
			fprintf(g, ";----------生成number%s的string代码----------\n", f2.name)
			fprintf(g, ""+
				"\tmov eax,[@s_esp]\n"+
				"\tmov [@s_esp],esp\n"+
				"\tmov esp,eax\n")
			if f2.localAddr == 0 { //全局的
				fprintf(g, "\tmov eax,[@var_%s]\n", f2.name)
			} else { //局部的
				if f2.localAddr < 0 {
					fprintf(g, "\tmov eax,[ebp%d]\n", f2.localAddr)
				} else {
					fprintf(g, "\tmov eax,[ebp+%d]\n", f2.localAddr)
				}
			}
			fprintf(g, ""+
				"\tsub esp,1;先把数字的长度位置空出来\n"+
				"tmov ecx,0\n\tmov [esp],cl\n"+
				"tmov esi,esp\n")
			fprintf(g, "\tmov [ebp%d],esp\n", rec.localAddr) //存入数据指针
			//确定数字的正负
			fprintf(g, "\tmov edi,0\n") //保存eax符号：0+ 1-
			fprintf(g, "\tcmp eax,0\n")
			fprintf(g, "\tjge %s\n", labNumSignExt)
			fprintf(g, "%s:\n", labNumSign)
			fprintf(g, "\tneg eax\n")
			fprintf(g, "\tmov edi,1\n")
			fprintf(g, "%s:\n", labNumSignExt)

			fprintf(g, "\tmov ebx,10\n")
			fprintf(g, "%s:\n", labLop)
			fprintf(g, ""+
				"\tmov edx,0\n"+
				"tidiv ebx\n"+
				"\tmov cl,[esi]\n"+
				"\tinc cl\n"+
				"\tmov [esi],cl\n"+
				"\tsub esp,1\n"+
				"\tadd dl,48\n"+
				"\tmov [esp],dl\n"+
				"\tcmp eax,0\n")
			fprintf(g, "\tjne %s\n", labLop)

			fprintf(g, "\tcmp edi,0\n")
			fprintf(g, "\tje %s\n", labExt)
			fprintf(g, "\tsub esp,1\n\tmov ecx,%d\n\tmov [esp],cl\n", '-')
			fprintf(g, "\tmov cl,[esi]\n\tinc cl\n\tmov [esi],cl\n")
			fprintf(g, "%s:\n", labExt)
			fprintf(g, ""+
				"\tmov eax,[@s_esp]\n"+
				"\tmov [@s_esp],esp\n"+
				"\tmov esp,eax\n")
		} else if f2.kind == "char" {
			fprintf(g, ";----------生成char%s的string代码----------\n", f2.name)
			fprintf(g, ""+
				"\tmov eax,[@s_esp]\n"+
				"\tmov [@s_esp],esp\n"+
				"\tmov esp,eax\n")
			if f2.localAddr == 0 { //全局的
				fprintf(g, "\tmov eax,[@var_%s]\n", f2.name)
			} else { //局部的
				if f2.localAddr < 0 {
					fprintf(g, "\tmov eax,[ebp%d]\n", f2.localAddr)
				} else {
					fprintf(g, "\tmov eax,[ebp+%d]\n", f2.localAddr)
				}
				fprintf(g, "\tsub esp,1\n\tmov bl,1\n\tmov [esp],bl\n\tmov [ebp%d],esp\n", rec.localAddr) //存入数据指针
				fprintf(g, "\tsub esp,1\n\tmov [esp],al\n")
				fprintf(g, ""+
					"\tmov eax,[@s_esp]\n"+
					"\tmov [@s_esp],esp\n"+
					"\tmov esp,eax\n")
			}
		}
		fprintf(g, ";--------------------------------------------------\n")
		if f1.kind == "string" {
			labLop := g.id("lab", "", "cpystr1")
			labExt := g.id("lab", "", "cpystr1_exit")
			if f1.strValId == -1 { //动态string
				fprintf(g, ";----------生成动态string%s的代码----------\n", f1.name)
				fprintf(g, "\tmov eax,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,eax\n") //esp<=>[@s_esp]
				if f1.localAddr < 0 {
					fprintf(g, "\tmov ebx,[ebp%d]\n\tmov eax,0\n\tmov al,[ebx]\n", f1.localAddr)
				} else {
					fprintf(g, "\tmov ebx,[ebp+%d]\n\tmov eax,0\n\tmov al,[ebx]\n", f1.localAddr)
				}
				fprintf(g, "\tcmp eax,0\n")
				fprintf(g, "\tje %s\n", labExt)

				fprintf(g, "\tmov ebx,[ebp%d];\n", rec.localAddr) //将结果字符串的长度追加

				fprintf(g, "\tmov edx,0\n\tmov dl,[ebx]\n")
				fprintf(g, "\tadd edx,eax\n")
				fprintf(g, "\tmov [ebx],dl\n")

				fprintf(g, "\tmov ecx,0\n")
				if f1.localAddr < 0 {
					fprintf(g, "\tmov esi,[ebp%d]\n\tsub esi,1\n", f1.localAddr) //消除偏移
				} else {
					fprintf(g, "\tmov esi,[ebp+%d]\n\tsub esi,1\n", f1.localAddr) //消除偏移
				}
				fprintf(g, "\tneg eax\n")
				//仅仅是测试字符串总长是否超过255，超出部分忽略
				fprintf(g, "\tcmp edx,255\n")
				fprintf(g, "\tjna %s\n", labLop)
				fprintf(g, "\tcall @str2long\n")

				fprintf(g, "%s:\n", labLop)
				fprintf(g, "\tcmp ecx,eax\n")
				fprintf(g, "\tje %s\n", labExt)
				fprintf(g, "\tmov dl,[esi+ecx]\n")
				fprintf(g, "\tsub esp,1\n\tmov [esp],dl\n")
				fprintf(g, "\tdec ecx\n")
				fprintf(g, "\tjmp %s\n", labLop)
				fprintf(g, "%s:\n", labExt)
				fprintf(g, "\tmov eax,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,eax\n") //esp<=>[@s_esp]
			} else if f1.strValId > 0 { //常量string
				fprintf(g, ";----------生成常量string%s的代码----------\n", f1.name)
				fprintf(g, "\tmov eax,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,eax\n") //esp<=>[@s_esp]
				fprintf(g, "\tmov eax,@str_%d_len\n", f1.strValId)
				fprintf(g, "\tcmp eax,0\n")

				fprintf(g, "\tje %s\n", labExt)
				fprintf(g, "\tmov ebx,[ebp%d];\n", rec.localAddr) //将结果字符串的长度追加
				fprintf(g, "\tmov edx,0\n\tmov dl,[ebx]\n")
				fprintf(g, "\tadd edx,eax\n")
				fprintf(g, "\tmov [ebx],dl\n")

				fprintf(g, "\tmov ecx,@str_%d_len\n\tdec ecx\n", f1.strValId)
				fprintf(g, "\tmov esi,@str_%d\n", f1.strValId)
				//仅仅是测试字符串总长是否超过255，超出报错
				fprintf(g, "\tcmp edx,255\n")
				fprintf(g, "\tjna %s\n", labLop)
				fprintf(g, "\tcall @str2long\n")

				fprintf(g, "%s:\n", labLop)
				fprintf(g, "\tcmp ecx,-1\n")
				fprintf(g, "\tje %s\n", labExt)
				fprintf(g, "\tmov al,[esi+ecx]\n")
				fprintf(g, "\tsub esp,1\n\tmov [esp],al\n")
				fprintf(g, "\tdec ecx\n")
				fprintf(g, "\tjmp %s\n", labLop)
				fprintf(g, "%s:\n", labExt)
				fprintf(g, "\tmov eax,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,eax\n") //esp<=>[@s_esp]
			} else if f1.strValId == -2 { //全局string
				fprintf(g, ";----------生成全局string%s的代码----------\n", f1.name)
				fprintf(g, "\tmov eax,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,eax\n") //esp<=>[@s_esp]
				if convert_buffer == 0 {
					fprintf(g, "\tmov eax,0\n\tmov al,[@str_%s_len]\n", f1.name)
				} else {
					fprintf(g, "\tmov eax,0\n\tmov al,[%s_len]\n", f1.name)
				}
				fprintf(g, "\tcmp eax,0\n")

				fprintf(g, "\tje %s\n", labExt)
				fprintf(g, "\tmov ebx,[ebp%d];\n", rec.localAddr) //将结果字符串的长度追加
				fprintf(g, "\tmov edx,0\n\tmov dl,[ebx]\n")
				fprintf(g, "\tadd edx,eax\n")
				fprintf(g, "\tmov [ebx],dl\n")

				fprintf(g, "\tsub eax,1\n\tmov ecx,eax\n")
				if convert_buffer == 0 {
					fprintf(g, "\tmov esi,@str_%s\n", f1.name)
				} else {
					fprintf(g, "\tmov esi,%s\n", f1.name)
				}
				//仅仅是测试字符串总长是否超过255，超出报错
				fprintf(g, "\tcmp edx,255\n")
				fprintf(g, "\tjna %s\n", labLop)
				fprintf(g, "\tcall @str2long\n")

				fprintf(g, "%s:\n", labLop)
				fprintf(g, "\tcmp ecx,-1\n")
				fprintf(g, "\tje %s\n", labExt)
				fprintf(g, "\tmov al,[esi+ecx]\n")
				fprintf(g, "\tsub esp,1\n\tmov [esp],al\n")
				fprintf(g, "\tdec ecx\n")
				fprintf(g, "\tjmp %s\n", labLop)
				fprintf(g, "%s:\n", labExt)
				fprintf(g, "\tmov eax,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,eax\n") //esp<=>[@s_esp]
			}
		} else if f1.kind == "int" {
			labLop := g.id("lab", "", "numtostr1")
			labExt := g.id("lab", "", "numtostr1_exit")
			labNumSign := g.id("lab", "", "numsign1")
			labNumSignExt := g.id("lab", "", "numsign1_exit")
			lab2long := g.id("lab", "", "numsign1_add")
			fprintf(g, ";----------生成number%s的string代码----------\n", f1.name)
			fprintf(g, ""+
				"\tmov eax,[@s_esp]\n"+
				"\tmov [@s_esp],esp\n"+
				"\tmov esp,eax\n")
			if f1.localAddr == 0 { //全局的
				fprintf(g, "\tmov eax,[@var_%s]\n", f1.name)
			} else { //局部的
				if f1.localAddr < 0 {
					fprintf(g, "\tmov eax,[ebp%d]\n", f1.localAddr)
				} else {
					fprintf(g, "\tmov eax,[ebp+%d]\n", f1.localAddr)
				}
			}
			fprintf(g, "\tmov esi,[ebp%d];\n", rec.localAddr) //将临时字符串的长度地址记录下来

			//确定数字的正负
			fprintf(g, "\tmov edi,0\n") //保存eax符号：0+ 1-
			fprintf(g, "\tcmp eax,0\n")
			fprintf(g, "\tjge %s\n", labNumSignExt)
			fprintf(g, "%s:\n", labNumSign)
			fprintf(g, "\tneg eax\n")
			fprintf(g, "\tmov edi,1\n")
			fprintf(g, "%s:\n", labNumSignExt)

			//累加长度，压入数据
			fprintf(g, "\tmov ebx,10\n")
			fprintf(g, "%s:\n", labLop)
			fprintf(g, ""+
				"\tmov edx,0\n"+
				"\tidiv ebx\n"+
				"\tmov cl,[esi]\n"+
				"\tinc cl\n"+
				"\tmov [esi],cl\n"+
				"\tsub esp,1\n"+
				"\tadd dl,48\n"+
				"\tmov [esp],dl\n"+
				"\tcmp eax,0\n")
			fprintf(g, "\tjne %s\n", labLop)

			//添加符号
			fprintf(g, "\tcmp edi,0\n")
			fprintf(g, "\tje %s\n", lab2long)
			fprintf(g, "\tsub esp,1\n\tmov ecx,%d\n\tmov [esp],cl\n", '-')
			fprintf(g, "\tmov cl,[esi]\n\tinc cl\n\tmov [esi],cl\n")

			fprintf(g, "%s:\n", lab2long)
			//仅仅是测试字符串总长是否超过255，超出报错
			fprintf(g, "\tcmp cl,255\n")
			fprintf(g, "\tjna %s\n", labExt)
			fprintf(g, "\tcall @str2long\n")
			fprintf(g, "%s:\n", labExt)
			fprintf(g, ""+
				"\tmov eax,[@s_esp]\n"+
				"\tmov [@s_esp],esp\n"+
				"\tmov esp,eax\n")
		} else if f1.kind == "char" {
			labExt := g.id("lab", "", "chtostr2_exit")
			fprintf(g, ";----------生成char%s的string代码----------\n", f1.name)
			fprintf(g, ""+
				"\tmov eax,[@s_esp]\n"+
				"\tmov [@s_esp],esp\n"+
				"\tmov esp,eax\n")
			if f1.localAddr == 0 { //全局的
				fprintf(g, "\tmov eax,[@var_%s]\n", f1.name)
			} else { //局部的
				if f1.localAddr < 0 {
					fprintf(g, "\tmov eax,[ebp%d]\n", f1.localAddr)
				} else {
					fprintf(g, "\tmov eax,[ebp+%d]\n", f1.localAddr)
				}
			}
			fprintf(g, "\tmov esi,[ebp%d];\n", rec.localAddr) //将临时字符串的长度地址记录下来

			//累加长度，压入数据
			fprintf(g, ""+
				"\tmov cl,[esi]\n"+
				"\tinc cl\n"+
				"\tmov [esi],cl\n"+
				"\tsub esp,1\n"+
				"\tmov [esp],al\n")

			//仅仅是测试字符串总长是否超过255，超出报错
			fprintf(g, "\tcmp cl,255\n")
			fprintf(g, "\tjna %s\n", labExt)
			fprintf(g, "\tcall @str2long\n")
			fprintf(g, "%s:\n", labExt)
			fprintf(g, ""+
				"\tmov eax,[@s_esp]\n"+
				"\tmov [@s_esp],esp\n"+
				"\tmov esp,eax\n")
		}
		fprintf(g, ";--------------------------------------------------\n")
		break
	case "int": //需要考虑+ - * / 类型：int 算术运算
		//cout<<"算术运算"<<endl;
		if f1.localAddr == 0 { //全局的
			fprintf(g, "\tmov eax,[@var_%s]\n", f1.name)
		} else { //局部的
			if f1.localAddr < 0 {
				fprintf(g, "\tmov eax,[ebp%d]\n", f1.localAddr)
			} else {
				fprintf(g, "\tmov eax,[ebp+%d]\n", f1.localAddr)
			}
		}
		if f2.localAddr == 0 { //全局的
			fprintf(g, "\tmov ebx,[@var_%s]\n", f2.name)
		} else { //局部的
			if f2.localAddr < 0 {
				fprintf(g, "\tmov ebx,[ebp%d]\n", f2.localAddr)
			} else {
				fprintf(g, "\tmov ebx,[ebp+%d]\n", f2.localAddr)
			}
		}

		switch op {
		case ADD:
			fprintf(g, "\tadd eax,ebx\n")
			break
		case SUB:
			fprintf(g, "\tsub eax,ebx\n")
			break
		case MUL:
			fprintf(g, "\timul ebx\n")
			break
		case QUO:
			fprintf(g, "\tmov edx,0\n")
			fprintf(g, "\tidiv ebx\n")
			break
		}
		fprintf(g, "\tmov [ebp%d],eax\n", rec.localAddr)
		break

	case "char": // 比较运算
		if f1.kind == "string" {
			//cout<<"字符串比较运算"<<endl;
		} else {
			//cout<<"基本比较运算"<<endl;
			labLop := g.id("lab", "", "base_cmp")
			labExt := g.id("lab", "", "base_cmp_exit")
			if f1.localAddr == 0 { //全局的
				fprintf(g, "\tmov eax,[@var_%s]\n", f1.name)
			} else { //局部的
				if f1.localAddr < 0 {
					fprintf(g, "\tmov eax,[ebp%d]\n", f1.localAddr)
				} else {
					fprintf(g, "\tmov eax,[ebp+%d]\n", f1.localAddr)
				}
			}
			if f1.localAddr == 0 { //全局的
				fprintf(g, "\tmov ebx,[@var_%s]\n", f2.name)
			} else { //局部的
				if f2.localAddr < 0 {
					fprintf(g, "\tmov ebx,[ebp%d]\n", f2.localAddr)
				} else {
					fprintf(g, "\tmov ebx,[ebp+%d]\n", f2.localAddr)
				}
			}

			fprintf(g, "\tcmp eax,ebx\n")
			switch op {
			case GTR:
				fprintf(g, "\tjg %s\n", labLop)
				break
			case GEQ:
				fprintf(g, "\tjge %s\n", labLop)
				break
			case LSS:
				fprintf(g, "\tjl %s\n", labLop)
				break
			case LEQ:
				fprintf(g, "\tjle %s\n", labLop)
				break
			case EQL:
				fprintf(g, "\tje %s\n", labLop)
				break
			case NEQ:
				fprintf(g, "\tjne %s\n", labLop)
				break
			}
			fprintf(g, "\tmov eax,0\n")
			fprintf(g, "\tjmp %s\n", labExt)
			fprintf(g, "%s:\n", labLop)
			fprintf(g, "\tmov eax,1\n")
			fprintf(g, "%s:\n", labExt)
			fprintf(g, "\tmov [ebp%d],eax\n", rec.localAddr)
		}
		break
	}
	return rec
}

/**
 * 产生返回语句的代码——返回值是内容的复制，即使它返回的是string类型《参数传递的string也是副本》
 */
func (g *codegen) ret(ret *ProgDec, vn *int) {
	if ret != nil {
		if ret.kind == "string" {
			empty := ProgDec{kind: "string"}
			ret = g.exp(ADD, &empty, ret, vn)
		}
		if ret.localAddr < 0 {
			fprintf(g, "\tmov eax,[ebp%d]\n", ret.localAddr) //将副本字符串的地址放在eax中
		} else {
			fprintf(g, "\tmov eax,[ebp+%d]\n", ret.localAddr) //将副本字符串的地址放在eax中
		}
	} else {
		if ret.localAddr == 0 { //全局的
			fprintf(g, "\tmov eax,[@var_%s]\n", ret.name)
		} else {
			if ret.localAddr < 0 {
				fprintf(g, "\tmov eax,[ebp%d]\n", ret.localAddr)
			} else {
				fprintf(g, "\tmov eax,[ebp+%d]\n", ret.localAddr)
			}
		}
	}

	//函数末尾清理
	fprintf(g, "\tmov ebx,[@s_ebp]\n\tmov [@s_esp],ebx\n")                //s_leave
	fprintf(g, "\tmov ebx,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,ebx\n") //esp<=>[@s_esp]
	fprintf(g, "\tpop ebx\n\tmov [@s_ebp],ebx\n")                         //s_ebp
	fprintf(g, "\tmov ebx,[@s_esp]\n\tmov [@s_esp],esp\n\tmov esp,ebx\n") //esp<=>[@s_esp]
	fprintf(g, "\tmov esp,ebp\n\tpop ebp\n\tret\n")                       //leave
}

var convert_buffer = 0 // 标志是不是对缓冲区的数据进行操作
func (g *codegen) input(i *ProgDec, vn *int) {
	if i == nil {
		return
	}
	if i.kind == "void" {
		panic("void类型不能作为输入的对象。\n")
	}

	fprintf(g, "\t;为%s产生输入代码\n", i.name)
	fprintf(g, "\tmov ecx,@buffer\n\tmov edx,255\n\tmov ebx,0\n\tmov eax,3\n\tint 128\n") //输入到缓冲区
	///还没有更好的解决方式～～～2010-5-17 21：18
	//计算缓冲区的具体字符个数，第一个\n之前的所有字符,同时计算如是数字的值，放在eax
	fprintf(g, "\tcall @procBuf\n")
	//eax-数字的值,bl-字符，ecx-串长度
	if i.kind == "string" {
		//将buffer临时作为全局string变量输入到指定p_i
		gBuf := ProgDec{
			name:     "@buffer",
			kind:     "string",
			strValId: -2,
		}
		convert_buffer = 1
		g.assign(i, &gBuf, vn)
		convert_buffer = 0
	} else if i.kind == "int" {
		if i.localAddr == 0 { // 全局的
			fprintf(g, "\tmov [@var_%s],eax\n", i.name)
		} else {
			if i.localAddr < 0 {
				fprintf(g, "\tmov [ebp%d],eax\n", i.localAddr)
			} else {
				fprintf(g, "\tmov [ebp+%d],eax\n", i.localAddr)
			}
		}
	} else { // 猜测应该是 char 类型
		if i.localAddr == 0 { // 全局的
			fprintf(g, "\tmov [@var_%s],bl\n", i.name)
		} else {
			if i.localAddr < 0 {
				fprintf(g, "\tmov [ebp%d],bl\n", i.localAddr)
			} else {
				fprintf(g, "\tmov [ebp+%d],bl\n", i.localAddr)
			}
		}
	}
}

func (g *codegen) output(o *ProgDec, vn *int) {
	if o == nil {
		return
	}
	fprintf(g, "\t;为%s产生输出代码\n", o.name)
	//强制产生副本
	empty := ProgDec{kind: "string"}
	o = g.exp(ADD, &empty, o, vn)
	fprintf(g, "\tmov ecx,[ebp%d]\n\tmov edx,0\n\tmov dl,[ecx]\n\tsub ecx,edx\n\tmov ebx,1\n\tmov eax,4\n\tint 128\n", o.localAddr)
}
