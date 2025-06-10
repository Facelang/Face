package internal

// instr 表示一条汇编指令
type instr struct {
	Opcode string    // 操作码
	Src    *operand  // 源操作数
	Dst    *operand  // 目标操作数
	Size   int       // 操作数大小(byte/word/dword/qword)
}
