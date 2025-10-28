package utils

import (
    "testing"
)

// TestCompositeFlatFields 验证平铺字段的组合约束覆盖
func TestCompositeFlatFields(t *testing.T) {
    cfg := `
[constraints]
enable = true

[constraints.address_bundle]
type = "composite"
fields = ["province", "city"]
data = [["上海", "上海市"], ["北京", "北京市"], ["广东", "广州市"]]
`
    fp := writeTempConfig(t, cfg)
    _, err := LoadConfigWithConstraints(fp)
    if err != nil {
        t.Fatalf("加载配置失败: %v", err)
    }
    // 清理全局约束，防止影响其他测试
    t.Cleanup(func(){ globalConstraintConfig = nil })

    data := map[string]any{
        "province": "未知",
        "city":     "未知",
    }

    // 重置字段顺序，确保按当前数据生成
    originalKeyOrder = nil
    cases := GenerateTestCasesWithVariationRate(data, 5, 0.0, true)
    validPairs := map[[2]string]bool{
        {"上海", "上海市"}: true,
        {"北京", "北京市"}: true,
        {"广东", "广州市"}: true,
    }

    for _, tc := range cases {
        p, pok := tc["province"].(string)
        c, cok := tc["city"].(string)
        if !pok || !cok {
            t.Fatalf("字段类型错误: province=%T, city=%T", tc["province"], tc["city"])
        }
        if !validPairs[[2]string{p, c}] {
            t.Fatalf("组合约束未按同组生成: province=%s city=%s", p, c)
        }
    }
}

// TestCompositeNestedFields 验证嵌套对象字段的组合约束覆盖
func TestCompositeNestedFields(t *testing.T) {
    cfg := `
[constraints]
enable = true

[constraints.address_bundle]
type = "composite"
fields = ["province", "city"]
data = [["上海", "上海市"], ["浙江", "杭州市"]]
`
    fp := writeTempConfig(t, cfg)
    _, err := LoadConfigWithConstraints(fp)
    if err != nil {
        t.Fatalf("加载配置失败: %v", err)
    }
    t.Cleanup(func(){ globalConstraintConfig = nil })

    data := map[string]any{
        "user": map[string]any{
            "address": map[string]any{
                "province": "未知",
                "city":     "未知",
            },
        },
    }

    originalKeyOrder = nil
    cases := GenerateTestCasesWithVariationRate(data, 5, 0.0, true)
    validPairs := map[[2]string]bool{
        {"上海", "上海市"}: true,
        {"浙江", "杭州市"}: true,
    }

    for _, tc := range cases {
        user, ok := tc["user"].(map[string]any)
        if !ok { t.Fatalf("user 字段类型错误: %T", tc["user"]) }
        addr, ok := user["address"].(map[string]any)
        if !ok { t.Fatalf("user.address 字段类型错误: %T", user["address"]) }
        p, pok := addr["province"].(string)
        c, cok := addr["city"].(string)
        if !pok || !cok {
            t.Fatalf("字段类型错误: province=%T, city=%T", addr["province"], addr["city"])
        }
        if !validPairs[[2]string{p, c}] {
            t.Fatalf("组合约束未按同组生成(嵌套): province=%s city=%s", p, c)
        }
    }
}

// TestCompositeArrayObjects 验证数组对象内字段的组合约束覆盖（同一用例内保持一致）
func TestCompositeArrayObjects(t *testing.T) {
    cfg := `
[constraints]
enable = true

[constraints.address_bundle]
type = "composite"
fields = ["province", "city"]
data = [["四川", "成都市"], ["福建", "福州市"]]
`
    fp := writeTempConfig(t, cfg)
    _, err := LoadConfigWithConstraints(fp)
    if err != nil {
        t.Fatalf("加载配置失败: %v", err)
    }
    t.Cleanup(func(){ globalConstraintConfig = nil })

    data := map[string]any{
        "addresses": []any{
            map[string]any{"province": "未知", "city": "未知"},
            map[string]any{"province": "未知", "city": "未知"},
        },
    }

    originalKeyOrder = nil
    cases := GenerateTestCasesWithVariationRate(data, 3, 0.0, true)
    validPairs := map[[2]string]bool{
        {"四川", "成都市"}: true,
        {"福建", "福州市"}: true,
    }

    for _, tc := range cases {
        arr, ok := tc["addresses"].([]any)
        if !ok { t.Fatalf("addresses 字段类型错误: %T", tc["addresses"]) }
        if len(arr) != 2 { t.Fatalf("addresses 长度错误: 期望2, 实际%d", len(arr)) }
        a0, ok0 := arr[0].(map[string]any)
        a1, ok1 := arr[1].(map[string]any)
        if !ok0 || !ok1 { t.Fatalf("addresses 元素类型错误") }
        p0, _ := a0["province"].(string); c0, _ := a0["city"].(string)
        p1, _ := a1["province"].(string); c1, _ := a1["city"].(string)

        if !validPairs[[2]string{p0, c0}] { t.Fatalf("组合约束未按同组生成(数组-0): %s %s", p0, c0) }
        if !validPairs[[2]string{p1, c1}] { t.Fatalf("组合约束未按同组生成(数组-1): %s %s", p1, c1) }
        // 同一用例内，当前实现为单组覆盖，两个元素应一致
        if p0 != p1 || c0 != c1 { t.Fatalf("同一用例内数组元素不一致: (%s,%s) vs (%s,%s)", p0, c0, p1, c1) }
    }
}

// TestCompositeNestedPathFields 验证 fields 使用嵌套路径时的组合约束覆盖
func TestCompositeNestedPathFields(t *testing.T) {
    cfg := `
[constraints]
enable = true

[constraints.address_bundle]
type = "composite"
fields = ["user.province", "user.city"]
data = [["浙江", "杭州市"], ["江苏", "南京市"]]
`
    fp := writeTempConfig(t, cfg)
    _, err := LoadConfigWithConstraints(fp)
    if err != nil {
        t.Fatalf("加载配置失败: %v", err)
    }
    t.Cleanup(func(){ globalConstraintConfig = nil })

    data := map[string]any{
        "user": map[string]any{
            "province": "未知",
            "city": "未知",
        },
    }

    originalKeyOrder = nil
    cases := GenerateTestCasesWithVariationRate(data, 5, 0.0, true)
    validPairs := map[[2]string]bool{
        {"浙江", "杭州市"}: true,
        {"江苏", "南京市"}: true,
    }

    for _, tc := range cases {
        user, ok := tc["user"].(map[string]any)
        if !ok { t.Fatalf("user 字段类型错误: %T", tc["user"]) }
        p, _ := user["province"].(string)
        c, _ := user["city"].(string)
        if !validPairs[[2]string{p, c}] {
            t.Fatalf("组合约束未按同组生成(嵌套路径 fields): %s %s", p, c)
        }
    }
}

// TestCompositeFieldNameNormalization 验证字段名大小写/连字符归一化匹配
func TestCompositeFieldNameNormalization(t *testing.T) {
    cfg := `
[constraints]
enable = true

[constraints.address_bundle]
type = "composite"
fields = ["Province", "city-name"]
data = [["湖北", "武汉市"], ["湖南", "长沙市"]]
`
    fp := writeTempConfig(t, cfg)
    _, err := LoadConfigWithConstraints(fp)
    if err != nil {
        t.Fatalf("加载配置失败: %v", err)
    }
    t.Cleanup(func(){ globalConstraintConfig = nil })

    data := map[string]any{
        "province":  "未知",
        "city_name": "未知",
    }

    originalKeyOrder = nil
    cases := GenerateTestCasesWithVariationRate(data, 4, 0.0, true)
    // 只需验证生成值属于两组之一，说明匹配成功
    validPairs := map[[2]string]bool{
        {"湖北", "武汉市"}: true,
        {"湖南", "长沙市"}: true,
    }

    for _, tc := range cases {
        p, _ := tc["province"].(string)
        c, _ := tc["city_name"].(string)
        if !validPairs[[2]string{p, c}] {
            t.Fatalf("组合约束字段归一化匹配失败: province=%s city_name=%s", p, c)
        }
    }
}

// TestCompositeConstraintValidationErrors 验证非法组合约束配置的错误
func TestCompositeConstraintValidationErrors(t *testing.T) {
    // 重复字段
    c1 := FieldConstraint{ Type: "composite", Fields: []string{"a", "a"}, Data: [][]string{{"x", "y"}} }
    errs := validateFieldConstraint("comp1", c1, nil)
    if len(errs) == 0 { t.Fatalf("重复字段未报错") }

    // 数据行长度不匹配
    c2 := FieldConstraint{ Type: "composite", Fields: []string{"p", "q"}, Data: [][]string{{"x"}} }
    errs = validateFieldConstraint("comp2", c2, nil)
    if len(errs) == 0 { t.Fatalf("数据行长度不匹配未报错") }

    // 缺少必要字段
    c3 := FieldConstraint{ Type: "composite", Fields: []string{}, Data: [][]string{} }
    errs = validateFieldConstraint("comp3", c3, nil)
    if len(errs) == 0 { t.Fatalf("缺少必要字段未报错") }
}