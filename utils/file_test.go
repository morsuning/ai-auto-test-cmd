package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReadPromptFile 测试读取提示词文件功能
func TestReadPromptFile(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 测试用例1: 正常的UTF-8文件
	t.Run("ValidUTF8File", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "valid_prompt.txt")
		content := "这是一个测试提示词\n包含中文和英文 English text\n用于测试UTF-8编码"
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}

		result, err := ReadPromptFile(filePath)
		if err != nil {
			t.Errorf("ReadPromptFile失败: %v", err)
		}

		expected := "这是一个测试提示词\n包含中文和英文 English text\n用于测试UTF-8编码"
		if result != expected {
			t.Errorf("期望内容: %s, 实际内容: %s", expected, result)
		}
	})

	// 测试用例2: 文件不存在
	t.Run("FileNotExist", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "nonexistent.txt")
		_, err := ReadPromptFile(filePath)
		if err == nil {
			t.Error("期望返回错误，但没有错误")
		}
		if !contains(err.Error(), "提示词文件不存在") {
			t.Errorf("错误信息不正确: %v", err)
		}
	})

	// 测试用例3: 空文件
	t.Run("EmptyFile", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "empty.txt")
		err := os.WriteFile(filePath, []byte(""), 0644)
		if err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}

		_, err = ReadPromptFile(filePath)
		if err == nil {
			t.Error("期望返回错误，但没有错误")
		}
		if !contains(err.Error(), "内容为空") {
			t.Errorf("错误信息不正确: %v", err)
		}
	})

	// 测试用例4: 只包含空白字符的文件
	t.Run("WhitespaceOnlyFile", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "whitespace.txt")
		err := os.WriteFile(filePath, []byte("   \n\t  \r\n  "), 0644)
		if err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}

		_, err = ReadPromptFile(filePath)
		if err == nil {
			t.Error("期望返回错误，但没有错误")
		}
		if !contains(err.Error(), "内容为空") {
			t.Errorf("错误信息不正确: %v", err)
		}
	})

	// 测试用例5: 包含前后空白字符的有效文件
	t.Run("FileWithWhitespace", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "whitespace_content.txt")
		content := "  \n\t  这是有效内容  \n\r  "
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}

		result, err := ReadPromptFile(filePath)
		if err != nil {
			t.Errorf("ReadPromptFile失败: %v", err)
		}

		expected := "这是有效内容"
		if result != expected {
			t.Errorf("期望内容: %s, 实际内容: %s", expected, result)
		}
	})

	// 测试用例6: 非UTF-8编码文件（模拟）
	t.Run("InvalidUTF8File", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "invalid_utf8.txt")
		// 创建包含无效UTF-8字节序列的文件
		invalidUTF8 := []byte{0xFF, 0xFE, 0xFD} // 无效的UTF-8字节序列
		err := os.WriteFile(filePath, invalidUTF8, 0644)
		if err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}

		_, err = ReadPromptFile(filePath)
		if err == nil {
			t.Error("期望返回错误，但没有错误")
		}
		if !contains(err.Error(), "不是有效的UTF-8编码") {
			t.Errorf("错误信息不正确: %v", err)
		}
	})
}

// contains 检查字符串是否包含子字符串（辅助函数）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}()))
}