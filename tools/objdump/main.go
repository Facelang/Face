package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/marshal/objdump/elf"
	"github.com/marshal/objdump/x86"
)

func main() {
	// 命令行参数
	outputFormat := flag.String("format", "intel", "输出格式: intel 或 att")
	showAll := flag.Bool("all", false, "显示所有节的内容")
	showText := flag.Bool("text", true, "显示代码段内容")
	showSymbols := flag.Bool("symbols", false, "显示符号表")
	showHeaders := flag.Bool("headers", false, "显示ELF头信息")
	flag.Parse()

	// 检查是否提供了文件参数
	if flag.NArg() < 1 {
		fmt.Println("用法: objdump [选项] <文件>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// 获取输入文件路径
	inputFile := flag.Arg(0)

	// 检查文件是否存在
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		log.Fatalf("错误: 文件 %s 不存在", inputFile)
	}

	// 解析ELF文件
	elfFile, err := elf.ParseELF(inputFile)
	if err != nil {
		log.Fatalf("解析ELF文件失败: %v", err)
	}

	// 显示文件头信息
	if *showHeaders {
		fmt.Printf("文件: %s\n", filepath.Base(inputFile))
		fmt.Printf("格式: ELF%d\n", elfFile.Class)
		fmt.Printf("架构: %s\n", elfFile.Machine)
		fmt.Println("ELF头信息:")
		elfFile.PrintELFHeader()
		fmt.Println()
	}

	// 显示符号表
	if *showSymbols {
		fmt.Println("符号表:")
		elfFile.PrintSymbolTable()
		fmt.Println()
	}

	// 显示代码段内容
	if *showText {
		fmt.Println("代码段反汇编:")
		// 设置反汇编格式
		syntax := x86.INTEL
		if *outputFormat == "att" {
			syntax = x86.ATT
		}

		disasm := x86.NewDisassembler(syntax)
		elfFile.DisassembleTextSection(disasm)
		fmt.Println()
	}

	// 显示所有节的内容
	if *showAll {
		fmt.Println("所有节信息:")
		elfFile.PrintAllSections()
	}
}
