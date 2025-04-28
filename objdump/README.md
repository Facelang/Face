# Go Objdump

一个用Go语言编写的x86_64机器码逆向解析工具，用于解析ELF文件并将代码段反汇编为Intel或AT&T汇编语法。

## 功能

- 支持ELF文件格式解析
- 支持x86_64架构（32位和64位）
- 支持Intel和AT&T汇编语法
- 显示代码段反汇编
- 显示符号表
- 显示ELF头信息
- 显示所有段信息

## 目标文件测试
项目在 x86_64 linux 环境下，使用 objdump 逆向编译为 intel 风格汇编代码，可以用于项目测试，具体如下：
```bash
objdump -d -M intel -j .text main.o

main.o:     file format elf64-x86-64


Disassembly of section .text:

0000000000000000 <main>:
   0:   55                      push   rbp
   1:   48 89 e5                mov    rbp,rsp
   4:   48 83 ec 10             sub    rsp,0x10
   8:   c7 45 fc 02 00 00 00    mov    DWORD PTR [rbp-0x4],0x2
   f:   8b 15 00 00 00 00       mov    edx,DWORD PTR [rip+0x0]        # 15 <main+0x15>
  15:   8b 45 fc                mov    eax,DWORD PTR [rbp-0x4]
  18:   01 d0                   add    eax,edx
  1a:   89 45 f8                mov    DWORD PTR [rbp-0x8],eax
  1d:   8b 45 f8                mov    eax,DWORD PTR [rbp-0x8]
  20:   89 c6                   mov    esi,eax
  22:   48 8d 05 00 00 00 00    lea    rax,[rip+0x0]        # 29 <main+0x29>
  29:   48 89 c7                mov    rdi,rax
  2c:   b8 00 00 00 00          mov    eax,0x0
  31:   e8 00 00 00 00          call   36 <main+0x36>
  36:   b8 00 00 00 00          mov    eax,0x0
  3b:   c9                      leave
  3c:   c3                      ret
```

## 使用方法

```bash
# 基本用法
./objdump [选项] <文件>

# 选项
-format string   输出格式: intel 或 att (默认 "intel")
-all             显示所有节的内容
-text            显示代码段内容 (默认 true)
-symbols         显示符号表
-headers         显示ELF头信息
```

## 示例

```bash
# 使用Intel语法反汇编可执行文件
./objdump -format=intel -text test/hello

# 显示ELF头信息
./objdump -headers test/hello

# 显示符号表
./objdump -symbols test/hello

# 显示所有信息
./objdump -all -headers -symbols test/hello
```

## 构建测试程序

在`test`目录下有一个简单的C语言程序，可以用来测试工具：

```bash
# 编译测试程序
cd test
gcc -o hello hello.c

# 使用objdump工具反汇编
../objdump -format=intel -text -headers hello
```

## 依赖项

- golang.org/x/arch/x86/x86asm - 用于x86_64指令解码 