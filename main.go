package main

import "face-lang/compiler/provider/asm"

func main() {
	_ = asm.Program("example/ass/common.s")
	println("完成编译！")
}
