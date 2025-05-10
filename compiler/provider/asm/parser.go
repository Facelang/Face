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
	fileName  string          // 解析文件
	tempFile  *os.File        // 临时输出文件
	err       error           // 解析异常
	lexer     *lexer          // 读取器
	progTable *TemporaryTable // 缓存解析表
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
		fileName:  file,
		tempFile:  temp,
		err:       err,
		lexer:     lex,
		progTable: NewTemporaryTable(),
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

func (p *parser) reg(token Token) byte { // 返回寄存器宽度
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
		p.inst(token) // 进入语法处理
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
	case A_EQU: // equ 常量？伪指令，所有使用到该符号的，全部替换为值，不存在地址。
		// 这里需要被替换为值
		// 关于 equ 语法说明，equ 支持表达式：可以是数字、地址、其他符号、算术表达式等
		// equ 定义的符号在汇编时就被替换为具体值，不会占用内存，也不会生成机器码。
		// 不能对 equ 定义的符号赋新值（它不是变量）。
		// equ 只能用于常量表达式，不能用于运行时可变的值。
		// todo 完整的逻辑需要支持 数字，其他符号，表达式， 【最终获得运算后的值】
		// todo equ 引用其它符号，必须提前申明！
		p.require(NUMBER) // todo 当前只支持数字
		ProcessTable.AddLabel(NewRecWithAddr(id, p.number()))
	case COLON: // 代码段（label）, main: 一般是函数名作为一个单独的记号
		ProcessTable.AddLabel(NewRec(id, false))
		return
	default: // 变量支持
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
	cont := make([]int, 255)         // 数据缓存
	contLen := 0                     // 这里计算数据占用宽度
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
		for _, ch := range []byte(p.lexer.id) {
			(*cont)[*contLen] = int(ch)
			*contLen++
		}
	case IDENT: // 如果是一个变量， 说明是一个引用，需要添加重定位，
		lb := ProcessTable.GetLabel(p.id())
		(*cont)[*contLen] = lb.Addr
		ProcessTable.AddRelOff(*contLen*size, p.id(), R_386_32)
		*contLen++
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

// inst 处理指令
func (p *parser) inst(token Token) {
	if token >= I_MOV && token <= I_LEA { // i_mov,i_cmp,i_sub,i_add,i_lea,//2p
		instr := NewInstrRec(token, 2)
		instr.OprList[0] = p.opr() // 取第一个操作数
		p.require(COMMA)
		instr.OprList[1] = p.opr() // 取第二个操作数
		ProcessTable.PushInstr(instr)

		// 生成指令
		//Gen2Op(token, type1, type2, length)
	} else if token >= I_CALL && token <= I_POP { // i_call,i_int,i_imul,i_idiv,i_neg,i_inc,i_dec,i_jmp,i_je,i_jg,i_jl,i_jge,i_jle,i_jne,i_jna,i_push,i_pop,//1p
		instr := NewInstrRec(token, 1)
		instr.OprList[0] = p.opr()
		ProcessTable.PushInstr(instr)
		//Gen1op(token, kind, size)
	} else if token == I_RET {
		instr := NewInstrRec(token, 0)
		ProcessTable.PushInstr(instr)
		//Gen0op(token)
	} else {
		fmt.Printf("opcode err[line:%d]\n", p.lexer.line)
	}
}

// Operand 处理操作数
func (p *parser) opr() *OperandRecord {
	opr := &OperandRecord{}
	token := p.NextToken()
	switch token {
	case NUMBER:
		opr.Type = OPRTP_IMM
		opr.Value = int64(p.number())
		opr.Length = 4 // 代表 4*8
	case IDENT: // 变量名 立即数
		opr.Type = OPRTP_IMM
		opr.RelLabel = ProcessTable.GetLabel(p.id())
	case LBRACK: // 内存寻址
		opr.Type = OPRTP_MEM
		p.addr(opr)
		p.require(RBRACK)
	case SUB: // 负立即数
		p.require(NUMBER)
		opr.Type = OPRTP_IMM
		opr.Value = int64(-p.number())
		opr.Length = 4 // 代表 4*8
	default: // 寄存器操作数 todo 双寄存器需要特殊处理
		regLen := p.reg(token)
		opr.Type = OPRTP_REG
		opr.Value = int64(token-BR_AL) - int64((1-regLen%4)*8) // 这里保持寄存器编号
		opr.Length = int(regLen)
	}
	return opr
}

func (p *parser) addr(opr *OperandRecord) { // [立即数， 变量， 寄存器] 间接寻址
	token := p.NextToken()
	switch token {
	case NUMBER: // 直接寻址 特例 [00 xxx 101 disp32]
		opr.Value = int64(p.number())
		opr.Length = 4
		opr.ModRm.Mod = 0
		opr.ModRm.Rm = 0b101
	case IDENT: // 直接寻址 特例 [变量]
		opr.ModRm.Mod = 0
		opr.ModRm.Rm = 0b101
		opr.RelLabel = ProcessTable.GetLabel(p.id())
	default: // 寄存器寻址 [eax, edi]
		p.regaddr(opr, token)
	}
}

// todo 偏移寻址会不会出现符号引用？变量？
func (p *parser) regaddr(opr *OperandRecord, basereg Token) { // 可能存在基于寄存器 + 偏移
	token := p.NextToken()
	if token == ADD || token == SUB { // 有变址寄存器
		p.regaddrtail(opr, basereg, token)
	} else { // 寄存器间址 00 xxx rrr <esp ebp特殊考虑>
		if basereg == DR_ESP { //[esp] 特例 需要引导 sib 表示
			opr.ModRm.Mod = 0
			opr.ModRm.Rm = 0b100
			opr.SIB.Scale = 0     // 随便啦
			opr.SIB.Index = 0b100 //(esp) 代表不存在变址
			opr.SIB.Base = 0b100  // 代表esp
		} else if basereg == DR_EBP { //[ebp],特例 改写为 [ebp+0]
			opr.ModRm.Mod = 1    // 1字节偏移
			opr.ModRm.Rm = 0b101 //
			opr.Value = 0
			opr.Length = 1 // 1位偏移
		} else { // 一般寄存器
			opr.ModRm.Mod = 0
			opr.ModRm.Rm = byte(int(basereg-BR_AL) % 8)
		}
		p.lexer.Back(token)
	}
}

func (p *parser) regaddrtail(opr *OperandRecord, basereg Token, sign Token) { // 基址 + 偏移[立即数/另外一个寄存器]
	token := p.NextToken()
	switch token {
	case NUMBER: // 寄存器基址 + 偏移 disp8/disp32， mod = 01/10
		num := p.number()
		if sign == SUB {
			num = -num
		}

		if num >= -128 && num <= 127 { // 8位
			opr.Length = 1
			opr.ModRm.Mod = 0b01
		} else { // 32位偏移
			opr.Length = 4
			opr.ModRm.Mod = 0b10
		}
		opr.Value = int64(num)
		opr.ModRm.Rm = byte(int(basereg-BR_AL) % 8)

		if opr.ModRm.Rm == 4 { // 0b100(esp) 引导SIB
			opr.SIB.Scale = 0     // 随便啦
			opr.SIB.Index = 0b100 //(esp) 代表不存在变址
			opr.SIB.Base = 0b100  // 代表esp
		}
	default: // 基址变址寻址 [base+index*2^scale+disp],不会发生在esp和ebp上，没有生成这样的指令
		regLen := p.reg(token) // kind和type都是寄存器宽度 1/4
		opr.ModRm.Mod = 0      // 不继续解析 disp 偏移
		opr.ModRm.Rm = 4       // 0b100
		opr.SIB.Scale = 0      // 不继续解析
		opr.SIB.Index = byte(token-BR_AL) - (1-regLen%4)*8
		opr.SIB.Base = byte(basereg-BR_AL) - (1-regLen%4)*8
	}
}

/**
 * 寻址模式归纳
 * 000   001   010   011   100        101        110   111
 * eax   ecx   edx   ebx   esp[sib]   ebp[rip]   esi   edi
 * mod = 00 寄存器间接寻址          [eax]
 *       01 寄存器 + 8位偏移        [eax+4]
 *       10 寄存器 + 32位偏移       [eax+0x123456]
 * mod = 11 寄存器操作数            eax
 *
 * 特例1：mod=00, r/m=101(ebp) 代表直接寻址 [0x12345678], ebp 使用 [ebp+0] 表示
 *             mov ecx, [0x12345678] // 没有寄存器
 *             mov eax, [ebp+0] // 替代仅[ebp]表达式
 * 特例2：mod!=11, r/m=100(esp) 代表 基址（寄存器）+变址（寄存器）+偏移（立即数） 【引导SIB】
 *             mod=00 代表没有偏移， mod=01代表有8位偏移, mod=10代表有32位偏移
 *             [base+index*2^scale+disp]
 *      不存在变址时：index=100(esp) 指令集规定 esp 不能作为变址寄存器！
 *                 [base+esp*2^scale+disp] 属于不合法指令
 *      不存在基址时：mod=00,base=101(ebp),且强制包含32位偏移
 *                 [ebp+index*2^scale] 表达式需要替换为 [ebp+index*2^scale+disp8]?
 *      [esp+disp] 表达式替换 mod!=11, r/m=100, base=100,index=100(不存在变址),scale=0 => [esp+不存在的变址+disp]
 *
 *
 *
 *
 */
