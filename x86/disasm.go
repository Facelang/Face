package x86

import (
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/arch/x86/x86asm"
)

// SyntaxType 定义反汇编语法类型
type SyntaxType int

const (
	// Intel语法
	INTEL SyntaxType = iota
	// AT&T语法
	ATT
)

// Instruction 表示反汇编后的指令
type Instruction struct {
	Address uint64
	Bytes   string
	Text    string
}

// Disassembler 是x86_64指令的反汇编器
type Disassembler struct {
	Syntax SyntaxType
}

// NewDisassembler 创建新的反汇编器
func NewDisassembler(syntax SyntaxType) *Disassembler {
	return &Disassembler{
		Syntax: syntax,
	}
}

// Disassemble 对给定的二进制数据进行反汇编
func (d *Disassembler) Disassemble(code []byte, baseAddr uint64) ([]Instruction, error) {
	var instructions []Instruction
	pc := uint64(0)

	for pc < uint64(len(code)) {
		inst, err := x86asm.Decode(code[pc:], 64)
		if err != nil {
			// 如果解码失败，尝试作为数据字节处理
			instructions = append(instructions, Instruction{
				Address: baseAddr + pc,
				Bytes:   formatBytes(code[pc : pc+1]),
				Text:    fmt.Sprintf(".byte 0x%02x", code[pc]),
			})
			pc++
			continue
		}

		// 获取指令的机器码字节
		instBytes := code[pc : pc+uint64(inst.Len)]
		bytesStr := formatBytes(instBytes)

		// 格式化指令文本
		var text string
		switch d.Syntax {
		case INTEL:
			text = formatIntelSyntax(inst)
		case ATT:
			text = formatATTSyntax(inst)
		}

		instructions = append(instructions, Instruction{
			Address: baseAddr + pc,
			Bytes:   bytesStr,
			Text:    text,
		})

		pc += uint64(inst.Len)
	}

	return instructions, nil
}

// formatBytes 将字节序列格式化为十六进制字符串
func formatBytes(bytes []byte) string {
	if len(bytes) == 0 {
		return ""
	}

	// 将字节转换为十六进制字符串
	hexStr := hex.EncodeToString(bytes)

	// 每两个字符（一个字节）之间添加空格
	var result strings.Builder
	for i := 0; i < len(hexStr); i += 2 {
		if i > 0 {
			result.WriteByte(' ')
		}
		result.WriteString(hexStr[i : i+2])
	}

	return result.String()
}

// formatIntelSyntax 将指令格式化为Intel语法
func formatIntelSyntax(inst x86asm.Inst) string {
	// 指令名称
	text := inst.Op.String()

	// 添加操作数
	if len(inst.Args) > 0 {
		var args []string
		for _, arg := range inst.Args {
			if arg == nil {
				break
			}
			args = append(args, formatIntelArg(arg))
		}
		if len(args) > 0 {
			text = fmt.Sprintf("%-7s %s", text, strings.Join(args, ", "))
		}
	}

	return text
}

// formatATTSyntax 将指令格式化为AT&T语法
func formatATTSyntax(inst x86asm.Inst) string {
	// 指令名称（添加后缀）
	opName := inst.Op.String()
	text := fmt.Sprintf("%-7s", attOpName(opName))

	// 添加操作数（注意：AT&T语法中操作数顺序相反）
	var args []string
	for i := len(inst.Args) - 1; i >= 0; i-- {
		if inst.Args[i] == nil {
			continue
		}
		args = append(args, formatATTArg(inst.Args[i]))
	}

	if len(args) > 0 {
		text = fmt.Sprintf("%s %s", text, strings.Join(args, ", "))
	}

	return text
}

// attOpName 将操作码转换为AT&T语法（添加大小后缀）
func attOpName(opName string) string {
	// 这里是简化版本，完整版需要根据操作数类型添加适当的后缀
	lower := strings.ToLower(opName)

	// 添加常见指令的后缀
	if strings.HasPrefix(lower, "mov") ||
		strings.HasPrefix(lower, "add") ||
		strings.HasPrefix(lower, "sub") ||
		strings.HasPrefix(lower, "and") ||
		strings.HasPrefix(lower, "or") ||
		strings.HasPrefix(lower, "xor") {
		return lower + "q" // 64位后缀
	}

	return lower
}

// formatIntelArg 格式化Intel语法的操作数
func formatIntelArg(arg x86asm.Arg) string {
	switch a := arg.(type) {
	case x86asm.Reg:
		return a.String()
	case x86asm.Mem:
		// 构建内存引用字符串
		var base, index, disp string

		if a.Base != 0 {
			base = a.Base.String()
		}

		if a.Index != 0 {
			index = fmt.Sprintf("%s*%d", a.Index, a.Scale)
		}

		if a.Disp != 0 {
			if a.Disp < 0 {
				disp = fmt.Sprintf("-0x%x", -a.Disp)
			} else {
				disp = fmt.Sprintf("0x%x", a.Disp)
			}
		}

		// 组合内存引用部分
		var inner string
		if base != "" && index != "" {
			inner = fmt.Sprintf("%s+%s", base, index)
		} else if base != "" {
			inner = base
		} else if index != "" {
			inner = index
		}

		if inner != "" && disp != "" {
			if a.Disp >= 0 {
				inner = fmt.Sprintf("%s+%s", inner, disp)
			} else {
				inner = fmt.Sprintf("%s%s", inner, disp)
			}
		} else if inner == "" && disp != "" {
			inner = disp
		} else if inner == "" && disp == "" {
			inner = "0"
		}

		// 根据指令操作数大小确定内存大小前缀
		prefix := ""
		// 由于x86asm.Mem没有Size字段，我们根据指令类型和上下文来判断
		// 这里使用一个固定值作为示例，实际应用中可能需要更复杂的逻辑
		operandSize := getOperandSizeFromContext()

		switch operandSize {
		case 1:
			prefix = "byte ptr "
		case 2:
			prefix = "word ptr "
		case 4:
			prefix = "dword ptr "
		case 8:
			prefix = "qword ptr "
		case 10:
			prefix = "tbyte ptr "
		case 16:
			prefix = "xmmword ptr "
		case 32:
			prefix = "ymmword ptr "
		}

		return fmt.Sprintf("%s[%s]", prefix, inner)

	case x86asm.Imm:
		return fmt.Sprintf("0x%x", a)

	case x86asm.Rel:
		return fmt.Sprintf("0x%x", uint64(a))

	default:
		return fmt.Sprintf("%v", a)
	}
}

// 从上下文中获取操作数大小（这里是简化实现）
func getOperandSizeFromContext() int {
	// 在实际应用中，这应该基于指令和架构模式来确定
	// 这里默认返回8表示64位架构下的默认大小
	return 8
}

// formatATTArg 格式化AT&T语法的操作数
func formatATTArg(arg x86asm.Arg) string {
	switch a := arg.(type) {
	case x86asm.Reg:
		return "%" + a.String()

	case x86asm.Mem:
		// 构建内存引用字符串
		var base, index, disp string

		if a.Base != 0 {
			base = "%" + a.Base.String()
		}

		if a.Index != 0 {
			index = fmt.Sprintf("%%%s,%d", a.Index, a.Scale)
		}

		if a.Disp != 0 {
			if a.Disp < 0 {
				disp = fmt.Sprintf("-0x%x", -a.Disp)
			} else {
				disp = fmt.Sprintf("0x%x", a.Disp)
			}
		}

		// 组合内存引用部分
		var inner string
		if disp != "" {
			inner = disp
		}

		if base != "" {
			if inner != "" {
				inner = fmt.Sprintf("%s(%s", inner, base)
			} else {
				inner = fmt.Sprintf("(%s", base)
			}
		} else if inner != "" {
			inner = fmt.Sprintf("%s(", inner)
		} else {
			inner = "("
		}

		if index != "" {
			inner = fmt.Sprintf("%s,%s", inner, index)
		}

		inner = inner + ")"

		// 根据指令操作数大小确定内存大小前缀
		prefix := ""
		// 由于x86asm.Mem没有Size字段，我们根据指令类型和上下文来判断
		operandSize := getOperandSizeFromContext()

		switch operandSize {
		case 1:
			prefix = "b"
		case 2:
			prefix = "w"
		case 4:
			prefix = "l"
		case 8:
			prefix = "q"
		}

		if prefix != "" {
			return fmt.Sprintf("%s%s", prefix, inner)
		}
		return inner

	case x86asm.Imm:
		return fmt.Sprintf("$0x%x", a)

	case x86asm.Rel:
		return fmt.Sprintf("0x%x", uint64(a))

	default:
		return fmt.Sprintf("%v", a)
	}
}
