package elf

import (
	"bytes"
	"encoding/binary"
	"os"
)

type fileWriter struct {
	name  string           // 文件名
	w     *bytes.Buffer    // 文件输出
	err   error            // 错误记录
	order binary.ByteOrder // 读取器
}

type FileWriter = *fileWriter

func (f *fileWriter) Write(data any) error {
	defer func() {
		f.err = nil
	}()
	if f.err != nil {
		return f.err
	}
	return binary.Write(f.w, f.order, data)
}

func (f *fileWriter) Flush() error {
	if f.err != nil {
		return f.err
	}
	w, err := os.Create(f.name) // 可以覆盖
	if err != nil {
		return err
	}
	_, err = w.Write(f.w.Bytes())
	return err
}

func NewWriter(file string, order binary.ByteOrder) FileWriter {
	//w, err := os.Create(file) // 可以覆盖
	w, err := bytes.NewBuffer(nil), error(nil)
	return &fileWriter{name: file, w: w, err: err, order: order}
}

// FileWrite 输出elf 文件
func FileWrite(file *File, target string) error {
	w := NewWriter(target, file.Endian())
	_ = w.Write(file.Ehdr) //elf文件头

	// 可执行文件
	if file.Ehdr.Type == Elf32_Half(ET_EXEC) {
		//程序头表
		for _, phdr := range file.PhdrTab {
			_ = w.Write(phdr)
		}
		// 【数据段】最重要的部分
		pad := [1]byte{0}
		for _, seg := range file.ProgSegList {
			padnum := seg.Offset - seg.Begin
			for ; padnum != 0; padnum-- { //填充
				_ = w.Write(pad)
			}
			if seg.Name == ".bss" {
				continue
			}
			var oldBlock *Block = nil
			instPad := [1]byte{0x90}
			for i := 0; i < len(seg.Blocks); i++ {
				b := seg.Blocks[i]
				if oldBlock != nil {
					padnum = b.Offset - (oldBlock.Offset + oldBlock.Size)
					for ; padnum != 0; padnum-- { //填充
						_ = w.Write(instPad)
					}
				}
				oldBlock = b
				_ = w.Write(b.Data)
			}
		}
	} else {
		// 【数据段】最重要的部分
		var prev *ProgSeg = nil
		pad := [1]byte{0}
		for _, seg := range file.ProgSegList {
			if prev != nil {
				padnum := seg.Offset - (prev.Offset + prev.Size)
				for ; padnum != 0; padnum-- { //填充
					_ = w.Write(pad)
				}
			}
			if seg.Name == ".bss" {
				continue
			}
			for i := 0; i < len(seg.Blocks); i++ {
				b := seg.Blocks[i]
				_ = w.Write(b.Data)
			}
			prev = seg
		}
	}

	// 最后写段表字符串
	_ = w.Write(file.Shstrtab)

	// 段表
	for _, sh := range file.ShdrNames {
		_ = w.Write(file.ShdrTab[sh])
	}

	// 符号表
	for _, sym := range file.SymNames {
		_ = w.Write(file.SymTab[sym])
	}

	// 字符串表
	_ = w.Write(file.Strtab)

	// 重定位表
	for _, rel := range file.RelTab {
		_ = w.Write(rel.Rel)
	}

	return w.Flush() // 最后一部再写入文件
}
