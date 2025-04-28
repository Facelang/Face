package asm

import (
	"fmt"
	"os"
	"strconv"
)

type ParserFactory interface {
	Parse() (interface{}, error)
	NextToken() Token
}

type parser struct {
	fileName string   // 解析文件
	tempFile *os.File // 临时输出文件
	err      error    // 解析异常
	//gen       *codegen // 代码生成器
	lexer *lexer // 读取器
	//progFn    *ProgFunc
	//progTable *ProgTable
}

func NewFileParser(file string) ParserFactory {
	errFunc := func(file string, line, col, off int, msg string) {
		return
	}

	lex := NewLexer()
	err := lex.init(file, errFunc)
	if err != nil {
		return &parser{err: err}
	}
	temp, err := os.Create(file + ".t")
	if err != nil {
		return &parser{err: err}
	}

	p := &parser{
		fileName: file,
		tempFile: temp,
		err:      err,
		lexer:    lex,
	}

	return p
}

func (p *parser) id() string {
	return p.lexer.id
}

func (p *parser) number() int {
	v, _ := strconv.Atoi(p.id())
	return v
}

func (p *parser) reg() int {
	token := p.NextToken()
	if token >= BR_AL && token <= BR_BH { // 8位寄存器
		return 1
	}
	if token >= DR_EAX && token <= DR_EDI { // 32位寄存器
		return 4
	}
	p.err = fmt.Errorf("[reg](%d,%d): %s, %s，需要一个寄存器！",
		p.lexer.line, p.lexer.col, token.String(), p.lexer.id)
	panic(p.err)
}

func (p *parser) require(next Token) string {
	token := p.lexer.NextToken()
	if token != next {
		p.err = fmt.Errorf("[%d,%d]: %s, %s，需要一个 %s 类型！",
			p.lexer.line, p.lexer.col, token.String(), p.lexer.id, next)
		panic(p.err)
	}
	return p.lexer.id
}

func (p *parser) NextToken() Token {
	token := p.lexer.NextToken()
	for token == COMMENT {
		token = p.lexer.NextToken()
	}
	return token
}

func (p *parser) Parse() (interface{}, error) {
	return p.program()
}

// .：语法结束符，表示规则的终结（EOF）。
// {} 表示 零个或多个 顶层声明，因此顶层声明也是可选的。
// program = decl { program } .
func (p *parser) program() (interface{}, error) {
	if p.err != nil {
		return nil, p.err
	}
	token := p.NextToken()
	if token == EOF {
		ProcessTable.SwitchSeg(p.id())
		ProcessTable.Exports()
		return nil, nil
	}

	switch token {
	case IDENT: // 两种情况，段落定义，变量定义
		p.lbtail(p.id())
	case A_SEC: // 1. 先读到的，代码段
		p.require(IDENT)
		ProcessTable.SwitchSeg(p.id())
	case A_GLB: // todo 暂时没有被处理
		// 定义全局入口符号，这里默认是_start,链接器默认也是lit，这里不做具体处理
		p.require(IDENT)
		break
	default:
		// 进入语法处理
		p.lexer.Back(token)
		p.inst()
	}
	return p.program()
}

// 段落定义，变量定义;
// 如果是数据定义：变量名称 [ times 次数 ] 类型 值
// 最后都是添加符号，函数、变量（和值）
func (p *parser) lbtail(id string) {
	token := p.NextToken()
	switch token {
	case A_TIMES: // 需要重复
		p.require(NUMBER)
		times := p.number()
		p.values(id, times, p.size()) // size 代表数据类型 db dd 字符、字、双字
		return
		//case A_EQU: // [暂时忽略] equ 常量？伪指令，所有使用到该符号的，全部替换为值，不存在地址。
		//	p.require(NUMBER)
		//	ProcessTable.AddLabel(NewRecWithAddr(id, p.number()))
		return
	case COLON: // 代码段（label）, main: 一般是函数名作为一个单独的记号
		ProcessTable.AddLabel(NewRec(id, false))
		return
	default:
		p.lexer.Back(token)
		p.values(id, 1, p.size()) // 单个变量定义，直接解析
	}
}

func (p *parser) size() int {
	token := p.NextToken()
	switch token {
	case A_DB:
		return 1
	case A_DW:
		return 2
	case A_DD:
		return 4
	default:
		p.err = fmt.Errorf("[size](%d,%d): %s, %s，需要获取数据占用空间！",
			p.lexer.line, p.lexer.col, token.String(), p.lexer.id)
		panic(p.err)
	}
}

func (p *parser) values(id string, times, size int) {
	cont := make([]int, 255)
	contLen := 0
	p.valType(&cont, &contLen, size) // 这里获得的是值， 数字、字符串、引用名
	p.valTail(&cont, &contLen, size) // 判断逗号 多个值，依次存入 cont 中
	ProcessTable.AddLabel(NewRecWithData(id, times, size, cont, contLen))
}

// ValType 处理数据类型
func (p *parser) valType(cont *[]int, contLen *int, size int) {
	token := p.NextToken()
	switch token {
	case NUMBER:
		(*cont)[*contLen] = p.number()
		*contLen++
	case STRING:
		for _, ch := range p.lexer.id {
			(*cont)[*contLen] = int(ch)
			*contLen++
		}
	//case IDENT: // [暂时忽略] 如果是一个变量， 说明是一个引用，需要添加重定位，
	//	lb := ProcessTable.GetLabel(p.id())
	//	(*cont)[*contLen] = lb.Addr
	//	ObjFile.addRel(
	//		ProcessTable.CurSegName,
	//		ProcessTable.CurSegOff+*contLen*size, p.id(), R_386_32)
	//	*contLen++
	default:
		p.err = fmt.Errorf("[valType](%d,%d): %s, %s，数据类型获取异常！",
			p.lexer.line, p.lexer.col, token.String(), p.lexer.id)
		panic(p.err)
	}
}

// ValTail 处理数据值后续部分
func (p *parser) valTail(cont *[]int, contLen *int, size int) {
	token := p.NextToken()
	if token == COMMA {
		p.valType(cont, contLen, size)
		p.valTail(cont, contLen, size)
	} else {
		p.lexer.Back(token)
	}
}

func (p *parser) mem() {
	p.require(LBRACK)
	p.addr()
	p.require(RBRACK)
}

func (p *parser) addr() {
	token := p.NextToken()
	switch token {
	case NUMBER: // 直接寻址 [00 xxx 101 disp32]
		instr.modrm.mod = 0
		instr.modrm.rm = 5
		instr.setDisp(p.number(), 4)
	case IDENT: // 直接寻址 [变量]
		instr.modrm.mod = 0
		instr.modrm.rm = 5
		lr := ProgTable.get(p.id())
		instr.setDisp(lr.Addr, 4)
		if ScanLop == 2 { // 第二次扫描记录重定位项
			if !lr.IsEqu { // 不是equ
				// 记录符号
				RelLb = lr
			}
		}
	default: // 寄存器寻址 [eax, edi]
		p.lexer.Back(token)
		p.regaddr(token, p.reg())
	}
}

func (p *parser) regaddr(basereg Token, kind int) {
	token := p.NextToken()
	if token == ADD || token == SUB { // 有变址寄存器
		p.lexer.Back(token)
		p.off()
		p.regaddrtail(basereg, kind, token)
	} else { // 寄存器间址 00 xxx rrr <esp ebp特殊考虑>
		if basereg == DR_ESP { //[esp]
			instr.modrm.mod = 0
			instr.modrm.rm = 4 // 引导SIB
			instr.sib.scale = 0
			instr.sib.index = 4
			instr.sib.base = 4
		} else if basereg == DR_EBP { //[ebp],生成汇编代码中未出现
			instr.modrm.mod = 1 // 8-bit 0 disp，或者mod=2 32-bit 0 disp
			instr.modrm.rm = 5
			instr.setDisp(0, 1)
		} else { // 一般寄存器
			instr.modrm.mod = 0
			instr.modrm.rm = int(basereg-BR_AL) - (1-kind%4)*8
		}
		p.lexer.Back(token)
	}
}

func (p *parser) off() {
	token := p.NextToken()
	if token == ADD || token == SUB {
	} else {
		msg, _ := fmt.Printf("addr err![line:%d]\n", p.lexer.line)
		panic(msg)
	}
}

func (p *parser) regaddrtail(basereg Token, kind int, sign Token) {
	token := p.NextToken()
	switch token {
	case NUMBER: // 寄存器基址寻址 01/10 xxx rrr disp8/disp32
		num := p.number()
		if sign == SUB {
			num = -num
		}
		if num >= -128 && num <= 127 {
			instr.modrm.mod = 1
			instr.setDisp(num, 1)
		} else {
			instr.modrm.mod = 2
			instr.setDisp(num, 4)
		}
		instr.modrm.rm = int(basereg-BR_AL) - (1-kind%4)*8

		if basereg == DR_ESP { // sib
			instr.modrm.rm = 4 // 引导SIB
			instr.sib.scale = 0
			instr.sib.index = 4
			instr.sib.base = 4
		}
	default: // 基址变址寻址 00 xxx 100 00=scale rrr2=index rrr1=base,不会发生在esp和ebp上，没有生成这样的指令
		p.lexer.Back(token)
		typei := p.reg()
		instr.modrm.mod = 0
		instr.modrm.rm = 4
		instr.sib.scale = 0
		instr.sib.index = int(token-BR_AL) - (1-typei%4)*8
		instr.sib.base = int(basereg-BR_AL) - (1-kind%4)*8
	}
}

// inst 处理指令
func (p *parser) inst() {
	instr.init()
	token := p.NextToken()
	if token >= I_MOV && token <= I_LEA { // i_mov,i_cmp,i_sub,i_add,i_lea,//2p
		// 双操作数指令
		// 读取第一个操作数
		regNum1 := 0
		type1 := 0
		len1 := 0
		p.opr(&regNum1, &type1, &len1) // regNum = 1, s_type=0, len = 4
		p.require(COMMA)
		// 读取第二个操作数
		regNum2 := 0
		type2 := 0
		len2 := 0
		p.opr(&regNum2, &type2, &len2) // s_type = 地址类型：寄存器，立即数，内存； len = 寄存器宽度

		// 生成指令
		//common.Gen2Op(token, type1, type2, length)
	} else if token >= I_CALL && token <= I_POP { // i_call,i_int,i_imul,i_idiv,i_neg,i_inc,i_dec,i_jmp,i_je,i_jg,i_jl,i_jge,i_jle,i_jne,i_jna,i_push,i_pop,//1p
		regNum, kind, size := 0, 0, 0
		p.opr(&regNum, &kind, &size)
		Gen1op(token, kind, size)
	} else if token == I_RET {
		Gen0op(token)
	} else {
		fmt.Printf("opcode err[line:%d]\n", p.lexer.line)
	}
}

var instr = NewInst()

// Operand 处理操作数
func (p *parser) opr(regNum *int, opType *int, opLen *int) {
	token := p.NextToken()
	switch token {
	case NUMBER:
		*opType = OPR_IMMD
		instr.Imm32 = p.number()
	case IDENT: // 变量名 立即数
		*opType = OPR_IMMD
		lr := ProcessTable.GetLabel(p.id()) // 从 lb_map 中取一条记录，如果没有，新插入一条（外部定义）
		instr.Imm32 = lr.Addr
		if ScanLop == 2 {
			if !lr.IsEqu { // 不是equ
				// 记录符号
				RelLb = lr
			}
		}
	case LBRACK: // 内存寻址
		*opType = OPR_MEMR
		p.lexer.Back(token)
		p.mem()
	case SUB: // 负立即数
		*opType = OPR_IMMD
		p.require(NUMBER)
		instr.Imm32 = -p.number()
	default: // 寄存器操作数 11 rm=des reg=src
		*opType = OPR_REGS
		p.lexer.Back(token)
		*opLen = p.size() // 寄存器宽度，根据 token, 寄存器名字判断
		if *regNum != 0 { // 双reg，将原来reg写入rm作为目的操作数，本次写入reg
			instr.modrm.mod = 3                                 // 双寄存器模式
			instr.modrm.rm = instr.modrm.reg                    // 因为统一采用opcode rm,r 的指令格式，比如mov rm32,r32就使用0x89,若是使用opcode r,rm 形式则不需要
			instr.modrm.reg = int(token-BR_AL) - (1-*opLen%4)*8 // 计算寄存器的编码
		} else { // 第一次出现reg，临时在reg中，若双reg这次是目的寄存器，需要交换位置
			instr.modrm.reg = int(token-BR_AL) - (1-*opLen%4)*8 // 计算寄存器的编码
		}
		*regNum++
	}
}
