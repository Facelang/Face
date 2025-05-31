package internal

type ProgType byte

const (
	Unknown ProgType = iota
	Instr            // 指令
	Label            // 符号定义
	Section          // 段标记
	Global           // 全局符号
	Local            // 本地符号
	Type             // .type 指定类型
	Size             // .size 指定大小
)

type Program struct {
	Type ProgType
	Name string
	Pc   int64
}
