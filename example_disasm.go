package main

//
//import (
//	"flag"
//	"fmt"
//
//	"golang.org/x/arch/x86/x86asm"
//)
//
//func main() {
//	outputFormat := flag.String("format", "intel", "输出格式: intel 或 att")
//	flag.Parse()
//
//	fmt.Println("x86_64样例机器码反汇编演示")
//	fmt.Println("============================")
//
//	// 一些简单的x86_64指令机器码
//	code := []byte{
//		0x55,             // push rbp
//		0x48, 0x89, 0xe5, // mov rbp, rsp
//		0x48, 0x83, 0xec, 0x10, // sub rsp, 0x10
//		0xc7, 0x45, 0xfc, 0x00, 0x00, 0x00, 0x00, // mov DWORD PTR [rbp-0x4], 0x0
//		0xb8, 0x00, 0x00, 0x00, 0x00, // mov eax, 0x0
//		0x48, 0x83, 0xc4, 0x10, // add rsp, 0x10
//		0x5d, // pop rbp
//		0xc3, // ret
//	}
//
//	// 设置汇编语法
//	useIntel := true
//	if *outputFormat == "att" {
//		useIntel = false
//		fmt.Println("使用AT&T语法:")
//	} else {
//		fmt.Println("使用Intel语法:")
//	}
//
//	// 反汇编
//	disassemble(code, useIntel)
//}
//
//// 反汇编代码
//func disassemble(code []byte, useIntel bool) {
//	// 打印表头
//	fmt.Println("地址     | 机器码                  | 汇编指令")
//	fmt.Println("---------+--------------------------+------------------")
//
//	// 逐字节反汇编
//	pc := uint64(0)
//	for pc < uint64(len(code)) {
//		inst, err := x86asm.Decode(code[pc:], 64)
//		if err != nil {
//			// 如果解码失败，尝试作为数据字节处理
//			fmt.Printf("0x%08x | %-24s | %s\n",
//				pc,
//				formatBytes(code[pc:pc+1]),
//				fmt.Sprintf(".byte 0x%02x", code[pc]))
//			pc++
//			continue
//		}
//
//		// 获取指令的机器码字节
//		instBytes := code[pc : pc+uint64(inst.Len)]
//		bytesStr := formatBytes(instBytes)
//
//		// 格式化指令文本
//		var text string
//		if useIntel {
//			text = formatIntelSyntax(inst)
//		} else {
//			text = formatATTSyntax(inst)
//		}
//
//		// 打印指令
//		fmt.Printf("0x%08x | %-24s | %s\n", pc, bytesStr, text)
//
//		pc += uint64(inst.Len)
//	}
//}
//
//// 格式化字节序列为十六进制字符串
//func formatBytes(bytes []byte) string {
//	result := ""
//	for i, b := range bytes {
//		if i > 0 {
//			result += " "
//		}
//		result += fmt.Sprintf("%02x", b)
//	}
//	return result
//}
//
//// 格式化Intel语法的指令
//func formatIntelSyntax(inst x86asm.Inst) string {
//	// 指令名称
//	text := inst.Op.String()
//
//	// 添加操作数
//	if len(inst.Args) > 0 {
//		var args []string
//		for _, arg := range inst.Args {
//			if arg == nil {
//				break
//			}
//			args = append(args, formatArg(arg, true))
//		}
//		if len(args) > 0 {
//			text = fmt.Sprintf("%-7s %s", text, formatArgs(args, true))
//		}
//	}
//
//	return text
//}
//
//// 格式化AT&T语法的指令
//func formatATTSyntax(inst x86asm.Inst) string {
//	// 指令名称
//	text := inst.Op.String() + "q" // 简化处理：添加q后缀表示64位操作
//
//	// 添加操作数（注意：AT&T语法中操作数顺序相反）
//	var args []string
//	for i := len(inst.Args) - 1; i >= 0; i-- {
//		if inst.Args[i] == nil {
//			continue
//		}
//		args = append(args, formatArg(inst.Args[i], false))
//	}
//
//	if len(args) > 0 {
//		text = fmt.Sprintf("%-7s %s", text, formatArgs(args, false))
//	}
//
//	return text
//}
//
//// 格式化操作数参数
//func formatArgs(args []string, isIntel bool) string {
//	if isIntel {
//		// Intel: op1, op2, op3
//		result := ""
//		for i, arg := range args {
//			if i > 0 {
//				result += ", "
//			}
//			result += arg
//		}
//		return result
//	} else {
//		// AT&T: op3, op2, op1
//		result := ""
//		for i, arg := range args {
//			if i > 0 {
//				result += ", "
//			}
//			result += arg
//		}
//		return result
//	}
//}
//
//// 格式化单个操作数
//func formatArg(arg x86asm.Arg, isIntel bool) string {
//	switch a := arg.(type) {
//	case x86asm.Reg:
//		if isIntel {
//			return a.String()
//		} else {
//			return "%" + a.String()
//		}
//
//	case x86asm.Mem:
//		if isIntel {
//			// Intel格式: dword ptr [rbp-0x4]
//			var base, index, disp string
//
//			if a.Base != 0 {
//				base = a.Base.String()
//			}
//
//			if a.Index != 0 {
//				index = fmt.Sprintf("%s*%d", a.Index, a.Scale)
//			}
//
//			if a.Disp != 0 {
//				if a.Disp < 0 {
//					disp = fmt.Sprintf("-0x%x", -a.Disp)
//				} else {
//					disp = fmt.Sprintf("0x%x", a.Disp)
//				}
//			}
//
//			// 组合内存引用部分
//			var inner string
//			if base != "" && index != "" {
//				inner = fmt.Sprintf("%s+%s", base, index)
//			} else if base != "" {
//				inner = base
//			} else if index != "" {
//				inner = index
//			}
//
//			if inner != "" && disp != "" {
//				if a.Disp >= 0 {
//					inner = fmt.Sprintf("%s+%s", inner, disp)
//				} else {
//					inner = fmt.Sprintf("%s%s", inner, disp)
//				}
//			} else if inner == "" && disp != "" {
//				inner = disp
//			} else if inner == "" && disp == "" {
//				inner = "0"
//			}
//
//			// 根据指令推断的内存大小添加前缀
//			return fmt.Sprintf("dword ptr [%s]", inner)
//		} else {
//			// AT&T格式: -0x4(%rbp)
//			var base, index, disp string
//
//			if a.Base != 0 {
//				base = "%" + a.Base.String()
//			}
//
//			if a.Index != 0 {
//				index = fmt.Sprintf("%%%s,%d", a.Index, a.Scale)
//			}
//
//			if a.Disp != 0 {
//				if a.Disp < 0 {
//					disp = fmt.Sprintf("-0x%x", -a.Disp)
//				} else {
//					disp = fmt.Sprintf("0x%x", a.Disp)
//				}
//			}
//
//			// 组合内存引用部分
//			var result string
//			if disp != "" {
//				result = disp
//			}
//
//			if base != "" {
//				if result != "" {
//					result = fmt.Sprintf("%s(%s", result, base)
//				} else {
//					result = fmt.Sprintf("(%s", base)
//				}
//			} else if result != "" {
//				result = fmt.Sprintf("%s(", result)
//			} else {
//				result = "("
//			}
//
//			if index != "" {
//				result = fmt.Sprintf("%s,%s", result, index)
//			}
//
//			result = result + ")"
//
//			// 添加大小前缀 (l for 32-bit)
//			return fmt.Sprintf("l%s", result)
//		}
//
//	case x86asm.Imm:
//		if isIntel {
//			return fmt.Sprintf("0x%x", a)
//		} else {
//			return fmt.Sprintf("$0x%x", a)
//		}
//
//	case x86asm.Rel:
//		// 相对地址跳转
//		addr := uint64(a)
//		if isIntel {
//			return fmt.Sprintf("0x%x", addr)
//		} else {
//			return fmt.Sprintf("0x%x", addr)
//		}
//
//	default:
//		return fmt.Sprintf("%v", a)
//	}
//}
