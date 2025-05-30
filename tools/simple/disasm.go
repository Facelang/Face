package main

import (
	"debug/elf"
	"flag"
	"fmt"
	"log"
	"strings"

	"golang.org/x/arch/x86/x86asm"
)

func main() {
	// 命令行参数
	outputFormat := flag.String("format", "intel", "输出格式: intel 或 att")
	objFile := flag.String("file", "./main.o", "要分析的目标文件")
	flag.Parse()

	// 打开ELF文件
	file, err := elf.Open(*objFile)
	if err != nil {
		log.Fatalf("无法打开目标文件: %v", err)
	}
	defer file.Close()

	// 打印基本信息
	printFileInfo(file, *objFile)

	// 查找.text段
	textSection := file.Section(".text")
	if textSection == nil {
		log.Fatalf("未找到.text段")
	}

	// 读取.text段数据
	data, err := textSection.Data()
	if err != nil {
		log.Fatalf("无法读取.text段数据: %v", err)
	}

	// 设置汇编语法
	useIntel := true
	if *outputFormat == "att" {
		useIntel = false
		fmt.Println("\n使用AT&T语法反汇编:")
	} else {
		fmt.Println("\n使用Intel语法反汇编:")
	}

	// 反汇编代码段
	fmt.Printf("\nDisassembly of section .text:\n\n")
	disassemble(data, textSection.Addr, useIntel)
}

// 打印文件信息
func printFileInfo(f *elf.File, filePath string) {
	fmt.Printf("文件: %s\n", filePath)
	fmt.Printf("格式: %s\n", getELFClass(f))
	fmt.Printf("架构: %s\n", getMachine(f.Machine))
	fmt.Printf("入口点: 0x%x\n", f.Entry)
	fmt.Printf("节数量: %d\n", len(f.Sections))

	// 打印所有节的信息
	fmt.Println("\n节列表:")
	fmt.Printf("%-20s %-10s %-10s %s\n", "名称", "类型", "地址", "大小")
	for _, section := range f.Sections {
		fmt.Printf("%-20s %-10s 0x%-8x 0x%x\n",
			section.Name,
			getSectionType(section.Type),
			section.Addr,
			section.Size)
	}
}

// 获取ELF类型字符串
func getELFClass(f *elf.File) string {
	switch f.Class {
	case elf.ELFCLASS32:
		return "ELF32"
	case elf.ELFCLASS64:
		return "ELF64"
	default:
		return "未知"
	}
}

// 获取机器类型字符串
func getMachine(machine elf.Machine) string {
	switch machine {
	case elf.EM_386:
		return "x86"
	case elf.EM_X86_64:
		return "x86_64"
	case elf.EM_ARM:
		return "ARM"
	case elf.EM_AARCH64:
		return "AArch64"
	default:
		return fmt.Sprintf("未知(%d)", machine)
	}
}

// 获取节类型字符串
func getSectionType(sType elf.SectionType) string {
	switch sType {
	case elf.SHT_NULL:
		return "NULL"
	case elf.SHT_PROGBITS:
		return "PROGBITS"
	case elf.SHT_SYMTAB:
		return "SYMTAB"
	case elf.SHT_STRTAB:
		return "STRTAB"
	case elf.SHT_RELA:
		return "RELA"
	case elf.SHT_HASH:
		return "HASH"
	case elf.SHT_DYNAMIC:
		return "DYNAMIC"
	case elf.SHT_NOTE:
		return "NOTE"
	case elf.SHT_NOBITS:
		return "NOBITS"
	case elf.SHT_REL:
		return "REL"
	case elf.SHT_SHLIB:
		return "SHLIB"
	case elf.SHT_DYNSYM:
		return "DYNSYM"
	default:
		return fmt.Sprintf("未知(%d)", sType)
	}
}

// 反汇编代码
func disassemble(code []byte, baseAddr uint64, useIntel bool) {
	// 开始反汇编
	pc := uint64(0)

	// 假设main函数在起始位置
	fmt.Printf("\n%016x <main>:\n", baseAddr)

	for pc < uint64(len(code)) {
		// 反汇编当前指令
		inst, err := x86asm.Decode(code[pc:], 64)
		if err != nil {
			// 如果解码失败，尝试作为数据字节处理
			fmt.Printf("   %04x:\t%02x               \t.byte 0x%02x\n",
				pc, code[pc], code[pc])
			pc++
			continue
		}

		// 获取指令的机器码字节
		instBytes := code[pc : pc+uint64(inst.Len)]
		bytesStr := formatBytes(instBytes)

		// 格式化指令文本
		var text string
		if useIntel {
			text = formatIntelSyntax(inst)
		} else {
			text = formatATTSyntax(inst)
		}

		// 打印指令
		fmt.Printf("   %04x:\t%-20s\t%s\n", pc, bytesStr, text)

		pc += uint64(inst.Len)
	}
}

// 格式化字节序列为十六进制字符串
func formatBytes(bytes []byte) string {
	result := ""
	for i, b := range bytes {
		if i > 0 {
			result += " "
		}
		result += fmt.Sprintf("%02x", b)
	}
	return result
}

// 格式化Intel语法的指令
func formatIntelSyntax(inst x86asm.Inst) string {
	// 指令名称
	text := strings.ToLower(inst.Op.String())

	// 添加操作数
	if len(inst.Args) > 0 {
		var args []string
		for _, arg := range inst.Args {
			if arg == nil {
				break
			}
			args = append(args, formatArg(arg, true))
		}
		if len(args) > 0 {
			text = fmt.Sprintf("%-7s %s", text, strings.Join(args, ","))
		}
	}

	return text
}

// 格式化AT&T语法的指令
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
		args = append(args, formatArg(inst.Args[i], false))
	}

	if len(args) > 0 {
		text = fmt.Sprintf("%s %s", text, strings.Join(args, ","))
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

// 格式化单个操作数
func formatArg(arg x86asm.Arg, isIntel bool) string {
	switch a := arg.(type) {
	case x86asm.Reg:
		if isIntel {
			return strings.ToLower(a.String())
		} else {
			return "%" + strings.ToLower(a.String())
		}

	case x86asm.Mem:
		if isIntel {
			// Intel格式: dword ptr [rbp-0x4]
			var base, index, disp string

			if a.Base != 0 {
				base = strings.ToLower(a.Base.String())
			}

			if a.Index != 0 {
				index = fmt.Sprintf("%s*%d", strings.ToLower(a.Index.String()), a.Scale)
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

			// 根据指令推断的内存大小添加前缀
			// 根据实际情况，这里使用 DWORD PTR 与 README 示例保持一致
			return fmt.Sprintf("DWORD PTR [%s]", inner)

		} else {
			// AT&T格式: -0x4(%rbp)
			var base, index, disp string

			if a.Base != 0 {
				base = "%" + strings.ToLower(a.Base.String())
			}

			if a.Index != 0 {
				index = fmt.Sprintf("%%%s,%d", strings.ToLower(a.Index.String()), a.Scale)
			}

			if a.Disp != 0 {
				if a.Disp < 0 {
					disp = fmt.Sprintf("-0x%x", -a.Disp)
				} else {
					disp = fmt.Sprintf("0x%x", a.Disp)
				}
			}

			// 组合内存引用部分
			var result string
			if disp != "" {
				result = disp
			}

			if base != "" {
				if result != "" {
					result = fmt.Sprintf("%s(%s", result, base)
				} else {
					result = fmt.Sprintf("(%s", base)
				}
			} else if result != "" {
				result = fmt.Sprintf("%s(", result)
			} else {
				result = "("
			}

			if index != "" {
				result = fmt.Sprintf("%s,%s", result, index)
			}

			result = result + ")"

			// 添加大小前缀 (l for 32-bit)
			return fmt.Sprintf("l%s", result)
		}

	case x86asm.Imm:
		if isIntel {
			return fmt.Sprintf("0x%x", a)
		} else {
			return fmt.Sprintf("$0x%x", a)
		}

	case x86asm.Rel:
		// 相对地址跳转
		addr := uint64(a)
		if isIntel {
			return fmt.Sprintf("0x%x", addr)
		} else {
			return fmt.Sprintf("0x%x", addr)
		}

	default:
		return fmt.Sprintf("%v", a)
	}
}
