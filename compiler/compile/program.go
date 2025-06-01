package compile

import (
	"fmt"
	"github.com/facelang/face/compiler/compile/internal"
)

func Program(file string) {
	p := internal.Parser(file)
	_, err := p.Parse()
	if err != nil {
		panic(err)
	}
}

func PrintTokenList(file string) {
	p := internal.Parser(file)
	token, content, line, col := p.NextToken()
	for token != EOF {
		fmt.Printf("%s：%s [%d, %d]\n", token.String(), content, line, col)
		token, content, line, col = p.NextToken()
	}
}
