// Package utils 提供配置文件读取功能
package utils

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// DifyConfig Dify相关配置
type DifyConfig struct {
	URL    string `toml:"url"`     // Dify API Base URL
	APIKey string `toml:"api_key"` // Dify API Key
}

// Config 应用配置结构
type Config struct {
	Dify DifyConfig `toml:"dify"` // Dify配置
}

// LoadConfig 从指定文件加载配置
func LoadConfig(configFile string) (*Config, error) {
	// 检查文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configFile)
	}

	// 读取配置文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析TOML配置
	var config Config
	err = toml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// LoadDefaultConfig 加载默认配置文件(config.toml)
func LoadDefaultConfig() (*Config, error) {
	return LoadConfig("config.toml")
}
