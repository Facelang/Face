package internal

import (
	"log"
	"os"
)

// NewLexer returns a lexer for the named file and the given link context.
func NewLexer(name string) TokenReader { // 封装后的读取器
	input := NewInput(name)
	fd, err := os.Open(name)
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	input.Push(NewTokenizer(name, fd, fd))
	return input
}
