package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadConfig 测试配置文件加载功能
func TestLoadConfig(t *testing.T) {
	// 创建临时配置文件
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.toml")

	// 写入测试配置内容
	configContent := `[dify]
url = "http://test.example.com/v1"
api_key = "app-test123456789"
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("创建测试配置文件失败: %v", err)
	}

	// 测试加载配置
	config, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("加载配置文件失败: %v", err)
	}

	// 验证配置内容
	if config.Dify.URL != "http://test.example.com/v1" {
		t.Errorf("URL配置错误，期望: %s, 实际: %s", "http://test.example.com/v1", config.Dify.URL)
	}

	if config.Dify.APIKey != "app-test123456789" {
		t.Errorf("API Key配置错误，期望: %s, 实际: %s", "app-test123456789", config.Dify.APIKey)
	}
}

// TestLoadConfigFileNotExist 测试配置文件不存在的情况
func TestLoadConfigFileNotExist(t *testing.T) {
	_, err := LoadConfig("nonexistent_config.toml")
	if err == nil {
		t.Error("期望配置文件不存在时返回错误，但没有返回错误")
	}
}

// TestLoadConfigInvalidFormat 测试无效配置文件格式
func TestLoadConfigInvalidFormat(t *testing.T) {
	// 创建临时配置文件
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid_config.toml")

	// 写入无效的TOML内容
	invalidContent := `[dify
url = "http://test.example.com/v1"
api_key = "app-test123456789"
`
	err := os.WriteFile(configFile, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("创建测试配置文件失败: %v", err)
	}

	// 测试加载无效配置
	_, err = LoadConfig(configFile)
	if err == nil {
		t.Error("期望无效配置文件格式时返回错误，但没有返回错误")
	}
}

// TestLoadDefaultConfig 测试加载默认配置文件
func TestLoadDefaultConfig(t *testing.T) {
	// 创建临时的config.toml文件
	originalWd, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// 写入默认配置内容
	configContent := `[dify]
url = "http://localhost/v1"
api_key = "app-default123"
`
	err := os.WriteFile("config.toml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("创建默认配置文件失败: %v", err)
	}

	// 测试加载默认配置
	config, err := LoadDefaultConfig()
	if err != nil {
		t.Fatalf("加载默认配置文件失败: %v", err)
	}

	// 验证配置内容
	if config.Dify.URL != "http://localhost/v1" {
		t.Errorf("默认URL配置错误，期望: %s, 实际: %s", "http://localhost/v1", config.Dify.URL)
	}

	if config.Dify.APIKey != "app-default123" {
		t.Errorf("默认API Key配置错误，期望: %s, 实际: %s", "app-default123", config.Dify.APIKey)
	}
}