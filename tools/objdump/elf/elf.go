package elf

import (
	"debug/elf"
	"fmt"

	"github.com/marshal/objdump/x86"
)

// ELFFile 结构体表示解析后的ELF文件
type ELFFile struct {
	Path    string
	File    *elf.File
	Class   int    // 32位或64位
	Machine string // 架构类型
}

// ParseELF 解析指定路径的ELF文件
func ParseELF(filePath string) (*ELFFile, error) {
	file, err := elf.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开ELF文件: %v", err)
	}

	elfFile := &ELFFile{
		Path:  filePath,
		File:  file,
		Class: int(file.Class),
	}

	// 设置架构类型
	switch file.Machine {
	case elf.EM_X86_64:
		elfFile.Machine = "x86_64"
	case elf.EM_386:
		elfFile.Machine = "x86"
	default:
		elfFile.Machine = fmt.Sprintf("未知架构(%d)", file.Machine)
	}

	return elfFile, nil
}

// PrintELFHeader 打印ELF头信息
func (e *ELFFile) PrintELFHeader() {
	hdr := e.File.FileHeader
	fmt.Printf("  入口点地址: 0x%x\n", hdr.Entry)
	fmt.Printf("  开始于: 0x%x\n", 0)

	// 获取节的数量
	fmt.Printf("  节表项数量: %d\n", len(e.File.Sections))

	// 打印文件类型
	fmt.Printf("  类型: ")
	switch hdr.Type {
	case elf.ET_NONE:
		fmt.Println("NONE (未知类型)")
	case elf.ET_REL:
		fmt.Println("REL (可重定位文件)")
	case elf.ET_EXEC:
		fmt.Println("EXEC (可执行文件)")
	case elf.ET_DYN:
		fmt.Println("DYN (共享对象文件)")
	case elf.ET_CORE:
		fmt.Println("CORE (核心转储文件)")
	default:
		fmt.Printf("未知类型(0x%x)\n", hdr.Type)
	}
}

// PrintSymbolTable 打印符号表
func (e *ELFFile) PrintSymbolTable() {
	symbols, err := e.File.Symbols()
	if err != nil {
		fmt.Printf("获取符号表失败: %v\n", err)
		return
	}

	if len(symbols) == 0 {
		fmt.Println("没有符号表信息")
		return
	}

	fmt.Println("  Num    Value        Size Type    Bind   Vis      Ndx Name")
	for i, sym := range symbols {
		fmt.Printf("  %-6d 0x%-10x %5d %-7s %-6s %-8s %3s %s\n",
			i,
			sym.Value,
			sym.Size,
			symbolType(sym.Info),
			symbolBind(sym.Info),
			symbolVisibility(sym.Other),
			sectionIndexToString(sym.Section),
			sym.Name)
	}
}

// DisassembleTextSection 反汇编.text段
func (e *ELFFile) DisassembleTextSection(disasm *x86.Disassembler) {
	textSection := e.File.Section(".text")
	if textSection == nil {
		fmt.Println("未找到.text段")
		return
	}

	// 获取文本段数据
	data, err := textSection.Data()
	if err != nil {
		fmt.Printf("读取.text段数据失败: %v\n", err)
		return
	}

	// 获取文本段地址
	baseAddr := textSection.Addr

	// 反汇编代码
	instructions, err := disasm.Disassemble(data, baseAddr)
	if err != nil {
		fmt.Printf("反汇编失败: %v\n", err)
		return
	}

	// 输出反汇编结果
	for _, inst := range instructions {
		fmt.Printf("  %08x: %-20s\t%s\n", inst.Address, inst.Bytes, inst.Text)
	}
}

// PrintAllSections 打印所有段的信息
func (e *ELFFile) PrintAllSections() {
	fmt.Println("段列表:")
	fmt.Println("  Idx Name           Size     VMA              Type")
	for i, section := range e.File.Sections {
		fmt.Printf("  %3d %-15s %08x  %016x  %s\n",
			i,
			section.Name,
			section.Size,
			section.Addr,
			sectionType(section.Type))
	}
}

// 辅助函数：获取符号类型字符串
func symbolType(info uint8) string {
	switch elf.ST_TYPE(info) {
	case elf.STT_NOTYPE:
		return "NOTYPE"
	case elf.STT_OBJECT:
		return "OBJECT"
	case elf.STT_FUNC:
		return "FUNC"
	case elf.STT_SECTION:
		return "SECTION"
	case elf.STT_FILE:
		return "FILE"
	case elf.STT_COMMON:
		return "COMMON"
	case elf.STT_TLS:
		return "TLS"
	default:
		return fmt.Sprintf("未知(%d)", elf.ST_TYPE(info))
	}
}

// 辅助函数：获取符号绑定字符串
func symbolBind(info uint8) string {
	switch elf.ST_BIND(info) {
	case elf.STB_LOCAL:
		return "LOCAL"
	case elf.STB_GLOBAL:
		return "GLOBAL"
	case elf.STB_WEAK:
		return "WEAK"
	default:
		return fmt.Sprintf("未知(%d)", elf.ST_BIND(info))
	}
}

// 辅助函数：获取符号可见性字符串
func symbolVisibility(other uint8) string {
	switch elf.ST_VISIBILITY(other) {
	case elf.STV_DEFAULT:
		return "DEFAULT"
	case elf.STV_INTERNAL:
		return "INTERNAL"
	case elf.STV_HIDDEN:
		return "HIDDEN"
	case elf.STV_PROTECTED:
		return "PROTECTED"
	default:
		return fmt.Sprintf("未知(%d)", elf.ST_VISIBILITY(other))
	}
}

// 辅助函数：获取段类型字符串
func sectionType(secType elf.SectionType) string {
	switch secType {
	case elf.SHT_NULL:
		return "NULL"
	case elf.SHT_PROGBITS:
		return "PROGBITS"
	case elf.SHT_SYMTAB:
		return "SYMTAB"
	case elf.SHT_STRTAB:
		return "STRTAB"
	case elf.SHT_RELA:
		return "RELA"
	case elf.SHT_HASH:
		return "HASH"
	case elf.SHT_DYNAMIC:
		return "DYNAMIC"
	case elf.SHT_NOTE:
		return "NOTE"
	case elf.SHT_NOBITS:
		return "NOBITS"
	case elf.SHT_REL:
		return "REL"
	case elf.SHT_SHLIB:
		return "SHLIB"
	case elf.SHT_DYNSYM:
		return "DYNSYM"
	default:
		return fmt.Sprintf("未知(%d)", secType)
	}
}

// 辅助函数：段索引转字符串
func sectionIndexToString(index elf.SectionIndex) string {
	switch index {
	case elf.SHN_UNDEF:
		return "UND"
	case elf.SHN_ABS:
		return "ABS"
	case elf.SHN_COMMON:
		return "COM"
	default:
		return fmt.Sprintf("%d", index)
	}
}
