/*
Copyright © 2025 API自动化测试命令行工具

*/
package utils

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
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