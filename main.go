package main

import (
	"face-lang/compiler/factory"
)

func main() {
	//wd, _ := os.Getwd()
	//filePath := filepath.Join(wd, "example/hello/hello.c")
	file := "example/hello/hello.c"
	factory.Program(file)
	println("\n\n 解析完成")
}
