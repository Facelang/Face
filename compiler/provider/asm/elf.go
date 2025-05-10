package asm

import (
	"fmt"
	"os"
	"strings"
)

// Elf32_Half
// Type for a 16-bit quantity.
type Elf32_Half = uint16
type Elf64_Half = uint16

// Elf32_Word
// Types for signed and unsigned 32-bit quantities.
type Elf32_Word = uint32

type Elf32_Sword = int32

type Elf64_Word = uint32

type Elf64_Sword = int32

// Elf32_Xword
// Types for signed and unsigned 64-bit quantities.
type Elf32_Xword = uint64

type Elf32_Sxword = int64

type Elf64_Xword = uint64

type Elf64_Sxword = int64

// Elf32_Addr
// Type of addresses.
type Elf32_Addr = uint32

type Elf64_Addr = uint64

// Elf32_Off
// Type of file offsets.
type Elf32_Off = uint32

type Elf64_Off = uint64

// Elf32_Section
// Type for section indices, which are 16-bit quantities.
type Elf32_Section = uint16

type Elf64_Section = uint16

// Elf32_Versym
// Type for version symbol information.
type Elf32_Versym = Elf32_Half

type Elf64_Versym = Elf64_Half

const EI_NIDENT = 16

// Elf32_Ehdr ELF32文件头结构
type Elf32_Ehdr struct {
	E_ident     [16]byte // 魔数和相关信息
	E_type      uint16   // 文件类型
	E_machine   uint16   // 架构类型 4
	E_version   uint32   // 文件版本
	E_entry     uint32   // 入口点虚拟地址
	E_phoff     uint32   // 程序头表偏移
	E_shoff     uint32   // 节头表偏移
	E_flags     uint32   // 处理器特定标志 44
	E_ehsize    uint16   // ELF头部大小
	E_phentsize uint16   // 程序头表项大小
	E_phnum     uint16   // 程序头表项数量
	E_shentsize uint16   // 节头表项大小
	E_shnum     uint16   // 节头表项数量
	E_shstrndx  uint16   // 节头字符串表索引
}

// Elf32_Shdr 段表项结构
type Elf32_Shdr struct {
	Sh_name      uint32 // 段名（4字节，存在于字符串表中的偏移量， shstrtab 也是一个段， e_shstrndx ）
	Sh_type      uint32 // 段类型 (1表示程序段.text.data 2表示符号段.symtab 3表示串表段.shstrtab 8表示内容段.bss 9表示重定位表段.rel.text.rel.data)
	Sh_flags     uint32 // 段标志 (0表示默认 1表示可写 2表示段加载后需要为之分配空间 4表示可执行)
	Sh_addr      uint32 // 段虚拟地址 可重定位文件默认为零， 可执行文件由链接器计算地址
	Sh_offset    uint32 // 段在文件中的偏移
	Sh_size      uint32 // 段的大小，字节单位， SHT_NOBITS 代表没有数据（此时指代加载后占用的内存大小）
	Sh_link      uint32 // 段的链接信息，一般用于描述符号标段和重定位表段的链接信息。
	Sh_info      uint32 // 附加信息
	Sh_addralign uint32 // 对齐要求
	Sh_entsize   uint32 // 表项大小
}

// Elf32_Sym ELF32符号表项结构
type Elf32_Sym struct {
	St_name  uint32 // 符号名称在字符串表中的偏移
	St_value uint32 // 符号值
	St_size  uint32 // 符号大小
	St_info  uint8  // 符号类型和绑定属性
	St_other uint8  // 未使用
	St_shndx uint16 // 符号所在节的索引
}

// Elf32_Rel ELF32重定位表项结构
type Elf32_Rel struct {
	R_offset uint32 // 重定位发生的位置
	R_info   uint32 // 重定位类型和符号索引
}

// ELF文件结构// elf文件类，包含elf文件的重要内容，处理elf文件
type ElfFile struct {
	Ehdr      Elf32_Ehdr             // ELF文件头
	ShdrTab   map[string]*Elf32_Shdr // 段表映射
	ShdrNames []string               // 段名列表,  段表名和索引的映射关系，方便符号查询自己的段信息
	SymTab    map[string]*Elf32_Sym  // 符号表映射
	SymNames  []string               // 符号名列表, 符号名与符号表项索引的映射关系，对于重定位表生成重要
	//RelTab       []RelInfo              // 重定位信息列表
	Shstrtab     []byte       // 段表字符串表数据
	ShstrtabSize int          // 段表字符串表长
	Strtab       []byte       // 字符串表数据
	StrtabSize   int          // 字符串表长
	RelTextTab   []*Elf32_Rel // -
	RelDataTab   []*Elf32_Rel // -
}

// ELF文件常量
const (
	// 文件类型
	ET_REL  = 1 // 可重定位文件
	ET_EXEC = 2 // 可执行文件

	// 机器类型
	EM_386 = 3 // Intel 80386

	// 节头类型
	SHT_NULL     = 0  // 未使用
	SHT_PROGBITS = 1  // 程序数据
	SHT_SYMTAB   = 2  // 符号表
	SHT_STRTAB   = 3  // 字符串表
	SHT_RELA     = 4  // 带附加信息的重定位项
	SHT_HASH     = 5  // 符号哈希表
	SHT_DYNAMIC  = 6  // 动态链接信息
	SHT_NOTE     = 7  // 文件信息
	SHT_NOBITS   = 8  // 不占文件空间的数据
	SHT_REL      = 9  // 重定位项
	SHT_SHLIB    = 10 // 保留
	SHT_DYNSYM   = 11 // 动态链接符号表

	// 节头标志
	SHF_WRITE     = 0x1 // 可写
	SHF_ALLOC     = 0x2 // 占用内存
	SHF_EXECINSTR = 0x4 // 包含可执行代码

	// 符号绑定类型
	STB_LOCAL  = 0 // 局部符号
	STB_GLOBAL = 1 // 全局符号

	// 符号类型
	STT_NOTYPE  = 0 // 未知类型
	STT_OBJECT  = 1 // 数据对象
	STT_FUNC    = 2 // 函数
	STT_SECTION = 3 // 节
	STT_FILE    = 4 // 文件名

	// 特殊节索引
	SHN_UNDEF = 0 // 未定义的符号
)

//func NewRelInfo(seg string, addr int, lb string, t int) *RelInfo {
//	return &RelInfo{Offset: addr, TarSeg: seg, LbName: lb, Type: t}
//}

func NewElfFile() *ElfFile {
	elf := &ElfFile{
		ShdrTab:   make(map[string]*Elf32_Shdr),
		SymTab:    make(map[string]*Elf32_Sym),
		ShdrNames: make([]string, 0),
		SymNames:  make([]string, 0),
		//RelTab:     make([]RelInfo, 0),
		Shstrtab:   make([]byte, 0),
		Strtab:     make([]byte, 0),
		RelTextTab: make([]*Elf32_Rel, 0),
		RelDataTab: make([]*Elf32_Rel, 0),
	}

	// 初始化ELF文件头
	elf.Ehdr = Elf32_Ehdr{
		E_ident:     [16]byte{}, // 这个字段比较复杂
		E_type:      ET_REL,     // 文件类型： 1表示可重定位, 2表示可执行 3表示共享目标 4 表示核心转储  0 表示无效
		E_machine:   EM_386,     // 机器类型
		E_version:   1,          // 文件版本 一般取1
		E_entry:     0,          // 程序入口的线性地址，一般用于可以执行文件， 可重定向文件该字段为 0
		E_phoff:     0,          // 程序头表在文件内的偏移地址， 标识了程序头表在文件内的位置
		E_flags:     0,          // 文件平台相关属性， 一般默认为 0
		E_ehsize:    52,         // 文件头的大小
		E_phentsize: 0,          // 程序头表项的大小
		E_phnum:     0,          // 程序头表项的个数，确定程序头表在文件[e_phoff: e_phoff + e_phentsize*e_phnum] 的数据块中
		E_shentsize: 40,         // 段表项的大小
		E_shnum:     0,          // 段表项的个数， 确定数据区块存在于 [e_shoff:e_shoff+e_shentsize*eshnum] 中
		E_shstrndx:  0,          // .shstrtab的索引
	}

	// 初始化ELF魔数
	elf.Ehdr.E_ident[0] = 0x7F // \DEL
	elf.Ehdr.E_ident[1] = 'E'  //
	elf.Ehdr.E_ident[2] = 'L'  //
	elf.Ehdr.E_ident[3] = 'F'  //
	elf.Ehdr.E_ident[4] = 1    // 32位格式 64位(2) 0表示无效
	elf.Ehdr.E_ident[5] = 1    // 小端序 大端(2) 0表示无效
	elf.Ehdr.E_ident[6] = 1    // ELF版本 默认为1
	// 后面9字节在ELF标准中未定义， 一般用于平台相关的扩展标志
	// 第8字节 取0 表示 unix 系统
	// 第9字节 取0 表示系统 ABI 版本为 0
	// 其它字节默认为 0

	// 添加空节表项
	elf.addShdrFunc("", 0, 0, 0, 0, 0, 0, 0, 0, 0)

	// 添加空符号表项
	elf.addSym2("", &Elf32_Sym{})

	return elf
}

func (e *ElfFile) getSegIndex(segName string) int {
	index := 0
	for i := 0; i < len(e.ShdrNames); i++ {
		if e.ShdrNames[i] == segName { // 找到段
			break
		}
		index += 1
	}
	return index
}

func (e *ElfFile) getSymIndex(symName string) int {
	index := 0
	for i := 0; i < len(e.SymNames); i++ {
		if e.SymNames[i] == symName { // 找到段
			break
		}
		index += 1
	}
	return index
}

// sh_name和sh_offset都需要重新计算
func (e *ElfFile) addShdr(shName string, size uint32) {
	off := uint32(52 + ProcessTable.DataLen)
	if shName == ".text" {
		e.addShdrFunc(shName, SHT_PROGBITS, SHF_ALLOC|SHF_EXECINSTR, 0, off, size, 0, 0, 4, 0)
	} else if shName == ".data" {
		e.addShdrFunc(shName, SHT_PROGBITS, SHF_ALLOC|SHF_WRITE, 0, off, size, 0, 0, 4, 0)
	} else if shName == ".bss" {
		e.addShdrFunc(shName, SHT_NOBITS, SHF_ALLOC|SHF_WRITE, 0, off, size, 0, 0, 4, 0)
	}
}

// 添加一个段表项
func (e *ElfFile) addShdrFunc(shName string, sh_type, sh_flags Elf32_Word,
	sh_addr Elf32_Addr, sh_offset Elf32_Off,
	sh_size, sh_link, sh_info, sh_addralign, sh_entsize Elf32_Word) {
	sh := &Elf32_Shdr{}
	sh.Sh_name = 0
	sh.Sh_type = sh_type
	sh.Sh_flags = sh_flags
	sh.Sh_addr = sh_addr
	sh.Sh_offset = sh_offset
	sh.Sh_size = sh_size
	sh.Sh_link = sh_link
	sh.Sh_info = sh_info
	sh.Sh_addralign = sh_addralign
	sh.Sh_entsize = sh_entsize
	// 这里 ShdrTab 可以直接使用数组，ShdrNames写入Name时，直接返回索引。写入 Sh_name
	e.ShdrTab[shName] = sh
	e.ShdrNames = append(e.ShdrNames, shName)
}

func (e *ElfFile) addSym(lb *LabelRecord) {
	//解析符号的全局性局部性，避免符号冲突
	//bool glb=false;
	//string name=lb->lbName;
	///*
	//	//对于@while_ @if_ @lab_ @cal_开头和@s_stack的都是局部符号，可以不用导出，但是为了objdump方便而导出
	//*/
	//if(name.find("@lab_")==0||name.find("@if_")==0||name.find("@while_")==0||name.find("@cal_")==0||name=="@s_stack")
	//return;
	glb := false
	if lb.SegName == ".text" { // 代码段
		if lb.LbName == "@str2long" || lb.LbName == "@procBuf" {
			glb = true
		} else if lb.LbName[0] != '@' { //不带@符号的，都是定义的函数或者_start,全局的
			glb = true
		}
	} else if lb.SegName == ".data" { //数据段
		strings.Contains(lb.LbName, "@str_")
		if lb.LbName[:5] == "@str_" { //@str_开头符号
			glb = !(lb.LbName[5] >= '0' && lb.LbName[5] <= '9') //不是紧跟数字，全局str
		} else { //其他类型全局符号
			glb = true
		}
	} else if lb.SegName == "" { // 外部符号
		glb = lb.Externed //false
	}
	sym := &Elf32_Sym{}
	sym.St_name = 0
	sym.St_value = uint32(lb.Addr)                       //符号段偏移,外部符号地址为0
	sym.St_size = uint32(lb.Times * lb.Len * lb.ContLen) //函数无法通过目前的设计确定，而且不必关心
	if glb {                                             //统一记作无类型符号，和链接器lit协议保持一致
		sym.St_info = ((STB_GLOBAL) << 4) + ((STT_NOTYPE) & 0xf) //全局符号
	} else {
		sym.St_info = ((STB_LOCAL) << 4) + ((STT_NOTYPE) & 0xf) //局部符号，避免名字冲突
	}
	sym.St_other = 0
	if lb.Externed {
		sym.St_shndx = 0 // /* End of a chain.  */
	} else {
		sym.St_shndx = uint16(e.getSegIndex(lb.SegName))
	}
	e.addSym2(lb.LbName, sym)
}

func (e *ElfFile) addSym2(stName string, s *Elf32_Sym) {
	sym := &Elf32_Sym{}
	e.SymTab[stName] = sym
	if stName == "" {
		sym.St_name = 0
		sym.St_value = 0
		sym.St_size = 0
		sym.St_info = 0
		sym.St_other = 0
		sym.St_shndx = 0
	} else {
		sym.St_name = 0
		sym.St_value = s.St_value
		sym.St_size = s.St_size
		sym.St_info = s.St_info
		sym.St_other = s.St_other
		sym.St_shndx = s.St_shndx
	}
	e.SymNames = append(e.SymNames, stName)
}

//func (e *ElfFile) addRel(seg string, addr int, lb string, t int) {
//	rel := NewRelInfo(seg, addr, lb, t)
//	e.RelTab = append(e.RelTab, *rel)
//}

// WriteElf 写入ELF文件
func (e *ElfFile) WriteElf(outName string) {
	//_ = Fout.Close()
	//// 打开输出文件
	outFile, err := os.Create(outName)
	if err != nil {
		fmt.Printf("无法创建输出文件: %s\n", err)
		return
	}
	//Fout = outFile

	defer func() {
		_ = outFile.Close()
	}()

	// 组装ELF文件
	e.assembleElfFile()

	// 写入ELF头
	e.writeElfHeaderToFile(outFile)

	//输出.text
	//fclose(fin);
	//fin=fopen((fName+".t").c_str(),"r");//临时输出文件，供代码段使用
	//char buffer[1024]={0};
	//int f=-1;
	//while(f!=0)
	//{
	//	f=fread(buffer,1,1024,fin);
	//	fwrite(buffer,1,f,fout);
	//}

	// 写入.text段
	e.writeTextSectionToFile(outFile, outName)

	// 填充.text和.data段之间的空隙
	e.PadSeg(outFile, ".text", ".data")

	// 写入.data段，在main.go中调用semantic.Table.Write()

	// 填充.data和.bss段之间的空隙
	e.PadSeg(outFile, ".data", ".bss")

	// 写入其余部分
	e.writeElfTailToFile(outFile)

	fmt.Printf("ELF文件 %s 已生成\n", outName)
}

// PadSeg 填充段间的空隙
func (e *ElfFile) PadSeg(outFile *os.File, first, second string) {
	pad := []byte{0}
	padNum := int(e.ShdrTab[second].Sh_offset - (e.ShdrTab[first].Sh_offset + e.ShdrTab[first].Sh_size))

	for i := 0; i < padNum; i++ {
		_, _ = outFile.Write(pad)
	}
}

// assembleElfFile 组装ELF文件
func (e *ElfFile) assembleElfFile() {

	// 准备段表字符串表
	//".rel.text".length()+".rel.data".length()+".bss".length()
	//".shstrtab".length()+".symtab".length()+".strtab".length()+5;//段表字符串表大小
	shstrtabSize := 51 // 所有段名的总长度，包括结束符
	e.Shstrtab = make([]byte, shstrtabSize)

	// 填充段表字符串表
	shstrIndex := make(map[string]int)
	index := 0

	// .rel.text
	shstrIndex[".rel.text"] = index
	copy(e.Shstrtab[index:], ".rel.text\x00")
	shstrIndex[".text"] = index + 4
	index += 10

	// .rel.data
	shstrIndex[".rel.data"] = index
	copy(e.Shstrtab[index:], ".rel.data\x00")
	shstrIndex[".data"] = index + 4
	index += 10

	// .bss
	shstrIndex[".bss"] = index
	copy(e.Shstrtab[index:], ".bss\x00")
	index += 5

	// .shstrtab
	shstrIndex[".shstrtab"] = index
	copy(e.Shstrtab[index:], ".shstrtab\x00")
	index += 10

	// .symtab
	shstrIndex[".symtab"] = index
	copy(e.Shstrtab[index:], ".symtab\x00")
	index += 8

	// .strtab
	shstrIndex[".strtab"] = index
	copy(e.Shstrtab[index:], ".strtab\x00")
	index += 8

	// 计算所有段的偏移和大小
	offset := 52 + ProcessTable.DataLen // ELF头 + 代码和数据 header+(.text+pad+.data+pad)数据偏移，.shstrtab偏移

	// 添加.shstrtab段
	e.addShdrFunc(".shstrtab", SHT_STRTAB, 0, 0, uint32(offset), uint32(shstrtabSize), SHN_UNDEF, 0, 1, 0)
	offset += shstrtabSize
	e.Ehdr.E_shoff = uint32(offset) // 设置段表偏移

	//-----添加符号表
	offset += 9 * 40 // 9个段表项，每个40字节（8个段+空段，段表字符串表偏移，符号表表偏移）
	// .symtab,sh_link 代表.strtab索引，默认在.symtab之后,sh_info不能确定
	symtabSize := (len(e.SymNames)) * 16 // 每个符号16字节
	e.addShdrFunc(".symtab", SHT_SYMTAB, 0, 0, uint32(offset), uint32(symtabSize), 7, 1, 4, 16)
	e.ShdrTab[".symtab"].Sh_link = uint32(e.getSegIndex(".symtab") + 1) //.strtab默认在.symtab之后
	offset += symtabSize                                                //.strtab偏移
	//-----添加.strtab
	e.StrtabSize = 0                       //字符串表大小
	for i := 0; i < len(e.SymNames); i++ { //遍历所有符号
		e.StrtabSize += len(e.SymNames[i]) + 1
	}

	//填充strtab数据
	e.addShdrFunc(".strtab", SHT_STRTAB, 0, 0, uint32(offset), uint32(e.StrtabSize), SHN_UNDEF, 0, 1, 0) //.strtab
	e.Strtab = make([]byte, e.StrtabSize)
	offset += e.StrtabSize

	//串表与符号表名字更新
	// 填充字符串表
	strtabIndex := 1 // 索引1开始
	for _, name := range e.SymNames {
		e.SymTab[name].St_name = uint32(strtabIndex)
		copy(e.Strtab[strtabIndex:], name+"\x00")
		strtabIndex += len(name) + 1
	}

	// 添加字符串表段

	// 处理重定位表
	relTextSize := 0
	relDataSize := 0
	//for _, rel := range e.RelTab {
	//	symIndex := e.getSymIndex(rel.LbName)
	//	relData := &Elf32_Rel{
	//		R_offset: uint32(rel.Offset),
	//		R_info:   uint32(((symIndex) << 8) + ((rel.Type) & 0xff)),
	//	}
	//	if rel.TarSeg == ".text" {
	//		e.RelTextTab = append(e.RelTextTab, relData)
	//		relTextSize += 8 // 每个重定位项8字节
	//	} else if rel.TarSeg == ".data" {
	//		e.RelTextTab = append(e.RelDataTab, relData)
	//		relDataSize += 8
	//	}
	//}

	//-----添加.rel.text
	e.addShdrFunc(".rel.text", SHT_REL, 0, 0, uint32(offset), Elf32_Word(relTextSize), Elf32_Word(e.getSegIndex(".symtab")), Elf32_Word(e.getSegIndex(".text")), 1, 8) //.rel.text
	offset += relTextSize
	//-----添加.rel.data
	e.addShdrFunc(".rel.data", SHT_REL, 0, 0, Elf32_Off(offset), Elf32_Word(relDataSize), Elf32_Word(e.getSegIndex(".symtab")), Elf32_Word(e.getSegIndex(".data")), 1, 8) //.rel.data

	// 更新段名在字符串表中的偏移
	for name, shdr := range e.ShdrTab {
		if idx, ok := shstrIndex[name]; ok {
			shdr.Sh_name = uint32(idx)
		}
	}
}

// writeElfHeaderToFile 写入ELF头
func (e *ElfFile) writeElfHeaderToFile(outFile *os.File) {
	// 写入ELF头部魔数
	_, _ = outFile.Write(e.Ehdr.E_ident[:])

	// 写入ELF头部其他字段
	WriteUint16(outFile, e.Ehdr.E_type)
	WriteUint16(outFile, e.Ehdr.E_machine)
	WriteUint32(outFile, e.Ehdr.E_version)
	WriteUint32(outFile, e.Ehdr.E_entry)
	WriteUint32(outFile, e.Ehdr.E_phoff)
	WriteUint32(outFile, e.Ehdr.E_shoff)
	WriteUint32(outFile, e.Ehdr.E_flags)
	WriteUint16(outFile, e.Ehdr.E_ehsize)
	WriteUint16(outFile, e.Ehdr.E_phentsize)
	WriteUint16(outFile, e.Ehdr.E_phnum)
	WriteUint16(outFile, e.Ehdr.E_shentsize)
	WriteUint16(outFile, e.Ehdr.E_shnum)
	WriteUint16(outFile, e.Ehdr.E_shstrndx)
}

// writeTextSectionToFile 写入.text段
func (e *ElfFile) writeTextSectionToFile(outFile *os.File, finName string) {
	// 打开临时文件
	tempName := finName[:strings.LastIndex(finName, ".")] + ".t"
	tempFile, err := os.Open(tempName)
	if err != nil {
		fmt.Printf("无法打开临时文件: %s\n", err)
		return
	}
	defer tempFile.Close()

	// 复制临时文件内容到输出文件
	buffer := make([]byte, 1024)
	for {
		n, err := tempFile.Read(buffer)
		if err != nil || n == 0 {
			break
		}
		_, _ = outFile.Write(buffer[:n])
	}
}

// writeElfTailToFile 写入ELF文件尾部
func (e *ElfFile) writeElfTailToFile(outFile *os.File) {
	// 写入.shstrtab段
	outFile.Write(e.Shstrtab)

	// 写入段表
	for _, name := range e.ShdrNames {
		shdr := e.ShdrTab[name]
		WriteUint32(outFile, shdr.Sh_name)
		WriteUint32(outFile, shdr.Sh_type)
		WriteUint32(outFile, shdr.Sh_flags)
		WriteUint32(outFile, shdr.Sh_addr)
		WriteUint32(outFile, shdr.Sh_offset)
		WriteUint32(outFile, shdr.Sh_size)
		WriteUint32(outFile, shdr.Sh_link)
		WriteUint32(outFile, shdr.Sh_info)
		WriteUint32(outFile, shdr.Sh_addralign)
		WriteUint32(outFile, shdr.Sh_entsize)
	}

	// 写入符号表
	for _, name := range e.SymNames {
		sym := e.SymTab[name]
		WriteUint32(outFile, sym.St_name)
		WriteUint32(outFile, sym.St_value)
		WriteUint32(outFile, sym.St_size)
		outFile.Write([]byte{sym.St_info, sym.St_other})
		WriteUint16(outFile, sym.St_shndx)
	}

	// 写入字符串表
	outFile.Write(e.Strtab)

	// 写入重定位表
	// 先写.rel.text
	//for _, rel := range e.RelTab {
	//	if rel.TarSeg == ".text" {
	//		symIdx := e.getSymIndex(rel.LbName)
	//		info := uint32(symIdx<<8) | uint32(rel.Type)
	//		WriteUint32(outFile, uint32(rel.Offset))
	//		WriteUint32(outFile, info)
	//	}
	//}
	//
	//// 再写.rel.data
	//for _, rel := range e.RelTab {
	//	if rel.TarSeg == ".data" {
	//		symIdx := e.getSymIndex(rel.LbName)
	//		info := uint32(symIdx<<8) | uint32(rel.Type)
	//		WriteUint32(outFile, uint32(rel.Offset))
	//		WriteUint32(outFile, info)
	//	}
	//}
}

var ObjFile = NewElfFile()
