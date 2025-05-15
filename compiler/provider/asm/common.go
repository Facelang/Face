package asm

import (
	"os"
)

type ErrFunc func(file string, line, col, off int, msg string)

// 缓存溢出
//func overflow(b *buffer) error {
//	return fmt.Errorf(
//		"Cache Overflow. The Maximum Limit 64M "+
//			"In File(%s %d,%d)", b.file, b.line+1, b.col+1)
//}

type RelFunc func(string, int) bool

// var Fout *os.File = nil // 输出文件指针
// var ScanLop = 0              // 扫描次数，1表示第一遍，2表示第二遍
//var RelLb *LabelRecord = nil //

// 操作数类型
const OPR_IMMD = 1 // 立即数
const OPR_MEMR = 3 // 内存
const OPR_REGS = 2 // 寄存器

// WriteBytes 写入字节（按照小端顺序输出不大于4字节长度数据）
// 按照小端顺序（little endian）输出指定长度数据
// len=1：输出第4字节
// len=2:输出第3,4字节
// len=4:输出第1,2,3,4字节
//func WriteBytes(file *os.File, value, length int) {
//	ProcessTable.CurSegOff += length
//	//if ScanLop == 2 {
//	bytes := make([]byte, length)
//	for i := 0; i < length; i++ {
//		bytes[i] = byte((value >> (i * 8)) & 0xFF)
//	}
//	_, _ = file.Write(bytes)
//	//InLen += length
//	//}
//}

//func WriteBytes(w io.Writer, value, length int) {
//	ProcessTable.CurSegOff += length
//	bytes := make([]byte, length)
//	for i := 0; i < length; i++ {
//		bytes[i] = byte((value >> (i * 8)) & 0xFF)
//	}
//	_, err := w.Write(bytes)
//	if err != nil {
//		panic(err)
//	}
//}

// WriteUint16 辅助函数，写入16位无符号整数（小端序）
func WriteUint16(file *os.File, val uint16) {
	_, _ = file.Write([]byte{byte(val), byte(val >> 8)})
}

// WriteUint32 辅助函数，写入32位无符号整数（小端序）
func WriteUint32(file *os.File, val uint32) {
	_, _ = file.Write([]byte{byte(val), byte(val >> 8), byte(val >> 16), byte(val >> 24)})
}
