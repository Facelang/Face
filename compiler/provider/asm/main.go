package asm

import (
	"fmt"
	"strings"
)

// Program todo 应该支持多个文件同时解析！ 暂时支持一个
func Program(src string) error {
	target := fmt.Sprintf("%s.o", src)
	if strings.HasSuffix(src, ".s") {
		target = fmt.Sprintf("%s.o", src[:len(src)-2])
	}

	fp := NewFileParser(src)
	process, err := fp.Parse()
	if err != nil {
		return err
	}
	println("完成解析！")
	if err = process.LocalRel(); err != nil {
		return err
	}
	export := process.ExportElf()
	return export.WriteFile(target)
	//ObjFile.WriteElf(f + ".o")
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
