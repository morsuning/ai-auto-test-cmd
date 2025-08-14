// Package utils 提供配置文件读取功能
package utils

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// LLMConfig LLM相关配置
type LLMConfig struct {
	URL        string `toml:"url"`         // LLM API Base URL
	APIKey     string `toml:"api_key"`     // LLM API Key
	UserPrompt string `toml:"user_prompt"` // 自定义提示词
}

// RequestConfig 请求相关配置
type RequestConfig struct {
	URL        string   `toml:"url"`          // 目标URL
	Method     string   `toml:"method"`       // 请求方法
	File       string   `toml:"file"`         // CSV测试用例文件
	SavePath   string   `toml:"save_path"`    // 结果保存路径
	Timeout    int      `toml:"timeout"`      // 请求超时时间
	Concurrent int      `toml:"concurrent"`   // 并发请求数
	AuthBearer string   `toml:"auth_bearer"`  // Bearer Token认证
	AuthBasic  string   `toml:"auth_basic"`   // Basic Auth认证
	AuthAPIKey string   `toml:"auth_api_key"` // API Key认证
	Headers    []string `toml:"headers"`      // 自定义HTTP头
}

// TestCaseConfig 用例设置
type TestCaseConfig struct {
	Num             int    `toml:"num"`              // 用例生成数量
	Output          string `toml:"output"`           // 输出文件路径
	PositiveExample string `toml:"positive_example"` // 正例报文（支持多行字符串）
	Type            string `toml:"type"`             // 正例报文类型（xml或json）
}

// Config 应用配置结构
type Config struct {
	LLM         LLMConfig                  `toml:"llm"`          // LLM配置
	Request     RequestConfig              `toml:"request"`      // 请求配置
	TestCase    TestCaseConfig             `toml:"testcase"`     // 用例设置
	Constraints map[string]FieldConstraint `toml:"constraints"`  // 约束配置
	BuiltinData BuiltinData                `toml:"builtin_data"` // 内置数据
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

// LoadConfigWithConstraints 从指定文件加载配置并设置约束系统
func LoadConfigWithConstraints(configFile string) (*Config, error) {
	config, err := LoadConfig(configFile)
	if err != nil {
		return nil, err
	}

	// 如果配置文件中包含约束配置，设置全局约束配置
	if len(config.Constraints) > 0 || len(config.BuiltinData.FirstNames) > 0 {
		constraintConfig := &ConstraintConfig{
			Constraints: config.Constraints,
			BuiltinData: config.BuiltinData,
		}

		// 验证约束配置
		if err := ValidateConstraintConfig(constraintConfig); err != nil {
			return nil, fmt.Errorf("约束配置验证失败: %w", err)
		}

		// 设置全局约束配置
		globalConstraintConfig = constraintConfig
	}

	return config, nil
}

// LoadDefaultConfig 加载默认配置文件(config.toml)
func LoadDefaultConfig() (*Config, error) {
	return LoadConfig("config.toml")
}
