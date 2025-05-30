package elf

// Section 段信息
type Section struct {
	Name   string // 名称
	Offset int    // 偏移
	Length int    // 内容大小
}
