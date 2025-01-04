package factory

import (
	"fmt"
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

func (fn *ProgFunc) addArg(v *ProgDec) {
	fn.args = append(fn.args, v.kind)
	fn.addLocalVar(v)
}

func (fn *ProgFunc) addLocalVar(v *ProgDec) {
	//局部变量定义的缓冲机制，defined==1之前，是不写入符号表的，因为此时还不能确定是不是函数定义

	if fn.defined == 0 { //还是参数声明
		fn.localVars = append(fn.localVars, v)
	} else {
		fn.localVars = append(fn.localVars, v)
		// 局部变量的地址按照ebp-4*count的方式变化,修改
		v.localAddr = -4 * (len(fn.localVars) - len(fn.args))
		//代码中为局部变量开辟临时空间
		//genLocvar(0);
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

func (fn *ProgFunc) createTempVar(kind string, hasVal bool, vn *int) *ProgDec {
	// 创建临时变量记录
	temp := &ProgDec{kind: kind}
	switch kind {
	case "int":
		if hasVal {
			temp.intVal = num // 取值
		}
	case "char":
		temp.charVal = ' '
	case "string":
		if hasVal {
			temp.strValId = table.addstring()
		} else {
			temp.strValId = -1
		}
	}
	temp.name = genName("temp", kind, "")
	*vn += 1

	fn.localVars = append(fn.localVars, temp)
	table.addVar(temp)

	temp.localAddr = -4 * (len(fn.localVars) - len(fn.args))
	genLocvar(temp.strValId)
	return temp
}

type ProgTable struct {
	fnRecList   map[string]*ProgFunc // 变量声明列表
	varRecList  map[string]*ProgDec  // 函数声明列表
	stringTable []*string            // 串空间
	realArgList []*ProgDec           // 函数调用的实参列表，用于检查参数调用匹配和实参代码生成
}

func (t *ProgTable) addFn(f *ProgFunc) {
	if _, ok := t.fnRecList[f.name]; ok {
		// 记录异常
	}
	t.fnRecList[f.name] = f
	fmt.Printf("\t\t\t全局变量 <%s>(%s)\n", f.kind, f.name)
}

func (t *ProgTable) addVar(v *ProgDec) {
	if _, ok := t.varRecList[v.name]; ok {
		// 记录异常
	}
	t.varRecList[v.name] = v
	fmt.Printf("\t\t\t全局变量 <%s>(%s)\n", v.kind, v.name)
}

func (t *ProgTable) addRealArg(arg *ProgDec, vn *int) {
	if arg.kind == "string" {
		empty := ProgDec{
			kind: "string",
		}
		arg = genExp(ADD, &empty, arg, vn)
	}
	t.realArgList = append(t.realArgList, arg)
}

func (t *ProgTable) getVar(name string) *ProgDec {
	if v, ok := t.varRecList[name]; ok {
		return v
	}
	panic("函数被调用之前没有合法的声明。\\n")
}

func (t *ProgTable) genCall(name string, vn *int) *ProgDec {
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
			ret := t.realArgList[i]
			if ret.kind == "string" {
				_, _ = fmt.Fprintf(nil, "\tmov eax,[ebp%d]\n", ret.localAddr)
			} else {
				if ret.localAddr == 0 { // 全局的？
					_, _ = fmt.Fprintf(nil, "\tmov eax,[@var%s]\n", ret.name)
				} else { // 局部的？
					if ret.localAddr < 0 {
						_, _ = fmt.Fprintf(nil, "\tmov eax,[@ebp%d]\n", ret.localAddr)
					} else {
						_, _ = fmt.Fprintf(nil, "\tmov eax,[@ebp+%d]\n", ret.localAddr)
					}
				}
			}
			_, _ = fmt.Fprintf(nil, "\tpush eax\n")
		}
		// 调用代码
		_, _ = fmt.Fprintf(nil, "\tcall %s\n", name)
		_, _ = fmt.Fprintf(nil, "\tadd esp,%d\n", 4*l)

		// 产生函数返回代码
		// 非void函数在函数返回的时候将eax的数据放到临时变量中，为调用代码使用
		if fn.kind != "void" {
			// 创建临时变量
			//pRec=tfun.create_tmpvar(pfun->type,0,var_num);//创建临时变量
			v := &ProgDec{}
			_, _ = fmt.Fprintf(nil, "\tmov [ebp%d],eax\n", v.localAddr)
			if fn.kind == "string" { //返回的是临时string，必须拷贝
				empty := ProgDec{
					kind: "string",
				}
				v = genExp(ADD, &empty, v, vn)
			}
		}

		// 清除实际参数
		for ; m > 0; m-- {
			t.realArgList = t.realArgList[:len(t.realArgList)-1]
		}
	}
	panic("变量在使用之前没有合法的声明。\\n")
}

// 测试局部变量，参数的名字是否重复，主要应对变量的重复定义，全局变量和函数不需要调用他
func (t *ProgTable) exist(name string) bool {
	_, ok1 := t.fnRecList[name]
	_, ok2 := t.varRecList[name]
	return ok1 || ok2
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
