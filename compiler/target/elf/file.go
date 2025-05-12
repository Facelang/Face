package link

import (
	"encoding/binary"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
)

// Elf_Magic elf 文件魔术信息(32bit/64bit 通用)
type Elf_Magic [EI_NIDENT]byte

// 32位 ELF 文件头结构
//type Header32 struct {
//	Magic      [4]byte // ELF 魔数
//	Class      byte    // 文件类型 (32/64位)
//	Data       byte    // 数据编码方式
//	Version    byte    // ELF 版本
//	OSABI      byte    // 操作系统 ABI
//	ABIVersion byte    // ABI 版本
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

func NewElfEhdr(reader BytesReader, bits int) *Elf32_Ehdr {
	return &Elf32_Ehdr{
		Type:      reader.Uint16(),
		Machine:   reader.Uint16(),
		Version:   reader.Uint32(),
		Entry:     reader.Uint32(), // auto
		Phoff:     reader.Uint32(), // auto
		Shoff:     reader.Uint32(), // auto
		Flags:     reader.Uint32(),
		Ehsize:    reader.Uint16(),
		Phentsize: reader.Uint16(),
		Phnum:     reader.Uint16(),
		Shentsize: reader.Uint16(),
		Shnum:     reader.Uint16(),
		Shstrndx:  reader.Uint16(),
	}
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

func NewElfShdr(reader BytesReader, bits int) *Elf32_Shdr {
	return &Elf32_Shdr{
		Name:      reader.Uint32(),
		Type:      reader.Uint32(),
		Flags:     reader.Uint32(), // auto
		Addr:      reader.Uint32(), // auto
		Offset:    reader.Uint32(), // auto
		Size:      reader.Uint32(), // auto
		Link:      reader.Uint32(),
		Info:      reader.Uint32(),
		Addralign: reader.Uint32(), // auto
		Entsize:   reader.Uint32(), // auto
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

func NewElfSym(reader BytesReader, bits int) *Elf32_Sym {
	return &Elf32_Sym{
		Name:  reader.Uint32(),
		Value: reader.Uint32(),
		Size:  reader.Uint32(),
		Info:  reader.Byte(),
		Other: reader.Byte(),
		Shndx: reader.Uint16(),
	}
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

func NewElfRel(reader BytesReader, bits int) *Elf32_Rel {
	return &Elf32_Rel{
		Offset: reader.Uint32(),
		Info:   reader.Uint32(),
	}
}

// ElfFile elf文件类，包含elf文件的重要内容，处理elf文件
type ElfFile struct {
	Ehdr         *Elf32_Ehdr            // ELF文件头
	PhdrTab      []interface{}          // 程序头表！
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
}

func (file *ElfFile) Objdump() {

}

func (file *ElfFile) Objdump() {
	fmt.Printf("\nELF 文件头信息：\n")
	fmt.Printf("\t魔术信息：")
	for _, b := range file.Ehdr.Magic {
		fmt.Printf("%02x ", b) // %02x 保证两位数，不足补零
	}
	fmt.Printf("\n")
	fmt.Printf("\t文件类型：%d\n", file.Ehdr.Type)
	fmt.Printf("\t架构：%d\n", file.Ehdr.Machine)
	fmt.Printf("\t版本号：%d\n", file.Ehdr.Version)
	fmt.Printf("\t入口地址：0x%x(%d)\n", file.Ehdr.Entry, file.Ehdr.Entry)
	fmt.Printf("\t程序头表偏移地址：0x%x(%d)\n", file.Ehdr.Phoff, file.Ehdr.Phoff)
	fmt.Printf("\t段表偏移地址：0x%x(%d)\n", file.Ehdr.Shoff, file.Ehdr.Shoff)
	fmt.Printf("\tFlags 标志信息：%d\n", file.Ehdr.Flags)
	fmt.Printf("\t文件头大小：%d bytes\n", file.Ehdr.Ehsize)
	fmt.Printf("\t程序头表项大小：%d bytes\n", file.Ehdr.Phentsize)
	fmt.Printf("\t程序头表项数：%d\n", file.Ehdr.Phnum)
	fmt.Printf("\t段表项大小：%d bytes\n", file.Ehdr.Shentsize)
	fmt.Printf("\t段表项数：%d\n", file.Ehdr.Ehsize)
	fmt.Printf("\t节头字符串表索引：%d\n", file.Ehdr.Shstrndx)

	offset := int(file.Ehdr.Shoff)
	shnum := int(file.Ehdr.Shnum)
	shentsize := int(file.Ehdr.Shentsize)
	fmt.Printf("\nELF 段表信息[开始：0x%x, 长度：0x%x]：\n", offset, shnum*shentsize)
	w := tablewriter.NewWriter(os.Stdout)
	w.SetAlignment(tablewriter.ALIGN_RIGHT)
	w.SetHeader([]string{
		"开始地址", "序号", "名称", "类型", "标志", "地址", "位置偏移", "空间大小", "链接", "附加", "对齐", "表项大小",
	})
	for i, name := range file.ShdrNames {
		e := file.ShdrTab[name]
		w.Append([]string{
			fmt.Sprintf("0x%x", offset+i*shentsize),
			fmt.Sprintf("[%d]", i),
			name,
			SectionTypeName(e.Type),
			SectionFlagName(uint64(e.Flags)),
			fmt.Sprintf("0x%x", e.Addr),
			fmt.Sprintf("0x%x", e.Offset),
			fmt.Sprintf("%d bytes", e.Size),
			fmt.Sprintf("%d", e.Link),
			fmt.Sprintf("%d", e.Info),
			fmt.Sprintf("%d", e.Addralign),
			fmt.Sprintf("%d bytes", e.Entsize),
		})
	}
	w.Render()

	// 打印符号表【段】
	symTabInfo := file.ShdrTab[".symtab"]
	offset = int(symTabInfo.Offset)
	fmt.Printf("\nELF 符号表信息[开始：0x%x, 长度：0x%x]：\n", offset, symTabInfo.Size)
	w = tablewriter.NewWriter(os.Stdout)
	w.SetHeader([]string{
		"开始地址", "序号", "名称", "地址", "空间大小", "类型和绑定", "其它", "所在节",
	})
	for i, name := range file.SymNames {
		e := file.SymTab[name]
		segment := fmt.Sprintf("%d", e.Shndx)
		if e.Shndx > 0 {
			sh := file.ShdrNames[e.Shndx]
			segment = fmt.Sprintf("%s,%s", segment, sh)
		}
		w.Append([]string{
			fmt.Sprintf("0x%x", offset+i*16),
			fmt.Sprintf("[%d]", i),
			name,
			fmt.Sprintf("0x%x", e.Value),
			fmt.Sprintf("%d bytes", e.Size),
			fmt.Sprintf("0b%b", e.Info),
			fmt.Sprintf("0b%b", e.Other),
			segment,
		})
	}
	w.Render()

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
