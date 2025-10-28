package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// helper to write temp config file
func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	fp := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
		t.Fatalf("写入临时配置失败: %v", err)
	}
	return fp
}

// helper for benchmark to write temp config file
func writeTempConfigB(b *testing.B, content string) string {
	b.Helper()
	dir := b.TempDir()
	fp := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
		b.Fatalf("写入临时配置失败: %v", err)
	}
	return fp
}

// TestCustomTypeValuesInline 测试内联值的自定义类型生成
func TestCustomTypeValuesInline(t *testing.T) {
	cfg := `
[constraints]
enable = true

[constraints.types.status_text]
values = ["pending", "paid", "shipped"]

[constraints.order_status]
type = "status_text"
`
	fp := writeTempConfig(t, cfg)
	conf, err := LoadConfigWithConstraints(fp)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	if !IsConstraintsEnabled(conf) {
		t.Fatalf("约束系统未启用")
	}
	c := FindFieldConstraint("order_status")
	if c == nil {
		t.Fatalf("未找到字段约束: order_status")
	}
	val := GenerateConstrainedValue(c, "")
	s, ok := val.(string)
	if !ok {
		t.Fatalf("生成的值不是字符串: %T", val)
	}
	allowed := map[string]bool{"pending": true, "paid": true, "shipped": true}
	if !allowed[s] {
		t.Fatalf("生成值不在允许集合内: %s", s)
	}
}

// TestCustomTypeDataset 测试引用自定义数据集的自定义类型生成
func TestCustomTypeDataset(t *testing.T) {
	cfg := `
[constraints]
enable = true

[constraints.datasets]
region_codes = ["CN-01", "CN-02", "US-01"]

[constraints.types.region_code]
dataset = "region_codes"

[constraints.region]
type = "region_code"
`
	fp := writeTempConfig(t, cfg)
	_, err := LoadConfigWithConstraints(fp)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	// 清理全局约束，防止影响其他测试
	t.Cleanup(func() { globalConstraintConfig = nil })
	c := FindFieldConstraint("region")
	if c == nil {
		t.Fatalf("未找到字段约束: region")
	}
	val := GenerateConstrainedValue(c, "")
	s, ok := val.(string)
	if !ok {
		t.Fatalf("生成的值不是字符串: %T", val)
	}
	allowed := map[string]bool{"CN-01": true, "CN-02": true, "US-01": true}
	if !allowed[s] {
		t.Fatalf("生成值不在数据集中: %s", s)
	}
}

// TestCustomTypePatternFilter 测试正则过滤生效
func TestCustomTypePatternFilter(t *testing.T) {
	cfg := `
[constraints]
enable = true

[constraints.types.alpha_only]
values = ["ABC", "123XYZ", "HELLO", "456"]
pattern = "^[A-Z]+$"

[constraints.sample]
type = "alpha_only"
`
	fp := writeTempConfig(t, cfg)
	_, err := LoadConfigWithConstraints(fp)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	c := FindFieldConstraint("sample")
	if c == nil {
		t.Fatalf("未找到字段约束: sample")
	}
	// 多次生成，确保匹配正则
	re := regexp.MustCompile("^[A-Z]+$")
	for i := 0; i < 10; i++ {
		val := GenerateConstrainedValue(c, "")
		s, ok := val.(string)
		if !ok {
			t.Fatalf("生成的值不是字符串: %T", val)
		}
		if !re.MatchString(s) {
			t.Fatalf("生成值未匹配正则: %s", s)
		}
	}
}

// TestCustomTypeInvalidPatternIgnored 测试无效正则被忽略但不影响生成
func TestCustomTypeInvalidPatternIgnored(t *testing.T) {
	cfg := `
[constraints]
enable = true

[constraints.types.alphanum]
values = ["A1", "B2", "C3"]
pattern = "["

[constraints.sample]
type = "alphanum"
`
	fp := writeTempConfig(t, cfg)
	_, err := LoadConfigWithConstraints(fp)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	c := FindFieldConstraint("sample")
	if c == nil {
		t.Fatalf("未找到字段约束: sample")
	}
	allowed := map[string]bool{"A1": true, "B2": true, "C3": true}
	for i := 0; i < 10; i++ {
		val := GenerateConstrainedValue(c, "")
		s, ok := val.(string)
		if !ok {
			t.Fatalf("生成的值不是字符串: %T", val)
		}
		if !allowed[s] {
			t.Fatalf("生成值未在允许集合内: %s", s)
		}
	}
}

// TestCustomTypeUnknownDatasetFallback 测试未知数据集时回退到values集合
func TestCustomTypeUnknownDatasetFallback(t *testing.T) {
	cfg := `
[constraints]
enable = true

[constraints.types.region_or_fallback]
dataset = "unknown_ds"
values = ["X-01", "Y-02"]

[constraints.region]
type = "region_or_fallback"
`
	fp := writeTempConfig(t, cfg)
	_, err := LoadConfigWithConstraints(fp)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	c := FindFieldConstraint("region")
	if c == nil {
		t.Fatalf("未找到字段约束: region")
	}
	allowed := map[string]bool{"X-01": true, "Y-02": true}
	for i := 0; i < 10; i++ {
		val := GenerateConstrainedValue(c, "")
		s, ok := val.(string)
		if !ok {
			t.Fatalf("生成的值不是字符串: %T", val)
		}
		if !allowed[s] {
			t.Fatalf("生成值未在允许集合内: %s", s)
		}
	}
}

// TestCustomTypeEmptySpecReturnsOriginal 当values和dataset都为空时应返回原值
func TestCustomTypeEmptySpecReturnsOriginal(t *testing.T) {
	cfg := `
[constraints]
enable = true

[constraints.types.empty_spec]

[constraints.sample]
type = "empty_spec"
`
	fp := writeTempConfig(t, cfg)
	_, err := LoadConfigWithConstraints(fp)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	c := FindFieldConstraint("sample")
	if c == nil {
		t.Fatalf("未找到字段约束: sample")
	}
	orig := "original-value"
	val := GenerateConstrainedValue(c, orig)
	s, ok := val.(string)
	if !ok {
		t.Fatalf("生成的值不是字符串: %T", val)
	}
	if s != orig {
		t.Fatalf("当类型候选为空时应返回原值，得到: %s", s)
	}
}

// TestLocalGenNestedCustomTypesIntegration 集成测试：嵌套字段上的自定义类型
func TestLocalGenNestedCustomTypesIntegration(t *testing.T) {
	cfg := `
[constraints]
enable = true

[constraints.datasets]
region_codes = ["US-01", "CN-01"]

[constraints.types.status_text]
values = ["pending", "paid", "shipped"]

[constraints.types.region_code]
dataset = "region_codes"

[constraints.user.region]
type = "region_code"

[constraints.user.profile.status_text]
type = "status_text"

[constraints.order.status_text]
type = "status_text"
`
	fp := writeTempConfig(t, cfg)
	_, err := LoadConfigWithConstraints(fp)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 构造嵌套正例数据
	data := map[string]any{
		"user": map[string]any{
			"region":  "unknown",
			"profile": map[string]any{"status_text": "unset"},
		},
		"order": map[string]any{
			"status_text": "pending",
		},
	}

	// 重置字段顺序以避免使用之前解析的顺序
	originalKeyOrder = nil
	cases := GenerateTestCasesWithVariationRate(data, 3, 0.0, true)
	allowedStatus := map[string]bool{"pending": true, "paid": true, "shipped": true}
	allowedRegion := map[string]bool{"US-01": true, "CN-01": true}

	for _, tc := range cases {
		user, ok := tc["user"].(map[string]any)
		if !ok {
			t.Fatalf("user 字段类型错误: %T", tc["user"])
		}
		region, ok := user["region"].(string)
		if !ok || !allowedRegion[region] {
			t.Fatalf("region 值非法: %v", user["region"])
		}

		profile, ok := user["profile"].(map[string]any)
		if !ok {
			t.Fatalf("user.profile 字段类型错误: %T", user["profile"])
		}
		st1, ok := profile["status_text"].(string)
		if !ok || !allowedStatus[st1] {
			t.Fatalf("user.profile.status_text 值非法: %v", profile["status_text"])
		}

		order, ok := tc["order"].(map[string]any)
		if !ok {
			t.Fatalf("order 字段类型错误: %T", tc["order"])
		}
		st2, ok := order["status_text"].(string)
		if !ok || !allowedStatus[st2] {
			t.Fatalf("order.status_text 值非法: %v", order["status_text"])
		}
	}
}

// BenchmarkGenerateCustomTypeValue 性能基准：自定义类型生成
func BenchmarkGenerateCustomTypeValue(b *testing.B) {
	cfg := `
[constraints]
enable = true

[constraints.datasets]
region_codes = ["CN-01", "US-02", "JP-01", "DE-01"]

[constraints.types.region_code]
dataset = "region_codes"
pattern = "^[A-Z]{2}-\\d{2}$"

[constraints.user_region]
type = "region_code"
`
	fp := writeTempConfigB(b, cfg)
	_, err := LoadConfigWithConstraints(fp)
	if err != nil {
		b.Fatalf("加载配置失败: %v", err)
	}
	c := FindFieldConstraint("user_region")
	if c == nil {
		b.Fatalf("未找到字段约束: user_region")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateConstrainedValue(c, "")
	}
}
