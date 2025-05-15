package main

import "face-lang/compiler/provider/asm"

func main() {
	//buf, err := os.ReadFile("common.t")
	//if err != nil {
	//	panic(err)
	//}
	//for _, b := range buf {
	//	fmt.Printf("%d, ", b)
	//}
	_ = asm.Program("example/ass/hello.s")
	println("完成编译！")
	//file, _ := elf.ReadElf("example/ass/common.o")
	//file.Objdump()
	//err := link.Link("example/ass/hello", "example/ass/common.o", "example/ass/test.o")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("链接完成！")

}
