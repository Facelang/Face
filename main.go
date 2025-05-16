package main

import (
	"face-lang/compiler/provider/link"
	"fmt"
)

func main() {
	//buf, err := os.ReadFile("common.t")
	//if err != nil {
	//	panic(err)
	//}
	//for _, b := range buf {
	//	fmt.Printf("%d, ", b)
	//}
	//_ = asm.Program("example/hello.s")
	//println("完成编译！")
	//file, _ := elf.ReadElf("common.o")
	//file.Objdump()

	// 汇编器没有问题！
	err := link.Link("example/hello", "example/common.s.o", "example/hello.s.o")
	if err != nil {
		panic(err)
	}
	fmt.Println("链接完成！")

}
