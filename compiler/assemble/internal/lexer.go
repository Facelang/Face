package internal

import (
	"github.com/facelang/face/internal/reader"
	"github.com/facelang/face/internal/tokens"
	"text/scanner"
	"unicode"
)

// Whitespace 对比 map, switch 位掩码 比较效率最高, 忽略 \n
const Whitespace = 1<<'\t' | 1<<'\r' | 1<<' '

type lexer struct {
	reader *reader.Reader
	token  tokens.Token
	ident  string
}

func (lex *lexer) NextToken() tokens.Token {
	ch, chw := lex.reader.ReadRune()
	if chw == 0 {
		return tokens.EOF
	}

	// skip white space
	for Whitespace&(1<<ch) != 0 {
		ch, chw = lex.reader.ReadRune()
	}

	if chw == 0 {
		return tokens.EOF
	}

	lex.ident = ""

	// start collecting token text
	lex.reader.TextReady()

	if '0' <= ch && ch <= '9' { // 数字
		return GetDecimal(lex, ch)
	}

	if CheckIdent(ch, 0) { // 符号
		for i := 1; CheckIdent(ch, i); i++ {
			ch, chw = lex.reader.ReadRune()
		}
		lex.ident = lex.reader.ReadText()
		return tokens.IDENT
	}

	// determine token value
	switch ch {
	case '\n':
		return tokens.RETURN
	case '"':
		ident, _ := reader.String(lex.reader, '"')
		lex.ident = ident
		return tokens.STRING
	case '\'':
		lex.ident = reader.Char(lex.reader)
		return tokens.Char
	case '`':
		lex.ident = reader.RawString(lex.reader)
		return tokens.STRING
	case ';': // todo at&t 语法使用 # 作为注解
		lex.ident = reader.Comment(lex.reader)
		return tokens.COMMENT
	default:
		return tokens.Token(ch)
	}
}

// NextLine 读取一行指令
// section {name} | global {name} | label: | instr operand1, ...
//
//		示例如下：
//				section .data
//	   				msg db 'Hello, World!', 0
//
//				section .text
//
//				global _start
//
//				_start:
//					mov eax, 4          ; 系统调用号 (write)
//					mov ebx, 1          ; 文件描述符 (stdout)
//					mov ecx, msg        ; 消息地址
//					mov edx, 13         ; 消息长度
//					int 0x80            ; 调用内核
//
//					mov eax, 1          ; 系统调用号 (exit)
//					xor ebx, ebx        ; 返回码 0
//					int 0x80            ; 调用内核
func (lex *lexer) NextLine() []tokens.Token {
	for {
		token := lex.NextToken()
		if token == tokens.EOF {
			return nil // todo
		}
		if token == tokens.COMMENT { // 跳过注解
			continue
		}
		if token == tokens.RETURN { // 跳过前置换行符
			continue
		}
	}

	token := lex.NextToken()

	for {
		token = lex.NextToken()
		tok = p.nextToken()
		p.lineNum = p.lex.Line()
		switch tok {
		case '\n', ';':
			continue
		case scanner.EOF:
			return "", "", nil, false
		}
		break
	}
}

func (p *Parser) line(scratch [][]lex.Token) (word, cond string, operands [][]lex.Token, ok bool) {
next: // 读一行指令。。。
	// Skip newlines.
	var tok lex.ScanToken
	for {
		tok = p.nextToken()
		// We save the line number here so error messages from this instruction
		// are labeled with this line. Otherwise we complain after we've absorbed
		// the terminating newline and the line numbers are off by one in errors.
		p.lineNum = p.lex.Line()
		switch tok {
		case '\n', ';':
			continue
		case scanner.EOF:
			return "", "", nil, false
		}
		break
	}
	// First item must be an identifier.
	if tok != scanner.Ident {
		p.errorf("expected identifier, found %q", p.lex.Text())
		return "", "", nil, false // Might as well stop now.
	}
	word, cond = p.lex.Text(), "" // 指令名称, 条件后缀？
	operands = scratch[:0]        // 空切片？ 容量和 scratch 切片相同。
	// Zero or more comma-separated operands, one per loop.
	nesting := 0
	colon := -1
	for tok != '\n' && tok != ';' { // 读到一行结束， 必须是 \n 或者 ;
		// Process one operand.
		var items []lex.Token              // 什么意思
		if cap(operands) > len(operands) { // 复用长度和容量
			// Reuse scratch items slice.
			items = operands[:cap(operands)][len(operands)][:0] // opr[:cap][len][:0] // len 是最后一个  应该等同于opr[len][:0]， cap 避免越界？
		} else {
			items = make([]lex.Token, 0, 3) // 创建新的
		}
		for {
			tok = p.nextToken()
			if len(operands) == 0 && len(items) == 0 { // 判断是一个 点 . (cond = cond + . + text), cond 开始为空
				if p.arch.InFamily(sys.ARM, sys.ARM64, sys.AMD64, sys.I386, sys.Loong64, sys.RISCV64) && tok == '.' {
					// Suffixes: ARM conditionals, Loong64 vector instructions, RISCV rounding mode or x86 modifiers.
					tok = p.nextToken()
					str := p.lex.Text()
					if tok != scanner.Ident {
						p.errorf("instruction suffix expected identifier, found %s", str)
					}
					cond = cond + "." + str
					continue
				}
				if tok == ':' { // 读到了 : , 说明是一个 label 段落标记
					// Labels.
					p.pendingLabels = append(p.pendingLabels, word)
					goto next // 返回重新解析， cond 被丢
				}
			}
			if tok == scanner.EOF {
				p.errorf("unexpected EOF")
				return "", "", nil, false
			}
			// Split operands on comma. Also, the old syntax on x86 for a "register pair"
			// was AX:DX, for which the new syntax is DX, AX. Note the reordering.
			if tok == '\n' || tok == ';' || (nesting == 0 && (tok == ',' || tok == ':')) {
				if tok == ':' { // 中断的一种特殊情况 nesting = 0 && tok = :
					// Remember this location so we can swap the operands below.
					if colon >= 0 {
						p.errorf("invalid ':' in operand")
						return word, cond, operands, true
					}
					colon = len(operands)
				}
				break
			}
			if tok == '(' || tok == '[' {
				nesting++
			}
			if tok == ')' || tok == ']' {
				nesting--
			}
			items = append(items, lex.Make(tok, p.lex.Text()))
		}
		if len(items) > 0 {
			operands = append(operands, items)
			if colon >= 0 && len(operands) == colon+2 {
				// AX:DX becomes DX, AX.
				operands[colon], operands[colon+1] = operands[colon+1], operands[colon]
				colon = -1
			}
		} else if len(operands) > 0 || tok == ',' || colon >= 0 {
			// Had a separator with nothing after.
			p.errorf("missing operand")
		}
	}
	return word, cond, operands, true
}

func CheckIdent(ch rune, i int) bool {
	return ch == '_' || unicode.IsLetter(ch) ||
		unicode.IsDigit(ch) && i > 0 // 第一个字符必须是字母或下划线
}

func GetDecimal(lex *lexer, ch rune) tokens.Token {
	token, val := reader.Decimal(lex.reader, ch)
	lex.ident = val
	return token
}

func NewLexer(file string) *lexer { // 封装后的读取器
	return &lexer{reader: reader.FileReader(file)}
}
