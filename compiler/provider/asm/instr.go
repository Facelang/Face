package asm

import (
	"io"
)

// 双操作码指令操作码表
var i2Opcode = []byte{
	//      8位操作数                 |      32位操作数
	// r,r  r,rm|rm,r r,im  r,r  r,rm|rm,r r,im // 操作码 目的操作数, 源操作数
	0x88, 0x8a, 0x88, 0xb0, 0x89, 0x8b, 0x89, 0xb8, //  寄存器到寄存器， 内存到寄存器， 寄存器到内存， 立即数到寄存器
	0x38, 0x3a, 0x38, 0x80, 0x39, 0x3b, 0x39, 0x81, // cmp
	0x28, 0x2a, 0x28, 0x80, 0x29, 0x2b, 0x29, 0x81, // sub
	0x00, 0x02, 0x00, 0x80, 0x01, 0x03, 0x01, 0x81, // add
	0x00, 0x00, 0x00, 0x00, 0x00, 0x8d, 0x00, 0x00, // lea
}

// 单操作码指令操作码表
var i1opcode = []uint16{
	0xe8, 0xcd /*0xfe,*/, 0xf7, 0xf7, 0xf7, 0x40, 0x48, 0xe9, // call,int,imul,idiv,neg,inc,dec,jmp<rel32>
	0x0f84, 0x0f8f, 0x0f8c, 0x0f8d, 0x0f8e, 0x0f85, 0x0f86, // je,jg,jl,jge,jle,jne,jna<rel32>
	// 0xeb,//jmp rel8
	// 0x74,0x7f,0x7c,0x7d,0x7e,0x75,0x76,//je,jg,jl,jge,jle,jne,jna<rel8>
	/*0x68,*/ 0x50, // push
	0x58,           // pop
}

// 零操作码指令操作码表
var i0Opcode = []byte{
	0xc3, // ret
}

// 重定位类型常量
const (
	R_386_32   = 1 // 绝对寻址
	R_386_PC32 = 2 // 相对寻址
)

type OperandType byte

const OPRTP_IMM OperandType = 1
const OPRTP_REG OperandType = 2
const OPRTP_MEM OperandType = 4 // 地址类型，需要寻址

type OperandModRm struct {
	Mod, RegOp, Rm byte
}

type OperandSIB struct {
	Scale, Index, Base byte
}

type OperandRecord struct {
	Type     OperandType  // 操作数类型(1 立即数， 2寄存器, 4寻址类型（表示会用到 ModRM 字段） )
	Value    int64        // 立即数？地址
	Length   int          // 操作数宽度
	ModRm    OperandModRm // 扩展寻址类型
	SIB      OperandSIB   // 扩展 基址+变址+偏移寻址类型
	RelLabel *LabelRecord // 符号引用， 如果为nil，则表示无需重定位
}

// InstrRecord 指令记录表
type InstrRecord struct {
	Prefix  byte             // 指令前缀
	Name    Token            // 指令名称
	OprLen  byte             // 操作数长度 0/1/2
	OprList []*OperandRecord // 操作数参数列表
}

func NewInstrRec(name Token, oprLen byte) *InstrRecord {
	return &InstrRecord{
		Prefix:  0,
		Name:    name,
		OprLen:  oprLen,
		OprList: make([]*OperandRecord, oprLen),
	}
}

// WriteOut 根据指令输出二进制编码
func (i *InstrRecord) WriteOut(w io.Writer) int {
	byteCount := 0

	// 如果有前缀，则写入前缀
	if i.Prefix != 0 {
		WriteBytes(w, int(i.Prefix), 1)
		byteCount++
	}

	// 根据操作数长度和指令类型，生成不同的机器码
	switch i.OprLen {
	case 0: // 零操作数指令，如 ret
		byteCount += Gen0op(i.Name, w)
	case 1: // 单操作数指令
		byteCount += Gen1op(i.Name, i.OprList[0], w)
	case 2: // 双操作数指令，如 mov, add, sub, cmp, lea
		// 第一个操作数是目的操作数， 第二个才是源！
		byteCount += Gen2op(i.Name, i.OprList[1], i.OprList[0], w)
	}

	return byteCount
}

// ProcessRel 处理可能的重定位信息
func ProcessRel(lb *LabelRecord, relType int) bool {
	if lb == nil {
		return false
	}
	if relType == R_386_32 { // 绝对重定位
		if lb.IsEqu { // 只要是地址符号就必须重定位，宏除外
			//ObjFile.AddRel(CurSeg, ProcessTable.CurSegOff, RelLb.LbName, relType)
			return true
		}
	} else if relType == R_386_PC32 { // 相对重定位
		if lb.Externed { // 对于跳转，内部的不需要重定位，外部的需要重定位
			//ObjFile.AddRel(CurSeg, ProcessTable.CurSegOff, RelLb.LbName, relType)
			return true
		}
	}

	return false
}

// GenXop 扩展操作数有关指令生成， 需要根据内存寻址操作
func GenXop(opcode byte, opr *OperandRecord, w io.Writer) int {
	WriteBytes(w, int(opcode), 1) // mod = 0 代表没有偏移
	WriteModRM(w, opr.ModRm)
	if opr.ModRm.Mod == 0 {
		if opr.ModRm.Rm == 5 { //[disp32]
			ProcessRel(opr.RelLabel, R_386_32) // 可能是mov eax,[@buffer],后边disp8和disp32不会出现类似情况
			WriteBytes(w, 0, 4)
			return 6
		} else if opr.ModRm.Rm == 4 { // SIB
			WriteSIB(w, opr.SIB)
			return 3
		}
	} else if opr.ModRm.Rm == 4 {
		WriteSIB(w, opr.SIB)
		if opr.Length > 0 {
			WriteBytes(w, int(opr.Value), opr.Length)
			return 3 + opr.Length
		}
	}

	if opr.Length == 0 {
		return 2
	}
	WriteBytes(w, int(opr.Value), opr.Length)
	return 2 + opr.Length
}

// Gen2op 生成双操作数指令
func Gen2op(op Token, src, dest *OperandRecord, w io.Writer) int {
	index := -1
	if src.Type == OPRTP_IMM { // 根据源操作数决定指令编码 鉴别操作数种类
		index = 3
	} else {
		index = int((dest.Type-2)*2 + src.Type - 2)
	}
	// length 为寄存器宽度， 这里固定为4（32位，数组向后偏移4位）
	// （int(op-I_MOV)*8 + (1-length%4)*4 + index）
	index = int(op-I_MOV)*8 + index + 4
	opcode := i2Opcode[index]

	if src.Type == OPRTP_REG && dest.Type == OPRTP_REG { // 0x89 双操作数都是寄存器， 寄存器编号保存在 Value
		// d 位（direction） 决定数流方向，通常在操作码的最后一位
		// d 位（direction，方向位）主要出现在双操作数、且操作数可以互换方向的指令中。
		// d=1：reg 字段是目的，rm 字段是源
		// d=0：reg 字段是源，rm 字段是目的
		// w 位（width，宽度位）用来指示操作数的位宽（8位/16位/32位）。
		// 几乎所有支持8位和16/32位操作的指令都有 w 位
		dest.ModRm.Mod = 0b11
		dest.ModRm.Rm = byte(dest.Value)   // 第一个操作数 目的操作数（在前）
		dest.ModRm.RegOp = byte(src.Value) // 第二个操作数 源操作数（reg 一般保存第一个操作数？？？）
		// mov 0x89 10001001 d=0 reg 是源 | w=1 32位？

		WriteBytes(w, int(opcode), 1)
		WriteModRM(w, dest.ModRm)
		return 2
	} else if src.Type == OPRTP_IMM && dest.Type == OPRTP_REG { // 1011w reg (0xB8+寄存器) 立即数到寄存器
		// 立即数到内存 或 立即数到寄存器， 使用 [操作码+寄存器编号 32位立即数] 表示
		opc, length := GetOpcodeForReg(op, opcode, dest.Value)
		WriteBytes(w, opc, length)
		// 可能的重定位位置 mov eax,@buffer,也有可能是mov eax,@buffer_len，就不许要重定位，因为是宏
		ProcessRel(src.RelLabel, R_386_32) // 这里记录一个重定位（如果有）
		WriteBytes(w, int(src.Value), 4)   // todo 长度为寄存器宽度 一定要按照长度输出立即数
		return length + 4
	} else if src.Type == OPRTP_IMM && dest.Type == OPRTP_REG { // 立即数到内存
		// todo 暂时没有实现
		return 0
	} else if src.Type == OPRTP_MEM { // 内存到寄存器
		src.ModRm.RegOp = byte(dest.Value)
		return GenXop(opcode, src, w)
	} else if dest.Type == OPRTP_MEM { // 寄存器到内存
		dest.ModRm.RegOp = byte(dest.Value)
		return GenXop(opcode, dest, w)
	} else {
		panic("语法格式错误！")
	}

	// mov 指令
	// 1. 寄存器和寄存器/内存之间传送 reg/mem ←→ reg
	//    格式：mov r/m, reg 或 mov reg, r/m
	//    操作码：100010dw（即 0x88/0x89/0x8A/0x8B）（10001000/10001001/10001010/10001011）
	//           d：方向位（1=reg为目的，0=reg为源）
	//           w：宽度位（0=8位，1=16/32位）
	// 2. 立即数到寄存器
	//    格式：mov reg, imm
	//    操作码：1011w reg（0xB0~0xBF）
	//           w：宽度位
	//           reg：寄存器编号(0xB0-0xB7 8位寄存器， 0xB8-0xBf 16/32位寄存器)
	// 3. 立即数到内存/寄存器
	//    格式：mov r/m, imm
	//    操作码：1100011w（0xC6/0xC7）
	// 4. 段寄存器于通用寄存器/内存之间
	//    格式：mov Sreg, r/m16 或 mov r/m16, Sreg
	//    操作码：10001110（0x8E），10001100（0x8C）
	//           mov Sreg, r/m16 | 0x8E | 段寄存器 ← 通用寄存器/内存 |
	//           mov r/m16, Sreg | 0x8C | 通用寄存器/内存 ← 段寄存器 |
	// 5. 内存与累加器之间
	//    格式：mov AL/AX/EAX, moffs 或 mov moffs, AL/AX/EAX
	//    操作码：1010000w（0xA0/0xA1），1010001w（0xA2/0xA3）
	//| 指令 | 操作码 | 说明 |
	//|------------------|----------|----------------------------|
	//| mov AL, moffs8 | 0xA0 | 内存到AL |
	//| mov AX/EAX, moffs| 0xA1 | 内存到AX/EAX |
	//| mov moffs8, AL | 0xA2 | AL到内存 |
	//| mov moffs, AX/EAX| 0xA3 | AX/EAX到内存 |
	// 6. 特殊寄存器
	//    例如：mov eax, cr0、mov cr3, eax
	//    操作码：0F 20 /r、0F 22 /r 等（仅限特权级）
	// 终结：“88/89/8A/8B寄存器互传，B0~BF立即到寄存器，C6/C7立即到内存，A0~A3累加器与内存。”

}

// Gen1op 生成单操作数指令
func Gen1op(op Token, opr *OperandRecord, w io.Writer) int {
	opcode := int(i1opcode[op-I_CALL]) // 取指令编码
	byteCount := 0

	if op == I_CALL || (op >= I_JMP && op <= I_JNA) {
		// 跳转或调用指令
		if op == I_CALL || op == I_JMP {
			WriteBytes(w, opcode, 1)
			byteCount += 1
		} else {
			WriteBytes(w, opcode>>8, 1)
			WriteBytes(w, opcode, 1)
			byteCount += 2
		}
		// 需要计算一下地址
		relAddr := int(opr.Value) - (ProcessTable.CurSegOff + 4)
		if ProcessRel(opr.RelLabel, R_386_PC32) {
			relAddr = -4
		}
		WriteBytes(w, relAddr, 4)
		byteCount += 4
	} else if op == I_INT { // int 只能8位?
		WriteBytes(w, opcode, 1)
		WriteBytes(w, int(opr.Value), 1)
		byteCount += 2
	} else if op == I_PUSH { // push eax, 将寄存器或者立即数压入栈中， 可以操作立即数，寄存器，内存
		if opr.Type == OPR_IMMD { // 立即数, 操作数+立即数，占4位
			opcode = 0x68
			WriteBytes(w, opcode, 1)
			WriteBytes(w, int(opr.Value), 4)
			byteCount += 5
		} else { // 寄存器操作数, 只占一位
			opcode += int(opr.ModRm.RegOp)
			WriteBytes(w, opcode, 1)
			byteCount++
		}
	} else if op == I_POP { // pop 指令， 从栈弹出 到指定寄存器
		opcode += int(opr.ModRm.RegOp)
		WriteBytes(w, opcode, 1)
		byteCount++
	} else if op == I_INC || op == I_DEC { // inc 和 dec 不能操作立即数
		if opr.Length == 1 { // 为什么需要判断长度？
			opcode = 0xfe
			WriteBytes(w, opcode, 1)
			exchar := 0xc0
			if op == I_DEC {
				exchar = 0xc8
			}
			exchar += int(opr.ModRm.RegOp)
			WriteBytes(w, exchar, 1)
			byteCount += 2
		} else {
			opcode += int(opr.ModRm.RegOp)
			WriteBytes(w, opcode, 1)
			byteCount++
		}
	} else if op == I_NEG { // 取负号 0-x
		if opr.Length == 1 {
			opcode = 0xf6
		}
		exchar := 0xd8
		exchar += int(opr.ModRm.RegOp)
		WriteBytes(w, opcode, 1)
		WriteBytes(w, exchar, 1)
		byteCount += 2
	} else if op == I_IDIV || op == I_IMUL {
		WriteBytes(w, opcode, 1)
		exchar := 0xf8
		if op == I_IMUL {
			exchar = 0xe8
		}
		exchar += int(opr.ModRm.RegOp)
		WriteBytes(w, exchar, 1)
		byteCount += 1
	}
	return byteCount
}

func Gen0op(opt Token, w io.Writer) int {
	if opt != I_RET {
		return 0
	}
	WriteBytes(w, int(i0Opcode[0]), 1)
	return 1
}

func GetOpcodeForReg(op Token, opcode byte, reg int64) (int, int) {
	switch op {
	case I_MOV: // b0+rb MOV r/m8,imm8 b8+rd MOV r/m32,imm32
		return int(opcode) + int(reg), 1
	case I_CMP: // 80 /7 ib CMP r/m8,imm8 81 /7 id CMP r/m32,imm32
		return int(opcode)<<8&0xf8 + int(reg), 2
	case I_ADD: // 80 /0 ib ADD r/m8, imm8 81 /0 id ADD r/m32, imm32
		return int(opcode)<<8&0xc0 + int(reg), 2
	case I_SUB: // 80 /5 ib SUB r/m8, imm8 81 /5 id SUB r/m32, imm32
		return int(opcode)<<8&0xe8 + int(reg), 2
	default:
		panic("unhandled default case")
	}
}

// WriteModRM 输出ModRM字节
func WriteModRM(w io.Writer, o OperandModRm) {
	mrm := ((o.Mod & 0x00000003) << 6) + ((o.RegOp & 0x0000007) << 3) + (o.Rm & 0x00000007)
	WriteBytes(w, int(mrm), 1)
}

// WriteSIB 输出SIB字节
func WriteSIB(w io.Writer, o OperandSIB) {
	s := ((o.Scale & 0x00000003) << 6) + ((o.Index & 0x0000007) << 3) + (o.Base & 0x00000007)
	WriteBytes(w, int(s), 1)
}
