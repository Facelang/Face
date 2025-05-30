package internal

import (
	"github.com/facelang/face/internal/tokens"
	"io"
	"os"
)

type LineToken struct {
	tokens.Token        // 符号类型
	LiteralVal   string // 原始值
}

type Parser struct {
	lex           *lexer // 词法解析器
	lineNum       int    // Line number in source file.
	errorLine     int    // Line number of last error.
	errorCount    int    // Number of errors.
	sawCode       bool   // saw code in this file (as opposed to comments and blank lines)
	pc            int64  // virtual PC; count of Progs; doesn't advance for GLOBL or DATA.
	input         []lex.Token
	inputPos      int
	pendingLabels []string // Labels to attach to next instruction.
	labels        map[string]*obj.Prog
	toPatch       []Patch
	addr          []obj.Addr
	arch          *arch.Arch
	ctxt          *obj.Link
	firstProg     *obj.Prog
	lastProg      *obj.Prog
	dataAddr      map[string]int64 // Most recent address for DATA for this symbol.
	isJump        bool             // Instruction being assembled is a jump.
	allowABI      bool             // Whether ABI selectors are allowed.
	pkgPrefix     string           // Prefix to add to local symbols.
	errorWriter   io.Writer
}

func (p *Parser) Parse() *Program {
	scratch := make([][]lex.Token, 0, 3)
	for {
		word, cond, operands, ok := p.line(scratch) // operands = scratch 一维数组为每个参数， 逗号分割, 二维数组是具体的符号和 ident 两种
		if !ok {
			break
		}
		scratch = operands

		if p.pseudo(word, operands) { // 处理伪指令，段落、符号定义 DATA TEXT
			continue
		}
		i, present := p.arch.Instructions[word] // 这里取指令操作码
		if present {
			p.instruction(i, word, cond, operands) // 最重要！处理指令
			continue
		}
		p.errorf("unrecognized instruction %q", word)
	}
	if p.errorCount > 0 {
		return nil, false
	}
	p.patch() // todo 不知道用途 可能跟标签有关
	return p.firstProg, true
}

func NewParser(lex *lexer) *Parser {
	return &Parser{
		lex:         lex,
		labels:      make(map[string]*obj.Prog),
		dataAddr:    make(map[string]int64),
		errorWriter: os.Stderr,
		allowABI:    ctxt != nil && objabi.LookupPkgSpecial(ctxt.Pkgpath).AllowAsmABI,
		pkgPrefix:   pkgPrefix,
	}
}
