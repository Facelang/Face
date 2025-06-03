package internal

import (
	"fmt"
	"github.com/facelang/face/compiler/compile/internal/api"
	"github.com/facelang/face/compiler/compile/internal/parser"
	"os"
	"strconv"
)

// ProgDec 变量声明记录
type ProgDec struct {
	kind      string //类型
	name      string //名称
	intVal    int
	charVal   byte
	voidVal   int
	strValId  int //用strValId=-1来标示临时string
	localAddr int //局部变量相对与ebp指针的地址，或者临时string的索引地址
	extern    int //标示变量是不是外部的符号，针对全局变量来作用
	//var_record();//默认构造函数
	//void init(symbol dec_type,string dec_name);//声明初始化函数
	//void copy(const var_record*src);//拷贝函数
	//var_record(const var_record& v);//拷贝构造函数
	//~var_record();
}

type ProgFunc struct {
	kind      string     // 类型
	name      string     // 名称
	args      []string   // 参数类型列表
	localVars []*ProgDec // 局部变量列表,指向哈希表,仅仅为函数定义服务
	defined   int        // 函数是否给出定义
	flushed   int        // 函数参数是否已经缓冲写入，标记是否再清楚的时候清除缓冲区
	hadret    int        // 记录是否含有返回语句
}

func (fn *ProgFunc) addArg(gen *compile.codegen, v *ProgDec) {
	fn.args = append(fn.args, v.kind)
	fn.addLocalVar(gen, v) // 把参数当作变量临时存储在局部列表里边
}

// 添加一个局部变量
func (fn *ProgFunc) addLocalVar(gen *compile.codegen, v *ProgDec) {
	//局部变量定义的缓冲机制，defined==1之前，是不写入符号表的，因为此时还不能确定是不是函数定义

	if fn.defined == 0 { //还是参数声明
		fn.localVars = append(fn.localVars, v)
	} else {
		fn.localVars = append(fn.localVars, v)
		gen.parser.progTable.addVar(v)
		// 局部变量的地址按照ebp-4*count的方式变化,修改
		v.localAddr = -4 * (len(fn.localVars) - len(fn.args))
		//代码中为局部变量开辟临时空间
		gen.locvar(0)
	}
}

// 防止参数的名字在写入之前重复
func (fn *ProgFunc) exist(name string) bool {
	for _, v := range fn.localVars {
		if v.name == name {
			return true
		}
	}
	return false
}

func NumberVal(input string) int {
	r, _ := strconv.Atoi(input)
	return r
}

func (fn *ProgFunc) createTempVar(p *parser.parser, kind, val string, hasVal bool, vn *int) *ProgDec {
	// 创建临时变量记录
	temp := &ProgDec{kind: kind}
	switch kind {
	case "int":
		if hasVal {
			temp.intVal = NumberVal(val) // 取值
		}
	case "char":
		temp.charVal = ' '
	case "string":
		if hasVal {
			temp.strValId = p.progTable.addString(val)
		} else {
			temp.strValId = -1
		}
	}
	temp.name = p.gen.id("temp", kind, "")
	*vn += 1 // 局部变量计数
	// 添加临时变量入栈
	// 记录当前的局部变量
	fn.localVars = append(fn.localVars, temp)
	p.progTable.addVar(temp) // 应该可以不写入名字表，但是为了保持变量清除的一致性，作此操作
	// 计算地址
	temp.localAddr = -4 * (len(fn.localVars) - len(fn.args))
	// 代码中为局部变量开辟临时空间
	p.gen.locvar(temp.strValId)
	return temp
}

/**
 * 取得上一个进栈的变量的地址（esp相对于ebp偏移）
 */
func (fn *ProgFunc) getCurAddr() int {
	return -4 * (len(fn.localVars) - len(fn.args))
}

// 将参数写到符号表
func (fn *ProgFunc) flushargs(table *ProgTable) {
	for i := len(fn.args) - 1; i >= 0; i-- {
		v := fn.localVars[i]
		v.localAddr = 4 * (i + 2) // 修改参数地址，参数的地址按照ebp+4*count+4的方式变化
		if v.kind == "string" {
			v.strValId = -1 // 参数的类型统一为动态string
		}
		table.addVar(v)
	}
	fn.flushed = 1
}

// 弹出多个局部变量
func (fn *ProgFunc) poplocalvars(table *ProgTable, num int) {
	if num < 0 { // 函数定义结束 [全部清除]
		for i := 0; i < len(fn.args); i++ {
			table.delVar(fn.localVars[i].name) // 删除参数变量
		}
		fn.localVars = fn.localVars[:0]
		return
	}

	for i := 0; i < num; i++ {
		last := fn.localVars[len(fn.localVars)-1]
		fn.localVars = fn.localVars[:len(fn.localVars)-1]
		table.delVar(last.name)
	}
}

type ProgTable struct {
	fnRecList   map[string]*ProgFunc // 变量声明列表
	varRecList  map[string]*ProgDec  // 函数声明列表
	stringTable []*string            // 串空间
	realArgList []*ProgDec           // 函数调用的实参列表，用于检查参数调用匹配和实参代码生成
}

func (t *ProgTable) addFn(gen *compile.codegen, f *ProgFunc) {
	if _, ok := t.fnRecList[f.name]; ok {
		// TODO 判断已定义函数，和新的函数是否一直（参数数量，类型）
		// 为了简化语法，这里不允许重复定义
		_ = fmt.Errorf("(Func)重复定义【%s】\n", f.name)
		os.Exit(1)
	}
	t.fnRecList[f.name] = f
	if f.defined == 1 {
		f.flushargs(t)
		gen.funhead(f)
	}
}

func (t *ProgTable) addVar(v *ProgDec) {
	if _, ok := t.varRecList[v.name]; ok {
		// 为了简化语法，这里不允许重复定义
		_ = fmt.Errorf("(Var)重复定义【%s】\n", v.name)
		os.Exit(1)
	}
	t.varRecList[v.name] = v
}

var stringId = 0

func (t *ProgTable) addString(val string) int {
	stringId += 1
	t.stringTable = append(t.stringTable, &val)
	return stringId
}

func (t *ProgTable) getString(index int) string {
	return *t.stringTable[index]
}

func (t *ProgTable) addRealArg(gen *compile.codegen, arg *ProgDec, vn *int) {
	if arg.kind == "string" {
		empty := ProgDec{
			kind: "string",
		}
		arg = gen.exp(api.ADD, &empty, arg, vn)
	}
	t.realArgList = append(t.realArgList, arg)
}

func (t *ProgTable) getVar(name string) *ProgDec {
	if v, ok := t.varRecList[name]; ok {
		return v
	}
	panic("函数被调用之前没有合法的声明。\\n")
}

func (t *ProgTable) delVar(name string) {
	if _, ok := t.varRecList[name]; ok {
		delete(t.varRecList, name)
	}
}

// 测试局部变量，参数的名字是否重复，主要应对变量的重复定义，全局变量和函数不需要调用他
func (t *ProgTable) exist(name string) bool {
	_, ok1 := t.fnRecList[name]
	_, ok2 := t.varRecList[name]
	return ok1 || ok2
}

// 最后处理， 生成数据段中的静态数据区、文字池和辅助栈
func (t *ProgTable) over(gen *compile.codegen) {
	// hash_map<string, var_record*, string_hash>::iterator var_i,var_iend=var_map.end();
	compile.fprintf(gen, "section .data\n")

}

//struct fun_record { //函数声明记录
//fun_record();//默认构造函数
//fun_record(const fun_record& f);//拷贝构造函数
//void init(symbol dec_type,string dec_name);//初始化函数
//void addarg();//添加一个参数，默认使用tvar,同时修改pushlocalvar以防函数定义，要是声明就不管他的信息了
//int hasname(string id_name);//防止参数的名字在写入之前重复
//void pushlocalvar();//添加局部变量，默认使用tvar.
//int getCurAddr();//取得上一个进栈的变量的地址（相对于ebp）
//void flushargs();//将参数写到符号表
//void poplocalvars(int varnum);//弹出多个局部变量
//int equal(fun_record&f);
//var_record*create_tmpvar(symbol type,int hasVal,int &var_num);//根据常量添加一个临时变量，记得var_num++;
//~fun_record();
//};
