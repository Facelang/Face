package reader

import (
	"os"
	"testing"
	"unicode/utf8"
)

func TestReadRune(t *testing.T) {
	// 创建测试文件，包含各种 UTF-8 字符
	testContent := "Hello 世界\n你好 🌍\nTest 测试"
	testFile := "test_utf8.txt"

	// 写入测试文件
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// 创建 Reader
	reader := FileReader(testFile)

	// 读取所有字符并验证
	var result []rune
	for {
		ch, chw := reader.ReadRune()
		if chw == 0 {
			break
		}
		t.Logf("Read rune: %c (%U)", ch, ch)
		result = append(result, ch)
	}

	// 验证读取的字符是否正确
	expected := []rune(testContent)
	t.Logf("Got content: %q", result)
	t.Logf("Expected length: %d, Got length: %d", len(expected), len(result))

	if len(result) != len(expected) {
		t.Fatalf("Length mismatch: got %d, expected %d", len(result), len(expected))
	}

	for i, r := range result {
		if r != expected[i] {
			t.Errorf("Character mismatch at position %d: got %c (%U), expected %c (%U)",
				i, r, r, expected[i], expected[i])
		}
	}
}

func TestReadRuneASCII(t *testing.T) {
	// 测试纯 ASCII 字符
	testContent := "Hello World\nTest 123"
	testFile := "test_ascii.txt"

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	reader := FileReader(testFile)

	var result []rune
	for {
		r, chw := reader.ReadRune()
		if chw == 0 {
			break
		}
		result = append(result, r)
	}

	expected := []rune(testContent)
	if len(result) != len(expected) {
		t.Fatalf("Length mismatch: got %d, expected %d", len(result), len(expected))
	}

	for i, r := range result {
		if r != expected[i] {
			t.Errorf("Character mismatch at position %d: got %c, expected %c", i, r, expected[i])
		}
	}
}

func TestReadRuneEmoji(t *testing.T) {
	// 测试包含 emoji 的文本
	testContent := "Hello 🌍 World 🚀"
	testFile := "test_emoji.txt"

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	reader := FileReader(testFile)

	var result []rune
	for {
		r, chw := reader.ReadRune()
		if chw == 0 {
			break
		}
		result = append(result, r)
	}

	expected := []rune(testContent)
	if len(result) != len(expected) {
		t.Fatalf("Length mismatch: got %d, expected %d", len(result), len(expected))
	}

	for i, r := range result {
		if r != expected[i] {
			t.Errorf("Character mismatch at position %d: got %c (%U), expected %c (%U)",
				i, r, r, expected[i], expected[i])
		}
	}
}

func TestReadRuneInvalidUTF8(t *testing.T) {
	// 测试无效的 UTF-8 序列
	invalidUTF8 := []byte{0xFF, 0xFE, 0xFD} // 无效的 UTF-8 字节序列
	testFile := "test_invalid.txt"

	err := os.WriteFile(testFile, invalidUTF8, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	reader := FileReader(testFile)

	// 读取第一个字符，应该返回 RuneError
	r, chw := reader.ReadRune()
	if r != utf8.RuneError {
		t.Errorf("Expected RuneError for invalid UTF-8, got %c (%U)", r, r)
	}
	if chw != 1 {
		t.Error("Expected width 1 for invalid UTF-8")
	}
}
