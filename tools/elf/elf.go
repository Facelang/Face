package elf

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
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

// Elf_Magic elf 文件魔术信息(32bit/64bit 通用)
type Elf_Magic [EI_NIDENT]byte

func (m Elf_Magic) Bits() int {
	return int(m[4])
}

func (m Elf_Magic) Endian() int {
	return int(m[5])
}

func (m Elf_Magic) Objdump() {
	fmt.Printf("\t魔术信息：")
	for _, b := range m {
		fmt.Printf("%02x ", b) // %02x 保证两位数，不足补零
	}
	fmt.Printf("\n")
}

// Elf32_Ehdr ELF32文件头结构
type Elf32_Ehdr struct {
	E_Magic     Elf_Magic  // (16)魔数和相关信息
	E_Type      Elf32_Half // (2) 0 Unknown, 1 32-bit, 2 64-bit
	E_Machine   Elf32_Half // (2) 架构类型
	E_Version   Elf32_Word // (4) 0 或者 1
	E_Entry     Elf32_Addr // (8) [32/64] 入口点虚拟地址(32bit 占32位 64bit占64位)
	E_Phoff     Elf32_Off  // (8) [32/64] 程序头表偏移(按位占用地址宽度)
	E_Shoff     Elf32_Off  // (8) [32/64] 节头表偏移(按位占用地址宽度)
	E_Flags     Elf32_Word // (4) 处理器特定标志
	E_Ehsize    Elf32_Half // (2) ELF头部大小
	E_Phentsize Elf32_Half // (2) 程序头表项大小
	E_Phnum     Elf32_Half // (2) 程序头表项数量
	E_Shentsize Elf32_Half // (2) 节头表项大小
	E_Shnum     Elf32_Half // (2) 节头表项数量
	E_Shstrndx  Elf32_Half // (2) 节头字符串表索引
}

// Elf64_Ehdr ELF64文件头结构
type Elf64_Ehdr struct {
	E_Magic     Elf_Magic  // (16)魔数和相关信息
	E_Type      Elf64_Half // (2) 0 Unknown, 1 32-bit, 2 64-bit
	E_Machine   Elf64_Half // (2) 架构类型
	E_Version   Elf64_Word // (4) 0 或者 1
	E_Entry     Elf64_Addr // (8) [32/64] 入口点虚拟地址(32bit 占32位 64bit占64位)
	E_Phoff     Elf64_Off  // (8) [32/64] 程序头表偏移(按位占用地址宽度)
	E_Shoff     Elf64_Off  // (8) [32/64] 节头表偏移(按位占用地址宽度)
	E_Flags     Elf64_Word // (4) 处理器特定标志
	E_Ehsize    Elf64_Half // (2) ELF头部大小
	E_Phentsize Elf64_Half // (2) 程序头表项大小
	E_Phnum     Elf64_Half // (2) 程序头表项数量
	E_Shentsize Elf64_Half // (2) 节头表项大小
	E_Shnum     Elf64_Half // (2) 节头表项数量
	E_Shstrndx  Elf64_Half // (2) 节头字符串表索引
}

func (e Elf64_Ehdr) Objdump() {
	fmt.Printf("\nELF 文件头信息：\n")

	e.E_Magic.Objdump()
	fmt.Printf("\t文件类型：%d\n", e.E_Type)
	fmt.Printf("\t架构：%d\n", e.E_Machine)
	fmt.Printf("\t版本号：%d\n", e.E_Version)
	fmt.Printf("\t入口地址：0x%x(%d)\n", e.E_Entry, e.E_Entry)
	fmt.Printf("\t程序头表偏移地址：0x%x(%d)\n", e.E_Phoff, e.E_Phoff)
	fmt.Printf("\t段表偏移地址：0x%x(%d)\n", e.E_Shoff, e.E_Shoff)
	fmt.Printf("\tFlags 标志信息：%d\n", e.E_Flags)
	fmt.Printf("\t文件头大小：%d bytes\n", e.E_Ehsize)
	fmt.Printf("\t程序头表项大小：%d bytes\n", e.E_Phentsize)
	fmt.Printf("\t程序头表项数：%d\n", e.E_Phnum)
	fmt.Printf("\t段表项大小：%d bytes\n", e.E_Shentsize)
	fmt.Printf("\t段表项数：%d\n", e.E_Ehsize)
	fmt.Printf("\t节头字符串表索引：%d\n", e.E_Shstrndx)
}

// Elf32_Shdr 段表项结构
type Elf32_Shdr struct {
	Sh_Name      Elf32_Word // 段名（4字节，存在于字符串表中的偏移量， shstrtab 也是一个段， e_shstrndx ）
	Sh_Type      Elf32_Word // 段类型 (1表示程序段.text.data 2表示符号段.symtab 3表示串表段.shstrtab 8表示内容段.bss 9表示重定位表段.rel.text.rel.data)
	Sh_Flags     Elf32_Word // 段标志 (0表示默认 1表示可写 2表示段加载后需要为之分配空间 4表示可执行)
	Sh_Addr      Elf32_Addr // 段虚拟地址 可重定位文件默认为零， 可执行文件由链接器计算地址
	Sh_Offset    Elf32_Off  // 段在文件中的偏移
	Sh_Size      Elf32_Word // 段的大小，字节单位， SHT_NOBITS 代表没有数据（此时指代加载后占用的内存大小）
	Sh_Link      Elf32_Word // 段的链接信息，一般用于描述符号标段和重定位表段的链接信息。
	Sh_Info      Elf32_Word // 附加信息
	Sh_Addralign Elf32_Word // 对齐要求
	Sh_Entsize   Elf32_Word // 表项大小
}

// Elf64_Shdr 段表项结构
type Elf64_Shdr struct {
	Sh_Name      Elf64_Word  // 段名（4字节，存在于字符串表中的偏移量， shstrtab 也是一个段， e_shstrndx ）
	Sh_Type      Elf64_Word  // 段类型 (1表示程序段.text.data 2表示符号段.symtab 3表示串表段.shstrtab 8表示内容段.bss 9表示重定位表段.rel.text.rel.data)
	Sh_Flags     Elf64_Xword // 段标志 (0表示默认 1表示可写 2表示段加载后需要为之分配空间 4表示可执行)
	Sh_Addr      Elf64_Addr  // 段虚拟地址 可重定位文件默认为零， 可执行文件由链接器计算地址
	Sh_Offset    Elf64_Off   // 段在文件中的偏移
	Sh_Size      Elf64_Xword // 段的大小，字节单位， SHT_NOBITS 代表没有数据（此时指代加载后占用的内存大小）
	Sh_Link      Elf64_Word  // 段的链接信息，一般用于描述符号标段和重定位表段的链接信息。
	Sh_Info      Elf64_Word  // 附加信息
	Sh_Addralign Elf64_Xword // 对齐要求
	Sh_Entsize   Elf64_Xword // 表项大小
}

func (e Elf64_Shdr) Objdump(addr, i int, w *tablewriter.Table, strtab []byte) {
	w.Append([]string{
		fmt.Sprintf("0x%x", addr),
		fmt.Sprintf("[%d]", i),
		StringForTable(strtab, e.Sh_Name),
		SectionTypeName(e.Sh_Type),
		SectionFlagName(e.Sh_Flags),
		fmt.Sprintf("0x%x", e.Sh_Addr),
		fmt.Sprintf("0x%x", e.Sh_Offset),
		fmt.Sprintf("%d bytes", e.Sh_Size),
		fmt.Sprintf("%d", e.Sh_Link),
		fmt.Sprintf("%d", e.Sh_Info),
		fmt.Sprintf("%d", e.Sh_Addralign),
		fmt.Sprintf("%d bytes", e.Sh_Entsize),
	})
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

// RelInfo 重定位信息
type RelInfo struct {
	TarSeg string // 重定位目标段
	Offset int    // 重定位位置的偏移
	LbName string // 重定位符号的名称
	Type   int    // 重定位类型0-R_386_32；1-R_386_PC32
}

// ELF文件结构// elf文件类，包含elf文件的重要内容，处理elf文件
type ElfFile struct {
	Ehdr         Elf32_Ehdr             // ELF文件头
	ShdrTab      map[string]*Elf32_Shdr // 段表映射
	ShdrNames    []string               // 段名列表,  段表名和索引的映射关系，方便符号查询自己的段信息
	SymTab       map[string]*Elf32_Sym  // 符号表映射
	SymNames     []string               // 符号名列表, 符号名与符号表项索引的映射关系，对于重定位表生成重要
	RelTab       []RelInfo              // 重定位信息列表
	Shstrtab     []byte                 // 段表字符串表数据
	ShstrtabSize int                    // 段表字符串表长
	Strtab       []byte                 // 字符串表数据
	StrtabSize   int                    // 字符串表长
	//vector<Elf32_Rel *> relTextTab, relDataTab;
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

// Ehdr
//0000000 457f 464c 0102 0001 0000 0000 0000 0000
//0000010 0001 003e 0001 0000 0000 0000 0000 0000
//0000020 0000 0000 0000 0000 0278 0000 0000 0000
//0000030 0000 0000 0040 0000 0000 0040 000d 000c

// 0x40-:0x7d (61 bytes) .text 代码段
//0000040 4855 e589 8348 10ec 45c7 02fc 0000 8b00
//0000050 0015 0000 8b00 fc45 d001 4589 8bf8 f845
//0000060 c689 8d48 0005 0000 4800 c789 00b8 0000
//0000070 e800 0000 0000 00b8 0000 c900 00c3
// 对齐
//000007d 0000

//example/elf/main.o:     file format elf64-x86-64
//Disassembly of section .text:
//0000000000000000 <main>:
//0: 55                            pushq   %rbp
//1: 48 89 e5                      movq    %rsp, %rbp
//4: 48 83 ec 10                   subq    $0x10, %rsp
//8: c7 45 fc 02 00 00 00          movl    $0x2, -0x4(%rbp)
//f: 8b 15 00 00 00 00             movl    (%rip), %edx            # 0x15 <main+0x15>
//15: 8b 45 fc                      movl    -0x4(%rbp), %eax
//18: 01 d0                         addl    %edx, %eax
//1a: 89 45 f8                      movl    %eax, -0x8(%rbp)
//1d: 8b 45 f8                      movl    -0x8(%rbp), %eax
//20: 89 c6                         movl    %eax, %esi
//22: 48 8d 05 00 00 00 00          leaq    (%rip), %rax            # 0x29 <main+0x29>
//29: 48 89 c7                      movq    %rax, %rdi
//2c: b8 00 00 00 00                movl    $0x0, %eax
//31: e8 00 00 00 00                callq   0x36 <main+0x36>
//36: b8 00 00 00 00                movl    $0x0, %eax
//3b: c9                            leave
//3c: c3                            retq

// 0x80-0x84 (4 bytes) .data 数据段
//0000080 0001 0000

// .bss 为空

// 0x84-0x97 (19 bytes) .rodata
//0000084 aee8 e7a1 97ae bbe7 e693 9c9e
//0000090 bcef 259a 0a64 00

// 0x97-0xb7 (32 bytes) .comment
//0000097 00 4347 3a43 2820 6544
//00000a0 6962 6e61 3120 2e32 2e32 2d30 3431 2029
//00000b0 3231 322e 302e 00
// 对齐
//00000b7 00

// 0xb8-0xf0 (56 bytes) .eh_frame
//00000b8 0014 0000 0000 0000
//00000c0 7a01 0052 7801 0110 0c1b 0807 0190 0000
//00000d0 001c 0000 001c 0000 0000 0000 003d 0000
//00000e0 4100 100e 0286 0d43 7806 070c 0008 0000

// 0xf0-0x198 (168 bytes) .symtab
//00000f0 0000 0000 0000 0000 0000 0000 0000 0000
//0000100 0000 0000 0000 0000 0001 0000 0004 fff1
//0000110 0000 0000 0000 0000 0000 0000 0000 0000
//0000120 0000 0000 0003 0001 0000 0000 0000 0000
//0000130 0000 0000 0000 0000 0000 0000 0003 0005
//0000140 0000 0000 0000 0000 0000 0000 0000 0000
//0000150 0008 0000 0011 0003 0000 0000 0000 0000
//0000160 0004 0000 0000 0000 000a 0000 0012 0001
//0000170 0000 0000 0000 0000 003d 0000 0000 0000
//0000180 000f 0000 0010 0000 0000 0000 0000 0000
//0000190 0000 0000 0000 0000

// 0x198-0x1ae (22 bytes) .strtab
//0000198 6d00 6961 2e6e 0063
//00001a0 0061 616d 6e69 7000 6972 746e 0066
// 对齐
//00001ae 0000

// 0x1b0-0x1f8 (72bytes) .rela.text
//00001b0 0011 0000 0000 0000 0002 0000 0004 0000
//00001c0 fffc ffff ffff ffff 0025 0000 0000 0000
//00001d0 0002 0000 0003 0000 fffc ffff ffff ffff
//00001e0 0032 0000 0000 0000 0004 0000 0006 0000
//00001f0 fffc ffff ffff ffff

// 0x1f8-0x210 (2 4bytes) .rela.eh_frame
//00001f8 0020 0000 0000 0000
//0000200 0002 0000 0002 0000 0000 0000 0000 0000

// 0x210-0x271 (97 bytes) .shstrtab
//0000210 2e00 7973 746d 6261 2e00 7473 7472 6261
//0000220 2e00 6873 7473 7472 6261 2e00 6572 616c
//0000230 742e 7865 0074 642e 7461 0061 622e 7373
//0000240 2e00 6f72 6164 6174 2e00 6f63 6d6d 6e65
//0000250 0074 6e2e 746f 2e65 4e47 2d55 7473 6361
//0000260 006b 722e 6c65 2e61 6865 665f 6172 656d
//0000270 00
// 对齐
//0000271 00 0000 0000 0000

// 段表信息 0x278-0x5b8
//0000278 0000 0000 0000 0000
//* // (这个应该代表都是0)
//00002b0 0000 0000 0000 0000 0020 0000 0001 0000 // :0x2b8 NULL
//00002c0 0006 0000 0000 0000 0000 0000 0000 0000
//00002d0 0040 0000 0000 0000 003d 0000 0000 0000
//00002e0 0000 0000 0000 0000 0001 0000 0000 0000
//00002f0 0000 0000 0000 0000 001b 0000 0004 0000 // :0x2f8 .text [0x40(64)+61bytes] 头文件结束部分开始
//0000300 0040 0000 0000 0000 0000 0000 0000 0000 //        结束位置：0x7d(125)
//0000310 01b0 0000 0000 0000 0048 0000 0000 0000
//0000320 000a 0000 0001 0000 0008 0000 0000 0000
//0000330 0018 0000 0000 0000 0026 0000 0001 0000 // :0x338 .rela.text [0x1b0(432)+72bytes] 可重定位代码
//0000340 0003 0000 0000 0000 0000 0000 0000 0000 //        结束位置：0x1f8(504)
//0000350 0080 0000 0000 0000 0004 0000 0000 0000
//0000360 0000 0000 0000 0000 0004 0000 0000 0000
//0000370 0000 0000 0000 0000 002c 0000 0008 0000 // :0x378 .data [0x80(128)+4bytes] 可重定位代码
//0000380 0003 0000 0000 0000 0000 0000 0000 0000 //        结束位置：0x84(132)
//0000390 0084 0000 0000 0000 0000 0000 0000 0000
//00003a0 0000 0000 0000 0000 0001 0000 0000 0000
//00003b0 0000 0000 0000 0000 0031 0000 0001 0000 // :0x3b8 .bss [0x84]
//00003c0 0002 0000 0000 0000 0000 0000 0000 0000
//00003d0 0084 0000 0000 0000 0013 0000 0000 0000
//00003e0 0000 0000 0000 0000 0001 0000 0000 0000
//00003f0 0000 0000 0000 0000 0039 0000 0001 0000 // :0x3f8 .rodata [0x84+19bytes]
//0000400 0030 0000 0000 0000 0000 0000 0000 0000 //        结束位置：0x97(151)
//0000410 0097 0000 0000 0000 0020 0000 0000 0000
//0000420 0000 0000 0000 0000 0001 0000 0000 0000
//0000430 0001 0000 0000 0000 0042 0000 0001 0000 // :0x438 .comment [0x97+32bytes]
//0000440 0000 0000 0000 0000 0000 0000 0000 0000 //        结束位置：0xb7(183)
//0000450 00b7 0000 0000 0000 0000 0000 0000 0000
//0000460 0000 0000 0000 0000 0001 0000 0000 0000
//0000470 0000 0000 0000 0000 0057 0000 0001 0000 // :0x478 .note.GUN-stack [0xb7+0]
//0000480 0002 0000 0000 0000 0000 0000 0000 0000 //        结束位置：0xb7(183)
//0000490 00b8 0000 0000 0000 0038 0000 0000 0000
//00004a0 0000 0000 0000 0000 0008 0000 0000 0000
//00004b0 0000 0000 0000 0000 0052 0000 0004 0000 // :0x4b8 .eh_frame [0xb8(184)+56bytes]
//00004c0 0040 0000 0000 0000 0000 0000 0000 0000 //        结束位置：0xf0(240)
//00004d0 01f8 0000 0000 0000 0018 0000 0000 0000
//00004e0 000a 0000 0008 0000 0008 0000 0000 0000
//00004f0 0018 0000 0000 0000 0001 0000 0002 0000 // :0x4f8 .rela.eh_frame [0x1f8(504)+24bytes]
//0000500 0000 0000 0000 0000 0000 0000 0000 0000 //        结束位置：0x210(528)
//0000510 00f0 0000 0000 0000 00a8 0000 0000 0000
//0000520 000b 0000 0004 0000 0008 0000 0000 0000
//0000530 0018 0000 0000 0000 0009 0000 0003 0000 // :0x538 .symtab [0xf0(240)+168bytes]
//0000540 0000 0000 0000 0000 0000 0000 0000 0000 //        结束位置：0x198(408)
//0000550 0198 0000 0000 0000 0016 0000 0000 0000
//0000560 0000 0000 0000 0000 0001 0000 0000 0000
//0000570 0000 0000 0000 0000 0011 0000 0003 0000 // :0x578 .strtab [0x198+22bytes]
//0000580 0000 0000 0000 0000 0000 0000 0000 0000 //        结束位置：0x1ae(430)
//0000590 0210 0000 0000 0000 0061 0000 0000 0000
//00005a0 0000 0000 0000 0000 0001 0000 0000 0000
//00005b0 0000 0000 0000 0000                     // :0x5b8 .shstrtab [0x210(528)+97bytes] [结束：0x271(625)]

//Symbol table '.symtab' contains 7 entries:
//Num:    Value          Size Type    Bind   Vis      Ndx Name
//0: 0000000000000000     0 NOTYPE  LOCAL  DEFAULT  UND
//1: 0000000000000000     0 FILE    LOCAL  DEFAULT  ABS main.c
//2: 0000000000000000     0 SECTION LOCAL  DEFAULT    1 .text
//3: 0000000000000000     0 SECTION LOCAL  DEFAULT    5 .rodata
//4: 0000000000000000     4 OBJECT  GLOBAL DEFAULT    3 a
//5: 0000000000000000    61 FUNC    GLOBAL DEFAULT    1 main
//6: 0000000000000000     0 NOTYPE  GLOBAL DEFAULT  UND printf
//Admin@Facade:~/projects/elf$ readelf -r main.o
//
//Relocation section '.rela.text' at offset 0x1b0 contains 3 entries:
//Offset          Info           Type           Sym. Value    Sym. Name + Addend
//000000000011  000400000002 R_X86_64_PC32     0000000000000000 a - 4
//000000000025  000300000002 R_X86_64_PC32     0000000000000000 .rodata - 4
//000000000032  000600000004 R_X86_64_PLT32    0000000000000000 printf - 4
//
//Relocation section '.rela.eh_frame' at offset 0x1f8 contains 1 entry:
//Offset          Info           Type           Sym. Value    Sym. Name + Addend
//000000000020  000200000002 R_X86_64_PC32     0000000000000000 .text + 0
