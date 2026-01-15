// Package cmd 提供API自动化测试命令行工具的命令实现
package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/morsuning/ai-auto-test-cmd/models"
	"github.com/morsuning/ai-auto-test-cmd/utils"
)

// printRequestResponsePair 成对打印请求和响应详细信息(用于debug模式)
func printRequestResponsePair(testCaseNum int, req utils.HTTPRequest, result models.TestResult) {
	fmt.Print("\n" + strings.Repeat("=", 65) + "\n")
	fmt.Printf("📋 测试用例 %d\n", testCaseNum)
	fmt.Println(strings.Repeat("=", 65))

	// ========== HTTP REQUEST 部分 ==========
	fmt.Println("\n🔵 HTTP REQUEST")
	fmt.Println("┌─────────────────────────────────────────────────────────────")

	// 输出URL和方法
	fmt.Printf("│ URL:    %s\n", req.URL)
	fmt.Printf("│ Method: %s\n", req.Method)
	fmt.Printf("│ Timeout: %d秒\n", req.Timeout)
	fmt.Println("│")

	// 输出HTTP Headers
	fmt.Println("│ HTTP Headers:")
	if len(req.Headers) == 0 {
		fmt.Println("│   (无自定义请求头)")
	} else {
		for key, value := range req.Headers {
			fmt.Printf("│   %s: %s\n", key, value)
		}
	}
	fmt.Println("│")

	// 输出HTTP Body
	fmt.Println("│ HTTP Body:")
	if req.Body == "" {
		fmt.Println("│   (空请求体)")
	} else {
		// 格式化输出请求体,每行前加上"│   "
		bodyLines := strings.Split(req.Body, "\n")
		for _, line := range bodyLines {
			fmt.Printf("│   %s\n", line)
		}
	}

	fmt.Println("└─────────────────────────────────────────────────────────────")

	// ========== HTTP RESPONSE 部分 ==========
	fmt.Println("\n🟢 HTTP RESPONSE")
	fmt.Println("┌─────────────────────────────────────────────────────────────")

	// 输出基本信息
	fmt.Printf("│ 测试用例ID: %s\n", result.TestCaseID)
	fmt.Printf("│ 状态码:     %d\n", result.StatusCode)
	fmt.Printf("│ 耗时:       %dms\n", result.Duration)
	fmt.Printf("│ 执行结果:   %s\n", func() string {
		if result.Success {
			return "✅ 成功"
		}
		return "❌ 失败"
	}())
	fmt.Println("│")

	// 输出错误信息(如果有)
	if result.Error != "" {
		fmt.Println("│ 错误信息:")
		errorLines := strings.Split(result.Error, "\n")
		for _, line := range errorLines {
			fmt.Printf("│   %s\n", line)
		}
		fmt.Println("│")
	}

	// 输出HTTP响应头
	fmt.Println("│ HTTP Response Headers:")
	if len(result.ResponseHeaders) == 0 {
		fmt.Println("│   (无响应头)")
	} else {
		for key, values := range result.ResponseHeaders {
			for _, value := range values {
				fmt.Printf("│   %s: %s\n", key, value)
			}
		}
	}
	fmt.Println("│")

	// 输出响应体
	fmt.Println("│ HTTP Response Body:")
	if result.ResponseBody == "" {
		fmt.Println("│   (空响应体)")
	} else {
		// 尝试格式化JSON响应体
		var jsonData any
		if err := json.Unmarshal([]byte(result.ResponseBody), &jsonData); err == nil {
			// 如果是有效的JSON,进行格式化输出
			if formattedJSON, err := json.MarshalIndent(jsonData, "│   ", "  "); err == nil {
				// 格式化输出JSON,每行前加上"│   "
				jsonLines := strings.Split(string(formattedJSON), "\n")
				for _, line := range jsonLines {
					fmt.Printf("│   %s\n", line)
				}
			} else {
				// JSON格式化失败,直接输出原始内容
				responseLines := strings.Split(result.ResponseBody, "\n")
				for _, line := range responseLines {
					fmt.Printf("│   %s\n", line)
				}
			}
		} else {
			// 不是JSON格式,直接输出原始内容
			responseLines := strings.Split(result.ResponseBody, "\n")
			for _, line := range responseLines {
				fmt.Printf("│   %s\n", line)
			}
		}
	}

	fmt.Println("└─────────────────────────────────────────────────────────────")
	fmt.Println(strings.Repeat("=", 65))
}
