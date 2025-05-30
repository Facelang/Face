package elf

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
)

func Parse(file string) {
	reader := NewReader(file)

	ParseElf(reader)

	//hdr.E_Magic = [16]byte(reader.readMagic())
	//e, t := hdr.Endian(), hdr.Type()

	//hdr.Objdump()

}

func ParseElf(reader FileReader) interface{} {
	bytes, err := reader.segment(0, 16)
	if err != nil {
		panic(err)
	}
	magic := Elf_Magic(bytes.buf)
	bits := magic.Bits()
	endian := magic.Endian()

	if bits == 1 {
		bytes, err = reader.segment(16, 52-16)
	} else if bits == 2 {
		bytes, err = reader.segment(16, 64-16)
	} else {
		panic("不支持的文件结构！unsupport bits")
	}

	hdr := ParseElf_Ehdr(bytes, bits, endian)
	hdr.E_Magic = magic
	hdr.Objdump()

	ParseElf_Shdr(reader, hdr)

	if err != nil || bytes == nil {
		panic("文件解析异常！")
	}

	return nil
}

func ParseElf_Ehdr(bytes BytesReader, bits, endian int) *Elf64_Ehdr {
	hdr := &Elf64_Ehdr{}
	hdr.E_Type = bytes.readUint16(endian)
	hdr.E_Machine = bytes.readUint16(endian)
	hdr.E_Version = bytes.readUint32(endian)
	hdr.E_Entry = bytes.readUintAuto(bits, endian)
	hdr.E_Phoff = bytes.readUintAuto(bits, endian)
	hdr.E_Shoff = bytes.readUintAuto(bits, endian)
	hdr.E_Flags = bytes.readUint32(endian)
	hdr.E_Ehsize = bytes.readUint16(endian)
	hdr.E_Phentsize = bytes.readUint16(endian)
	hdr.E_Phnum = bytes.readUint16(endian)
	hdr.E_Shentsize = bytes.readUint16(endian)
	hdr.E_Shnum = bytes.readUint16(endian)
	hdr.E_Shstrndx = bytes.readUint16(endian)
	return hdr
}

func ParseElf_Shdr(reader FileReader, hdr *Elf64_Ehdr) {
	bits := hdr.E_Magic.Bits()
	endian := hdr.E_Magic.Endian()

	// 先要读取 .shstrtab
	offset := int(hdr.E_Shoff)
	shnum := int(hdr.E_Shnum)
	shstrndx := int(hdr.E_Shstrndx)
	shentsize := int(hdr.E_Shentsize)
	off := offset + shstrndx*shentsize
	bytes, err := reader.segment(off, shentsize)
	if err != nil {
		panic(err)
	}
	el := ParseElf_Shdr_El(bytes, bits, endian)

	// 读取字符串信息
	bytes, err = reader.segment(int(el.Sh_Offset), int(el.Sh_Size))
	if err != nil {
		panic(err)
	}

	// 读取完整段表
	shdrs := make([]*Elf64_Shdr, shnum)
	for index := 0; index < shnum; index++ {
		begin := offset + index*shentsize
		raw, e := reader.segment(begin, shentsize)
		if e != nil {
			panic(e)
		}
		r := ParseElf_Shdr_El(raw, bits, endian)
		shdrs[index] = r
	}

	fmt.Printf("\nELF 段表信息[开始：0x%x, 结束：0x%x]：\n", offset, offset+shnum*shentsize)
	w := tablewriter.NewWriter(os.Stdout)
	w.SetAlignment(tablewriter.ALIGN_RIGHT)
	w.SetHeader([]string{
		"开始地址", "序号", "名称", "类型", "标志", "地址", "位置偏移", "空间大小", "链接", "附加", "对齐", "表项大小",
	})

	for i, shdr := range shdrs {
		shdr.Objdump(offset+i*shentsize, i, w, bytes.buf)
	}

	//el.Objdump(w, bytes.buf)
	w.Render()

}

func ParseElf_Shdr_El(bytes BytesReader, bits, endian int) *Elf64_Shdr {
	shdr := &Elf64_Shdr{}
	shdr.Sh_Name = bytes.readUint32(endian)
	shdr.Sh_Type = bytes.readUint32(endian)
	shdr.Sh_Flags = bytes.readUintAuto(bits, endian)
	shdr.Sh_Addr = bytes.readUintAuto(bits, endian)
	shdr.Sh_Offset = bytes.readUintAuto(bits, endian)
	shdr.Sh_Size = bytes.readUintAuto(bits, endian)
	shdr.Sh_Link = bytes.readUint32(endian)
	shdr.Sh_Info = bytes.readUint32(endian)
	shdr.Sh_Addralign = bytes.readUintAuto(bits, endian)
	shdr.Sh_Entsize = bytes.readUintAuto(bits, endian)
	return shdr
}
