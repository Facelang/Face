package asm

import "os"

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
	Type     OperandType  // 操作数类型(1 立即数， 2寄存器, 寻址[8直接寻址mod=00,rm=101, 4 寄存器, , 8 寄存器+偏移, 16 基址+变址+偏移] )
	Value    int64        // 立即数？地址
	Length   int          // 操作数宽度
	ModRm    OperandModRm // 扩展寻址类型
	SIB      OperandSIB   // 扩展 基址+变址+偏移寻址类型
	RelLabel *LabelRecord // 符号引用
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

// 设置disp，自动检测disp长度（符号），及时是无符号地址值也无妨
//func (i *Inst) setDisp(d, dLen int) {
//	i.Disp = d
//	i.DispLen = dLen
//}
//
//// 按照记录的disp长度输出
//func (i *Inst) writeDisp() {
//	if i.DispLen == 0 {
//		return
//	}
//	WriteBytes(i.Disp, i.DispLen)
//	i.DispLen = 0 // 还原
//}

func GetCodeLen(token Token) int {
	return 1
}

// 获取指令模式
//
//	单指令：立即数, 寄存器, 内存
//	双指令：立即数到寄存器, 立即数到内存, 内存到寄存器, 寄存器到内存, 寄存器到寄存器
func (i *InstrRecord) OpType() int {

}

// CodeLen 计算指令输出占用长度
func (i *InstrRecord) CodeLen() int {
	l := GetCodeLen(i.Name)
	for _, opr := range i.OprList {
		switch opr.Type {
		case OPRTP_IMM:
			l += opr.Length
		case OPRTP_REG:
			// 直接使用寄存器表示
		case OPRTP_MEM:
			if opr.ModRm.Mod != 0b11 && opr.ModRm.Rm == 0b100 { // 需要SIB
				l += 2 + opr.Length
			} else { // 只使用 ModRM 字段
				l += 1 + opr.Length
			}
		}
	}
}

// WriteOut 根据指令输出二进制编码
func (i *InstrRecord) WriteOut(file *os.File) {
	//switch i.ArgLen {
	//case 0:
	//	Gen0op(0)
	//case 1:
	//
	//case 2:
	//
	//}

}

func Gen0op(opt Token) {
	if opt != I_RET {
		return
	}
	WriteBytes(int(i0Opcode[0]), 1)
}
