package elf

import (
	"encoding/binary"
	"fmt"
)

// Elf_Magic elf 文件魔术信息(32bit/64bit 通用)
type Elf_Magic [EI_NIDENT]byte

// 32位 ELF 文件头结构
//type Header32 struct {
//	Magic      [4]byte // ELF 魔数 0x7F, 0x45, 0x4C, 0x46 - 对应ASCII码为 \x7FELF
//	Class      byte    // 文件类型 (32/64位)
//	Data       byte    // 字节序 0x01：小端序(Little Endian)，低字节在前 0x02：大端序(Big Endian)，高字节在前
//	Version    byte    // ELF 版本 通常为0x01，表示原始ELF格式规范版本
//	OSABI      byte    // 操作系统 ABI(0x00：System V 0x01：HP-UX 0x02：NetBSD 0x03：Linux 0x06：Solaris 0x09：FreeBSD 0x0C：OpenBSD)
//	ABIVersion byte    // ABI 版本(通常依赖于特定的ABI，对于System V通常为0x00)
//  第10-16个字节 (9-15)：填充字节 这些字节为保留字节，通常填充为0，保留供将来使用
//	Type       uint16  // 文件类型
//	Machine    uint16  // 机器类型
//	Entry      uint32  // 程序入口点
//	Phoff      uint32  // 程序头表偏移
//	Shoff      uint32  // 节头表偏移
//	Flags      uint32  // 处理器特定标志
//	Ehsize     uint16  // ELF 头大小
//	Phentsize  uint16  // 程序头表项大小
//	Phnum      uint16  // 程序头表项数量
//	Shentsize  uint16  // 节头表项大小
//	Shnum      uint16  // 节头表项数量
//	Shstrndx   uint16  // 节名字符串表索引
//}

func (m Elf_Magic) Bits() int {
	return int(m[4])
}

func (m Elf_Magic) Endian() binary.ByteOrder {
	if m[5] == 1 {
		return binary.LittleEndian
	} else if m[5] == 2 {
		return binary.BigEndian
	}
	panic("不支持的字节序")
}

type Elf32_Phdr struct {
	Type   Elf32_Word
	Offset Elf32_Off
	VAddr  Elf32_Addr
	Paddr  Elf32_Addr
	Filesz Elf32_Word
	Memsz  Elf32_Word
	Flags  Elf32_Word
	Align  Elf32_Word
}

// Elf32_Ehdr ELF32文件头结构
type Elf32_Ehdr struct {
	Magic     Elf_Magic  // (16)魔数和相关信息
	Type      Elf32_Half // (2) 0 Unknown, 1 32-bit, 2 64-bit
	Machine   Elf32_Half // (2) 架构类型
	Version   Elf32_Word // (4) 0 或者 1
	Entry     Elf32_Addr // (8) [32/64] 入口点虚拟地址(32bit 占32位 64bit占64位)
	Phoff     Elf32_Off  // (8) [32/64] 程序头表偏移(按位占用地址宽度)
	Shoff     Elf32_Off  // (8) [32/64] 节头表偏移(按位占用地址宽度)
	Flags     Elf32_Word // (4) 处理器特定标志
	Ehsize    Elf32_Half // (2) ELF头部大小
	Phentsize Elf32_Half // (2) 程序头表项大小
	Phnum     Elf32_Half // (2) 程序头表项数量
	Shentsize Elf32_Half // (2) 节头表项大小
	Shnum     Elf32_Half // (2) 节头表项数量
	Shstrndx  Elf32_Half // (2) 节头字符串表索引
}

// Elf32_Shdr 段表项结构
type Elf32_Shdr struct {
	Name      Elf32_Word // 段名（4字节，存在于字符串表中的偏移量， shstrtab 也是一个段， shstrndx ）
	Type      Elf32_Word // 段类型 (1表示程序段.text.data 2表示符号段.symtab 3表示串表段.shstrtab 8表示内容段.bss 9表示重定位表段.rel.text.rel.data)
	Flags     Elf32_Word // 段标志 (0表示默认 1表示可写 2表示段加载后需要为之分配空间 4表示可执行)
	Addr      Elf32_Addr // 段虚拟地址 可重定位文件默认为零， 可执行文件由链接器计算地址
	Offset    Elf32_Off  // 段在文件中的偏移
	Size      Elf32_Word // 段的大小，字节单位， SHT_NOBITS 代表没有数据（此时指代加载后占用的内存大小）
	Link      Elf32_Word // 段的链接信息，一般用于描述符号标段和重定位表段的链接信息。
	Info      Elf32_Word // 附加信息
	Addralign Elf32_Word // 对齐要求
	Entsize   Elf32_Word // 表项大小
}

func NewShdr(Type SectionType, Flags SectionFlag, Offset, Size int) *Elf32_Shdr {
	return &Elf32_Shdr{
		Name:      0,
		Type:      Elf32_Word(Type),
		Flags:     Elf32_Word(Flags),
		Addr:      0,
		Offset:    Elf32_Off(Offset),
		Size:      Elf32_Word(Size),
		Link:      0,
		Info:      0,
		Addralign: 4,
		Entsize:   0,
	}
}

// Elf32_Sym ELF32符号表项结构
type Elf32_Sym struct {
	Name  uint32 // 符号名
	Value uint32 // 符号值
	Size  uint32 // 符号大小
	Info  byte   // 符号类型和绑定信息
	Other byte   // 保留
	Shndx uint16 // 符号所在节
}

// Elf32_Rel ELF32重定位表项结构
type Elf32_Rel struct {
	Offset uint32
	Info   uint32
}

type Elf32_RelInfo struct {
	SegName string     // 重定位的目标段名
	Rel     *Elf32_Rel // 重定位信息
	RelName string     // 符号名称
}

// File elf文件类，包含elf文件的重要内容，处理elf文件
type File struct {
	Ehdr         *Elf32_Ehdr            // ELF文件头
	PhdrTab      []*Elf32_Phdr          // 程序头表！
	ShdrTab      map[string]*Elf32_Shdr // 段表映射
	ShdrNames    []string               // 段名列表,  段表名和索引的映射关系，方便符号查询自己的段信息
	SymTab       map[string]*Elf32_Sym  // 符号表映射
	SymNames     []string               // 符号名列表, 符号名与符号表项索引的映射关系，对于重定位表生成重要
	RelTab       []*Elf32_RelInfo       // 重定位信息列表,// 省略 辅助数据 char *elf_dir;			   // 处理elf文件的目录
	Name         string                 // 文件名称
	Reader       BytesReader            // 缓存s
	Shstrtab     []byte                 // 段表字符串表数据
	ShstrtabSize int                    // 段表字符串表长
	Strtab       []byte                 // 字符串表数据
	StrtabSize   int                    // 字符串表长
	ProgSegList  []*ProgSeg             // 程序头表缓存数据
}

func NewElfFile(magic Elf_Magic, eType, eMachine Elf32_Half) *File {
	file := &File{
		Ehdr: &Elf32_Ehdr{
			Magic:     magic,                  // 这个字段比较复杂
			Type:      eType,                  // 文件类型： 1表示可重定位, 2表示可执行 3表示共享目标 4 表示核心转储  0 表示无效
			Machine:   eMachine,               // 机器类型
			Version:   Elf32_Word(EV_CURRENT), // 文件版本 一般取1
			Entry:     0,                      // 程序入口的线性地址，一般用于可以执行文件， 可重定向文件该字段为 0
			Phoff:     0,                      // 程序头表在文件内的偏移地址， 标识了程序头表在文件内的位置
			Flags:     0,                      // 文件平台相关属性， 一般默认为 0 (x86 应该没用到)
			Ehsize:    52,                     // 文件头的大小 (跟系统位数有关 32位52字节 64位64字节)
			Phentsize: 0,                      // 程序头表项的大小
			Phnum:     0,                      // 程序头表项的个数，确定程序头表在文件[phoff: phoff + phentsize*phnum] 的数据块中
			Shentsize: 40,                     // 段表项的大小
			Shnum:     0,                      // 段表项的个数， 确定数据区块存在于 [shoff:shoff+shentsize*eshnum] 中
			Shstrndx:  0,                      // .shstrtab的索引
		},
		ShdrTab:     make(map[string]*Elf32_Shdr),
		ShdrNames:   make([]string, 0),
		SymTab:      make(map[string]*Elf32_Sym),
		SymNames:    make([]string, 0),
		RelTab:      make([]*Elf32_RelInfo, 0),
		Shstrtab:    make([]byte, 0),
		Strtab:      make([]byte, 0),
		ProgSegList: make([]*ProgSeg, 0),
	}

	// 初始化ELF魔数
	//Ehdr.Magic[0] = 0x7F // DEL
	//Ehdr.Magic[1] = 'E'  // .
	//Ehdr.Magic[2] = 'L'  // .
	//Ehdr.Magic[3] = 'F'  // .
	//Ehdr.Magic[4] = 1    // Class 32位格式 64位(2) 0表示无效
	//Ehdr.Magic[5] = 1    // 小端序 大端(2) 0表示无效
	//Ehdr.Magic[6] = 1    // ELF版本 默认为1
	// 后面9字节在ELF标准中未定义， 一般用于平台相关的扩展标志
	// 第8字节 取0 表示 unix 系统
	// 第9字节 取0 表示系统 ABI 版本为 0
	// 其它字节默认为 0

	// 添加空节表项(重定位文件和可执行文件都有)
	file.AddShdr("", &Elf32_Shdr{})

	// 添加空符号表项
	file.AddSym("", nil)

	return file
}

func (e *File) Bits() int {
	return e.Ehdr.Magic.Bits()
}

func (e *File) Endian() binary.ByteOrder { return e.Ehdr.Magic.Endian() }

func (e *File) AddShdr(shName string, shdr *Elf32_Shdr) {
	if shdr != nil {
		e.ShdrTab[shName] = shdr
	}
	e.ShdrNames = append(e.ShdrNames, shName)
}

// AddShdrSec sh_name和sh_offset都需要重新计算 todo
func (e *File) AddShdrSec(section *Section, offset int) {
	if section.Name == ".text" {
		e.AddShdr(section.Name,
			NewShdr(SHT_PROGBITS, SHF_ALLOC|SHF_EXECINSTR, offset, section.Length),
		)
	} else if section.Name == ".data" {
		e.AddShdr(section.Name,
			NewShdr(SHT_PROGBITS, SHF_ALLOC|SHF_WRITE, offset, section.Length),
		)
	} else if section.Name == ".bss" { // 非必须
		// 关于 .bss 段： 用于存储未初始化的全局变量和静态变量
		// 特点：在程序价值时会被自动初始化为 0
		// 优势：节省可执行文件空间，只占用很少部分（通常只记录大小）
		// 场景：大小数组或缓冲区的申明， 未初始化的全局变量，未初始化的静态局部变量， 需要零初始化的数据结构
		// 语法: buffer: resw 1024 // 记录 Buffer 符号 需要 resw 宽度 * 1024 空间 (resw 等同于 dw)
		e.AddShdr(section.Name,
			NewShdr(SHT_NOBITS, SHF_ALLOC|SHF_WRITE, offset, section.Length),
		)
	}
}

// AddPhdrRec 添加程序头表
func (e *File) AddPhdr(t Elf32_Word, off Elf32_Off, vaddr Elf32_Addr, filesz, memsz, flags, align Elf32_Word) {
	ph := &Elf32_Phdr{
		Type:   t,
		Offset: off,
		VAddr:  vaddr,
		Filesz: filesz,
		Memsz:  memsz,
		Flags:  flags,
		Align:  align,
	}
	e.PhdrTab = append(e.PhdrTab, ph)
}

// AddProgSeg 添加程序头表, 同时添加段表
func (e *File) AddProgSeg(name string, seg *ProgSeg) {
	flags := PF_W | PF_R // 可读、可写
	filesz := seg.Size   // 占用磁盘大小（合并后的大小）
	if name == ".text" {
		flags = PF_X | PF_R //.text段可读可执行
	}
	if name == ".bss" {
		filesz = 0 // .bss段不占磁盘空间
	}

	seg.Name = name
	e.ProgSegList = append(e.ProgSegList, seg)
	e.AddPhdr(Elf32_Word(PT_LOAD), seg.Offset, seg.BaseAddr,
		filesz, seg.Size, Elf32_Word(flags), MemAlign)

	shType := SHT_PROGBITS
	shFlags := SHF_ALLOC | SHF_WRITE
	shAlign := 4 //4B
	if name == ".bss" {
		shType = SHT_NOBITS
	}
	if name == ".text" {
		shFlags = SHF_ALLOC | SHF_EXECINSTR
		shAlign = 16
	}
	// 添加程序头表也要添加对应的段
	//添加一个段表项，暂时按照4字节对齐
	shdr := NewShdr(shType, shFlags, int(seg.Offset), int(seg.Size))
	shdr.Addr = seg.BaseAddr
	shdr.Addralign = Elf32_Word(shAlign)
	e.AddShdr(name, shdr)
}

func (e *File) AddSym(name string, sym *Elf32_Sym) {
	target := &Elf32_Sym{
		Name:  0,
		Value: 0,
		Size:  0,
		Info:  0,
		Other: 0,
		Shndx: 0,
	}
	if name != "" {
		target.Value = sym.Value
		target.Size = sym.Size
		target.Info = sym.Info
		target.Other = sym.Other
		target.Shndx = sym.Shndx
	}
	e.SymTab[name] = target
	e.SymNames = append(e.SymNames, name)
}

func (e *File) AddRel(info *Elf32_RelInfo) {
	e.RelTab = append(e.RelTab, info)
}

func (e *File) GetSegIndex(seg string) int {
	for i, name := range e.ShdrNames {
		if name == seg {
			return i
		}
	}
	return -1
}

func (e *File) GetSymIndex(sym string) int {
	for i, name := range e.SymNames {
		if name == sym {
			return i
		}
	}
	return -1
}

func (e *File) ReadData(offset Elf32_Off, size Elf32_Word) []byte {
	return e.Reader.Data(int(offset), int(size))
}

func (e *File) ReadDataBy(seg string) []byte {
	section := e.ShdrTab[seg]
	return e.Reader.Data(int(section.Offset), int(section.Size))
}

func (e *File) WriteFile(target string) error {
	return FileWrite(e, target)
}

/*
	dir:输出目录
	flag:1-第一次写，文件头+PHT；2-第二次写，段表字符串表+段表+符号表+字符串表；
*/
//void Elf_file::writeElf(const char*dir,int flag)
//{
//if(flag==1)
//{
//FILE*fp=fopen(dir,"w+");
//fwrite(&ehdr,ehdr.e_ehsize,1,fp);//elf文件头
//if(!phdrTab.empty())//程序头表
//{
//for(int i=0;i<phdrTab.size();++i)
//fwrite(phdrTab[i],ehdr.e_phentsize,1,fp);
//}
//fclose(fp);
//}
//else if(flag==2)
//{
//FILE*fp=fopen(dir,"a+");
//fwrite(shstrtab,shstrtabSize,1,fp);//.shstrtab
//for(int i=0;i<shdrNames.size();++i)//段表
//{
//Elf32_Shdr*sh=shdrTab[shdrNames[i]];
//fwrite(sh,ehdr.e_shentsize,1,fp);
//}
//for(int i=0;i<symNames.size();++i)//符号表
//{
//Elf32_Sym*sym=symTab[symNames[i]];
//fwrite(sym,sizeof(Elf32_Sym),1,fp);
//}
//fwrite(strtab,strtabSize,1,fp);//.strtab
//fclose(fp);
//}
//}

func (e *File) Objdump() {
	fmt.Printf("\nELF 文件头信息：\n")
	fmt.Printf("\t魔术信息：")
	for _, b := range e.Ehdr.Magic {
		fmt.Printf("%02x ", b) // %02x 保证两位数，不足补零
	}
	fmt.Printf("\n")
	fmt.Printf("\t文件类型：%d\n", e.Ehdr.Type)
	fmt.Printf("\t架构：%d\n", e.Ehdr.Machine)
	fmt.Printf("\t版本号：%d\n", e.Ehdr.Version)
	fmt.Printf("\t入口地址：0x%x(%d)\n", e.Ehdr.Entry, e.Ehdr.Entry)
	fmt.Printf("\t程序头表偏移地址：0x%x(%d)\n", e.Ehdr.Phoff, e.Ehdr.Phoff)
	fmt.Printf("\t段表偏移地址：0x%x(%d)\n", e.Ehdr.Shoff, e.Ehdr.Shoff)
	fmt.Printf("\tFlags 标志信息：%d\n", e.Ehdr.Flags)
	fmt.Printf("\t文件头大小：%d bytes\n", e.Ehdr.Ehsize)
	fmt.Printf("\t程序头表项大小：%d bytes\n", e.Ehdr.Phentsize)
	fmt.Printf("\t程序头表项数：%d\n", e.Ehdr.Phnum)
	fmt.Printf("\t段表项大小：%d bytes\n", e.Ehdr.Shentsize)
	fmt.Printf("\t段表项数：%d\n", e.Ehdr.Ehsize)
	fmt.Printf("\t节头字符串表索引：%d\n", e.Ehdr.Shstrndx)

	//offset := int(e.Ehdr.Shoff)
	//shnum := int(e.Ehdr.Shnum)
	//shentsize := int(e.Ehdr.Shentsize)
	//fmt.Printf("\nELF 段表信息[开始：0x%x, 长度：0x%x]：\n", offset, shnum*shentsize)
	//w := tablewriter.NewWriter(os.Stdout)
	//w.SetAlignment(tablewriter.ALIGN_RIGHT)
	//w.SetHeader([]string{
	//	"开始地址", "序号", "名称", "类型", "标志", "地址", "位置偏移", "空间大小", "链接", "附加", "对齐", "表项大小",
	//})
	//for i, name := range e.ShdrNames {
	//	e := e.ShdrTab[name]
	//	w.Append([]string{
	//		fmt.Sprintf("0x%x", offset+i*shentsize),
	//		fmt.Sprintf("[%d]", i),
	//		name,
	//		SectionTypeName(e.Type),
	//		SectionFlagName(uint64(e.Flags)),
	//		fmt.Sprintf("0x%x", e.Addr),
	//		fmt.Sprintf("0x%x", e.Offset),
	//		fmt.Sprintf("%d bytes", e.Size),
	//		fmt.Sprintf("%d", e.Link),
	//		fmt.Sprintf("%d", e.Info),
	//		fmt.Sprintf("%d", e.Addralign),
	//		fmt.Sprintf("%d bytes", e.Entsize),
	//	})
	//}
	//w.Render()
	//
	//// 打印符号表【段】
	//symTabInfo := e.ShdrTab[".symtab"]
	//offset = int(symTabInfo.Offset)
	//fmt.Printf("\nELF 符号表信息[开始：0x%x, 长度：0x%x]：\n", offset, symTabInfo.Size)
	//w = tablewriter.NewWriter(os.Stdout)
	//w.SetHeader([]string{
	//	"开始地址", "序号", "名称", "地址", "空间大小", "类型和绑定", "其它", "所在节",
	//})
	//for i, name := range e.SymNames {
	//	sym := e.SymTab[name]
	//	segment := fmt.Sprintf("%d", sym.Shndx)
	//	if sym.Shndx > 0 {
	//		sh := e.ShdrNames[sym.Shndx]
	//		segment = fmt.Sprintf("%s,%s", segment, sh)
	//	}
	//	w.Append([]string{
	//		fmt.Sprintf("0x%x", offset+i*16),
	//		fmt.Sprintf("[%d]", i),
	//		name,
	//		fmt.Sprintf("0x%x", sym.Value),
	//		fmt.Sprintf("%d bytes", sym.Size),
	//		fmt.Sprintf("0b%b", sym.Info),
	//		fmt.Sprintf("0b%b", sym.Other),
	//		segment,
	//	})
	//}
	//w.Render()

	// 打印重定位表【段】

	// todo 循环遍历，依次打印段信息

}

//// GetData GetSectionData 获取节数据
//func (f *ElfFile) GetData(seg *Elf32_Shdr) ([]byte, error) {
//	offset := uint64(seg.Offset)
//	size := uint64(seg.Size)
//	os.Open() // 读取数据
//
//	data := make([]byte, size)
//	if _, err := f.FileHandle.Seek(int64(offset), 0); err != nil {
//		fmt.Printf("[DEBUG] 错误: 定位到节偏移失败: %v\n", err)
//		return nil, err
//	}
//	if _, err := io.ReadFull(f.FileHandle, data); err != nil {
//		fmt.Printf("[DEBUG] 错误: 读取节数据失败: %v\n", err)
//		return nil, err
//	}
//	fmt.Printf("[DEBUG] 成功读取节数据, 大小: %d\n", len(data))
//	return data, nil
//}
