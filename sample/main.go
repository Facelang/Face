package main

import (
	"flag"
	"fmt"

	"github.com/marshal/objdump/x86"
)

func main() {
	outputFormat := flag.String("format", "intel", "输出格式: intel 或 att")
	flag.Parse()

	fmt.Println("x86_64样例机器码反汇编演示")
	fmt.Println("============================")

	// 一些简单的x86_64指令机器码
	code := []byte{
		0x55,             // push rbp
		0x48, 0x89, 0xe5, // mov rbp, rsp
		0x48, 0x83, 0xec, 0x10, // sub rsp, 0x10
		0xc7, 0x45, 0xfc, 0x00, 0x00, 0x00, 0x00, // mov DWORD PTR [rbp-0x4], 0x0
		0xb8, 0x00, 0x00, 0x00, 0x00, // mov eax, 0x0
		0x48, 0x83, 0xc4, 0x10, // add rsp, 0x10
		0x5d, // pop rbp
		0xc3, // ret
	}

	// 设置汇编语法
	syntax := x86.INTEL
	if *outputFormat == "att" {
		syntax = x86.ATT
		fmt.Println("使用AT&T语法:")
	} else {
		fmt.Println("使用Intel语法:")
	}

	// 反汇编
	disasm := x86.NewDisassembler(syntax)
	instructions, err := disasm.Disassemble(code, 0x0)
	if err != nil {
		fmt.Printf("反汇编失败: %v\n", err)
		return
	}

	// 打印结果
	fmt.Println("地址     | 机器码                  | 汇编指令")
	fmt.Println("---------+--------------------------+------------------")
	for _, inst := range instructions {
		fmt.Printf("0x%08x | %-24s | %s\n", inst.Address, inst.Bytes, inst.Text)
	}
}
