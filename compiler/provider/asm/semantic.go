package asm

import (
	"os"
)

// modrm字段
type ModRM struct {
	mod int // 0-1
	reg int // 2-4
	rm  int // 5-7
}

func (s *ModRM) init() {
	s.mod = -1
	s.reg = 0
	s.rm = 0
}

// sib字段
type SIB struct {
	scale int // 0-1
	index int // 2-4
	base  int // 5-7
}

func (s *SIB) init() {
	s.scale = -1
	s.index = 0
	s.base = 0
}

// 指令的其他部分
type Inst struct {
	OpCode  byte
	Disp    int
	DispLen int //偏移的长度
	Imm32   int
	modrm   *ModRM
	sib     *SIB
}

func NewInst() *Inst {
	i := new(Inst)
	i.init()
	return i
}

func (i *Inst) init() {
	i.OpCode = 0
	i.Disp = 0
	i.DispLen = 0
	i.Imm32 = 0

	i.modrm = new(ModRM)
	i.modrm.init()
	i.sib = new(SIB)
	i.sib.init()
}

// 设置disp，自动检测disp长度（符号），及时是无符号地址值也无妨
func (i *Inst) setDisp(d, dLen int) {
	i.Disp = d
	i.DispLen = dLen
}

// 按照记录的disp长度输出
func (i *Inst) writeDisp(file *os.File) {
	if i.DispLen == 0 {
		return
	}
	//WriteBytes(file, i.Disp, i.DispLen)
	i.DispLen = 0 // 还原
}
