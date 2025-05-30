package elf

import (
	"strconv"
	"strings"
)

type intName struct {
	i uint64
	s string
}

func stringName(i uint64, names []intName, goSyntax bool) string {
	for _, n := range names {
		if n.i == i {
			if goSyntax {
				return "elf." + n.s
			}
			return n.s
		}
	}

	// second pass - look for smaller to add with.
	// assume sorted already
	for j := len(names) - 1; j >= 0; j-- {
		n := names[j]
		if n.i < i {
			s := n.s
			if goSyntax {
				s = "elf." + s
			}
			return s + "+" + strconv.FormatUint(uint64(i-n.i), 10)
		}
	}

	return strconv.FormatUint(uint64(i), 10)
}

var shtStrings = []intName{
	{0, "SHT_NULL"},
	{1, "SHT_PROGBITS"},
	{2, "SHT_SYMTAB"},
	{3, "SHT_STRTAB"},
	{4, "SHT_RELA"},
	{5, "SHT_HASH"},
	{6, "SHT_DYNAMIC"},
	{7, "SHT_NOTE"},
	{8, "SHT_NOBITS"},
	{9, "SHT_REL"},
	{10, "SHT_SHLIB"},
	{11, "SHT_DYNSYM"},
	{14, "SHT_INIT_ARRAY"},
	{15, "SHT_FINI_ARRAY"},
	{16, "SHT_PREINIT_ARRAY"},
	{17, "SHT_GROUP"},
	{18, "SHT_SYMTAB_SHNDX"},
	{0x60000000, "SHT_LOOS"},
	{0x6ffffff5, "SHT_GNU_ATTRIBUTES"},
	{0x6ffffff6, "SHT_GNU_HASH"},
	{0x6ffffff7, "SHT_GNU_LIBLIST"},
	{0x6ffffffd, "SHT_GNU_VERDEF"},
	{0x6ffffffe, "SHT_GNU_VERNEED"},
	{0x6fffffff, "SHT_GNU_VERSYM"},
	{0x70000000, "SHT_LOPROC"},
	{0x7000002a, "SHT_MIPS_ABIFLAGS"},
	{0x7fffffff, "SHT_HIPROC"},
	{0x80000000, "SHT_LOUSER"},
	{0xffffffff, "SHT_HIUSER"},
}

func SectionTypeName(i uint32) string {
	return stringName(uint64(i), shtStrings, false)
}

var shfStrings = []intName{
	{0x1, "SHF_WRITE"},
	{0x2, "SHF_ALLOC"},
	{0x4, "SHF_EXECINSTR"},
	{0x10, "SHF_MERGE"},
	{0x20, "SHF_STRINGS"},
	{0x40, "SHF_INFO_LINK"},
	{0x80, "SHF_LINK_ORDER"},
	{0x100, "SHF_OS_NONCONFORMING"},
	{0x200, "SHF_GROUP"},
	{0x400, "SHF_TLS"},
	{0x800, "SHF_COMPRESSED"},
}

func SectionFlagName(i uint64) string {
	return stringName(i, shfStrings, false)
}

func StringForTable(bytes []byte, start uint32) string {
	builder := strings.Builder{}
	ch := bytes[start]
	offset := start
	for ch != 0 {
		builder.WriteByte(ch)
		offset += 1
		ch = bytes[offset]
	}
	return builder.String()
}
