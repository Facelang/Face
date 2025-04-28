package asm

import "fmt"

type ErrorFunc func(file string, line, col, off int, msg string)

// 缓存溢出
func overflow(b *buffer) error {
	return fmt.Errorf(
		"Cache Overflow. The Maximum Limit 64M "+
			"In File(%s %d,%d)", b.file, b.line+1, b.col+1)
}
