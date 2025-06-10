package utils

import (
	"encoding/binary"
	"math"
	"strconv"
)

func Float2Bytes(val float64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, math.Float64bits(val))
	return buf
}

func Float(lit string) float64 {
	val, err := strconv.ParseFloat(lit, 64)
	if err != nil {
		panic("无效的浮点数: " + lit)
	}
	return val
}

func FloatBytes(lit string) []byte {
	return Float2Bytes(Float(lit))
}

func Int2Bytes(val int64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(val))
	return buf
}

func Int(lit string) []byte {
	if lit == "" {
		return make([]byte, 8) // 返回8字节的0
	}
	var val int64
	if lit[0] == '0' {
		if len(lit) == 1 {
			return make([]byte, 8)
		}
		switch lit[1] {
		case 'b', 'B': // 二进制
			v, err := strconv.ParseInt(lit[2:], 2, 64)
			if err != nil {
				panic("无效的二进制数字: " + lit)
			}
			val = v
		case 'x', 'X': // 十六进制
			v, err := strconv.ParseInt(lit[2:], 16, 64)
			if err != nil {
				panic("无效的十六进制数字: " + lit)
			}
			val = v
		case 'o', 'O': // 八进制
			v, err := strconv.ParseInt(lit[2:], 8, 64)
			if err != nil {
				panic("无效的八进制数字: " + lit)
			}
			val = v
		default: // 八进制（以0开头）
			v, err := strconv.ParseInt(lit, 8, 64)
			if err != nil {
				panic("无效的八进制数字: " + lit)
			}
			val = v
		}
	} else {
		// 十进制
		v, err := strconv.ParseInt(lit, 10, 64)
		if err != nil {
			panic("无效的十进制数字: " + lit)
		}
		val = v
	}
	return val
}

func IntBytes(lit string) []byte {
	return Int2Bytes(Int(lit))
}
