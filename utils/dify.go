// Package utils 提供了一系列用于数据处理和测试用例生成的工具函数。
// 包含XML解析、JSON解析以及基于原始数据生成测试用例的功能。
package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DifyClient 表示Dify API客户端
type DifyClient struct {
	BaseURL   string
	Timeout   int
	UserAgent string
}

// DifyRequest 表示发送给Dify API的请求
type DifyRequest struct {
	Format       string      `json:"format"`       // 格式：xml或json
	RawMessage   string      `json:"raw_message"`  // 原始报文
	Count        int         `json:"count"`        // 生成数量
	Documentation interface{} `json:"documentation,omitempty"` // 接口文档（可选）
}

// DifyResponse 表示Dify API的响应
type DifyResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}

// NewDifyClient 创建新的Dify API客户端
func NewDifyClient(baseURL string, timeout int) *DifyClient {
	if timeout <= 0 {
		timeout = 60 // 默认60秒
	}

	return &DifyClient{
		BaseURL:   baseURL,
		Timeout:   timeout,
		UserAgent: "API-Auto-Test-CMD/1.0",
	}
}

// GenerateTestCases 通过Dify API生成测试用例
func (c *DifyClient) GenerateTestCases(req DifyRequest) ([]map[string]interface{}, error) {
	// 构建请求URL
	url := c.BaseURL

	// 将请求转换为JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.UserAgent)

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(c.Timeout) * time.Second,
	}

	// 发送请求
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var difyResp DifyResponse
	if err := json.Unmarshal(body, &difyResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查响应状态
	if !difyResp.Success {
		return nil, fmt.Errorf("API返回错误: %s", difyResp.Error)
	}

	// 解析测试用例数据
	testCasesData, ok := difyResp.Data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	// 转换为所需格式
	testCases := make([]map[string]interface{}, len(testCasesData))
	for i, item := range testCasesData {
		testCase, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("测试用例数据格式错误")
		}
		testCases[i] = testCase
	}

	return testCases, nil
}