// Package utils 提供约束系统功能，支持基于TOML配置的字段约束
package utils

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// FieldConstraint 字段约束配置
type FieldConstraint struct {
	Type         string   `toml:"type"`          // 约束类型
	Format       string   `toml:"format"`        // 格式（用于日期等）
	MinDate      string   `toml:"min_date"`      // 最小日期
	MaxDate      string   `toml:"max_date"`      // 最大日期
	MinDatetime  string   `toml:"min_datetime"`  // 最小日期时间（RFC 3339 Extended格式）
	MaxDatetime  string   `toml:"max_datetime"`  // 最大日期时间（RFC 3339 Extended格式）
	Timezone     string   `toml:"timezone"`      // 时区（如：+08:00, UTC, Asia/Shanghai）
	Min          *float64 `toml:"min"`           // 最小值
	Max          *float64 `toml:"max"`           // 最大值
	Precision    *int     `toml:"precision"`     // 精度（小数位数）
	KeepOriginal *bool    `toml:"keep_original"` // 是否保持原值不变
	Description  string   `toml:"description"`   // 描述
}

// BuiltinData 内置数据集
type BuiltinData struct {
	FirstNames   []string `toml:"first_names"`   // 姓氏
	LastNames    []string `toml:"last_names"`    // 名字
	Addresses    []string `toml:"addresses"`     // 地址
	EmailDomains []string `toml:"email_domains"` // 邮箱域名
}

// ConstraintConfig 约束配置
type ConstraintConfig struct {
	Constraints map[string]FieldConstraint // 字段约束映射
	BuiltinData BuiltinData                `toml:"builtin_data"` // 内置数据
}

// 全局约束配置
var globalConstraintConfig *ConstraintConfig

// ValidationError 验证错误信息
type ValidationError struct {
	Field   string // 字段名
	Message string // 错误信息
}

// Error 实现error接口
func (e ValidationError) Error() string {
	return fmt.Sprintf("字段 '%s': %s", e.Field, e.Message)
}

// ValidationErrors 多个验证错误
type ValidationErrors []ValidationError

// Error 实现error接口
func (errs ValidationErrors) Error() string {
	if len(errs) == 0 {
		return "无验证错误"
	}
	if len(errs) == 1 {
		return errs[0].Error()
	}

	var messages []string
	for _, err := range errs {
		messages = append(messages, err.Error())
	}
	return fmt.Sprintf("发现 %d 个配置错误:\n%s", len(errs), strings.Join(messages, "\n"))
}

// ValidateConstraintConfig 验证约束配置的格式和内容
func ValidateConstraintConfig(config *ConstraintConfig) error {
	var errors ValidationErrors

	// 验证每个字段约束
	for fieldName, constraint := range config.Constraints {
		if fieldErrors := validateFieldConstraint(fieldName, constraint); len(fieldErrors) > 0 {
			errors = append(errors, fieldErrors...)
		}
	}

	// 验证内置数据
	if builtinErrors := validateBuiltinData(config.BuiltinData); len(builtinErrors) > 0 {
		errors = append(errors, builtinErrors...)
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateFieldConstraint 验证单个字段约束
func validateFieldConstraint(fieldName string, constraint FieldConstraint) []ValidationError {
	var errors []ValidationError

	// 验证约束类型
	validTypes := []string{"date", "datetime", "chinese_name", "phone", "email", "chinese_address", "id_card", "integer", "float", "keep_original"}
	if constraint.Type == "" {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: "约束类型 'type' 不能为空",
		})
	} else {
		validType := false
		for _, vt := range validTypes {
			if constraint.Type == vt {
				validType = true
				break
			}
		}
		if !validType {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("无效的约束类型 '%s'，支持的类型: %s", constraint.Type, strings.Join(validTypes, ", ")),
			})
		}
	}

	// 根据类型进行特定验证
	switch constraint.Type {
	case "date":
		errors = append(errors, validateDateConstraint(fieldName, constraint)...)
	case "datetime":
		errors = append(errors, validateDatetimeConstraint(fieldName, constraint)...)
	case "integer":
		errors = append(errors, validateIntegerConstraint(fieldName, constraint)...)
	case "float":
		errors = append(errors, validateFloatConstraint(fieldName, constraint)...)
	}

	return errors
}

// validateDateConstraint 验证日期约束
func validateDateConstraint(fieldName string, constraint FieldConstraint) []ValidationError {
	var errors []ValidationError

	// 验证日期格式
	if constraint.Format != "" {
		// 尝试解析格式字符串
		testDate := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
		formatted := testDate.Format(constraint.Format)
		if _, err := time.Parse(constraint.Format, formatted); err != nil {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("无效的日期格式 '%s': %v", constraint.Format, err),
			})
		}
	}

	// 验证最小日期
	var minDate, maxDate time.Time
	if constraint.MinDate != "" {
		var err error
		minDate, err = time.Parse("20060102", constraint.MinDate)
		if err != nil {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("无效的最小日期格式 '%s'，应为 YYYYMMDD 格式", constraint.MinDate),
			})
		}
	}

	// 验证最大日期
	if constraint.MaxDate != "" {
		var err error
		maxDate, err = time.Parse("20060102", constraint.MaxDate)
		if err != nil {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("无效的最大日期格式 '%s'，应为 YYYYMMDD 格式", constraint.MaxDate),
			})
		} else if !minDate.IsZero() && maxDate.Before(minDate) {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("最大日期 '%s' 不能早于最小日期 '%s'", constraint.MaxDate, constraint.MinDate),
			})
		}
	}

	return errors
}

// validateDatetimeConstraint 验证日期时间约束
func validateDatetimeConstraint(fieldName string, constraint FieldConstraint) []ValidationError {
	var errors []ValidationError

	// RFC 3339 Extended格式：2006-01-02T15:04:05.000Z07:00
	rfc3339ExtendedFormat := "2006-01-02T15:04:05.000Z07:00"

	// 验证最小日期时间
	var minDatetime, maxDatetime time.Time
	if constraint.MinDatetime != "" {
		var err error
		minDatetime, err = time.Parse(rfc3339ExtendedFormat, constraint.MinDatetime)
		if err != nil {
			// 尝试其他RFC 3339格式
			minDatetime, err = time.Parse(time.RFC3339, constraint.MinDatetime)
			if err != nil {
				errors = append(errors, ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("无效的最小日期时间格式 '%s'，应为 RFC 3339 Extended 格式（如：2025-07-08T12:43:21.000+00:00）", constraint.MinDatetime),
				})
			}
		}
	}

	// 验证最大日期时间
	if constraint.MaxDatetime != "" {
		var err error
		maxDatetime, err = time.Parse(rfc3339ExtendedFormat, constraint.MaxDatetime)
		if err != nil {
			// 尝试其他RFC 3339格式
			maxDatetime, err = time.Parse(time.RFC3339, constraint.MaxDatetime)
			if err != nil {
				errors = append(errors, ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("无效的最大日期时间格式 '%s'，应为 RFC 3339 Extended 格式（如：2025-07-08T12:43:21.000+00:00）", constraint.MaxDatetime),
				})
			} else if !minDatetime.IsZero() && maxDatetime.Before(minDatetime) {
				errors = append(errors, ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("最大日期时间 '%s' 不能早于最小日期时间 '%s'", constraint.MaxDatetime, constraint.MinDatetime),
				})
			}
		} else if !minDatetime.IsZero() && maxDatetime.Before(minDatetime) {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("最大日期时间 '%s' 不能早于最小日期时间 '%s'", constraint.MaxDatetime, constraint.MinDatetime),
			})
		}
	}

	// 验证时区格式
	if constraint.Timezone != "" {
		if err := ValidateTimezone(constraint.Timezone); err != nil {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("无效的时区格式 '%s': %v。支持格式：UTC、±HH:MM（如：+08:00）或 IANA 时区名称（如：Asia/Shanghai）", constraint.Timezone, err),
			})
		}
	}

	return errors
}

// validateIntegerConstraint 验证整数约束
func validateIntegerConstraint(fieldName string, constraint FieldConstraint) []ValidationError {
	var errors []ValidationError

	// 验证最小值和最大值
	if constraint.Min != nil && constraint.Max != nil {
		if *constraint.Max < *constraint.Min {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("最大值 %.0f 不能小于最小值 %.0f", *constraint.Max, *constraint.Min),
			})
		}
	}

	// 验证精度字段不应用于整数类型
	if constraint.Precision != nil {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: "整数类型不应设置 'precision' 字段",
		})
	}

	return errors
}

// validateFloatConstraint 验证浮点数约束
func validateFloatConstraint(fieldName string, constraint FieldConstraint) []ValidationError {
	var errors []ValidationError

	// 验证最小值和最大值
	if constraint.Min != nil && constraint.Max != nil {
		if *constraint.Max < *constraint.Min {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("最大值 %.2f 不能小于最小值 %.2f", *constraint.Max, *constraint.Min),
			})
		}
	}

	// 验证精度
	if constraint.Precision != nil {
		if *constraint.Precision < 0 || *constraint.Precision > 10 {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("精度值 %d 超出有效范围 [0, 10]", *constraint.Precision),
			})
		}
	}

	return errors
}

// validateBuiltinData 验证内置数据
func validateBuiltinData(data BuiltinData) []ValidationError {
	var errors []ValidationError

	// 验证姓氏数据
	if len(data.FirstNames) == 0 {
		errors = append(errors, ValidationError{
			Field:   "builtin_data.first_names",
			Message: "姓氏列表不能为空",
		})
	} else {
		for i, name := range data.FirstNames {
			if strings.TrimSpace(name) == "" {
				errors = append(errors, ValidationError{
					Field:   "builtin_data.first_names",
					Message: fmt.Sprintf("第 %d 个姓氏不能为空", i+1),
				})
			}
		}
	}

	// 验证名字数据
	if len(data.LastNames) == 0 {
		errors = append(errors, ValidationError{
			Field:   "builtin_data.last_names",
			Message: "名字列表不能为空",
		})
	} else {
		for i, name := range data.LastNames {
			if strings.TrimSpace(name) == "" {
				errors = append(errors, ValidationError{
					Field:   "builtin_data.last_names",
					Message: fmt.Sprintf("第 %d 个名字不能为空", i+1),
				})
			}
		}
	}

	// 验证地址数据
	if len(data.Addresses) == 0 {
		errors = append(errors, ValidationError{
			Field:   "builtin_data.addresses",
			Message: "地址列表不能为空",
		})
	} else {
		for i, addr := range data.Addresses {
			if strings.TrimSpace(addr) == "" {
				errors = append(errors, ValidationError{
					Field:   "builtin_data.addresses",
					Message: fmt.Sprintf("第 %d 个地址不能为空", i+1),
				})
			}
		}
	}

	// 验证邮箱域名数据
	if len(data.EmailDomains) == 0 {
		errors = append(errors, ValidationError{
			Field:   "builtin_data.email_domains",
			Message: "邮箱域名列表不能为空",
		})
	} else {
		for i, domain := range data.EmailDomains {
			if strings.TrimSpace(domain) == "" {
				errors = append(errors, ValidationError{
					Field:   "builtin_data.email_domains",
					Message: fmt.Sprintf("第 %d 个邮箱域名不能为空", i+1),
				})
			} else if !strings.Contains(domain, ".") {
				errors = append(errors, ValidationError{
					Field:   "builtin_data.email_domains",
					Message: fmt.Sprintf("第 %d 个邮箱域名 '%s' 格式无效", i+1, domain),
				})
			}
		}
	}

	return errors
}

// LoadConstraintConfig 从TOML文件加载约束配置
func LoadConstraintConfig(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取约束配置文件失败: %w", err)
	}

	// 解析为通用map
	var rawConfig map[string]any
	err = toml.Unmarshal(data, &rawConfig)
	if err != nil {
		return fmt.Errorf("解析TOML配置文件失败: %w", err)
	}

	config := &ConstraintConfig{
		Constraints: make(map[string]FieldConstraint),
	}

	// 手动解析约束字段
	for key, value := range rawConfig {
		if key == "builtin_data" {
			// 解析内置数据
			builtinBytes, _ := toml.Marshal(map[string]any{"builtin_data": value})
			var temp struct {
				BuiltinData BuiltinData `toml:"builtin_data"`
			}
			toml.Unmarshal(builtinBytes, &temp)
			config.BuiltinData = temp.BuiltinData
		} else {
			// 解析字段约束
			constraintBytes, _ := toml.Marshal(map[string]any{key: value})
			var constraint FieldConstraint
			var temp map[string]FieldConstraint
			if toml.Unmarshal(constraintBytes, &temp) == nil {
				if c, exists := temp[key]; exists {
					constraint = c
					config.Constraints[key] = constraint
				}
			}
		}
	}

	// 验证配置
	if err := ValidateConstraintConfig(config); err != nil {
		return fmt.Errorf("约束配置验证失败: %w", err)
	}

	globalConstraintConfig = config
	return nil
}

// LoadDefaultConstraints 加载默认约束配置
func LoadDefaultConstraints() error {
	// 使用项目根目录下的constraints.toml文件
	return LoadConstraintConfig("constraints.toml")
}

// FindFieldConstraint 根据字段名查找约束配置
func FindFieldConstraint(fieldName string) *FieldConstraint {
	if globalConstraintConfig == nil {
		return nil
	}

	// 直接匹配字段名
	if constraint, exists := globalConstraintConfig.Constraints[fieldName]; exists {
		return &constraint
	}

	// 尝试小写匹配
	lowerFieldName := strings.ToLower(fieldName)
	if constraint, exists := globalConstraintConfig.Constraints[lowerFieldName]; exists {
		return &constraint
	}

	// 尝试匹配简单字段名（去掉路径前缀）
	if strings.Contains(fieldName, ".") {
		parts := strings.Split(fieldName, ".")
		simpleFieldName := parts[len(parts)-1] // 取最后一部分
		if constraint, exists := globalConstraintConfig.Constraints[simpleFieldName]; exists {
			return &constraint
		}

		// 尝试简单字段名的小写匹配
		lowerSimpleFieldName := strings.ToLower(simpleFieldName)
		if constraint, exists := globalConstraintConfig.Constraints[lowerSimpleFieldName]; exists {
			return &constraint
		}
	}

	return nil
}

// GenerateConstrainedValue 根据约束生成值
func GenerateConstrainedValue(constraint *FieldConstraint, originalValue any) any {
	if constraint == nil {
		return originalValue
	}

	// 检查是否设置了保持原值不变
	if constraint.KeepOriginal != nil && *constraint.KeepOriginal {
		return originalValue
	}

	switch constraint.Type {
	case "keep_original":
		return originalValue
	case "date":
		return generateDateValue(constraint)
	case "datetime":
		return generateDatetimeValue(constraint)
	case "chinese_name":
		return generateChineseName()
	case "phone":
		return generatePhoneNumber()
	case "email":
		return generateEmail()
	case "chinese_address":
		return generateChineseAddress()
	case "id_card":
		return generateIDCard()
	case "integer":
		return generateIntegerValue(constraint)
	case "float":
		return generateFloatValue(constraint)
	default:
		return originalValue
	}
}

// generateDateValue 生成日期值
func generateDateValue(constraint *FieldConstraint) string {
	format := constraint.Format
	if format == "" {
		format = "20060102" // 默认格式
	}

	// 解析日期范围
	minDate, _ := time.Parse("20060102", constraint.MinDate)
	maxDate, _ := time.Parse("20060102", constraint.MaxDate)

	if minDate.IsZero() {
		minDate = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if maxDate.IsZero() {
		maxDate = time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC)
	}

	// 如果最大日期小于或等于最小日期，返回最小日期
	if maxDate.Before(minDate) || maxDate.Equal(minDate) {
		return minDate.Format(format)
	}

	// 生成随机日期
	duration := maxDate.Sub(minDate)
	randomDuration := time.Duration(rand.Int63n(int64(duration)))
	randomDate := minDate.Add(randomDuration)

	return randomDate.Format(format)
}

// generateDatetimeValue 生成RFC 3339 Extended格式的日期时间值
func generateDatetimeValue(constraint *FieldConstraint) string {
	// 设置默认日期时间范围（包含时分秒）
	minDatetime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	maxDatetime := time.Date(2030, 12, 31, 23, 59, 59, 999000000, time.UTC)

	// RFC 3339 Extended格式
	rfc3339ExtendedFormat := "2006-01-02T15:04:05.000Z07:00"

	// 解析用户指定的日期时间范围
	if constraint.MinDatetime != "" {
		if parsed, err := time.Parse(rfc3339ExtendedFormat, constraint.MinDatetime); err == nil {
			minDatetime = parsed
		} else if parsed, err := time.Parse(time.RFC3339, constraint.MinDatetime); err == nil {
			minDatetime = parsed
		}
	}
	if constraint.MaxDatetime != "" {
		if parsed, err := time.Parse(rfc3339ExtendedFormat, constraint.MaxDatetime); err == nil {
			maxDatetime = parsed
		} else if parsed, err := time.Parse(time.RFC3339, constraint.MaxDatetime); err == nil {
			maxDatetime = parsed
		}
	}

	// 如果最大日期时间小于或等于最小日期时间，返回最小日期时间
	if maxDatetime.Before(minDatetime) || maxDatetime.Equal(minDatetime) {
		// 处理时区
		var targetTimezone *time.Location
		if constraint.Timezone != "" {
			if constraint.Timezone == "UTC" {
				targetTimezone = time.UTC
			} else if strings.HasPrefix(constraint.Timezone, "+") || strings.HasPrefix(constraint.Timezone, "-") {
				// 解析偏移量格式（如：+08:00, -05:00）
				if offset, err := parseTimezoneOffset(constraint.Timezone); err == nil {
					targetTimezone = time.FixedZone("Custom", offset)
				} else {
					targetTimezone = time.UTC // 默认使用UTC
				}
			} else {
				// IANA时区名称
				if loc, err := time.LoadLocation(constraint.Timezone); err == nil {
					targetTimezone = loc
				} else {
					targetTimezone = time.UTC // 默认使用UTC
				}
			}
		} else {
			// 默认使用UTC
			targetTimezone = time.UTC
		}
		// 转换到目标时区并返回
		return minDatetime.In(targetTimezone).Format("2006-01-02T15:04:05.000Z07:00")
	}

	// 生成随机日期时间（精确到毫秒）
	diff := maxDatetime.UnixNano() - minDatetime.UnixNano()
	randomNanos := rand.Int63n(diff + 1)
	randomDatetime := time.Unix(0, minDatetime.UnixNano()+randomNanos)

	// 处理时区
	var targetTimezone *time.Location
	if constraint.Timezone != "" {
		if constraint.Timezone == "UTC" {
			targetTimezone = time.UTC
		} else if strings.HasPrefix(constraint.Timezone, "+") || strings.HasPrefix(constraint.Timezone, "-") {
			// 解析偏移量格式（如：+08:00, -05:00）
			if offset, err := parseTimezoneOffset(constraint.Timezone); err == nil {
				targetTimezone = time.FixedZone("Custom", offset)
			} else {
				targetTimezone = time.UTC // 默认使用UTC
			}
		} else {
			// IANA时区名称
			if loc, err := time.LoadLocation(constraint.Timezone); err == nil {
				targetTimezone = loc
			} else {
				targetTimezone = time.UTC // 默认使用UTC
			}
		}
	} else {
		// 默认使用UTC
		targetTimezone = time.UTC
	}

	// 转换到目标时区
	randomDatetime = randomDatetime.In(targetTimezone)

	// 返回RFC 3339 Extended格式
	return randomDatetime.Format("2006-01-02T15:04:05.000Z07:00")
}

// ValidateTimezone 验证时区格式
func ValidateTimezone(timezone string) error {
	if timezone == "" {
		return nil
	}

	// UTC 时区
	if timezone == "UTC" {
		return nil
	}

	// 偏移量格式（如：+08:00, -05:00）
	if strings.HasPrefix(timezone, "+") || strings.HasPrefix(timezone, "-") {
		return validateTimezoneOffset(timezone)
	}

	// IANA 时区名称（如：Asia/Shanghai）
	if _, err := time.LoadLocation(timezone); err != nil {
		return fmt.Errorf("无法加载 IANA 时区 '%s': %v", timezone, err)
	}

	return nil
}

// validateTimezoneOffset 验证时区偏移量格式
func validateTimezoneOffset(offset string) error {
	if len(offset) != 6 || offset[3] != ':' {
		return fmt.Errorf("偏移量格式错误，应为 ±HH:MM 格式")
	}

	sign := offset[0]
	if sign != '+' && sign != '-' {
		return fmt.Errorf("偏移量符号错误，应为 + 或 -")
	}

	hours, err := strconv.Atoi(offset[1:3])
	if err != nil {
		return fmt.Errorf("小时部分格式错误: %v", err)
	}

	minutes, err := strconv.Atoi(offset[4:6])
	if err != nil {
		return fmt.Errorf("分钟部分格式错误: %v", err)
	}

	if hours < 0 || hours > 23 {
		return fmt.Errorf("小时值超出范围 (0-23): %d", hours)
	}

	if minutes < 0 || minutes > 59 {
		return fmt.Errorf("分钟值超出范围 (0-59): %d", minutes)
	}

	// 检查偏移量是否在合理范围内 (-12:00 到 +14:00)
	totalMinutes := hours*60 + minutes
	if sign == '-' {
		totalMinutes = -totalMinutes
	}

	if totalMinutes < -12*60 || totalMinutes > 14*60 {
		return fmt.Errorf("时区偏移量超出合理范围 (-12:00 到 +14:00): %s", offset)
	}

	return nil
}

// parseTimezoneOffset 解析时区偏移量（如：+08:00, -05:00）
func parseTimezoneOffset(offset string) (int, error) {
	if len(offset) != 6 || offset[3] != ':' {
		return 0, fmt.Errorf("invalid timezone offset format: %s", offset)
	}

	sign := 1
	if offset[0] == '-' {
		sign = -1
	} else if offset[0] != '+' {
		return 0, fmt.Errorf("invalid timezone offset sign: %s", offset)
	}

	hours, err := strconv.Atoi(offset[1:3])
	if err != nil {
		return 0, fmt.Errorf("invalid hours in timezone offset: %s", offset)
	}

	minutes, err := strconv.Atoi(offset[4:6])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes in timezone offset: %s", offset)
	}

	if hours < 0 || hours > 23 || minutes < 0 || minutes > 59 {
		return 0, fmt.Errorf("invalid timezone offset values: %s", offset)
	}

	return sign * (hours*3600 + minutes*60), nil
}

// generateChineseName 生成中文姓名
func generateChineseName() string {
	if globalConstraintConfig == nil || len(globalConstraintConfig.BuiltinData.FirstNames) == 0 {
		// 默认姓名
		defaultFirstNames := []string{"张", "王", "李", "赵", "刘"}
		defaultLastNames := []string{"伟", "芳", "娜", "敏", "静"}
		firstName := defaultFirstNames[rand.Intn(len(defaultFirstNames))]
		lastName := defaultLastNames[rand.Intn(len(defaultLastNames))]
		return firstName + lastName
	}

	firstName := globalConstraintConfig.BuiltinData.FirstNames[rand.Intn(len(globalConstraintConfig.BuiltinData.FirstNames))]
	lastName := globalConstraintConfig.BuiltinData.LastNames[rand.Intn(len(globalConstraintConfig.BuiltinData.LastNames))]
	return firstName + lastName
}

// generatePhoneNumber 生成手机号
func generatePhoneNumber() string {
	// 中国大陆手机号格式：1[3-9]xxxxxxxxx
	prefixes := []string{"13", "14", "15", "16", "17", "18", "19"}
	prefix := prefixes[rand.Intn(len(prefixes))]

	// 生成后9位数字
	suffix := fmt.Sprintf("%09d", rand.Intn(1000000000))
	return prefix + suffix
}

// generateEmail 生成邮箱地址
func generateEmail() string {
	domains := []string{"qq.com", "163.com", "126.com", "gmail.com", "sina.com"}
	if globalConstraintConfig != nil && len(globalConstraintConfig.BuiltinData.EmailDomains) > 0 {
		domains = globalConstraintConfig.BuiltinData.EmailDomains
	}

	// 生成用户名部分
	usernames := []string{"user", "demo", "test", "admin", "guest"}
	username := usernames[rand.Intn(len(usernames))]
	number := rand.Intn(1000)
	domain := domains[rand.Intn(len(domains))]

	return fmt.Sprintf("%s%d@%s", username, number, domain)
}

// generateChineseAddress 生成中文地址
func generateChineseAddress() string {
	defaultAddresses := []string{
		"北京市朝阳区建国门外大街1号",
		"上海市浦东新区陆家嘴环路1000号",
		"广州市天河区珠江新城花城大道85号",
		"深圳市南山区科技园南区深南大道9988号",
		"成都市高新区天府大道中段1388号",
	}

	addresses := defaultAddresses
	if globalConstraintConfig != nil && len(globalConstraintConfig.BuiltinData.Addresses) > 0 {
		addresses = globalConstraintConfig.BuiltinData.Addresses
	}

	return addresses[rand.Intn(len(addresses))]
}

// generateIDCard 生成身份证号
func generateIDCard() string {
	// 简化的身份证号生成（前6位地区码 + 8位生日 + 3位顺序码 + 1位校验码）
	areaCodes := []string{"110101", "310101", "440101", "500101", "510101"}
	areaCode := areaCodes[rand.Intn(len(areaCodes))]

	// 生成生日（1980-2005年）
	year := 1980 + rand.Intn(26)
	month := 1 + rand.Intn(12)
	day := 1 + rand.Intn(28)
	birthday := fmt.Sprintf("%04d%02d%02d", year, month, day)

	// 生成顺序码
	sequence := fmt.Sprintf("%03d", rand.Intn(1000))

	// 简单的校验码（随机生成）
	checkCodes := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "X"}
	checkCode := checkCodes[rand.Intn(len(checkCodes))]

	return areaCode + birthday + sequence + checkCode
}

// generateIntegerValue 生成整数值
func generateIntegerValue(constraint *FieldConstraint) int {
	min := 1
	max := 100

	if constraint.Min != nil {
		min = int(*constraint.Min)
	}
	if constraint.Max != nil {
		max = int(*constraint.Max)
	}

	// 如果最大值小于最小值，返回最小值
	if max < min {
		return min
	}

	// 如果最大值等于最小值，直接返回该值
	if max == min {
		return min
	}

	// 生成min到max之间的随机整数（包含边界）
	return min + rand.Intn(max-min+1)
}

// generateFloatValue 生成浮点数值
func generateFloatValue(constraint *FieldConstraint) float64 {
	min := 0.01
	max := 999.99
	precision := 2

	if constraint.Min != nil {
		min = *constraint.Min
	}
	if constraint.Max != nil {
		max = *constraint.Max
	}
	if constraint.Precision != nil {
		precision = *constraint.Precision
	}

	// 如果最大值小于最小值，返回最小值
	if max < min {
		return applyFloatPrecision(min, precision)
	}

	// 如果最大值等于最小值，直接返回该值
	if max == min {
		return applyFloatPrecision(min, precision)
	}

	// 生成随机浮点数
	value := min + rand.Float64()*(max-min)

	return applyFloatPrecision(value, precision)
}

// applyFloatPrecision 应用浮点数精度
func applyFloatPrecision(value float64, precision int) float64 {
	// 应用精度
	multiplier := 1.0
	for i := 0; i < precision; i++ {
		multiplier *= 10
	}

	// 四舍五入到指定精度
	return float64(int(value*multiplier+0.5)) / multiplier
}

// init 初始化随机数种子
func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}
