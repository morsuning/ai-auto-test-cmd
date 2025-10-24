// Package utils 提供了一系列用于数据处理和测试用例生成的工具函数。
// 包含XML解析、JSON解析以及基于原始数据生成测试用例的功能。
package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPRequest HTTP请求结构体
type HTTPRequest struct {
	URL       string            `json:"url"`       // 请求URL
	Method    string            `json:"method"`    // 请求方法
	Headers   map[string]string `json:"headers"`   // 请求头
	Body      string            `json:"body"`      // 请求体
	Timeout   int               `json:"timeout"`   // 超时时间（秒）
	IgnoreTLS bool              `json:"ignore_tls"` // 忽略TLS证书验证
}

// HTTPResponse 表示HTTP响应的结构
type HTTPResponse struct {
	StatusCode int
	Headers    map[string][]string
	Body       string
	Error      error
	Duration   time.Duration
}

// SendRequest 发送HTTP请求
func SendRequest(req HTTPRequest) HTTPResponse {
	start := time.Now()
	response := HTTPResponse{}

	// 设置请求方法和URL
	httpMethod := req.Method
	if httpMethod == "" {
		httpMethod = "GET"
	}

	// 创建请求
	httpReq, err := http.NewRequest(httpMethod, req.URL, bytes.NewBufferString(req.Body))
	if err != nil {
		response.Error = fmt.Errorf("创建请求失败: %v", err)
		return response
	}

	// 设置请求头
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// 设置超时时间
	timeout := 30 // 默认30秒
	if req.Timeout > 0 {
		timeout = req.Timeout
	}

	// 创建客户端
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// 如果配置了忽略TLS证书验证，则直接使用不安全的客户端
	if req.IgnoreTLS {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	// 发送请求
	resp, err := client.Do(httpReq)
	if err != nil {
		response.Error = fmt.Errorf("发送请求失败: %v", err)
		return response
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Error = fmt.Errorf("读取响应失败: %v", err)
		return response
	}

	// 设置响应信息
	response.StatusCode = resp.StatusCode
	response.Headers = resp.Header
	response.Body = string(body)
	response.Duration = time.Since(start)

	return response
}

// SendConcurrentRequests 并发发送多个HTTP请求
func SendConcurrentRequests(requests []HTTPRequest, concurrency int) []HTTPResponse {
	if concurrency <= 0 {
		concurrency = 1
	}

	total := len(requests)
	responses := make([]HTTPResponse, total)

	// 创建通道
	jobs := make(chan int, total)
	results := make(chan struct {
		index    int
		response HTTPResponse
	}, total)

	// 启动工作协程
	for w := 1; w <= concurrency; w++ {
		go func() {
			for j := range jobs {
				resp := SendRequest(requests[j])
				results <- struct {
					index    int
					response HTTPResponse
				}{j, resp}
			}
		}()
	}

	// 发送任务
	for j := 0; j < total; j++ {
		jobs <- j
	}
	close(jobs)

	// 收集结果
	for a := 0; a < total; a++ {
		result := <-results
		responses[result.index] = result.response
	}

	return responses
}
