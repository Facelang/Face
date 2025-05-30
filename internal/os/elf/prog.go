package elf

import "encoding/binary"

// Block 表示一个数据块
type Block struct {
	Data   []byte
	Offset uint32
	Size   uint32
}

// ProgSeg 表示段的列表, 还有两个方法： allocAddr, relocAddr
type ProgSeg struct {
	Name      string   // 段名称
	BaseAddr  uint32   // 分配基地址
	Offset    uint32   // 合并后的文件偏移
	Size      uint32   // 合并后大小
	Begin     uint32   // 对齐前开始位置偏移
	OwnerList []*File  // 拥有该段的文件序列
	Blocks    []*Block // 记录合并后的数据块序列
}

// AllocAddr 分配地址空间 base 是基址， off 是偏移
func (s *ProgSeg) AllocAddr(name string, base *uint32, off *uint32) {
	s.Begin = *off //记录对齐前偏移

	// 虚拟地址对齐，让所有的段按照4KB字节对齐
	if name != ".bss" {
		*base += (MemAlign - *base%MemAlign) % MemAlign
	}

	// 偏移地址对齐，让一般段按照4字节对齐，文本段按照16字节对齐
	align := uint32(DiscAlign)
	if name == ".text" {
		align = 16
	}
	*off += (align - *off%align) % align
	// 这里 off 的偏移和 base 的偏移不同

	// 使虚址和偏移按照4KB模同余
	*base = *base - *base%MemAlign + *off%MemAlign // todo 有些看不懂了

	// 累加地址和偏移
	s.BaseAddr = *base
	s.Offset = *off
	s.Size = 0
	for _, file := range s.OwnerList { // 拥有该段的所有文件，合并数据
		s.Size += (DiscAlign - s.Size%DiscAlign) % DiscAlign // 对齐每个小段，按照4字节，数据靠后
		seg := file.ShdrTab[name]
		//读取需要合并段的数据
		if name != ".bss" {
			buf := file.ReadData(seg.Offset, seg.Size)
			block := &Block{
				Data:   buf,
				Offset: s.Size, // 数据靠前靠后是否有区别？
				Size:   seg.Size,
			}
			s.Blocks = append(s.Blocks, block) // 添加到数据块
		}
		//修改每个文件中对应段的addr（seg 记录虚拟地址， 代表每一段数据在程序运行时加载到不同的地址段）
		seg.Addr = *base + s.Size //修改每个文件的段虚拟，为了方便计算符号或者重定位的虚址，不需要保存合并后文件偏移
		s.Size += seg.Size        //累加段大小
	}
	*base += s.Size // 基址也需要更新
	if name != ".bss" {
		*off += s.Size
	}
}

// RelocAddr 根据提供的重定位信息重定位地址
func (s *ProgSeg) RelocAddr(relAddr uint32, relocType uint8, symAddr uint32) {
	relOffset := relAddr - s.BaseAddr //同类合并段的数据偏移

	// 查找修正地址所在位置的数据块
	var targetBlock *Block
	for _, block := range s.Blocks {
		if block.Offset <= relOffset && block.Offset+block.Size > relOffset {
			targetBlock = block
			break
		}
	}
	if targetBlock == nil {
		return
	}

	//处理字节为b->data[relOffset-b->offset]
	// 获取需要修改的地址位置
	offset := relOffset - targetBlock.Offset
	if offset+4 > uint32(len(targetBlock.Data)) {
		return
	}

	// 获取当前地址值
	currentAddr := binary.LittleEndian.Uint32(targetBlock.Data[offset:])

	// 根据重定位类型进行修正
	switch relocType {
	case uint8(R_386_32): // 绝对地址修正
		binary.LittleEndian.PutUint32(targetBlock.Data[offset:], symAddr)
	case uint8(R_386_PC32): // 相对地址修正
		newAddr := symAddr - relAddr + currentAddr
		binary.LittleEndian.PutUint32(targetBlock.Data[offset:], newAddr)
	}
}
