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
	URL             string   `toml:"url"`               // 目标URL
	Method          string   `toml:"method"`            // 请求方法
	File            string   `toml:"file"`              // CSV测试用例文件
	SavePath        string   `toml:"save_path"`         // 结果保存路径
	Timeout         int      `toml:"timeout"`           // 请求超时时间
	Concurrent      int      `toml:"concurrent"`        // 并发请求数
	AuthBearer      string   `toml:"auth_bearer"`       // Bearer Token认证
	AuthBasic       string   `toml:"auth_basic"`        // Basic Auth认证
	AuthAPIKey      string   `toml:"auth_api_key"`      // API Key认证
	Headers         []string `toml:"headers"`           // 自定义HTTP头
	Query           []string `toml:"query"`             // GET请求的URL查询参数
	IgnoreTLSErrors bool     `toml:"ignore_tls_errors"` // 忽略TLS证书验证错误
}

// TestCaseConfig 用例设置
type TestCaseConfig struct {
	Num             int     `toml:"num"`              // 用例生成数量
	Output          string  `toml:"output"`           // 输出文件路径
	PositiveExample string  `toml:"positive_example"` // 正例报文（支持多行字符串）
	Type            string  `toml:"type"`             // 正例报文类型（xml或json）
	VariationRate   float64 `toml:"variation_rate"`   // 随机化因子，控制数据变化程度（0.0-1.0，默认0.5）
}

// ConstraintsConfig 约束系统配置
type ConstraintsConfig struct {
	Enable         *bool                      `toml:"enable"`       // 约束系统开关
	BuiltinData    BuiltinData                `toml:"builtin_data"` // 内置数据
	Constraints    map[string]FieldConstraint // 约束配置（手动解析）
	CustomTypes    map[string]CustomTypeSpec  `toml:"types"`    // 自定义类型
	CustomDatasets map[string][]string        `toml:"datasets"` // 自定义数据集
}

// Config 应用配置结构
type Config struct {
	LLM         LLMConfig         `toml:"llm"`          // LLM配置
	Request     RequestConfig     `toml:"request"`      // 请求配置
	TestCase    TestCaseConfig    `toml:"testcase"`     // 用例设置
	Constraints ConstraintsConfig `toml:"constraints"`  // 约束系统配置
	BuiltinData BuiltinData       `toml:"builtin_data"` // 内置数据（向后兼容）
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

	// 先解析为通用map以处理constraints节点
	var rawConfig map[string]any
	err = toml.Unmarshal(data, &rawConfig)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 解析TOML配置
	var config Config
	err = toml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 手动解析constraints节点
	if constraintsNode, exists := rawConfig["constraints"]; exists {
		if constraintsMap, ok := constraintsNode.(map[string]any); ok {
			// 初始化约束映射
			config.Constraints.Constraints = make(map[string]FieldConstraint)

			for key, value := range constraintsMap {
				// 跳过开关与在结构体中已由toml自动解析的节点
				if key == "enable" || key == "builtin_data" || key == "types" || key == "datasets" {
					continue
				}

				// 递归解析字段约束（支持嵌套路径，如 user.profile.status_text）
				collectConstraintsRecursive(key, value, &ConstraintConfig{
					Constraints:    config.Constraints.Constraints,
					BuiltinData:    config.Constraints.BuiltinData,
					CustomTypes:    config.Constraints.CustomTypes,
					CustomDatasets: config.Constraints.CustomDatasets,
				})
			}
		}
	}

	return &config, nil
}

// LoadConfigWithConstraints 从指定文件加载配置并设置约束系统
func LoadConfigWithConstraints(configFile string) (*Config, error) {
	config, err := LoadConfig(configFile)
	if err != nil {
		return nil, err
	}

	// 检查约束系统是否启用
	constraintsEnabled := IsConstraintsEnabled(config)

	// 如果约束系统启用且配置文件中包含约束配置，设置全局约束配置
	if constraintsEnabled && (len(config.Constraints.Constraints) > 0 || len(config.Constraints.BuiltinData.FirstNames) > 0 || len(config.BuiltinData.FirstNames) > 0) {
		// 合并约束配置（优先使用constraints节点下的配置，向后兼容builtin_data）
		constraints := config.Constraints.Constraints
		builtinData := config.Constraints.BuiltinData

		// 向后兼容：如果constraints节点下没有builtin_data，使用根节点下的
		if len(builtinData.FirstNames) == 0 && len(config.BuiltinData.FirstNames) > 0 {
			builtinData = config.BuiltinData
		}

		constraintConfig := &ConstraintConfig{
			Constraints:    constraints,
			BuiltinData:    builtinData,
			CustomTypes:    config.Constraints.CustomTypes,
			CustomDatasets: config.Constraints.CustomDatasets,
		}

		// 验证约束配置
		if err := ValidateConstraintConfig(constraintConfig); err != nil {
			return nil, fmt.Errorf("约束配置验证失败: %w", err)
		}

		// 设置全局约束配置
		globalConstraintConfig = constraintConfig
	} else {
		// 约束系统未启用，清空全局约束配置
		globalConstraintConfig = nil
	}

	return config, nil
}

// IsConstraintsEnabled 检查约束系统是否启用
func IsConstraintsEnabled(config *Config) bool {
	// 如果明确设置了constraints.enable，使用该设置
	if config.Constraints.Enable != nil {
		return *config.Constraints.Enable
	}

	// 默认值：如果有约束配置则启用，否则禁用
	return len(config.Constraints.Constraints) > 0
}

// LoadDefaultConfig 加载默认配置文件(config.toml)
func LoadDefaultConfig() (*Config, error) {
	return LoadConfig("config.toml")
}
