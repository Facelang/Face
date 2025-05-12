package link

import (
	"encoding/binary"
	"io"
	"os"
	"strings"
)

type bytesReader struct {
	buf    []byte           // 字节数组
	r, e   int              // 读取器游标
	reader binary.ByteOrder // 读取器
}

type BytesReader = *bytesReader

// 返回值是否为 eof
func (r *bytesReader) read() (byte, error) {
	if r.r == r.e { // 读到头了
		return 0, io.EOF
	}

	ch := r.buf[r.r]
	r.r += 1
	return ch, nil
}

func (r *bytesReader) Byte() byte {
	defer func() {
		r.r += 1
	}()
	return r.buf[r.r]
}

func (r *bytesReader) Uint16() uint16 {
	defer func() {
		r.r += 2
	}()
	return r.reader.Uint16(r.buf[r.r : r.r+2])
}

func (r *bytesReader) Uint32() uint32 {
	defer func() {
		r.r += 4
	}()
	return r.reader.Uint32(r.buf[r.r : r.r+4])
}

func (r *bytesReader) Uint64() uint64 {
	defer func() {
		r.r += 8
	}()
	return r.reader.Uint64(r.buf[r.r : r.r+8])
}

func (r *bytesReader) UintAuto(bits int) uint64 {
	if bits == 1 {
		return uint64(r.Uint32())
	} else if bits == 2 {
		return r.Uint64()
	}
	panic("不支持的系统位数！")
}

func (r *bytesReader) Offset(index int) {
	r.r = index
}

func (r *bytesReader) Data(begin, length int) []byte {
	//if begin+length > r.e {
	//	return nil, io.EOF
	//}
	return r.buf[begin : begin+length]
}

func (r *bytesReader) Party(begin, length int) BytesReader {
	//if begin+length > r.e {
	//	return nil, io.EOF
	//}
	return NewReader(r.buf[begin:begin+length], r.reader)
}

func NewReader(data []byte, reader binary.ByteOrder) BytesReader {
	return &bytesReader{
		buf:    data,
		r:      0,
		e:      len(data),
		reader: reader,
	}
}

// ReadElf 打开 ELF 文件, 需要记录端序
func ReadElf(file string) (*ElfFile, error) {
	elf := &ElfFile{Name: file}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	magic := Elf_Magic(data[:EI_NIDENT])
	reader := NewReader(data, magic.Endian())
	reader.Offset(EI_NIDENT)
	ehdr := NewElfEhdr(reader, magic.Bits())
	ehdr.Magic = magic
	elf.Ehdr = ehdr
	elf.Reader = reader

	// -------------------------------------------
	// 先解析段表字符串信息
	// -------------------------------------------
	offset := int(ehdr.Shoff)
	shentsize := int(ehdr.Shentsize)
	off := offset + int(ehdr.Shstrndx)*shentsize
	next := reader.Party(off, shentsize) // 这里需要解析为指定数据结构

	shstrtab := NewElfShdr(next, ehdr.Magic.Bits()) // 这个是表头， 记录字符串信息的
	shstrTabData := reader.Data(int(shstrtab.Offset), int(shstrtab.Size))
	elf.Shstrtab = shstrTabData
	elf.ShstrtabSize = int(shstrtab.Size)

	// -------------------------------------------
	// 解析段表
	// -------------------------------------------
	// 读取完整段表
	shdrTab := make(map[string]*Elf32_Shdr, int(ehdr.Shnum))
	shdrNames := make([]string, int(ehdr.Shnum))
	for index := 0; index < int(ehdr.Shnum); index++ {
		begin := offset + index*shentsize
		raw := reader.Party(begin, shentsize)
		shdr := NewElfShdr(raw, ehdr.Magic.Bits())
		name := StringTableName(shstrTabData, shdr.Name)
		shdrTab[name] = shdr
		shdrNames[index] = name
		//if name == "" { //删除空段表项
		//	shdrTab[name] = nil
		//} else {
		//	shdrTab[name] = shdr
		//}
	}
	elf.ShdrTab = shdrTab
	elf.ShdrNames = shdrNames

	strTab := shdrTab[".strtab"]
	strTabData := reader.Data(int(strTab.Offset), int(strTab.Size))
	elf.Strtab = strTabData
	elf.StrtabSize = int(strTab.Size)

	symTab := shdrTab[".symtab"]
	symTabSize := 16                           // todo 这个表达式不正确 2 ^ int(symTab.Entsize)      // 16
	symTabLen := int(symTab.Size) / symTabSize // ➗ 16
	symTabList := make(map[string]*Elf32_Sym, symTabLen)
	symNames := make([]string, symTabLen)
	for i := 0; i < symTabLen; i++ {
		begin := int(symTab.Offset) + i*symTabSize
		next := reader.Party(begin, symTabSize)
		sym := NewElfSym(next, ehdr.Magic.Bits())
		name := StringTableName(strTabData, sym.Name)
		symNames[i] = name
		symTabList[name] = sym
		//if name == "" { //无名符号，对于链接没有意义,按照链接器设计需要记录全局和局部符号，避免名字冲突
		//	symTabList[name] = nil
		//} else {
		//	symTabList[name] = sym //加入符号表
		//}
	}
	elf.SymTab = symTabList
	elf.SymNames = symNames

	elf.RelTab = make([]*Elf32_RelInfo, 0)
	for name, relTab := range shdrTab { //所有段的重定位项整合
		if strings.HasPrefix(name, ".rel") { // 重定位段
			relTabLen := int(relTab.Size) / 8
			for i := 0; i < relTabLen; i++ {
				begin := int(relTab.Offset) + i*8
				next := reader.Party(begin, 8)
				rel := NewElfRel(next, ehdr.Magic.Bits())
				sym := symNames[int(rel.Info>>8)]
				relName := StringTableName(strTabData, symTabList[sym].Name)
				elf.RelTab = append(elf.RelTab, &Elf32_RelInfo{
					SegName: name[4:],
					Rel:     rel,
					RelName: relName,
				})
			}
		}
	}

	return elf, nil
}
