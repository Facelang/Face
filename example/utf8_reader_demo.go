package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"face-lang/compiler/common/reader"
)

func main() {
	// 创建一个包含各种 UTF-8 字符的测试文件
	testContent := `// 这是一个包含多种字符的测试文件
package main

import "fmt"

func main() {
	// 英文字符
	message := "Hello World"
	
	// 中文字符
	中文 := "你好世界"
	
	// 日文字符
	日本語 := "こんにちは"
	
	// emoji 字符
	emoji := "🌍🚀💻"
	
	fmt.Println(message)
	fmt.Println(中文)
	fmt.Println(日本語)
	fmt.Println(emoji)
}`

	// 创建测试文件
	testFile := "test_utf8_source.go"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		fmt.Printf("创建测试文件失败: %v\n", err)
		return
	}
	defer os.Remove(testFile)

	fmt.Printf("=== UTF-8 文件读取器演示 ===\n\n")
	fmt.Printf("读取文件: %s\n", testFile)
	fmt.Printf("文件大小: %d 字节\n", len([]byte(testContent)))
	fmt.Printf("字符数量: %d 个\n\n", len([]rune(testContent)))

	// 获取绝对路径
	absPath, err := filepath.Abs(testFile)
	if err != nil {
		fmt.Printf("获取绝对路径失败: %v\n", err)
		return
	}

	// 使用我们的 UTF-8 读取器
	fileReader := reader.FileReader(absPath)
	if fileReader.GetFile() == nil {
		fmt.Printf("创建文件读取器失败\n")
		return
	}

	fmt.Printf("开始逐字符读取...\n\n")

	lineNum := 1
	colNum := 1
	charCount := 0

	for {
		r, err := fileReader.ReadRune()
		if err != nil {
			if err == io.EOF {
				fmt.Printf("\n文件读取完成！\n")
				break
			}
			fmt.Printf("读取错误: %v\n", err)
			break
		}

		charCount++

		// 显示字符信息
		if r == '\n' {
			fmt.Printf("第%d行第%d列: \\n (换行符, U+%04X)\n", lineNum, colNum, r)
			lineNum++
			colNum = 1
		} else if r == '\t' {
			fmt.Printf("第%d行第%d列: \\t (制表符, U+%04X)\n", lineNum, colNum, r)
			colNum++
		} else if r < 32 || r == 127 {
			fmt.Printf("第%d行第%d列: 控制字符 (U+%04X)\n", lineNum, colNum, r)
			colNum++
		} else {
			fmt.Printf("第%d行第%d列: '%c' (U+%04X)\n", lineNum, colNum, r, r)
			colNum++
		}

		// 只显示前50个字符，避免输出过长
		if charCount >= 50 {
			fmt.Printf("...(省略剩余字符)\n")
			break
		}
	}

	fmt.Printf("\n=== 统计信息 ===\n")
	fmt.Printf("已读取字符数: %d\n", charCount)
	fmt.Printf("当前行号: %d\n", lineNum)
	fmt.Printf("当前列号: %d\n", colNum)

	// 演示错误处理 - 读取不存在的文件
	fmt.Printf("\n=== 错误处理演示 ===\n")
	badReader := reader.FileReader("不存在的文件.txt")
	_, err = badReader.ReadRune()
	if err != nil {
		fmt.Printf("预期的错误: %v\n", err)
	}

	fmt.Printf("\n演示完成！\n")
}
