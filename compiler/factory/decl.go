package factory

import "fmt"

type ProgTable struct {
	fnRecList  map[string]*FnRecord
	varRecList map[string]*VarRecord
}

func (t *ProgTable) addFn(f *FnRecord) {
	if _, ok := t.fnRecList[f.name]; ok {
		// 记录异常
	}
	t.fnRecList[f.name] = f
	fmt.Printf("\t\t\t全局变量 <%s>(%s)\n", f.kind, f.name)
}

func (t *ProgTable) addVar(v *VarRecord) {
	if _, ok := t.varRecList[v.name]; ok {
		// 记录异常
	}
	t.varRecList[v.name] = v
	fmt.Printf("\t\t\t全局变量 <%s>(%s)\n", v.kind, v.name)
}

type FnRecord struct {
	kind      string       // 类型
	name      string       // 名称
	args      []string     // 参数类型列表
	localVars []*VarRecord // 局部变量列表,指向哈希表,仅仅为函数定义服务
	defined   int          // 函数是否给出定义
	flushed   int          // 函数参数是否已经缓冲写入，标记是否再清楚的时候清除缓冲区
	hadret    int          // 记录是否含有返回语句
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

// VarRecord 变量声明记录
type VarRecord struct {
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
