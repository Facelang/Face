package main

import "time"

type LimitCtl struct {
	Max        int   // 最大限制
	Interval   int   // 时间间隔
	IsFull     bool  // 是否填充满
	expires    []int // 过期时间列表
	start, end int   // 指针
}

func (l *LimitCtl) Accept() bool {
	t := int(time.Now().UnixNano())
	println(l.start, l.end, (l.expires[l.start]-t)/1e6)
	if t < l.expires[l.start] { // 可以插入
		// 判断最大数量控制, 数据写满, 判断下一次写入，是否填满

		// 这里时保存下一次状态， 直接返回会少缓存一次
		if l.end == l.start {
			return false
		}
	} else {
		// 第一条记录已过期, 丢一条
		l.IsFull = false
		l.start = (l.start + 1) % l.Max
	}

	// 剩余情况，写入一条
	l.expires[l.end] = t + l.Interval // 添加到末尾
	l.end = (l.end + 1) % l.Max       // 移动指针, 环形数组，需要取模

	return true
}

func NewLimit(max, interval int) *LimitCtl {
	return &LimitCtl{Max: max, Interval: interval, expires: make([]int, max), start: 0, end: 1}
}

var timeLimit []time.Duration
var maxLimit int = 200

func main() {
	limit := NewLimit(4, int(time.Second))

	for i := 0; i < 30; i++ {
		if limit.Accept() {
			println("PASS")
		} else {
			println("限流！")
		}
		time.Sleep(time.Millisecond * 100)
	}
}
