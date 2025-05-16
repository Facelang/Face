package asm

import (
	"fmt"
)

// Program todo 应该支持多个文件同时解析！ 暂时支持一个
func Program(src string) error {
	target := fmt.Sprintf("%s.o", src)
	//if strings.HasSuffix(src, ".s") {
	//	target = fmt.Sprintf("%s.o", src[:len(src)-2])
	//}

	// todo 流程需要进一步优化
	// 解析 -> 校验(略) -> 代码生成 -> elf 文件组装
	fp := NewFileParser(src)
	process, err := fp.Parse()
	if err != nil {
		return err
	}
	println("完成解析！")
	if err = process.LocalRel(); err != nil {
		return err
	}
	println("局部重定位完成！")

	export := process.ExportElf()
	println("elf 文件装载完成！")

	// shoff = 511 应该 515
	// shnum = 10 应该9 【对比头表， shNames 多了一个空. 】
	//      .rel.text off = 1141, size = 32
	//      .rel.data off = 应该 1173, size = 8
	// shstrndx = 65535 应该 4
	return export.WriteFile(target)
}

//func Program(file string) {
//	p := Parser(file)
//	_, err := p.Parse()
//	if err != nil {
//		panic(err)
//	}
//}

func PrintTokenList(f string) {
	//fp, ok := NewFileParser(f).(*parser)
	//if !ok {
	//	return
	//}
	//
	//token := fp.lexer.NextToken()
	//currentLine := 0
	//lineBuffer := strings.Builder{}
	//for token != EOF {
	//	content := fp.lexer.id
	//	//line, col := fp.lexer.line, fp.lexer.col
	//	if fp.lexer.line > currentLine {
	//		fmt.Printf("%s\n", lineBuffer.String())
	//		lineBuffer.Reset()
	//		currentLine = fp.lexer.line
	//	}
	//
	//	lineBuffer.WriteString(fmt.Sprintf("%s(%s) ", content, token.String()))
	//	//fmt.Printf("%s：%s [%d, %d]\n", token.String(), content, line, col)
	//	token = fp.lexer.NextToken()
	//}
}
