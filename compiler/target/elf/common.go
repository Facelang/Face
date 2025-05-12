package link

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

const EI_NIDENT = 16 // 无论是32位还是16位，都是16字节！

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
