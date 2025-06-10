package internal

import "fmt"

type LabelType uint8

const UNDEFINED_LABEL LabelType = 0 // 未定义
const TEXT_LABEL LabelType = 1      // 代码段符号
const EQU_LABEL LabelType = 2       // 常量
const LOCAL_LABEL LabelType = 3     // 局部变量
const EXTERNAL_LABEL LabelType = 4  // 外部变量, 提前申明的

type label struct {
	Name    string    // 标签名
	Type    LabelType // 标签类型
	Addr    int       // 地址
	Index   int       // 添加顺序， 从1开始
	Section string    // 段名
	Times   int       // 重复次数
	Size    int       // 字节长度
	Cont    []int     // 内容
	ContLen int       // 内容长度
	RelInfo bool      // 记录重定位信息
}

// AddLabel 添加符号到符号表; 一共三处，equ 常量 仅数字 NewRecWithEqu， 变量 NewRecWithData,  代码段 TextLabel
func (p *parser) AddLabel(name string, rec *label) {
	rec.Name = name // 缓存一次，减少后续查找名字
	if rec.Type == TEXT_LABEL || rec.Type == LOCAL_LABEL {
		rec.Addr = p.seg.Offset
		rec.Section = p.seg.Name
	}

	// 更新地址, 除了具体的变量定义，这里都是 0， 没有变化
	p.seg.Offset += rec.Times * rec.Size * rec.ContLen

	if i, ok := p.labelNames[name]; ok {
		labelRec := p.labelList[i]
		if labelRec.Type == UNDEFINED_LABEL {
			p.labelList[i] = rec // 直接替换
		} else {
			_ = fmt.Errorf("符号: %s 重复定义！", name)
		}
	} else {
		p.labelList = append(p.labelList, rec)
		p.labelNames[name] = len(p.labelList) - 1
	}
}

// GetLabel 获取符号
func (p *parser) GetLabel(name string) *label {
	if i, ok := p.labelNames[name]; ok {
		return p.labelList[i]
	}

	// 只有符号引用符号时， 才会被创建
	// 未知符号，添加为外部符号(待重定位)
	rec := NewLabel(UNDEFINED_LABEL)
	rec.Name = name
	p.labelList = append(p.labelList, rec)
	p.labelNames[name] = len(p.labelList) - 1
	return rec
}

func NewLabel(lType LabelType) *label {
	return &label{Type: lType}
}
