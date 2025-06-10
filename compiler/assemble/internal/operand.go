package internal

// 操作数类型
const (
	REGISTER  = iota // 寄存器
	IMMEDIATE        // 立即数
	MEMORY          // 内存引用
	SYMBOL          // 符号/标签
)

type ExpType byte

const EXP_ADD ExpType = 1
const EXP_SUB ExpType = 1
const EXP_ADD ExpType = 1
const EXP_ADD ExpType = 1

type OprType byte

const OPRTP_IMM OprType = 1
const OPRTP_REG OprType = 2
const OPRTP_MEM OprType = 3 // 地址类型，需要寻址
const OPRTP_REL OprType = 4 // 符号类型，需要重定位

type Operand interface {
	operand()
}

type Express interface {
}

// operand 表示一个操作数
type operand struct {
	Type  int    // 操作数类型
	Value string // 操作数值
	Base  string // 基址寄存器(用于内存引用)
	Index string // 变址寄存器(用于内存引用)
	Scale int    // 比例因子(用于内存引用)
}

type ExpOpr struct {
	ExpList []Express
}

type GenOpr struct {
	Type   OprType // 操作数类型(1 立即数， 2寄存器, 4寻址类型（表示会用到 ModRM 字段） )
	Value  int64   // 立即数？地址
	Length int     // 操作数宽度
}

type RelOpr struct {
	Label string // 符号名称
}
