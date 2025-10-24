package utils

import (
	"errors"
	"testing"
)

// TestValidateTimezone 测试时区验证功能
func TestValidateTimezone(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		wantErr  bool
	}{
		// 基本有效时区
		{"空时区", "", false},
		{"UTC时区", "UTC", false},

		// 偏移量格式
		{"正偏移量", "+08:00", false},
		{"负偏移量", "-05:00", false},
		{"零偏移量", "+00:00", false},

		// 无效偏移量格式
		{"无效偏移量格式1", "+8:00", true},
		{"无效偏移量格式2", "+08:0", true},
		{"无效偏移量格式3", "08:00", true},
		{"超出范围偏移量", "+25:00", true},

		// IANA 时区名称（在支持的系统上应该有效）
		{"亚洲上海", "Asia/Shanghai", false},
		{"美国纽约", "America/New_York", false},
		{"欧洲伦敦", "Europe/London", false},

		// 无效的 IANA 格式
		{"无斜杠", "AsiaShanghai", true},
		{"以斜杠开头", "/Asia/Shanghai", true},
		{"以斜杠结尾", "Asia/Shanghai/", true},
		{"连续斜杠", "Asia//Shanghai", true},
		{"无效字符", "Asia/Shang@hai", true},
		{"过多层级", "Asia/China/Shanghai/District", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimezone(tt.timezone)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTimezone() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestIsWindowsTimezoneValidationError 测试Windows时区验证错误检测
func TestIsWindowsTimezoneValidationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "unknown time zone错误",
			err:  errors.New("unknown time zone Asia/Shanghai"),
			want: true,
		},
		{
			name: "cannot find错误",
			err:  errors.New("cannot find time zone Asia/Shanghai"),
			want: true,
		},
		{
			name: "zoneinfo路径错误",
			err:  errors.New("open /usr/share/zoneinfo/Asia/Shanghai: no such file"),
			want: true,
		},
		{
			name: "其他错误",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "网络错误",
			err:  errors.New("network timeout"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isWindowsTimezoneValidationError(tt.err); got != tt.want {
				t.Errorf("isWindowsTimezoneValidationError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsValidIANATimezoneFormat 测试IANA时区格式验证
func TestIsValidIANATimezoneFormat(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		want     bool
	}{
		// 有效的IANA时区格式
		{"标准亚洲时区", "Asia/Shanghai", true},
		{"美洲时区", "America/New_York", true},
		{"欧洲时区", "Europe/London", true},
		{"澳洲时区", "Australia/Sydney", true},
		{"太平洋时区", "Pacific/Auckland", true},
		{"三级时区", "America/Argentina/Buenos_Aires", true},
		{"包含下划线", "America/Los_Angeles", true},
		{"包含连字符", "America/Port-au-Prince", true},
		{"数字时区", "GMT+8", false}, // 这不是标准IANA格式

		// 无效的IANA时区格式
		{"无斜杠", "AsiaShanghai", false},
		{"以斜杠开头", "/Asia/Shanghai", false},
		{"以斜杠结尾", "Asia/Shanghai/", false},
		{"连续斜杠", "Asia//Shanghai", false},
		{"只有一个部分", "Asia/", false},
		{"空字符串", "", false},
		{"只有斜杠", "/", false},
		{"无效字符-@", "Asia/Shang@hai", false},
		{"无效字符-空格", "Asia/Shang hai", false},
		{"无效字符-特殊符号", "Asia/Shanghai!", false},
		{"过多层级", "Asia/China/Shanghai/District", false},
		{"只有一层", "Asia", false},

		// 边界情况
		{"最短有效格式", "A/B", true},
		{"非标准区域但格式正确", "Custom/Zone", true},
		{"区域名为空", "/Shanghai", false},
		{"位置名为空", "Asia/", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidIANATimezoneFormat(tt.timezone); got != tt.want {
				t.Errorf("isValidIANATimezoneFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestValidateTimezoneOffset 测试时区偏移量验证
func TestValidateTimezoneOffset(t *testing.T) {
	tests := []struct {
		name    string
		offset  string
		wantErr bool
	}{
		// 有效偏移量
		{"标准正偏移", "+08:00", false},
		{"标准负偏移", "-05:00", false},
		{"零偏移", "+00:00", false},
		{"最大正偏移", "+14:00", false},
		{"最大负偏移", "-12:00", false},
		{"30分钟偏移", "+05:30", false},
		{"45分钟偏移", "+09:45", false},

		// 无效偏移量
		{"格式错误-缺少冒号", "+0800", true},
		{"格式错误-单位数小时", "+8:00", true},
		{"格式错误-单位数分钟", "+08:0", true},
		{"格式错误-无符号", "08:00", true},
		{"格式错误-错误符号", "*08:00", true},
		{"超出范围-小时过大", "+25:00", true},
		{"超出范围-分钟过大", "+08:60", true},
		{"超出范围-负小时过大", "-25:00", true},
		{"超出范围-总偏移过大", "+15:00", true},
		{"超出范围-总偏移过小", "-13:00", true},
		{"格式错误-过长", "+08:00:00", true},
		{"格式错误-过短", "+08:", true},
		{"格式错误-非数字小时", "+ab:00", true},
		{"格式错误-非数字分钟", "+08:ab", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTimezoneOffset(tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTimezoneOffset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// BenchmarkValidateTimezone 性能测试
func BenchmarkValidateTimezone(b *testing.B) {
	timezones := []string{
		"UTC",
		"+08:00",
		"Asia/Shanghai",
		"America/New_York",
		"Europe/London",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tz := range timezones {
			_ = ValidateTimezone(tz)
		}
	}
}

// BenchmarkIsValidIANATimezoneFormat 性能测试
func BenchmarkIsValidIANATimezoneFormat(b *testing.B) {
	timezones := []string{
		"Asia/Shanghai",
		"America/New_York",
		"Europe/London",
		"Australia/Sydney",
		"Pacific/Auckland",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tz := range timezones {
			isValidIANATimezoneFormat(tz)
		}
	}
}

// TestGenerateBankCard 测试银行卡号生成功能
func TestGenerateBankCard(t *testing.T) {
	// 创建银行卡约束
	constraint := &FieldConstraint{
		Type: "bank_card",
	}

	// 测试多次生成，确保格式正确
	for i := 0; i < 10; i++ {
		result := GenerateConstrainedValue(constraint, "")
		cardNumber, ok := result.(string)
		if !ok {
			t.Errorf("生成的银行卡号不是字符串类型: %T", result)
			continue
		}

		// 验证长度（银行卡号通常为15-19位）
		if len(cardNumber) < 15 || len(cardNumber) > 19 {
			t.Errorf("银行卡号长度错误: 期望15-19位，实际%d位，卡号: %s", len(cardNumber), cardNumber)
		}

		// 验证是否为纯数字
		for _, char := range cardNumber {
			if char < '0' || char > '9' {
				t.Errorf("银行卡号包含非数字字符: %s", cardNumber)
			}
		}
	}
}

// TestBankCardConstraintValidation 测试银行卡约束验证
func TestBankCardConstraintValidation(t *testing.T) {
	// 测试有效的银行卡约束
	constraint := FieldConstraint{
		Type: "bank_card",
	}

	errors := validateFieldConstraint("test_bank_card", constraint)
	if len(errors) != 0 {
		t.Errorf("有效的银行卡约束验证失败: %v", errors)
	}
}
