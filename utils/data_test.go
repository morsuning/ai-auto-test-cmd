package utils

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// TestXMLKeepOriginalPreservesLeadingZeros 验证在使用 keep_original 约束时，XML中的前导零不会丢失
func TestXMLKeepOriginalPreservesLeadingZeros(t *testing.T) {
    // 加载约束配置（使用测试用例中的约束文件）
    constraintsPath := filepath.Join("..", "test_cases", "test_keep_original_constraints.toml")
    if err := LoadConstraintConfig(constraintsPath); err != nil {
        t.Fatalf("加载约束配置失败: %v", err)
    }

    // 读取测试XML
    xmlPath := filepath.Join("..", "test_cases", "test_keep_original.xml")
    xmlBytes, err := os.ReadFile(xmlPath)
    if err != nil {
        t.Fatalf("读取测试XML失败: %v", err)
    }
    xmlStr := string(xmlBytes)

    // 解析XML
    data, err := ParseXML(xmlStr)
    if err != nil {
        t.Fatalf("解析XML失败: %v", err)
    }

    // 生成1条测试用例（启用约束）
    cases := GenerateTestCasesWithVariationRate(data, 1, 0.5, true)
    if len(cases) != 1 {
        t.Fatalf("生成的测试用例数量不正确: %d", len(cases))
    }

    // 转换为XML行
    rows := ConvertToXMLRows(cases)
    if len(rows) < 2 {
        t.Fatalf("XML行生成失败，行数: %d", len(rows))
    }

    // 第2行是第一条用例的XML字符串
    xmlOut := rows[1][0]

    // 验证 Head 节点下的前导零保留（TxnTm 和 SvcScn）
    if !strings.Contains(xmlOut, "<TxnTm>000137</TxnTm>") {
        t.Errorf("TxnTm 前导零未保留，输出: %s", xmlOut)
    }
    if !strings.Contains(xmlOut, "<SvcScn>09</SvcScn>") {
        t.Errorf("SvcScn 前导零未保留，输出: %s", xmlOut)
    }
}