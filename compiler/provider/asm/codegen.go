package asm

// 双操作码指令操作码表
var i2Opcode = []byte{
	//      8位操作数                 |      32位操作数
	// r,r  r,rm|rm,r r,im|r,r  r,rm|rm,r r,im
	0x88, 0x8a, 0x88, 0xb0, 0x89, 0x8b, 0x89, 0xb8, // mov
	0x38, 0x3a, 0x38, 0x80, 0x39, 0x3b, 0x39, 0x81, // cmp
	0x28, 0x2a, 0x28, 0x80, 0x29, 0x2b, 0x29, 0x81, // sub
	0x00, 0x02, 0x00, 0x80, 0x01, 0x03, 0x01, 0x81, // add
	0x00, 0x00, 0x00, 0x00, 0x00, 0x8d, 0x00, 0x00, // lea
}

// 单操作码指令操作码表
var i1opcode = []uint16{
	0xe8, 0xcd /*0xfe,*/, 0xf7, 0xf7, 0xf7, 0x40, 0x48, 0xe9, // call,int,imul,idiv,neg,inc,dec,jmp<rel32>
	0x0f84, 0x0f8f, 0x0f8c, 0x0f8d, 0x0f8e, 0x0f85, 0x0f86, // je,jg,jl,jge,jle,jne,jna<rel32>
	// 0xeb,//jmp rel8
	// 0x74,0x7f,0x7c,0x7d,0x7e,0x75,0x76,//je,jg,jl,jge,jle,jne,jna<rel8>
	/*0x68,*/ 0x50, // push
	0x58,           // pop
}

// 零操作码指令操作码表
var i0Opcode = []byte{
	0xc3, // ret
}

// 重定位类型常量
const (
	R_386_32   = 1 // 绝对寻址
	R_386_PC32 = 2 // 相对寻址
)

// WriteModRM 输出ModRM字节
func WriteModRM() {
	if instr.modrm.mod != -1 { // 有效
		mrm := byte(((instr.modrm.mod & 0x00000003) << 6) + ((instr.modrm.reg & 0x0000007) << 3) + (instr.modrm.rm & 0x00000007))
		WriteBytes(int(mrm), 1)
	}
}

// WriteSIB 输出SIB字节
func WriteSIB() {
	if instr.sib.scale != -1 {
		s := byte(((instr.sib.scale & 0x00000003) << 6) + ((instr.sib.index & 0x0000007) << 3) + (instr.sib.base & 0x00000007))
		WriteBytes(int(s), 1)
	}
}

// ProcessRel 处理可能的重定位信息
func ProcessRel(relType int) bool {
	if ScanLop == 1 || RelLb == nil {
		RelLb = nil
		return false
	}

	flag := false
	if relType == R_386_32 { // 绝对重定位
		if RelLb.IsEqu { // 只要是地址符号就必须重定位，宏除外
			//ObjFile.AddRel(CurSeg, ProcessTable.CurSegOff, RelLb.LbName, relType)
			flag = true
		}
	} else if relType == R_386_PC32 { // 相对重定位
		if RelLb.Externed { // 对于跳转，内部的不需要重定位，外部的需要重定位
			//ObjFile.AddRel(CurSeg, ProcessTable.CurSegOff, RelLb.LbName, relType)
			flag = true
		}
	}

	RelLb = nil
	return flag
}

// Gen2op 生成双操作数指令
func Gen2op(opt Token, desType, srcType, length int) {
	// lb_record::curAddr=0;
	// 测试信息
	//     cout<<"len="<<len<<"(1-Byte;4-DWord)\n";
	//     cout<<"des:type="<<des_t<<"(1-imm;2-mem;3-reg)\n";
	//     cout<<"src:type="<<src_t<<"(1-imm;2-mem;3-reg)\n";
	//  cout<<"ModR/M="<<modrm.mod<<" "<<modrm.reg<<" "<<modrm.rm<<endl;
	// cout<<"SIB="<<sib.scale<<" "<<sib.index<<" "<<sib.base<<endl;
	// cout<<"disp32="<<instr.disp32<<",disp8="<<(int)instr.disp8<<"(<-"<<instr.disptype<<":(0-disp8;1-disp32) imm32="<<instr.imm32<<endl;
	// 计算操作码索引 (mov,8,reg,reg)=000 (mov,8,reg,mem)=001 (mov,8,mem,reg)=010 (mov,8,reg,imm)=011
	//(mov,32,reg,reg)=100 (mov,32,reg,mem)=101 (mov,32,mem,reg)=110 (mov,32,reg,imm)=111  [0-7]*(i_lea-i_mov)

	index := -1
	if srcType == OPR_IMMD { // 鉴别操作数种类
		index = 3
	} else {
		index = (desType-2)*2 + srcType - 2
	}
	// 附加指令名称和长度
	index = int(opt-I_MOV)*8 + (1-length%4)*4 + index
	opcode := i2Opcode[index] // 获取机器码

	switch instr.modrm.mod {
	case -1: // reg,imm // 初始值
		switch opt {
		case I_MOV: // b0+rb MOV r/m8,imm8 b8+rd MOV r/m32,imm32
			opcode += byte(instr.modrm.reg)
			WriteBytes(int(opcode), 1)
			break
		case I_CMP: // 80 /7 ib CMP r/m8,imm8 81 /7 id CMP r/m32,imm32
			WriteBytes(int(opcode), 1)
			exchar := 0xf8
			exchar += instr.modrm.reg
			WriteBytes(exchar, 1)
			break
		case I_ADD: // 80 /0 ib ADD r/m8, imm8 81 /0 id ADD r/m32, imm32
			WriteBytes(int(opcode), 1)
			exchar := 0xc0
			exchar += instr.modrm.reg
			WriteBytes(exchar, 1)
			break
		case I_SUB: // 80 /5 ib SUB r/m8, imm8 81 /5 id SUB r/m32, imm32
			WriteBytes(int(opcode), 1)
			exchar := 0xe8
			exchar += instr.modrm.reg
			WriteBytes(exchar, 1)
			break
		}
		// 可能的重定位位置 mov eax,@buffer,也有可能是mov eax,@buffer_len，就不许要重定位，因为是宏
		ProcessRel(R_386_32)
		WriteBytes(instr.Imm32, length) // 一定要按照长度输出立即数
	case 0: // [reg],reg 或 reg,[reg]
		WriteBytes(int(opcode), 1)
		WriteModRM()
		if instr.modrm.rm == 5 { //[disp32]
			ProcessRel(R_386_32) // 可能是mov eax,[@buffer],后边disp8和disp32不会出现类似情况
			instr.writeDisp()    // 地址肯定是4字节长度
		} else if instr.modrm.rm == 4 { // SIB
			WriteSIB()
		}
	case 1: //[reg+disp8],reg reg,[reg+disp8]
		WriteBytes(int(opcode), 1)
		WriteModRM()
		if instr.modrm.rm == 4 {
			WriteSIB()
		}
		instr.writeDisp()
		break
	case 2: //[reg+disp32],reg reg,[reg+disp32]
		WriteBytes(int(opcode), 1)
		WriteModRM()
		if instr.modrm.rm == 4 {
			WriteSIB()
		}
		instr.writeDisp()
		break
	case 3: // reg,reg
		WriteBytes(int(opcode), 1)
		WriteModRM()
		break
	}

}

// Gen1op 生成单操作数指令
func Gen1op(opt Token, oprType, length int) {
	opcode := int(i1opcode[opt-I_CALL])
	if opt == I_CALL || opt >= I_JMP && opt <= I_JNA {
		// 统一使用长地址跳转，短跳转不好定位
		if opt == I_CALL || opt == I_JMP {
			WriteBytes(opcode, 1)
		} else {
			WriteBytes(opcode>>8, 1)
			WriteBytes(opcode, 1)
		}

		rel := instr.Imm32 - (ProcessTable.CurSegOff + 4) // 调用符号地址相对于下一条指令地址的偏移，因此加4
		ret := ProcessRel(R_386_PC32)                     // 处理可能的相对重定位信息，call fun,如果fun是本地定义的函数就不会重定位了
		if ret {                                          // 相对重定位成功，说明之前计算的偏移错误
			rel = -4 // 对于链接器必须的初始值
		}
		WriteBytes(rel, 4)
	} else if opt == I_INT {
		WriteBytes(opcode, 1)
		WriteBytes(instr.Imm32, 1)
	} else if opt == I_PUSH {
		if oprType == OPR_IMMD {
			opcode = 0x68
			WriteBytes(opcode, 1)
			WriteBytes(instr.Imm32, 4)
		} else {
			opcode += instr.modrm.reg
			WriteBytes(opcode, 1)
		}
	} else if opt == I_INC {
		if length == 1 {
			opcode = 0xfe
			WriteBytes(opcode, 1)
			exchar := 0xc0
			exchar += instr.modrm.reg
			WriteBytes(exchar, 1)
		} else {
			opcode += instr.modrm.reg
			WriteBytes(opcode, 1)
		}
	} else if opt == I_DEC {
		if length == 1 {
			opcode = 0xfe
			WriteBytes(opcode, 1)
			exchar := 0xc8
			exchar += instr.modrm.reg
			WriteBytes(exchar, 1)
		} else {
			opcode += instr.modrm.reg
			WriteBytes(opcode, 1)
		}
	} else if opt == I_NEG {
		if length == 1 {
			opcode = 0xf6
		}
		exchar := 0xd8
		exchar += instr.modrm.reg
		WriteBytes(opcode, 1)
		WriteBytes(exchar, 1)
	} else if opt == I_POP {
		opcode += instr.modrm.reg
		WriteBytes(opcode, 1)
	} else if opt == I_IMUL || opt == I_IDIV {
		WriteBytes(opcode, 1)
		exchar := 0xf8
		if opt == I_IMUL {
			exchar = 0xe8
		}
		exchar += instr.modrm.reg
		WriteBytes(exchar, 1)
	}

}

func Gen0op(opt Token) {
	if opt != I_RET {
		return
	}
	WriteBytes(int(i0Opcode[0]), 1)
}
