package internal

import (
	"bytes"
	"fmt"
	"github.com/facelang/face/compiler/compile/token"
	"github.com/facelang/face/internal/os/elf"
	"github.com/facelang/face/internal/utils"
	"go/ast"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/scanner"
)

// 重定位类型常量
const (
	R_386_32   = 1 // 绝对寻址
	R_386_PC32 = 2 // 相对寻址
)

type section struct {
	Name           string
	Offset, Length int
}

type relocate struct {
	Label   string // 重定位符号的名称
	Type    int    // 重定位类型0-R_386_32；1-R_386_PC32
	Offset  int    // 重定位位置的偏移
	Section string // 重定位目标段
}

type parser struct {
	*lexer                      // 词法解析器
	token        Token          // 符号类型
	error        error          // 错误信息
	declList     []*ast.GenDecl // 语句列表
	sec          *section       // 当前段
	secList      []*section     // 所有段列表
	instrList    []*instr       // 指令列表
	labelList    []*label       // 符号表
	labelNames   map[string]int // 符号表，名称映射
	relocateList []*relocate    // 重定位表

	//lineNum       int   // Line number in source file.
	//errorLine     int   // Line number of last error.
	//errorCount    int   // Number of errors.
	//sawCode       bool  // saw code in this file (as opposed to comments and blank lines)
	//pc            int64 // virtual PC; count of Progs; doesn't advance for GLOBL or DATA.
	//input         []lex.Token
	//inputPos      int
	//pendingLabels []string // Labels to attach to next instruction.
	//labels        map[string]*obj.Prog
	//toPatch       []Patch
	//addr          []obj.Addr
	//ctxt          *obj.Link
	//firstProg     *obj.Prog
	//lastProg      *obj.Prog
	//dataAddr      map[string]int64 // Most recent address for DATA for this symbol.
	//isJump        bool             // Instruction being assembled is a jump.
	//allowABI      bool             // Whether ABI selectors are allowed.
	//pkgPrefix     string           // Prefix to add to local symbols.
	//errorWriter   io.Writer
}

//func (p *Parser) prefix() (string, bool) {
//	var token tokens.Token
//	for {
//		token = p.lex.NextToken()
//		if token == tokens.EOF {
//			return "", false
//		}
//		if token != tokens.COMMENT {
//			break
//		}
//	}
//
//	if token == tokens.IDENT {
//		panic(fmt.Errorf("unexpected token %s", "IDENT"))
//	}
//
//	return p.lex.ident, true
//}

func (p *parser) _addRel(label string, relType int) {
	p.relocateList = append(
		p.relocateList,
		&relocate{
			Label:   label,        // 重定位符号的名称
			Type:    relType,      // 重定位类型0-R_386_32；1-R_386_PC32
			Offset:  p.sec.Offset, // 重定位位置的偏移
			Section: p.sec.Name,   // 重定位目标段
		},
	)
}

// 段落切换
func (p *parser) _switch(id string) {
	p.secList = append(
		p.secList,
		&section{
			Name:   p.sec.Name,
			Length: p.sec.Offset, // 结束位置，也代表大小, 先不记录偏移
		},
	)

	p.sec.Name = id  // 切换到下一个段
	p.sec.Offset = 0 // 清0段偏移
}

// ----------------------------------------------------------------------------------
// -- parser start

func (p *parser) errorf(format string, args ...interface{}) {
	p.error = fmt.Errorf(format, args...)
	panic(p.error)
}

func (p *parser) next() {
	p.token = p.lexer.NextToken()
	for p.token == COMMENT {
		p.token = p.lexer.NextToken()
	}
}

func (p *parser) got(token Token) bool {
	if p.token == token {
		p.next()
		return true
	}
	return false
}

func (p *parser) expect(tokens ...Token) Pos {
	pos := p.pos
	for _, tok := range tokens {
		if p.token == tok {
			p.next()
			return pos
		}
	}

	p.unexpect(tokens[0].String())
	return pos
}

func (p *parser) unexpect(except string) {
	found := token.TokenLabel(p.token, p.id)
	p.errorf("except %s, found %s", except, found)
}

// defineType 处理数据定义, 同时计算符号长度
func (p *parser) data(cont *[]int64, contLen *int64) {
	switch p.token {
	case IDENT: // 引用变量，变量必须已经被申明， 如果符号未定义，则记录重定位
		lb := p.GetLabel(p.id)
		if lb.Type == EQU_LABEL || lb.Type == LOCAL_LABEL {
			(*cont)[*contLen] = lb.Addr
		} else { // 未定义或非法符号, equ 做了单独处理！
			p._addRel(p.id, R_386_32)
		}
		*contLen++
		p.next()
	case INT:
		(*cont)[*contLen] = utils.IntBytes(p.id)
		*contLen++
		p.next()
	case FLOAT:
		(*cont)[*contLen] = utils.FloatBytes(p.id)
		*contLen++
		p.next()
	case STRING:
		for _, ch := range []byte(p.id) {
			(*cont)[*contLen] = int64(ch)
			*contLen++
		}
		p.next()
	default:
		// todo
		//p.errorf("[valType](%d,%d): %s, %，数据类型获取异常！", p.token.Message(p.id))
	}
}

func (p *parser) define(id string, times, size int) {
	lb := NewLabel(LOCAL_LABEL)
	lb.Times = times
	lb.Size = size
	lb.Cont = make([]int, 255) // 数据缓存
	lb.ContLen = 0
	p.data(&lb.Cont, &lb.ContLen) // 这里获得的是值， 数字、字符串、引用名

	// 看是否有连续定义, 例如："hello world", 13, 10
	token := p.NextToken()
	for token == COMMA {
		p.data(&lb.Cont, &lb.ContLen) // 这里获得的是值， 数字、字符串、引用名
		token = p.NextToken()
	}

	p.ProcTable.AddLabel(id, lb)
}

// 以符号名称开始的语句， 数据定义，或代码段标记
func (p *parser) labelDec(id string) {
	p.next()
	switch p.token {
	case A_TIMES: // 需要重复
		p.expect(INT)
		repeat := utils.Int(p.id)
		size := p.size()
		p.define(id, utils.Int(p.id), p.size()) // size 代表数据类型 db dd 字符、字、双字
	case A_EQU: // equ 常量？伪指令，所有使用到该符号的，全部替换为值，不存在地址。
		// 这里需要被替换为值
		// 关于 equ 语法说明，equ 支持表达式：可以是数字、地址、其他符号、算术表达式等
		// equ 定义的符号在汇编时就被替换为具体值，不会占用内存，也不会生成机器码。
		// 不能对 equ 定义的符号赋新值（它不是变量）。
		// equ 只能用于常量表达式，不能用于运行时可变的值。
		// todo 完整的逻辑需要支持 数字，其他符号，表达式， 【最终获得运算后的值】
		// todo equ 引用其它符号，必须提前申明！
		p.require(NUMBER)                                 // todo 当前只支持数字
		p.ProcTable.AddLabel(id, NewLabelEqu(p.number())) // 直接添加符号
	case COLON: // 代码段（label）, main: 一般是函数名作为一个单独的记号
		p.ProcTable.AddLabel(id, NewLabelText()) // 作为一个段符号
	default: // 变量支持
		p.Lexer.Back(token) // db, dd, dw // 退回去重新读 p.size()

		p.values(id, 1, p.size()) // 单个变量定义，直接解析
	}
}

func (p *parser) ParseFile() (*ast.File, error) {
	if p.error != nil {
		return nil, p.error
	}

	p.next()

	for p.token > _literal {
		switch p.token {
		case IDENT: // 两种情况，段落定义，变量定义
			p.declList = append(p.declList, p.labelDec(p.id))
		case A_SEC: // 段定义
			p.require(IDENT)
			p._switch(p.id) // 切换到新的段
		case A_DATA: // 数据段定义
			p.next() // 跳过.data
			// 解析数据段内容
			for p.token > _literal {
				switch p.token {
				case A_BYTE, A_WORD, A_LONG, A_QUAD, A_ASCII, A_ASCIZ, A_STRING:
					// 解析数据定义伪指令
					decl := p.parseDataDirective()
					if decl != nil {
						p.declList = append(p.declList, decl)
					}
				case A_REPT:
					// 解析重复定义
					decl := p.parseReptDirective()
					if decl != nil {
						p.declList = append(p.declList, decl)
					}
				case IDENT:
					// 解析标签定义
					p.declList = append(p.declList, p.labelDec(p.id))
				case A_GLB: // 全局符号定义
					p.require(IDENT)
					// 添加到全局符号表
					p.ProcTable.AddLabel(p.id, NewLabelGlobal())
				default:
					p.errorf("unexpected token in data section: %s", p.token)
					return nil, p.error
				}
				p.next()
			}
		case A_TEXT: // 代码段定义
			p.next() // 跳过.text
			// 解析代码段内容
			for p.token > _literal {
				switch p.token {
				case IDENT:
					p.declList = append(p.declList, p.labelDec(p.id))
				default:
					p.inst(p.token) // 解析指令
				}
				p.next()
			}
		case A_GLB: // 全局符号定义
			p.require(IDENT)
			// 添加到全局符号表
			p.ProcTable.AddLabel(p.id, NewLabelGlobal())
		default:
			p.inst(p.token) // 解析指令
		}

		p.next()
	}

	p._switch("") // 结束最后一个段

	return &ast.File{
		Decls: p.declList,
	}, nil
}

// ExportLb 导出符号表
//func (proc *parser) ExportLb() {
//	for _, lb := range proc.MapLabel {
//		if !lb.IsEqu { // EQU定义的符号不导出
//			ObjFile.addSym(lb)
//		}
//	}
//}

//func (proc *parser) WriteData(file *os.File) {
//	for _, lb := range proc.DefLabelList {
//		lb.write(file)
//	}
//}

//// Codegen 代码生成, 生成代码，同时记录每个段的大小
//func (proc *parser) Codegen() error {
//	// 源码扫描完成，开始生成代码， 内部符号已存在
//	instrBuffer := bytes.NewBuffer(nil)
//	for _, instr := range proc.InstrList {
//		instr.WriteOut(instrBuffer, &proc.seg.Offset)
//	}
//	instrBuffer.Len() // 代码段大小
//
//	// important 符号表[可能]存在符号嵌套引用
//	//    但是所有嵌套引用，被引用的符号必须被声明
//	//        如果引用外部符号，记录地址， 下一个引用符合也只引用所在地址信息
//	//    逻辑上无需处理嵌套
//	for _, label := range proc.labelList {
//
//	}
//	//for _, instr := range proc.InstrList {
//	//	instr.WriteOut(instrBuffer, &proc.seg.Offset)
//	//}
//
//}

//func Check(src, dest []byte, name string) {
//	for i, ch := range src {
//		if len(dest) <= i {
//			fmt.Printf("错误：[0x%X],  两文件内容长度不一致: [%d, %d]", i, len(src), len(dest))
//			return
//		}
//		if ch != dest[i] {
//			fmt.Printf("错误:[0x%X, %d]（%X(%d) != %X(%d)）", i, i, ch, ch, dest[i], dest[i])
//
//			for i2, b := range src[i-4 : i+16] {
//				fmt.Printf("%d: [%d, %d] \n", i+i2-4, b, dest[i+i2-4])
//			}
//			fmt.Printf("\n")
//			return
//		}
//	}
//
//	fmt.Printf("校验完成，[%s]完全一致！\n", name)
//}

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
	if len(operands) != 2 && len(operands) != 3 {  // 参数至少是,2个或者,,3个
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

// parseDataDirective 解析数据定义伪指令
func (p *parser) parseDataDirective() *ast.GenDecl {
	switch p.token {
	case A_BYTE:  // .byte
		return p.parseByteDirective()
	case A_WORD:  // .word
		return p.parseWordDirective()
	case A_LONG:  // .long
		return p.parseLongDirective()
	case A_QUAD:  // .quad
		return p.parseQuadDirective()
	case A_ASCII: // .ascii
		return p.parseAsciiDirective()
	case A_ASCIZ: // .asciz
		return p.parseAscizDirective()
	case A_STRING: // .string
		return p.parseStringDirective()
	case A_REPT:  // .rept
		return p.parseReptDirective()
	default:
		p.errorf("unknown data directive: %s", p.token)
		return nil
	}
}

// parseByteDirective 解析.byte伪指令
func (p *parser) parseByteDirective() *ast.GenDecl {
	decl := &ast.GenDecl{
		Tok: token.DATA,
	}
	
	p.next() // 跳过.byte
	
	// 解析值列表
	for {
		switch p.token {
		case INT:
			// 解析整数值
			val := utils.Int(p.id)
			decl.Specs = append(decl.Specs, &ast.ValueSpec{
				Type: &ast.Ident{Name: "byte"},
				Values: []ast.Expr{&ast.BasicLit{
					Kind:  token.INT,
					Value: strconv.FormatInt(val, 10),
				}},
			})
		case STRING:
			// 解析字符串
			for _, ch := range []byte(p.id) {
				decl.Specs = append(decl.Specs, &ast.ValueSpec{
					Type: &ast.Ident{Name: "byte"},
					Values: []ast.Expr{&ast.BasicLit{
						Kind:  token.INT,
						Value: strconv.FormatInt(int64(ch), 10),
					}},
				})
			}
		case IDENT:
			// 解析符号引用
			decl.Specs = append(decl.Specs, &ast.ValueSpec{
				Type: &ast.Ident{Name: "byte"},
				Values: []ast.Expr{&ast.Ident{
					Name: p.id,
				}},
			})
		default:
			p.errorf("invalid value in .byte directive")
			return nil
		}
		
		p.next()
		if p.token != COMMA {
			break
		}
		p.next()
	}
	
	return decl
}

// parseAsciiDirective 解析.ascii伪指令
func (p *parser) parseAsciiDirective() *ast.GenDecl {
	decl := &ast.GenDecl{
		Tok: token.DATA,
	}
	
	p.next() // 跳过.ascii
	
	if p.token != STRING {
		p.errorf("expected string literal after .ascii")
		return nil
	}
	
	// 将字符串转换为字节数组
	for _, ch := range []byte(p.id) {
		decl.Specs = append(decl.Specs, &ast.ValueSpec{
			Type: &ast.Ident{Name: "byte"},
			Values: []ast.Expr{&ast.BasicLit{
				Kind:  token.INT,
				Value: strconv.FormatInt(int64(ch), 10),
			}},
		})
	}
	
	p.next()
	return decl
}

// parseAscizDirective 解析.asciz伪指令
func (p *parser) parseAscizDirective() *ast.GenDecl {
	decl := p.parseAsciiDirective()
	if decl == nil {