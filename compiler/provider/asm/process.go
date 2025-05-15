package asm

import (
	"bytes"
	"face-lang/compiler/target/elf"
	"fmt"
	"io"
	"os"
)

type Section struct {
	Name   string // 名称
	Offset int    // 偏移
}

type LabelType uint8

const UNDEFINED_LABEL LabelType = 0 // 未定义
const TEXT_LABEL LabelType = 1      // 代码段符号
const EQU_LABEL LabelType = 2       // 常量
const LOCAL_LABEL LabelType = 3     // 局部变量
const EXTERNAL_LABEL LabelType = 4  // 外部变量, 提前申明的

// LabelRecord 符号记录
type LabelRecord struct {
	Name    string    // 标签名
	Type    LabelType // 标签类型
	Addr    int       // 地址
	Index   int       // 添加顺序， 从1开始
	Section string    // 段名
	Times   int       // 重复次数
	Size    int       // 字节长度
	Cont    []int     // 内容
	ContLen int       // 内容长度
	RelInfo bool      // 记录重定位信息
}

// Write 写入数据到输出文件
func (lb *LabelRecord) Write(file *os.File) {
	offset := 0
	for i := 0; i < lb.Times; i++ {
		for j := 0; j < lb.ContLen; j++ {
			WriteValue(file, &offset, lb.Cont[j], lb.Size)
		}
	}
}

func NewLabelRec(lType LabelType) *LabelRecord {
	return &LabelRecord{Type: lType}
}

func NewLabelEqu(value int) *LabelRecord {
	return &LabelRecord{Type: EQU_LABEL, Addr: value}
}

func NewLabelText() *LabelRecord {
	return &LabelRecord{Type: TEXT_LABEL}
}

// RelocateRecord 重定位信息
type RelocateRecord struct {
	Section string // 重定位目标段
	Offset  int    // 重定位位置的偏移
	Label   string // 重定位符号的名称
	Type    int    // 重定位类型0-R_386_32；1-R_386_PC32
}

// ProcessTable 过程表，主要记录解析过程的符号信息
type ProcessTable struct {
	Section         Section           // 当前段信息
	SectionList     []*Section        // 记录所有段信息
	DataLen         int               // 总数据长度
	InstrBuff       *bytes.Buffer     // 二进制缓冲区， 存放指令代码
	InstrList       []*InstrRecord    // 指令表
	LabelRecList    []*LabelRecord    // 符号表
	LabelRecNames   map[string]int    // 符号表(名称映射)
	RelocateRecList []*RelocateRecord // 重定位记录表
}

// NewProcessTable 创建新的符号表
func NewProcessTable() *ProcessTable {
	return &ProcessTable{
		Section:       Section{},
		DataLen:       0,
		InstrBuff:     bytes.NewBuffer(nil),
		InstrList:     make([]*InstrRecord, 0),
		LabelRecList:  make([]*LabelRecord, 0),
		LabelRecNames: make(map[string]int),
	}
}

// PushInstr 将指令解析后添加到临时列表， 指令生产过程需要处理符号地址引用
func (proc *ProcessTable) PushInstr(instr *InstrRecord) {
	proc.InstrList = append(proc.InstrList, instr)
	//instr.WriteOut(proc.InstrBuff, &proc.Section.Offset)
}

// AddLabel 添加符号到符号表; 一共三处，equ 常量 仅数字 NewRecWithEqu， 变量 NewRecWithData,  代码段 TextLabel
func (proc *ProcessTable) AddLabel(name string, rec *LabelRecord) {
	rec.Name = name // 缓存一次，减少后续查找名字
	rec.Addr = proc.Section.Offset
	rec.Section = proc.Section.Name

	// 更新地址, 除了具体的变量定义，这里都是 0， 没有变化
	proc.Section.Offset += rec.Times * rec.Size * rec.ContLen

	if i, ok := proc.LabelRecNames[name]; ok {
		labelRec := proc.LabelRecList[i]
		if labelRec.Type == UNDEFINED_LABEL {
			proc.LabelRecList[i] = rec // 直接替换
		} else {
			_ = fmt.Errorf("符号: %s 重复定义！", name)
		}
	} else {
		proc.LabelRecList = append(proc.LabelRecList, rec)
		proc.LabelRecNames[name] = len(proc.LabelRecList) - 1
	}
}

// GetLabel 获取符号
func (proc *ProcessTable) GetLabel(name string) *LabelRecord {
	if i, ok := proc.LabelRecNames[name]; ok {
		return proc.LabelRecList[i]
	}

	// 未知符号，添加为外部符号(待重定位)
	rec := NewLabelRec(UNDEFINED_LABEL)
	rec.Name = name
	proc.LabelRecList = append(proc.LabelRecList, rec)
	proc.LabelRecNames[name] = len(proc.LabelRecList) - 1
	return rec
}

func (proc *ProcessTable) AddRel(label string, relType int) {
	//Offset:  proc.Section.Offset + offset, // 重定位位置的偏移
	proc.RelocateRecList = append(
		proc.RelocateRecList,
		&RelocateRecord{
			Section: proc.Section.Name,   // 重定位目标段
			Offset:  proc.Section.Offset, // 重定位位置的偏移
			Label:   label,               // 重定位符号的名称
			Type:    relType,             // 重定位类型0-R_386_32；1-R_386_PC32
		},
	)
}

// Switch 切换段 段名在这里修改 curSeg = id
func (proc *ProcessTable) Switch(id string) {
	// 确保段对齐到4字节边界
	//proc.DataLen += (4 - proc.DataLen%4) % 4
	//if proc.Section.Off != ".bss" {
	//	proc.DataLen += proc.Section.Offset
	//}

	// 记录上一个段
	proc.SectionList = append(
		proc.SectionList,
		&Section{
			Name:   proc.Section.Name,
			Offset: proc.Section.Offset, // 结束位置，也代表大小
		},
	)

	proc.Section.Name = id  // 切换到下一个段
	proc.Section.Offset = 0 // 清0段偏移
}

// ExportLb 导出符号表
//func (proc *ProcessTable) ExportLb() {
//	for _, lb := range proc.MapLabel {
//		if !lb.IsEqu { // EQU定义的符号不导出
//			ObjFile.addSym(lb)
//		}
//	}
//}

//func (proc *ProcessTable) WriteData(file *os.File) {
//	for _, lb := range proc.DefLabelList {
//		lb.write(file)
//	}
//}

func ValueBytes(value, length int) []byte {
	temp := make([]byte, length)
	for i := 0; i < length; i++ {
		temp[i] = byte((value >> (i * 8)) & 0xFF)
	}
	return temp
}

func WriteBytes(w io.Writer, offset *int, value []byte, length int) {
	*offset += length
	_, err := w.Write(value)
	if err != nil {
		panic(err)
	}
}

func WriteValue(w io.Writer, offset *int, value, length int) {
	temp := ValueBytes(value, length)
	WriteBytes(w, offset, temp, length)
}

// LocalRel 局部重定位处理
// 因为只使用一次扫描， 所以，局部 常量/变量符号 需要替换。并移除局部重定位
//func (proc *ProcessTable) LocalRel() error {
//	text := make([]byte, proc.InstrBuff.Len())
//	copy(text, proc.InstrBuff.Bytes())
//	for i, rel := range proc.RelRecordList {
//		lb := proc.GetLabel(rel.LbName)
//		if rel.TarSeg != ".text" || lb.Externed { // 没有找到内部变量
//			continue
//		}
//		relAddr := lb.Addr
//		if rel.Type == R_386_PC32 { // 相对重定位
//			relAddr -= rel.Offset + 4 // lb.addr 符号位地址, rel 当前地址
//		}
//		// 修改指定位置的数据
//		copy(text[rel.Offset:rel.Offset+4], ValueBytes(relAddr, 4))
//		proc.RelRecordList[i] = nil // 移除重定位表
//	}
//	return nil
//}

// ExportElf 文件组装 生成 ELF 可执行文件
func (proc *ProcessTable) ExportElf() *elf.File {
	magic := elf.Elf_Magic{0x7f, 'E', 'L', 'F', 0x1, 0x1, 0x1}
	target := elf.NewElfFile(magic, elf.Elf32_Half(elf.ET_REL), elf.Elf32_Half(elf.EM_386))

	// 添加段表， 段表可以保存对应段信息
	for _, section := range proc.SectionList {
		proc.DataLen += (4 - proc.DataLen%4) % 4
		off := uint32(target.Ehdr.Ehsize) + uint32(proc.DataLen)
		if section.Name == ".text" {
			target.AddShdr(section.Name, elf.Elf32_Word(elf.SHT_PROGBITS),
				elf.Elf32_Word(elf.SHF_ALLOC|elf.SHF_EXECINSTR), 0,
				off, section.Offset, 0, 0, 4, 0)
		} else if section.Name == ".data" {
			target.AddShdr(section.Name, elf.Elf32_Word(elf.SHT_PROGBITS), elf.Elf32_Word(elf.SHF_ALLOC|elf.SHF_WRITE),
				0, off, size, 0, 0, 4, 0)
		} else if section.Name == ".bss" {
			target.AddShdr(section.Name, elf.Elf32_Word(elf.SHT_NOBITS), elf.Elf32_Word(elf.SHF_ALLOC|elf.SHF_WRITE),
				0, off, size, 0, 0, 4, 0)
		}
		if section.Name != ".bss" {
			proc.DataLen += section.Offset
		}
	}
	// 段表字符串表

	// 数据位置指针, 32 怎么来的
	curOff := uint32(target.Ehdr.Ehsize) + uint32(32*len(l.SegNames)) // 文件头52B+程序头表项32*个数

	// 计算节名字符串表大小， 书中直接将三个名字 和 l.SegNames 添加到 一个 AllSegNames, 然后添加到 shstrtab(shindex 根据名字记录索引，shstrindex 记录名字偏移)
	shstrtabSize := 26 // ".shstrtab".length()+".symtab".length()+".strtab".length()+3(空格)
	for _, name := range l.SegNames {
		seg := l.SegLists[name]       // 合并后的段
		shstrtabSize += len(name) + 1 // 考虑结束符'\0'
		// 可重定位文件：没有程序头表， 只有段表

		//shType := SHT_PROGBITS
		//shFlags := SHF_ALLOC | SHF_WRITE
		//shAlign := 4 //4B
		//if name == ".bss" {
		//	shType = SHT_NOBITS
		//}
		//if name == ".text" {
		//	shFlags = SHF_ALLOC | SHF_EXECINSTR
		//	shAlign = 16
		//}
		target.AddShdr(name, Elf32_Word(shType), Elf32_Word(shFlags),
			seg.BaseAddr, seg.Offset, seg.Size, 0, 0, Elf32_Word(shAlign), 0)

		//计算有效数据段的大小和偏移,最后一个决定
		curOff = seg.Offset //修正当前偏移，循环结束后保留的是.bss的基址
	}

	target.Ehdr.Phoff = 52
	target.Ehdr.Phentsize = 32 // 程序头表大小 不包含新添加的三个段
	target.Ehdr.Phnum = elf.Elf32_Half(len(l.SegNames))
	target.Shstrtab = make([]byte, shstrtabSize)
	target.ShstrtabSize = shstrtabSize

	// todo 可优化，段表字符串信息，可以在添加段表时生成
	index := 0
	shstrIndex := make(map[string]int) //段表串名与索引映射
	shstrIndex[".shstrtab"] = index
	copy(target.Shstrtab[index:], ".shstrtab")
	index += 10

	shstrIndex[".symtab"] = index
	copy(target.Shstrtab[index:], ".symtab")
	index += 8

	shstrIndex[".strtab"] = index
	copy(target.Shstrtab[index:], ".strtab")
	index += 8

	shstrIndex[""] = index - 1

	for _, name := range l.SegNames {
		shstrIndex[name] = index
		copy(target.Shstrtab[index:], name)
		index += len(name) + 1 // 留一个 \x0
	}

	// .shstrtab
	target.AddShdr(".shstrtab", elf.Elf32_Word(elf.SHT_STRTAB), 0, 0,
		curOff, elf.Elf32_Word(shstrtabSize), elf.Elf32_Word(elf.SHN_UNDEF), 0, 1, 0)
	target.Ehdr.Shstrndx = elf.Elf32_Half(target.GetSymIndex(".shstrtab"))
	curOff += uint32(shstrtabSize) //段表偏移
	target.Ehdr.Shoff = curOff
	target.Ehdr.Shentsize = 40
	target.Ehdr.Shnum = elf.Elf32_Half(4 + len(l.SegNames))

	//生成符号表项
	curOff += 40 * (4 + uint32(target.Ehdr.Shnum)) //符号表偏移
	target.AddShdr(".symtab", elf.Elf32_Word(elf.SHT_SYMTAB), 0, 0,
		curOff, elf.Elf32_Word((1+len(l.SymDef))*16), 0, 0, 1, 16)
	target.ShdrTab[".symtab"].Link = elf.Elf32_Word(target.GetSegIndex(".symtab") + 1) //。strtab默认在.symtab之后
	strtabSize := 0                                                                    //字符串表大小
	for _, link := range l.SymDef {                                                    //遍历所有符号
		strtabSize += len(link.Name) + 1
		sym := link.Prov.SymTab[link.Name]
		sym.Shndx = uint16(target.GetSegIndex(link.Prov.ShdrNames[sym.Shndx]))
		target.AddSym(link.Name, sym)
	}

	// 设置程序入口点
	target.Ehdr.Entry = target.SymTab[StartSymbol].Value

	// .strtab偏移
	curOff += uint32((1 + len(l.SymDef)) * 16)
	// 添加 .strtab
	target.AddShdr(".strtab", elf.Elf32_Word(elf.SHT_STRTAB), 0, 0,
		curOff, uint32(strtabSize), 0, 0, 1, 0)

	// 填充字符串表
	target.Strtab = make([]byte, strtabSize)
	target.StrtabSize = strtabSize

	//串表与索引映射
	index = 0
	strIndex := make(map[string]int)
	strIndex[""] = strtabSize - 1
	for _, link := range l.SymDef {
		strIndex[link.Name] = index
		name := fmt.Sprintf("%s\x00", link.Name)
		copy(target.Strtab[index:], name)
		index += len(link.Name) + 1
	}

	//更新符号表name
	for n, sym := range target.SymTab {
		sym.Name = uint32(strIndex[n])
	}

	//更新段表name
	for n, shdr := range target.ShdrTab {
		shdr.Name = uint32(strIndex[n])
	}

	return target

}
