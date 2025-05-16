package main

import (
	"fmt"
	"os"
)

func main() {
	// 文件对比
	name := os.Args[1]
	src, err := os.ReadFile(name + ".o")
	if err != nil {
		_, _ = fmt.Printf("源文件加载失败！%s\v", err.Error())
		return
	}
	dest, err := os.ReadFile(name + ".s.o")
	if err != nil {
		_, _ = fmt.Printf("目标文件加载失败！%s\v", err.Error())
		return
	}
	//if len(src) == 0 || len(src) != len(dest) {
	//	fmt.Printf("两文件内容长度不一致: [%d, %d]", len(src), len(dest))
	//}

	for i, ch := range src {
		if len(dest) <= i {
			fmt.Printf("错误：[0x%X],  两文件内容长度不一致: [%d, %d]", i, len(src), len(dest))
			return
		}
		if ch != dest[i] {
			fmt.Printf("错误:[0x%X, %d]（%X(%d) != %X）", i, i, ch, ch, dest[i])
			return
		}
	}

	fmt.Printf("校验完成，两文件完全一致！\n")

}
