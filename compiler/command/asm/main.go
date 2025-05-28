package main

import (
	"face-lang/compiler/command/asm/internal"
	"face-lang/compiler/provider/asm"
	"face-lang/compiler/target/elf"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	Debug      = flag.Bool("debug", false, "启用调试模式，默认不启用")
	OutputFile = flag.String("o", "", "输出文件，默认跟输入文件保持一致")
	// todo 可以指定平台信息， 支持跨平台编译
)

func Usage() {
	fmt.Fprintf(os.Stderr, "usage: asm [options] file.s ...\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	if flag.NArg() == 0 {
		flag.Usage()
	}

	if *OutputFile == "" {
		if flag.NArg() != 1 {
			flag.Usage()
		}
		input := filepath.Base(flag.Arg(0))
		input = strings.TrimSuffix(input, ".s")
		*OutputFile = fmt.Sprintf("%s.o", input)
	}

	for _, f := range flag.Args() {
		lexer := internal.NewLexer(f)
		parser := internal.NewParser(lexer)
		pList := new(obj.Plist)
		pList.Firstpc, ok = parser.Parse() // p.firstProg

		obj.Flushplist(ctxt, pList, nil)
	}

	buf, err := os.ReadFile("common.t")
	if err != nil {
		panic(err)
	}
	for _, b := range buf {
		fmt.Printf("%d, ", b)
	}
	_ = asm.Program("example/hello.s")
	println("完成编译！")
	file, _ := elf.ReadElf("common.o")
	file.Objdump()

}
