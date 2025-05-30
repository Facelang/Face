package parser

import (
	"testing"

	"github.com/facelang/face/compiler/common/reader"
	"github.com/facelang/face/compiler/common/tokens"
)

func TestDecimal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantTok  tokens.Token
		wantText string
		wantErr  bool
	}{
		// 十进制整数测试
		{"decimal integer", "123", tokens.INT, "123", false},
		{"decimal integer with underscore", "1_2_3", tokens.INT, "123", false},
		{"decimal zero", "0", tokens.INT, "0", false},

		// 八进制测试
		{"octal with prefix 0", "0123", tokens.INT, "0123", false},
		{"octal with prefix o", "0o123", tokens.INT, "0o123", false},
		{"invalid octal float", "0o1.2", tokens.INT, "0o1", true},

		// 十六进制测试
		{"hex with prefix x", "0x1A", tokens.INT, "0x1A", false},
		{"hex with prefix X", "0X1a", tokens.INT, "0X1a", false},
		{"hex float", "0x1.2p3", tokens.FLOAT, "0x1.2p3", false},
		{"hex float with capital P", "0x1.2P3", tokens.FLOAT, "0x1.2P3", false},
		{"hex float with negative exponent", "0x1.2p-3", tokens.FLOAT, "0x1.2p-3", false},

		// 二进制测试
		{"binary with prefix b", "0b1010", tokens.INT, "0b1010", false},
		{"binary with prefix B", "0B1010", tokens.INT, "0B1010", false},
		{"invalid binary float", "0b1.01", tokens.INT, "0b1", true},

		// 十进制浮点数测试
		{"decimal float", "123.456", tokens.FLOAT, "123.456", false},
		{"decimal float with exponent", "123.456e10", tokens.FLOAT, "123.456e10", false},
		{"decimal float with capital E", "123.456E10", tokens.FLOAT, "123.456E10", false},
		{"decimal float with negative exponent", "123.456e-10", tokens.FLOAT, "123.456e-10", false},
		{"decimal float with positive exponent", "123.456e+10", tokens.FLOAT, "123.456e+10", false},

		// 错误情况测试
		{"invalid hex exponent", "0x1.2e3", tokens.FLOAT, "0x1", true},
		{"invalid decimal exponent", "0o1.2e3", tokens.FLOAT, "0o1", true},
		{"hex float without exponent", "0x1.2", tokens.FLOAT, "0x1", true},
		{"no digits", "0x", tokens.INT, "0x", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 reader
			r := reader.BytesReader([]byte(tt.input))
			first, _ := r.ReadByte()

			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("Decimal() unexpected panic: %v", r)
					}
				}
			}()

			gotTok, gotText := Decimal(r, first)

			if !tt.wantErr {
				if gotTok != tt.wantTok {
					t.Errorf("Decimal() got token = %v, want %v", gotTok, tt.wantTok)
				}
				if gotText != tt.wantText {
					t.Errorf("Decimal() got text = %v, want %v", gotText, tt.wantText)
				}
			}
		})
	}
}
