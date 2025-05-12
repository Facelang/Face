package main

import "face-lang/compiler/provider/link"

func main() {
	//buf, err := os.ReadFile("common.t")
	//if err != nil {
	//	panic(err)
	//}
	//for _, b := range buf {
	//	fmt.Printf("%d, ", b)
	//}
	//_ = asm.Program("example/ass/hello.s")
	//println("完成编译！")
	file, _ := link.ReadElf("example/ass/common.o")
	file.Objdump()
}
