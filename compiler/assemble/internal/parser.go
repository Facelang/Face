package internal

import (
	"fmt"
	"github.com/facelang/face/internal/arch"
	"github.com/facelang/face/internal/tokens"
	"io"
	"os"
	"strconv"
	"text/scanner"
)

type LineToken struct {
	tokens.Token        // 符号类型
	LiteralVal   string // 原始值
}

type Parser struct {
	arch          *arch.Arch
	lex           *lexer   // 词法解析器
	section       string   // 当前段
	sections      []string // 所有段
	lineNum       int      // Line number in source file.
	errorLine     int      // Line number of last error.
	errorCount    int      // Number of errors.
	sawCode       bool     // saw code in this file (as opposed to comments and blank lines)
	pc            int64    // virtual PC; count of Progs; doesn't advance for GLOBL or DATA.
	input         []lex.Token
	inputPos      int
	pendingLabels []string // Labels to attach to next instruction.
	labels        map[string]*obj.Prog
	toPatch       []Patch
	addr          []obj.Addr
	ctxt          *obj.Link
	firstProg     *obj.Prog
	lastProg      *obj.Prog
	dataAddr      map[string]int64 // Most recent address for DATA for this symbol.
	isJump        bool             // Instruction being assembled is a jump.
	allowABI      bool             // Whether ABI selectors are allowed.
	pkgPrefix     string           // Prefix to add to local symbols.
	errorWriter   io.Writer
}

func (p *Parser) prefix() (string, bool) {
	var token tokens.Token
	for {
		token = p.lex.NextToken()
		if token == tokens.EOF {
			return "", false
		}
		if token != tokens.COMMENT {
			break
		}
	}

	if token == tokens.IDENT {
		panic(fmt.Errorf("unexpected token %s", "IDENT"))
	}

	return p.lex.ident, true
}

/*
*
.section .name
.global main               # 定义全局符号，使符号对其他文件可见
.local  local_func         # 定义局部符号，仅在当前文件可见
.type   main, @function    # 定义符号类型，@function表示这是一个函数
.size   main, .-main       # 定义符号大小，.-main表示从当前位置到main标签的距离
*/
func (p *Parser) pseudo(word string, args []LineToken) *Program {
	switch word {
	case ".section": // 分段

	case ".global":

	case ".local":

	case ".type":

	case ".size":

	case ".align":

	case "DATA":
		p.asmData(operands)
	case "FUNCDATA":
		p.asmFuncData(operands)
	case "GLOBL":
		p.asmGlobl(operands)
	case "PCDATA":
		p.asmPCData(operands)
	case "PCALIGN":
		p.asmPCAlign(operands)
	case "TEXT":
		p.asmText(operands) // 函数申明
	default: // 处理符号声明
		if len(args) > 0 && args[0].LiteralVal == ":" {
			// 说明是符号
		}
		return false
	}
	return true
}

// asmText assembles a TEXT pseudo-op.
// TEXT runtime·sigtramp(SB),4,$0-0
func (p *Parser) asmText(operands [][]lex.Token) { // 记录一个函数到代码段
	if len(operands) != 2 && len(operands) != 3 { // 参数至少是,2个或者,,3个
		p.errorf("expect two or three operands for TEXT")
		return
	}

	// Labels are function scoped. Patch existing labels and
	// create a new label space for this TEXT.
	p.patch()                             // todo， 多次被调用
	p.labels = make(map[string]*obj.Prog) // 每次都初始化？

	// Operand 0 is the symbol name in the form foo(SB).
	// That means symbol plus indirect on SB and no offset.
	nameAddr := p.address(operands[0]) // 计算地址？
	if !p.validSymbol("TEXT", &nameAddr, false) {
		return
	}
	name := symbolName(&nameAddr)
	next := 1

	// Next operand is the optional text flag, a literal integer.
	var flag = int64(0)
	if len(operands) == 3 {
		flag = p.evalInteger("TEXT", operands[1])
		next++
	}

	// Issue an error if we see a function defined as ABIInternal
	// without NOSPLIT. In ABIInternal, obj needs to know the function
	// signature in order to construct the morestack path, so this
	// currently isn't supported for asm functions.
	if nameAddr.Sym.ABI() == obj.ABIInternal && flag&obj.NOSPLIT == 0 {
		p.errorf("TEXT %q: ABIInternal requires NOSPLIT", name)
	}

	// Next operand is the frame and arg size.
	// Bizarre syntax: $frameSize-argSize is two words, not subtraction.
	// Both frameSize and argSize must be simple integers; only frameSize
	// can be negative.
	// The "-argSize" may be missing; if so, set it to objabi.ArgsSizeUnknown.
	// Parse left to right.
	op := operands[next]
	if len(op) < 2 || op[0].ScanToken != '$' {
		p.errorf("TEXT %s: frame size must be an immediate constant", name)
		return
	}
	op = op[1:]
	negative := false
	if op[0].ScanToken == '-' {
		negative = true
		op = op[1:]
	}
	if len(op) == 0 || op[0].ScanToken != scanner.Int {
		p.errorf("TEXT %s: frame size must be an immediate constant", name)
		return
	}
	frameSize := p.positiveAtoi(op[0].String())
	if negative {
		frameSize = -frameSize
	}
	op = op[1:]
	argSize := int64(abi.ArgsSizeUnknown)
	if len(op) > 0 {
		// There is an argument size. It must be a minus sign followed by a non-negative integer literal.
		if len(op) != 2 || op[0].ScanToken != '-' || op[1].ScanToken != scanner.Int {
			p.errorf("TEXT %s: argument size must be of form -integer", name)
			return
		}
		argSize = p.positiveAtoi(op[1].String())
	}
	p.ctxt.InitTextSym(nameAddr.Sym, int(flag), p.pos())
	prog := &obj.Prog{
		Ctxt: p.ctxt,
		As:   obj.ATEXT,
		Pos:  p.pos(),
		From: nameAddr,
		To: obj.Addr{
			Type:   obj.TYPE_TEXTSIZE,
			Offset: frameSize,
			// Argsize set below.
		},
	}
	nameAddr.Sym.Func().Text = prog
	prog.To.Val = int32(argSize)
	p.append(prog, "", true) // 添加一个代码段？
}

// asmData assembles a DATA pseudo-op.
// DATA masks<>+0x00(SB)/4, $0x00000000
func (p *Parser) asmData(operands [][]lex.Token) { // 记录一条数据到数据段
	if len(operands) != 2 {
		p.errorf("expect two operands for DATA")
		return
	}

	// Operand 0 has the general form foo<>+0x04(SB)/4.
	op := operands[0]
	n := len(op)
	if n < 3 || op[n-2].ScanToken != '/' || op[n-1].ScanToken != scanner.Int {
		p.errorf("expect /size for DATA argument")
		return
	}
	szop := op[n-1].String()
	sz, err := strconv.Atoi(szop)
	if err != nil {
		p.errorf("bad size for DATA argument: %q", szop)
	}
	op = op[:n-2]
	nameAddr := p.address(op)
	if !p.validSymbol("DATA", &nameAddr, true) {
		return
	}
	name := symbolName(&nameAddr)

	// Operand 1 is an immediate constant or address.
	valueAddr := p.address(operands[1])
	switch valueAddr.Type {
	case obj.TYPE_CONST, obj.TYPE_FCONST, obj.TYPE_SCONST, obj.TYPE_ADDR:
		// OK
	default:
		p.errorf("DATA value must be an immediate constant or address")
		return
	}

	// The addresses must not overlap. Easiest test: require monotonicity.
	if lastAddr, ok := p.dataAddr[name]; ok && nameAddr.Offset < lastAddr {
		p.errorf("overlapping DATA entry for %s", name)
		return
	}
	p.dataAddr[name] = nameAddr.Offset + int64(sz)

	switch valueAddr.Type {
	case obj.TYPE_CONST:
		switch sz {
		case 1, 2, 4, 8:
			nameAddr.Sym.WriteInt(p.ctxt, nameAddr.Offset, int(sz), valueAddr.Offset)
		default:
			p.errorf("bad int size for DATA argument: %d", sz)
		}
	case obj.TYPE_FCONST:
		switch sz {
		case 4:
			nameAddr.Sym.WriteFloat32(p.ctxt, nameAddr.Offset, float32(valueAddr.Val.(float64)))
		case 8:
			nameAddr.Sym.WriteFloat64(p.ctxt, nameAddr.Offset, valueAddr.Val.(float64))
		default:
			p.errorf("bad float size for DATA argument: %d", sz)
		}
	case obj.TYPE_SCONST:
		nameAddr.Sym.WriteString(p.ctxt, nameAddr.Offset, int(sz), valueAddr.Val.(string))
	case obj.TYPE_ADDR:
		if sz == p.arch.PtrSize {
			nameAddr.Sym.WriteAddr(p.ctxt, nameAddr.Offset, int(sz), valueAddr.Sym, valueAddr.Offset)
		} else {
			p.errorf("bad addr size for DATA argument: %d", sz)
		}
	}
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
