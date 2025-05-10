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
	_ = asm.Program("example/ass/common.s")
	println("完成编译！")
}
