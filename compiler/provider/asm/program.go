package asm

import (
	"fmt"
	"strings"
)

func Program(f string) error {
	//PrintTokenList(f)
	//return nil
	fp := NewFileParser(f)
	_, err := fp.Parse()

	textBuff := make([]byte, ProcessTable.InstrBuff.Len())
	copy(textBuff, ProcessTable.InstrBuff.Bytes())
	relFinish := map[int]struct{}{}
	for i, rel := range ProcessTable.RelRecordList {
		lb := ProcessTable.GetLabel(rel.LbName)
		if rel.TarSeg != ".text" || lb.Externed { // 没有找到内部变量
			continue
		}
		// 修改指定位置的数据
		copy(textBuff[rel.Offset:rel.Offset+4], ValueBytes(lb.Addr, 4))
		relFinish[i] = struct{}{}
	}
	println("完成解析！")
	//ObjFile.WriteElf(f + ".o")

	return err
}

//func Program(file string) {
//	p := Parser(file)
//	_, err := p.Parse()
//	if err != nil {
//		panic(err)
//	}
//}

func PrintTokenList(f string) {
	fp, ok := NewFileParser(f).(*parser)
	if !ok {
		return
	}

	token := fp.lexer.NextToken()
	currentLine := 0
	lineBuffer := strings.Builder{}
	for token != EOF {
		content := fp.lexer.id
		//line, col := fp.lexer.line, fp.lexer.col
		if fp.lexer.line > currentLine {
			fmt.Printf("%s\n", lineBuffer.String())
			lineBuffer.Reset()
			currentLine = fp.lexer.line
		}

		lineBuffer.WriteString(fmt.Sprintf("%s(%s) ", content, token.String()))
		//fmt.Printf("%s：%s [%d, %d]\n", token.String(), content, line, col)
		token = fp.lexer.NextToken()
	}
}
