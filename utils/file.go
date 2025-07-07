// Package utils 提供了一系列用于数据处理和测试用例生成的工具函数。
// 包含XML解析、JSON解析以及基于原始数据生成测试用例的功能。
package utils

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SaveToCSV 将数据保存为CSV文件
func SaveToCSV(data [][]string, filePath string) error {
	// 如果未指定文件路径，则使用默认路径
	if filePath == "" {
		filePath = "result.csv"
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %v", err)
		}
	}

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 写入CSV
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 检查数据是否为空
	if len(data) == 0 {
		return nil
	}

	// 写入数据
	for _, row := range data {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("写入CSV失败: %v", err)
		}
	}

	return nil
}

// ReadCSV 从CSV文件读取数据
func ReadCSV(filePath string) ([][]string, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 读取CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("读取CSV失败: %v", err)
	}

	return records, nil
}

// ReadFileContent 读取文件内容并返回字符串
func ReadFileContent(filePath string) (string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("文件不存在: %s", filePath)
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}

	return string(content), nil
}

// ValidateXMLFormat 验证字符串是否为有效的XML格式
func ValidateXMLFormat(content string) error {
	// 去除首尾空白字符
	content = strings.TrimSpace(content)
	
	// 检查是否为空
	if content == "" {
		return fmt.Errorf("XML内容不能为空")
	}
	
	// 尝试解析XML
	var xmlData interface{}
	err := xml.Unmarshal([]byte(content), &xmlData)
	if err != nil {
		return fmt.Errorf("无效的XML格式: %v", err)
	}
	
	return nil
}

// ValidateJSONFormat 验证字符串是否为有效的JSON格式
func ValidateJSONFormat(content string) error {
	// 去除首尾空白字符
	content = strings.TrimSpace(content)
	
	// 检查是否为空
	if content == "" {
		return fmt.Errorf("JSON内容不能为空")
	}
	
	// 尝试解析JSON
	var jsonData interface{}
	err := json.Unmarshal([]byte(content), &jsonData)
	if err != nil {
		return fmt.Errorf("无效的JSON格式: %v", err)
	}
	
	return nil
}

// ReadAndValidateFileContent 读取文件内容并根据指定格式进行验证
func ReadAndValidateFileContent(filePath string, format string) (string, error) {
	// 读取文件内容
	content, err := ReadFileContent(filePath)
	if err != nil {
		return "", err
	}
	
	// 根据格式进行验证
	switch strings.ToLower(format) {
	case "xml":
		if err := ValidateXMLFormat(content); err != nil {
			return "", fmt.Errorf("文件 %s 格式验证失败: %v", filePath, err)
		}
	case "json":
		if err := ValidateJSONFormat(content); err != nil {
			return "", fmt.Errorf("文件 %s 格式验证失败: %v", filePath, err)
		}
	default:
		return "", fmt.Errorf("不支持的格式: %s，仅支持 xml 或 json", format)
	}
	
	return content, nil
}