package asm

import (
	"bytes"
	"face-lang/compiler/target/elf"
	"fmt"
	"io"
	"os"
	"strings"
)

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
	Section         elf.Section       // 当前段信息
	SectionList     []*elf.Section    // 记录所有段信息
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
		Section:       elf.Section{},
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

	instr.Codegen(
		proc.InstrBuff,
		&proc.Section.Offset,
		func(label string, relType int) bool { // 处理重定位
			if label == "" {
				return false
			}
			proc.AddRel(label, relType)
			return true
		},
	)
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

	// 只有符号引用符号时， 才会被创建
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
		&elf.Section{
			Name:   proc.Section.Name,
			Length: proc.Section.Offset, // 结束位置，也代表大小, 先不记录偏移
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

//// Codegen 代码生成, 生成代码，同时记录每个段的大小
//func (proc *ProcessTable) Codegen() error {
//	// 源码扫描完成，开始生成代码， 内部符号已存在
//	instrBuffer := bytes.NewBuffer(nil)
//	for _, instr := range proc.InstrList {
//		instr.WriteOut(instrBuffer, &proc.Section.Offset)
//	}
//	instrBuffer.Len() // 代码段大小
//
//	// important 符号表[可能]存在符号嵌套引用
//	//    但是所有嵌套引用，被引用的符号必须被声明
//	//        如果引用外部符号，记录地址， 下一个引用符合也只引用所在地址信息
//	//    逻辑上无需处理嵌套
//	for _, label := range proc.LabelRecList {
//
//	}
//	//for _, instr := range proc.InstrList {
//	//	instr.WriteOut(instrBuffer, &proc.Section.Offset)
//	//}
//
//}

// LocalRel 局部重定位处理
// 因为只使用一次扫描， 所以，局部 常量/变量符号 需要替换。并移除局部重定位
func (proc *ProcessTable) LocalRel() error {
	text := make([]byte, proc.InstrBuff.Len())
	copy(text, proc.InstrBuff.Bytes())
	for i, rel := range proc.RelocateRecList {
		lb := proc.GetLabel(rel.Label)

		// 符号段的重定位，如果未定义，就一定是外部符号， 所以不存在重定位
		if rel.Section != ".text" {
			continue
		}

		//  todo 未定义符号，或外部符号， 严格来说，需要什么未外部符号， 暂时忽略
		if lb.Type == UNDEFINED_LABEL || lb.Type == EXTERNAL_LABEL {
			continue // 不做处理，都是要重定位的
		}

		if lb.Type == EQU_LABEL {
			proc.RelocateRecList[i] = nil // 移除重定位表, 保留绝对重定位
		}

		relAddr := lb.Addr
		if rel.Type == R_386_PC32 { // 相对重定位
			relAddr -= rel.Offset + 4     // lb.addr 符号位地址, rel 当前地址
			proc.RelocateRecList[i] = nil // 移除重定位表, 保留绝对重定位
		}

		// 修改指定位置的数据
		copy(text[rel.Offset:rel.Offset+4], ValueBytes(relAddr, 4))

		//if lb.Type == EQU_LABEL {
		//	proc.RelocateRecList[i] = nil // 移除重定位表
		//}
	}
	proc.InstrBuff.Reset()
	proc.InstrBuff.Write(text)
	return nil
}

// ExportElf 文件组装 生成 ELF 可执行文件 todo 内部重定位应该在组装文件之前
func (proc *ProcessTable) ExportElf() *elf.File {
	magic := elf.Elf_Magic{0x7f, 'E', 'L', 'F', 0x1, 0x1, 0x1}
	target := elf.NewElfFile(magic, elf.Elf32_Half(elf.ET_REL), elf.Elf32_Half(elf.EM_386))

	// 从头文件结束开始
	// 添加段表， 段表可以保存对应段信息 // 段表主要有 .text .data .bss
	offset := int(target.Ehdr.Ehsize)
	for _, section := range proc.SectionList {
		offset += (4 - offset%4) % 4 // 4字节对齐（头信息, 52/64 也是 4 字节对齐的，其它未知）
		switch section.Name {
		case ".text":
			target.ProgSegList = append(
				target.ProgSegList,
				&elf.ProgSeg{
					Name:   section.Name,
					Offset: uint32(offset),
					Size:   uint32(section.Length),
					Blocks: []*elf.Block{
						{
							Data: proc.InstrBuff.Bytes(),
							Size: uint32(proc.InstrBuff.Len()),
						},
					},
				},
			)
		case ".data":
			dataBuffer := bytes.NewBuffer(nil)
			for _, label := range proc.LabelRecList {
				if label.Section == ".data" && label.Type == LOCAL_LABEL {
					for t := 0; t < label.Times; t++ {
						for n := 0; n < label.ContLen; n++ {
							dataBuffer.Write(ValueBytes(label.Cont[n], label.Size))
						}
					}
				}
			}
			target.ProgSegList = append(
				target.ProgSegList,
				&elf.ProgSeg{
					Name:   section.Name,
					Offset: uint32(offset),
					Size:   uint32(section.Length),
					Blocks: []*elf.Block{
						{
							Data: dataBuffer.Bytes(),
							Size: uint32(dataBuffer.Len()),
						},
					},
				},
			)
		case ".bss":
			target.ProgSegList = append(
				target.ProgSegList,
				&elf.ProgSeg{
					Name:   section.Name,
					Offset: uint32(offset),
					Size:   uint32(section.Length),
					Blocks: nil,
				},
			)
		default:

		}
		target.AddShdrSec(section, offset)

		if section.Name != ".bss" { // .bss 65535 不添加
			offset += section.Length
		}
	}

	// 先添加符号表
	for _, label := range proc.LabelRecList {
		if label.Type == EQU_LABEL {
			continue
		}

		name := label.Name
		/*
			//对于@while_ @if_ @lab_ @cal_开头和@s_stack的都是局部符号，可以不用导出，但是为了objdump方便而导出
		*/
		if strings.HasPrefix(name, "@lab_") || strings.HasPrefix(name, "@if_") ||
			strings.HasPrefix(name, "@while_") ||
			strings.HasPrefix(name, "@cal_") || strings.HasPrefix(name, "@s_stack") {
			continue
		}

		//解析符号的全局性局部性，避免符号冲突
		glb := false

		if label.Section == ".text" { // 代码段
			if name == "@str2long" || name == "@procBuf" {
				glb = true
			} else if name[0] != '@' { //不带@符号的，都是定义的函数或者_start,全局的
				glb = true
			}
		} else if label.Section == ".data" { // 数据段
			if strings.HasPrefix(name, "@str_") { // @str_开头符号
				glb = !(name[5] >= '0' && name[5] <= '9') //不是紧跟数字，全局str
			} else { // 其他类型全局符号
				glb = true
			}
		} else {
			glb = label.Type == UNDEFINED_LABEL || label.Type == EXTERNAL_LABEL
		}

		sym := elf.Elf32_Sym{
			Name:  0,
			Value: uint32(label.Addr),
			Size:  uint32(label.Times * label.Size * label.ContLen),
		}
		if glb { // 统一记作无类型符号，和链接器lit协议保持一致
			sym.Info = elf.ST_INFO(elf.STB_GLOBAL, elf.STT_NOTYPE) //全局符号
		} else {
			sym.Info = elf.ST_INFO(elf.STB_LOCAL, elf.STT_NOTYPE) //局部符号，避免名字冲突
		}

		if label.Type == UNDEFINED_LABEL || label.Type == EXTERNAL_LABEL {
			sym.Shndx = 0 // STN_UNDEF
		} else {
			sym.Shndx = uint16(target.GetSegIndex(label.Section))
		}
		target.AddSym(label.Name, &sym)
	}

	// 段表字符串表
	//".rel.text".length()+".rel.data".length()+".bss".length()
	//".shstrtab".length()+".symtab".length()+".strtab".length()+5;//段表字符串表大小
	target.ShstrtabSize = 51
	target.Shstrtab = make([]byte, target.ShstrtabSize)

	shstrIndex := make(map[string]int) // 段名所在位置索引
	index := 0

	// .rel.text
	shstrIndex[".rel.text"] = index
	copy(target.Shstrtab[index:], ".rel.text\x00")

	shstrIndex[".text"] = index + 4 // 一个字符串，两处使用，节省空间
	index += 10

	// .rel.data
	shstrIndex[".rel.data"] = index
	copy(target.Shstrtab[index:], ".rel.data\x00")

	shstrIndex[".data"] = index + 4
	index += 10

	// .bss
	shstrIndex[".bss"] = index
	copy(target.Shstrtab[index:], ".bss\x00")
	index += 5

	// .shstrtab
	shstrIndex[".shstrtab"] = index
	copy(target.Shstrtab[index:], ".shstrtab\x00")
	index += 10

	// .symtab
	shstrIndex[".symtab"] = index
	copy(target.Shstrtab[index:], ".symtab\x00")
	index += 8

	// .strtab
	shstrIndex[".strtab"] = index
	copy(target.Shstrtab[index:], ".strtab\x00")
	index += 8

	shstrIndex[""] = index - 1

	// .shstrtab, 紧跟 .text.data.bss 段
	shstrTab := elf.NewShdr(elf.SHT_STRTAB, 0, offset, target.ShstrtabSize)
	shstrTab.Link = elf.Elf32_Word(elf.SHN_UNDEF)
	shstrTab.Addralign = 1
	target.AddShdr(".shstrtab", shstrTab)

	offset += target.ShstrtabSize

	target.Ehdr.Shoff = elf.Elf32_Off(offset)

	// ---- 添加符号表 ToDo 需要先构建符号表
	offset += 9 * int(target.Ehdr.Shentsize) // 什么意思？符号表表偏移 = 8个段+空段，段表字符串表偏移
	// .symtab,sh_link 代表.strtab索引，默认在.symtab之后,sh_info不能确定
	symtabSize := (len(target.SymNames)) * 16 // 每个符号16字节
	symtab := elf.NewShdr(elf.SHT_SYMTAB, 0, offset, symtabSize)
	symtab.Addralign = 1
	symtab.Entsize = 16
	target.AddShdr(".symtab", symtab)
	symtab.Link = elf.Elf32_Word(target.GetSegIndex(".symtab")) + 1 //.strtab默认在.symtab之后(当前索引+1)

	// ---- 字符串表 .strtab
	offset += symtabSize
	target.StrtabSize = 0                       //字符串表大小
	for i := 0; i < len(target.SymNames); i++ { //遍历所有符号
		target.StrtabSize += len(target.SymNames[i]) + 1
	}

	// 填充字符串表 & 串表与符号表名字更新
	target.Strtab = make([]byte, target.StrtabSize)
	strtabIndex := 0 // 索引1开始
	for _, name := range target.SymNames {
		target.SymTab[name].Name = uint32(strtabIndex) // 将符号名对应索引记录到符号表
		copy(target.Strtab[strtabIndex:], name+"\x00")
		strtabIndex += len(name) + 1
	}
	//shstrIndex[""] = index - 1

	strtab := elf.NewShdr(elf.SHT_STRTAB, 0, offset, target.StrtabSize)
	strtab.Link = elf.Elf32_Word(elf.SHN_UNDEF)
	strtab.Addralign = 1
	target.AddShdr(".strtab", strtab)

	// ---- 重定位表
	offset += target.StrtabSize
	relTextSize, relDataSize := 0, 0

	for _, rel := range proc.RelocateRecList {
		if rel == nil { // 已内部处理；无需重定位；可能出现空值
			continue
		}

		symIndex := target.GetSymIndex(rel.Label)
		relInfo := &elf.Elf32_RelInfo{
			SegName: rel.Section,
			Rel: &elf.Elf32_Rel{
				Offset: uint32(rel.Offset),
				Info:   uint32(((symIndex) << 8) + ((rel.Type) & 0xff)),
			},
			RelName: rel.Label, // 符号名称
		}

		if rel.Section == ".text" { // 应该有4个段，实际只有一个段， 应该记录所有相对重定位，EQU除外
			target.AddRel(relInfo)
			relTextSize += 8 // 每个重定位项8字节
		} else if rel.Section == ".data" {
			target.AddRel(relInfo)
			relDataSize += 8
		}
	}

	// 添加.rel.text
	relTextTab := elf.NewShdr(elf.SHT_REL, 0, offset, relTextSize)
	relTextTab.Link = elf.Elf32_Word(target.GetSegIndex(".symtab"))
	relTextTab.Info = elf.Elf32_Word(target.GetSegIndex(".text"))
	relTextTab.Addralign = 1
	relTextTab.Entsize = 8
	target.AddShdr(".rel.text", relTextTab)

	// 添加.rel.data
	offset += relTextSize
	relDataTab := elf.NewShdr(elf.SHT_REL, 0, offset, relDataSize)
	relDataTab.Link = elf.Elf32_Word(target.GetSegIndex(".symtab"))
	relDataTab.Info = elf.Elf32_Word(target.GetSegIndex(".data"))
	relDataTab.Addralign = 1
	relDataTab.Entsize = 8
	target.AddShdr(".rel.data", relDataTab)

	//更新符号表name
	//for n, sym := range target.SymTab {
	//	sym.Name = uint32(strtabIndex[n])
	//}

	//更新段表name
	for n, shdr := range target.ShdrTab {
		shdr.Name = uint32(shstrIndex[n])
	}

	target.Ehdr.Shentsize = 40
	target.Ehdr.Shnum = elf.Elf32_Half(len(target.ShdrTab))
	target.Ehdr.Shstrndx = elf.Elf32_Half(target.GetSegIndex(".shstrtab"))

	return target

}

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
