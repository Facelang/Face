package asm

//// 双操作码指令操作码表
//var i2Opcode = []byte{
//	//      8位操作数                 |      32位操作数
//	// r,r  r,rm|rm,r r,im|r,r  r,rm|rm,r r,im
//	0x88, 0x8a, 0x88, 0xb0, 0x89, 0x8b, 0x89, 0xb8, // mov
//	0x38, 0x3a, 0x38, 0x80, 0x39, 0x3b, 0x39, 0x81, // cmp
//	0x28, 0x2a, 0x28, 0x80, 0x29, 0x2b, 0x29, 0x81, // sub
//	0x00, 0x02, 0x00, 0x80, 0x01, 0x03, 0x01, 0x81, // add
//	0x00, 0x00, 0x00, 0x00, 0x00, 0x8d, 0x00, 0x00, // lea
//}
//
//// 单操作码指令操作码表
//var i1opcode = []uint16{
//	0xe8, 0xcd /*0xfe,*/, 0xf7, 0xf7, 0xf7, 0x40, 0x48, 0xe9, // call,int,imul,idiv,neg,inc,dec,jmp<rel32>
//	0x0f84, 0x0f8f, 0x0f8c, 0x0f8d, 0x0f8e, 0x0f85, 0x0f86, // je,jg,jl,jge,jle,jne,jna<rel32>
//	// 0xeb,//jmp rel8
//	// 0x74,0x7f,0x7c,0x7d,0x7e,0x75,0x76,//je,jg,jl,jge,jle,jne,jna<rel8>
//	/*0x68,*/ 0x50, // push
//	0x58,           // pop
//}
//
//// 零操作码指令操作码表
//var i0Opcode = []byte{
//	0xc3, // ret
//}
//
//// 重定位类型常量
//const (
//	R_386_32   = 1 // 绝对寻址
//	R_386_PC32 = 2 // 相对寻址
//)

//// ProcessRel 处理可能的重定位信息
//func ProcessRel(relType int) bool {
//	if ScanLop == 1 || RelLb == nil {
//		RelLb = nil
//		return false
//	}
//
//	flag := false
//	if relType == R_386_32 { // 绝对重定位
//		if RelLb.IsEqu { // 只要是地址符号就必须重定位，宏除外
//			//ObjFile.AddRel(CurSeg, ProcessTable.CurSegOff, RelLb.LbName, relType)
//			flag = true
//		}
//	} else if relType == R_386_PC32 { // 相对重定位
//		if RelLb.Externed { // 对于跳转，内部的不需要重定位，外部的需要重定位
//			//ObjFile.AddRel(CurSeg, ProcessTable.CurSegOff, RelLb.LbName, relType)
//			flag = true
//		}
//	}
//
//	RelLb = nil
//	return flag
//}
