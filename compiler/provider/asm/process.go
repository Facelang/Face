package asm

import "os"

// LabelRecord 符号记录
type LabelRecord struct {
	LbName    string // 标签名
	Addr      int    // 地址
	Externed  bool   // 是否为外部符号
	IsEqu     bool   // 是否为EQU定义的符号
	EquTarget string // 跳转别名
	SegName   string // 段名
	Times     int    // 重复次数
	Len       int    // 字节长度
	Cont      []int  // 内容
	ContLen   int    // 内容长度
}

func NewRec(name string, ex bool) *LabelRecord {
	lb := &LabelRecord{
		LbName:   name,
		Addr:     ProcessTable.CurSegOff,
		Externed: ex,
		IsEqu:    false,
		SegName:  ProcessTable.CurSegName,
		Times:    0,
		Len:      0,
		Cont:     nil,
		ContLen:  0,
	}

	if ex {
		lb.Addr = 0
		lb.SegName = ""
	}

	return lb
}

// NewRecWithAddr 创建EQU符号记录， 记录为跳转别名
func NewRecWithAddr(name, target string) *LabelRecord {
	return &LabelRecord{
		LbName:    name,
		SegName:   ProcessTable.CurSegName,
		EquTarget: target,
		IsEqu:     true,
		Externed:  false,
		Times:     0,
		Len:       0,
		Cont:      nil,
		ContLen:   0,
	}
}

// NewRecWithData 添加符号时，没有地址 Addr 默认为
func NewRecWithData(name string, times int, length int, cont []int, contLen int) *LabelRecord {
	lb := &LabelRecord{
		LbName:   name,
		Addr:     ProcessTable.CurSegOff,
		SegName:  ProcessTable.CurSegName,
		IsEqu:    false,
		Times:    times,
		Len:      length,
		ContLen:  contLen,
		Externed: false,
	}

	// 复制内容
	lb.Cont = make([]int, contLen) // 这里面是值
	copy(lb.Cont, cont)

	// 更新地址
	ProcessTable.CurSegOff += times * length * contLen

	return lb
}

// Write 写入数据到输出文件
func (lb *LabelRecord) write() {
	for i := 0; i < lb.Times; i++ {
		for j := 0; j < lb.ContLen; j++ {
			WriteBytes(lb.Cont[j], lb.Len)
		}
	}
}

// TemporaryTable 临时表，主要记录解析过程的符号信息
type TemporaryTable struct {
	CurSegOff    int                     // 当前段地址偏移
	CurSegName   string                  // 当前段名称
	DataLen      int                     // 总数据长度
	TempFile     *os.File                // 临时文件
	MapLabel     map[string]*LabelRecord // 符号映射表
	DefLabelList []*LabelRecord          // 定义的符号列表
}

// NewTemporaryTable 创建新的符号表
func NewTemporaryTable() *TemporaryTable {
	return &TemporaryTable{
		CurSegOff:    0,
		CurSegName:   "",
		DataLen:      0,
		MapLabel:     make(map[string]*LabelRecord),
		DefLabelList: make([]*LabelRecord, 0),
	}
}

// Exist 检查符号表中是否有指定名称的符号
func (t *TemporaryTable) Exist(name string) bool {
	_, exists := t.MapLabel[name]
	return exists
}

// AddLabel 添加符号到符号表
func (t *TemporaryTable) AddLabel(lb *LabelRecord) {
	//if common.ScanLop != 1 { // 只在第一遍添加新符号
	//	return
	//}

	if t.Exist(lb.LbName) { // 符号存在
		if t.MapLabel[lb.LbName].Externed == true && lb.Externed == false { // 本地符号覆盖外部符号
			t.MapLabel[lb.LbName] = lb
		}
	} else {
		t.MapLabel[lb.LbName] = lb
	}

	// 包含数据段内容的符号：数据段内除了不含数据（times==0）的符号，外部符号段名为空
	if lb.Times != 0 && lb.SegName == ".data" {
		t.DefLabelList = append(t.DefLabelList, lb)
	}
}

// GetLabel 获取符号
func (t *TemporaryTable) GetLabel(name string) *LabelRecord {
	if t.Exist(name) {
		return t.MapLabel[name]
	}

	// 未知符号，添加为外部符号(待重定位)
	lb := NewRec(name, true)
	t.MapLabel[name] = lb
	return lb
}

// SwitchSeg 切换段 段名在这里修改 curSeg = id
func (t *TemporaryTable) SwitchSeg(id string) {
	//if ScanLop == 1 {
	// 确保段对齐到4字节边界
	t.DataLen += (4 - t.DataLen%4) % 4

	// 记录上一个段
	ObjFile.addShdr(t.CurSegName, uint32(ProcessTable.CurSegOff))

	if t.CurSegName != ".bss" {
		t.DataLen += ProcessTable.CurSegOff
	}

	t.CurSegName = id          // 切换到下一个段
	ProcessTable.CurSegOff = 0 // 清0段偏移
}

// Exports 导出符号表
func (t *TemporaryTable) Exports() {
	for _, lb := range t.MapLabel {
		if !lb.IsEqu { // EQU定义的符号不导出
			ObjFile.addSym(lb)
		}
	}
}

// Write 写入符号数据
func (t *TemporaryTable) Write() {
	for _, lb := range t.DefLabelList {
		lb.write()
	}
}

var ProcessTable = NewTemporaryTable()
