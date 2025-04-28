package elf

import (
	"encoding/binary"
	"io"
	"os"
)

type bytesReader struct {
	buf  []byte // 字节数组
	r, e int    // 读取器游标
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

func (r *bytesReader) endian(e int) binary.ByteOrder {
	if e == 1 {
		return binary.LittleEndian
	} else if e == 2 {
		return binary.BigEndian
	}
	panic("不支持的字节序")
}

func (r *bytesReader) readUint16(e int) uint16 {
	defer func() {
		r.r += 2
	}()
	return r.endian(e).Uint16(r.buf[r.r : r.r+2])
}

func (r *bytesReader) readUint32(e int) uint32 {
	defer func() {
		r.r += 4
	}()
	return r.endian(e).Uint32(r.buf[r.r : r.r+4])
}

func (r *bytesReader) readUint64(e int) uint64 {
	defer func() {
		r.r += 8
	}()
	return r.endian(e).Uint64(r.buf[r.r : r.r+8])
}

func (r *bytesReader) readUintAuto(t, e int) uint64 {
	if t == 1 {
		return uint64(r.readUint32(e))
	} else if t == 2 {
		return r.readUint64(e)
	}
	panic("不支持的系统位数！")
}

type fileReader struct {
	buf  []byte // 缓存池
	size int    // 缓存大小
	err  error  // 错误信息
}

type FileReader = *fileReader

func (r *fileReader) init(file string) error {
	r.buf, r.err = os.ReadFile(file)
	r.size = len(r.buf)
	return r.err
}

func (r *fileReader) segment(begin, length int) (BytesReader, error) {
	if begin+length > r.size {
		return nil, io.EOF
	}

	return &bytesReader{
		buf: r.buf[begin : begin+length],
		r:   0, e: length,
	}, nil
}

func NewReader(file string) FileReader {
	buf := &fileReader{}
	_ = buf.init(file)
	return buf
}
