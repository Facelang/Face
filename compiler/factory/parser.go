package factory

import (
	"fmt"
	"os"
)

type parser struct {
	lex *lexer // 读取器
}

func (p *parser) init(file string) {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	p.lex.init(file, f, func(file string, line, col, off int, msg string) {
		return
	})
}

func (p *parser) parse() {
	p.program()
}

// .：语法结束符，表示规则的终结（EOF）。
// {} 表示 零个或多个 顶层声明，因此顶层声明也是可选的。
// program = segment { program } .
func (p *parser) program() {
	for {
		token := p.lex.scan()
		if token == EOF {
			println("文件解析结束！")
			return
		}
		fmt.Printf("[%d,%d,%d] %s %s \n",
			p.lex.line, p.lex.line, p.lex.off,
			token.String(), p.lex.content,
		)
		println(token.String())
	}

}
