// Package cmd 提供API自动化测试命令行工具的命令实现
package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/morsuning/ai-auto-test-cmd/models"
	"github.com/morsuning/ai-auto-test-cmd/utils"
)

// RequestParams 包含request命令的所有参数
type RequestParams struct {
	URL           string   // 目标URL
	Method        string   // 请求方法
	Save          bool     // 是否保存结果
	SavePath      string   // 结果保存路径
	Timeout       int      // 请求超时时间
	Concurrent    int      // 并发请求数
	Debug         bool     // 调试模式
	AuthBearer    string   // Bearer Token认证
	AuthBasic     string   // Basic Auth认证
	AuthAPIKey    string   // API Key认证
	CustomHeaders []string // 自定义HTTP头
	QueryParams   []string // URL查询参数
	IsXML         bool     // 使用XML格式
	IsJSON        bool     // 使用JSON格式
}

// validateRequestParams 验证request参数
func validateRequestParams(params RequestParams) error {
	// 验证URL
	if params.URL == "" {
		return fmt.Errorf("使用 -e 参数时必须在配置文件中指定 request.url")
	}

	// 验证请求体格式参数
	if !params.IsXML && !params.IsJSON {
		return fmt.Errorf("使用 -e 参数时必须指定请求体格式（--xml 或 --json）")
	}

	if params.IsXML && params.IsJSON {
		return fmt.Errorf("不能同时指定 --xml 和 --json 参数，请只选择一种格式")
	}

	// GET请求现在支持JSON和XML格式，不再有格式限制

	return nil
}

// executeGeneratedTestCases 执行生成的测试用例
func executeGeneratedTestCases(outputFile string, params RequestParams) error {
	fmt.Println("\n🚀 开始执行生成的测试用例...")

	// 确定内容类型
	contentType := "json"
	if params.IsXML {
		contentType = "xml"
	}

	// 打印执行信息
	fmt.Printf("目标URL: %s\n", params.URL)
	fmt.Printf("请求方法: %s\n", strings.ToUpper(params.Method))
	fmt.Printf("测试用例文件: %s\n", outputFile)
	fmt.Printf("内容类型: %s\n", contentType)
	fmt.Printf("并发数: %d\n", params.Concurrent)
	fmt.Printf("请求超时时间: %d秒\n", params.Timeout)
	fmt.Println()

	// 构建鉴权配置
	authConfig := AuthConfig{
		BearerToken:   params.AuthBearer,
		BasicAuth:     params.AuthBasic,
		APIKey:        params.AuthAPIKey,
		CustomHeaders: params.CustomHeaders,
	}

	// 执行批量请求
	if err := executeBatchRequestsWithAuth(params.URL, params.Method, outputFile, params.Save, params.SavePath, params.Timeout, params.Concurrent, contentType, params.Debug, authConfig, params.QueryParams); err != nil {
		return fmt.Errorf("执行测试用例失败: %v", err)
	}

	return nil
}

// executeTestCasesDirectly 直接执行测试用例数据，不依赖文件
func executeTestCasesDirectly(testCases []models.TestCase, params RequestParams) error {
	fmt.Println("\n🚀 开始执行生成的测试用例...")

	// 确定内容类型
	contentType := "json"
	if params.IsXML {
		contentType = "xml"
	}

	// 打印执行信息
	fmt.Printf("目标URL: %s\n", params.URL)
	fmt.Printf("请求方法: %s\n", strings.ToUpper(params.Method))
	fmt.Printf("测试用例数量: %d\n", len(testCases))
	fmt.Printf("内容类型: %s\n", contentType)
	fmt.Printf("并发数: %d\n", params.Concurrent)
	fmt.Printf("请求超时时间: %d秒\n", params.Timeout)
	fmt.Println()

	// 构建鉴权配置
	authConfig := AuthConfig{
		BearerToken:   params.AuthBearer,
		BasicAuth:     params.AuthBasic,
		APIKey:        params.AuthAPIKey,
		CustomHeaders: params.CustomHeaders,
	}

	// 构建HTTP请求
	useJSON := strings.ToLower(contentType) == "json"
	useXML := strings.ToLower(contentType) == "xml"
	requests, err := buildHTTPRequestsWithAuth(testCases, params.URL, params.Method, params.Timeout, useJSON, useXML, authConfig, params.QueryParams)
	if err != nil {
		return fmt.Errorf("构建HTTP请求失败: %v", err)
	}

	// 如果启用调试模式，输出请求详情
	if params.Debug {
		printDebugInfo(requests)
	}

	// 执行批量请求
	fmt.Println("🚀 开始执行批量请求...")
	start := time.Now()
	responses := utils.SendConcurrentRequests(requests, params.Concurrent)
	duration := time.Since(start)

	// 处理响应结果
	results := processResponses(testCases, responses, requests)

	// 显示结果统计
	displayResults(results, duration, params.Debug)

	// 保存结果（如果需要）
	if params.Save {
		savePath := params.SavePath
		if savePath == "" {
			savePath = "result.csv"
		}
		if err := saveResults(results, savePath); err != nil {
			return fmt.Errorf("保存结果失败: %v", err)
		}
	}

	return nil
}

// getFormatName 获取报文格式名称
func getFormatName(isXML, isJSON bool) string {
	if isXML {
		return "XML"
	} else if isJSON {
		return "JSON"
	}
	return "未指定"
}

// truncateString 截断字符串到指定长度，用于调试输出
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// AuthConfig 鉴权配置结构体
type AuthConfig struct {
	BearerToken   string   // Bearer Token认证
	BasicAuth     string   // Basic Auth认证（username:password格式）
	APIKey        string   // API Key认证
	CustomHeaders []string // 自定义HTTP头（Key: Value格式）
}

// buildHTTPRequestsWithAuth 构建HTTP请求列表（支持鉴权）
func buildHTTPRequestsWithAuth(testCases []models.TestCase, url, method string, timeout int, useJSON, useXML bool, authConfig AuthConfig, queryParams []string) ([]utils.HTTPRequest, error) {
	// 检查并添加默认协议
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
		fmt.Printf("ℹ️  URL 未指定协议，默认使用 HTTP: %s\n", url)
	}

	requests := make([]utils.HTTPRequest, len(testCases))

	for i, testCase := range testCases {
		// 构建请求体
		body := ""
		headers := make(map[string]string)

		// 应用鉴权配置
		if err := applyAuthConfig(headers, authConfig); err != nil {
			return nil, err
		}

		if strings.ToUpper(method) == "POST" {
			// POST请求，根据格式化数据
			if useXML {
				// XML格式
				if xmlContent, exists := testCase.Data["_xml_content"]; exists {
					// 直接使用XML内容
					body = fmt.Sprintf("%v", xmlContent)
					headers["Content-Type"] = "application/xml"
				} else {
					// 从字段数据转换为XML
					xmlData, err := convertToXML(testCase.Data)
					if err != nil {
						// 如果转换失败，回退到JSON
						jsonData, _ := json.Marshal(testCase.Data)
						body = string(jsonData)
						headers["Content-Type"] = "application/json"
					} else {
						body = xmlData
						headers["Content-Type"] = "application/xml"
					}
				}
			}
			if useJSON {
				// JSON格式（默认）
				if jsonContent, exists := testCase.Data["_json_content"]; exists {
					// 直接使用JSON内容
					body = fmt.Sprintf("%v", jsonContent)
					headers["Content-Type"] = "application/json"
				} else {
					// 从字段数据转换为JSON
					jsonData, _ := json.Marshal(testCase.Data)
					body = string(jsonData)
					headers["Content-Type"] = "application/json"
				}
			}
		} else if strings.ToUpper(method) == "GET" {
			// GET请求现在支持在body中放置JSON/XML数据
			if useXML {
				// XML格式
				if xmlContent, exists := testCase.Data["_xml_content"]; exists {
					// 直接使用XML内容
					body = fmt.Sprintf("%v", xmlContent)
					headers["Content-Type"] = "application/xml"
				} else {
					// 从字段数据转换为XML
					xmlData, err := convertToXML(testCase.Data)
					if err != nil {
						// 如果转换失败，回退到JSON
						jsonData, _ := json.Marshal(testCase.Data)
						body = string(jsonData)
						headers["Content-Type"] = "application/json"
					} else {
						body = xmlData
						headers["Content-Type"] = "application/xml"
					}
				}
			}
			if useJSON {
				// JSON格式
				if jsonContent, exists := testCase.Data["_json_content"]; exists {
					// 直接使用JSON内容
					body = fmt.Sprintf("%v", jsonContent)
					headers["Content-Type"] = "application/json"
				} else {
					// 从字段数据转换为JSON
					jsonData, _ := json.Marshal(testCase.Data)
					body = string(jsonData)
					headers["Content-Type"] = "application/json"
				}
			}

			// 查询参数将在最后统一处理
			headers["Accept"] = "application/json"
		} else {
			// 其他请求方法
			headers["Accept"] = "application/json"
		}
		// 构建最终URL（包含查询参数）
		finalURL := url
		if len(queryParams) > 0 {
			separator := "?"
			if strings.Contains(url, "?") {
				separator = "&"
			}
			finalURL = url + separator + strings.Join(queryParams, "&")
		}

		requests[i] = utils.HTTPRequest{
			URL:     finalURL,
			Method:  strings.ToUpper(method),
			Headers: headers,
			Body:    body,
			Timeout: timeout,
		}
	}
	return requests, nil
}

// applyAuthConfig 应用鉴权配置到HTTP头
func applyAuthConfig(headers map[string]string, authConfig AuthConfig) error {
	// 应用Bearer Token认证
	if authConfig.BearerToken != "" {
		headers["Authorization"] = "Bearer " + authConfig.BearerToken
	}

	// 应用Basic Auth认证
	if authConfig.BasicAuth != "" {
		// 解析username:password格式
		parts := strings.SplitN(authConfig.BasicAuth, ":", 2)
		if len(parts) == 2 {
			// 编码为Base64
			credentials := base64.StdEncoding.EncodeToString([]byte(authConfig.BasicAuth))
			headers["Authorization"] = "Basic " + credentials
		} else {
			fmt.Printf("⚠️  警告: Basic Auth格式不正确，应为 'username:password'，跳过Basic Auth认证\n")
		}
	}

	// 应用API Key认证
	if authConfig.APIKey != "" {
		parts := strings.SplitN(authConfig.APIKey, ":", 2)
		if len(parts) == 2 {
			headers[parts[0]] = parts[1]
		} else {
			// 如果格式不正确，默认使用X-API-Key作为header名
			headers["X-API-Key"] = authConfig.APIKey
		}
	}

	// 应用自定义HTTP头
	for _, header := range authConfig.CustomHeaders {
		// 解析Key: Value格式
		parts := strings.SplitN(header, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("自定义HTTP头格式错误: %s，正确格式应为 'HeaderName: HeaderValue'", header)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return fmt.Errorf("自定义HTTP头名称不能为空: %s", header)
		}
		headers[key] = value
	}
	return nil
}

// convertToXML 将数据转换为XML格式
func convertToXML(data map[string]any) (string, error) {
	// 由于Go的xml包对map支持有限，我们手动构建XML字符串
	var xmlBuilder strings.Builder
	xmlBuilder.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	xmlBuilder.WriteString("<data>\n")

	for key, value := range data {
		// 清理XML标签名（移除特殊字符）
		cleanKey := strings.ReplaceAll(key, " ", "_")
		cleanKey = strings.ReplaceAll(cleanKey, "-", "_")

		xmlBuilder.WriteString(fmt.Sprintf("  <%s>", cleanKey))

		// 根据值的类型进行处理
		switch v := value.(type) {
		case string:
			// 转义XML特殊字符
			escapedValue := strings.ReplaceAll(v, "&", "&amp;")
			escapedValue = strings.ReplaceAll(escapedValue, "<", "&lt;")
			escapedValue = strings.ReplaceAll(escapedValue, ">", "&gt;")
			escapedValue = strings.ReplaceAll(escapedValue, "\"", "&quot;")
			escapedValue = strings.ReplaceAll(escapedValue, "'", "&apos;")
			xmlBuilder.WriteString(escapedValue)
		case int, int32, int64, float32, float64, bool:
			xmlBuilder.WriteString(fmt.Sprintf("%v", v))
		default:
			// 对于复杂类型，尝试JSON序列化后转义
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				xmlBuilder.WriteString(fmt.Sprintf("%v", v))
			} else {
				escapedValue := strings.ReplaceAll(string(jsonBytes), "&", "&amp;")
				escapedValue = strings.ReplaceAll(escapedValue, "<", "&lt;")
				escapedValue = strings.ReplaceAll(escapedValue, ">", "&gt;")
				xmlBuilder.WriteString(escapedValue)
			}
		}

		xmlBuilder.WriteString(fmt.Sprintf("</%s>\n", cleanKey))
	}

	xmlBuilder.WriteString("</data>")
	return xmlBuilder.String(), nil
}

// printDebugInfo 打印调试信息
func printDebugInfo(requests []utils.HTTPRequest) {
	fmt.Println("\n=== 调试信息 ===")
	fmt.Printf("总请求数: %d\n\n", len(requests))

	for i, req := range requests {
		fmt.Printf("📋 请求 %d:\n", i+1)
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
			// 格式化输出请求体，每行前加上"│   "
			bodyLines := strings.Split(req.Body, "\n")
			for _, line := range bodyLines {
				fmt.Printf("│   %s\n", line)
			}
		}

		fmt.Println("└─────────────────────────────────────────────────────────────")
		fmt.Println()
	}

	fmt.Println("=== 调试信息结束 ===")
}

// processResponses 处理响应结果
func processResponses(testCases []models.TestCase, responses []utils.HTTPResponse, requests []utils.HTTPRequest) []models.TestResult {
	results := make([]models.TestResult, len(testCases))

	for i, response := range responses {
		result := models.TestResult{
			TestCaseID:   testCases[i].ID,
			StatusCode:   response.StatusCode,
			ResponseBody: response.Body,
			RequestBody:  "", // 默认为空，下面会设置
			Duration:     response.Duration.Milliseconds(),
		}

		// 设置原始请求报文
		if i < len(requests) {
			result.RequestBody = requests[i].Body
		}

		if response.Error != nil {
			result.Success = false
			result.Error = response.Error.Error()
		} else {
			// 简单判断：状态码2xx为成功
			result.Success = response.StatusCode >= 200 && response.StatusCode < 300
		}

		results[i] = result
	}

	return results
}

// displayResults 显示结果统计
func displayResults(results []models.TestResult, duration time.Duration, debug bool) {
	fmt.Println("\n=== 执行结果 ===")

	total := len(results)
	success := 0
	failed := 0

	for i, result := range results {
		if result.Success {
			success++
			fmt.Printf("✅ 测试用例 %d: 成功 (状态码: %d, 耗时: %dms)\n", i+1, result.StatusCode, result.Duration)
			// 在debug模式下，也输出成功响应的详细信息
			if debug {
				printResponseDetails(i+1, result)
			}
		} else {
			failed++
			if result.Error != "" {
				fmt.Printf("❌ 测试用例 %d: 失败 - %s\n", i+1, result.Error)
			} else {
				fmt.Printf("❌ 测试用例 %d: 失败 (状态码: %d, 耗时: %dms)\n", i+1, result.StatusCode, result.Duration)
			}
			// 在debug模式下，输出失败响应的详细信息
			if debug {
				printResponseDetails(i+1, result)
			}
		}
	}

	fmt.Println("\n=== 统计信息 ===")
	fmt.Printf("总计: %d\n", total)
	fmt.Printf("成功: %d\n", success)
	fmt.Printf("失败: %d\n", failed)
	fmt.Printf("成功率: %.2f%%\n", float64(success)/float64(total)*100)
	fmt.Printf("总耗时: %v\n", duration)
}

// saveResults 保存结果到文件
func saveResults(results []models.TestResult, savePath string) error {
	// 确定保存路径
	if savePath == "" {
		savePath = "result.csv"
	}

	// 如果指定的是目录，则在目录下创建默认文件名
	if info, err := os.Stat(savePath); err == nil && info.IsDir() {
		timestamp := time.Now().Format("20060102_150405")
		savePath = filepath.Join(savePath, fmt.Sprintf("test_result_%s.csv", timestamp))
	}

	fmt.Printf("💾 正在保存结果到: %s\n", savePath)

	// 构建CSV数据
	csvData := [][]string{
		{"测试用例ID", "原始请求报文", "响应体", "是否成功", "状态码", "错误信息", "耗时(ms)"},
	}

	for _, result := range results {
		row := []string{
			result.TestCaseID,
			result.RequestBody,
			result.ResponseBody,
			strconv.FormatBool(result.Success),
			strconv.Itoa(result.StatusCode),
			result.Error,
			strconv.FormatInt(result.Duration, 10),
		}
		csvData = append(csvData, row)
	}

	// 保存到CSV文件
	if err := utils.SaveToCSV(csvData, savePath); err != nil {
		return err
	}

	fmt.Printf("✅ 结果已保存到: %s\n", savePath)
	return nil
}

// printResponseDetails 打印响应详细信息（用于debug模式）
func printResponseDetails(testCaseNum int, result models.TestResult) {
	fmt.Printf("📄 测试用例 %d 响应详情:\n", testCaseNum)
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

	// 输出错误信息（如果有）
	if result.Error != "" {
		fmt.Println("│ 错误信息:")
		errorLines := strings.Split(result.Error, "\n")
		for _, line := range errorLines {
			fmt.Printf("│   %s\n", line)
		}
		fmt.Println("│")
	}

	// 输出响应体
	fmt.Println("│ 响应体:")
	if result.ResponseBody == "" {
		fmt.Println("│   (空响应体)")
	} else {
		// 尝试格式化JSON响应体
		var jsonData any
		if err := json.Unmarshal([]byte(result.ResponseBody), &jsonData); err == nil {
			// 如果是有效的JSON，进行格式化输出
			if formattedJSON, err := json.MarshalIndent(jsonData, "│   ", "  "); err == nil {
				// 格式化输出JSON，每行前加上"│   "
				jsonLines := strings.Split(string(formattedJSON), "\n")
				for _, line := range jsonLines {
					fmt.Printf("│   %s\n", line)
				}
			} else {
				// JSON格式化失败，直接输出原始内容
				responseLines := strings.Split(result.ResponseBody, "\n")
				for _, line := range responseLines {
					fmt.Printf("│   %s\n", line)
				}
			}
		} else {
			// 不是JSON格式，直接输出原始内容
			responseLines := strings.Split(result.ResponseBody, "\n")
			for _, line := range responseLines {
				fmt.Printf("│   %s\n", line)
			}
		}
	}

	fmt.Println("└─────────────────────────────────────────────────────────────")
	fmt.Println()
}
