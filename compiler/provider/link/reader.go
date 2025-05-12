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

func GetName(bytes []byte, start uint32) string {
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

//func NewReader(file string) FileReader {
//	buf := &fileReader{}
//	_ = buf.init(file)
//	return buf
//}
