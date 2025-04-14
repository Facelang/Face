package main

import (
	"face-lang/compiler/factory"
	"os"
	"strings"
)

func main() {
	//wd, _ := os.Getwd()
	//filePath := filepath.Join(wd, "example/hello/hello.c")
	file := "example/hello/hello.c"
	factory.Program(file)
	println(strings.Join(os.Environ(), "\n"))

	println("\n\n 解析完成")
}
