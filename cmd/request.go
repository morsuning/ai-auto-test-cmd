// Package cmd 提供API自动化测试命令行工具的命令实现
package cmd

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/morsuning/ai-auto-test-cmd/models"
	"github.com/morsuning/ai-auto-test-cmd/utils"
	"github.com/spf13/cobra"
)

// requestCmd 表示批量请求目标系统接口的命令
var requestCmd = &cobra.Command{
	Use:   "request",
	Short: "批量请求目标系统接口",
	Long: `通过命令及本地的CSV文件，批量请求目标系统接口，返回执行结果，并且可以保存。

示例：
  # 根据测试用例文件xxx.csv,批量使用POST方法请求目标系统http接口，发送JSON格式数据
  atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --json

  # 根据测试用例文件xxx.csv,批量使用POST方法请求目标系统http接口，发送XML格式数据
  atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --xml

  # 根据测试用例文件xxx.csv,批量使用GET方法请求目标系统http接口，结果默认保存至当前目录
  atc request -u https://xxx.system.com/xxx/xxx -m get -f xxx.csv --json -s

  # 根据测试用例文件xxx.csv,批量使用GET方法请求目标系统http接口，结果保存至指定目录及文件
  atc request -u https://xxx.system.com/xxx/xxx -m get -f xxx.csv --json -s /xxx/tool/result.csv`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取命令行参数
		url, _ := cmd.Flags().GetString("url")
		method, _ := cmd.Flags().GetString("method")
		filePath, _ := cmd.Flags().GetString("file")
		save, _ := cmd.Flags().GetBool("save")
		savePath, _ := cmd.Flags().GetString("save-path")
		timeout, _ := cmd.Flags().GetInt("timeout")
		concurrent, _ := cmd.Flags().GetInt("concurrent")
		debug, _ := cmd.Flags().GetBool("debug")
		
		// 获取请求体格式参数
		isXML, _ := cmd.Flags().GetBool("xml")
		isJSON, _ := cmd.Flags().GetBool("json")
		
		// 验证请求体格式参数
		if !isXML && !isJSON {
			fmt.Println("❌ 错误: 必须指定请求体格式，使用 --xml 或 --json 参数")
			os.Exit(1)
		}
		
		if isXML && isJSON {
			fmt.Println("❌ 错误: 不能同时指定 --xml 和 --json 参数，请只选择一种格式")
			os.Exit(1)
		}
		
		// 验证GET请求的格式约束
		if strings.ToUpper(method) == "GET" && isXML {
			fmt.Println("❌ 错误: GET请求只支持JSON格式，请使用 --json 参数")
			os.Exit(1)
		}
		
		// 确定内容类型
		contentType := "json"
		if isXML {
			contentType = "xml"
		}

		// 打印开始信息
		fmt.Println("=== API 自动化测试命令行工具 - 批量请求 ===")
		fmt.Printf("目标URL: %s\n", url)
		fmt.Printf("请求方法: %s\n", strings.ToUpper(method))
		fmt.Printf("测试用例文件: %s\n", filePath)
		fmt.Printf("内容类型: %s\n", contentType)
		fmt.Printf("并发数: %d\n", concurrent)
		fmt.Printf("请求超时时间: %d秒\n", timeout)
		fmt.Println()

		// 执行批量请求
		if err := executeBatchRequests(url, method, filePath, save, savePath, timeout, concurrent, contentType, debug); err != nil {
			fmt.Printf("❌ 执行失败: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(requestCmd)

	// 定义命令行参数
	requestCmd.Flags().StringP("url", "u", "", "目标URL（必需）")
	requestCmd.Flags().StringP("method", "m", "get", "请求方法（get/post，默认get）")
	requestCmd.Flags().StringP("file", "f", "", "测试用例文件路径（必需）")
	requestCmd.Flags().BoolP("save", "s", false, "是否保存结果")
	requestCmd.Flags().String("save-path", "", "结果保存路径（默认为当前目录下的result.csv）")
	requestCmd.Flags().IntP("timeout", "t", 30, "请求超时时间（秒，默认30）")
	requestCmd.Flags().IntP("concurrent", "c", 1, "并发请求数（默认1）")
	
	// 请求体格式参数（互斥）
	requestCmd.Flags().Bool("xml", false, "使用XML格式发送请求体")
	requestCmd.Flags().Bool("json", false, "使用JSON格式发送请求体")
	
	// 调试参数
	requestCmd.Flags().Bool("debug", false, "启用调试模式，输出详细的请求信息")

	// 标记必需的参数
	requestCmd.MarkFlagRequired("url")
	requestCmd.MarkFlagRequired("file")
}

// executeBatchRequests 执行批量请求
func executeBatchRequests(url, method, filePath string, save bool, savePath string, timeout, concurrent int, contentType string, debug bool) error {
	// 读取CSV文件
	fmt.Println("📖 正在读取测试用例文件...")
	data, err := utils.ReadCSV(filePath)
	if err != nil {
		return fmt.Errorf("读取CSV文件失败: %v", err)
	}

	if len(data) == 0 {
		return fmt.Errorf("CSV文件为空")
	}

	// 解析CSV数据为测试用例
	testCases, err := parseCSVToTestCases(data)
	if err != nil {
		return fmt.Errorf("解析测试用例失败: %v", err)
	}

	fmt.Printf("✅ 成功读取 %d 个测试用例\n\n", len(testCases))

	// 构建HTTP请求
	requests := buildHTTPRequests(testCases, url, method, timeout, contentType)

	// 如果启用调试模式，输出请求详情
	if debug {
		printDebugInfo(requests)
	}

	// 执行批量请求
	fmt.Println("🚀 开始执行批量请求...")
	start := time.Now()
	responses := utils.SendConcurrentRequests(requests, concurrent)
	duration := time.Since(start)

	// 处理响应结果
	results := processResponses(testCases, responses)

	// 显示结果统计
	displayResults(results, duration, debug)

	// 保存结果（如果需要）
	if save {
		if err := saveResults(results, savePath); err != nil {
			return fmt.Errorf("保存结果失败: %v", err)
		}
	}

	return nil
}

// parseCSVToTestCases 将CSV数据解析为测试用例
func parseCSVToTestCases(data [][]string) ([]models.TestCase, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("CSV文件至少需要包含标题行和一行数据")
	}

	headers := data[0]
	testCases := make([]models.TestCase, 0, len(data)-1)

	// 检查是否是XML单列格式（只有一列且列名为XML）
	isXMLFormat := len(headers) == 1 && strings.ToUpper(headers[0]) == "XML"
	// 检查是否是JSON单列格式（只有一列且列名为JSON）
	isJSONFormat := len(headers) == 1 && strings.ToUpper(headers[0]) == "JSON"

	for i, row := range data[1:] {
		if len(row) != len(headers) {
			return nil, fmt.Errorf("第%d行数据列数与标题行不匹配", i+2)
		}

		var testData map[string]interface{}
		
		if isXMLFormat {
			// XML格式：直接使用XML字符串
			testData = map[string]interface{}{
				"_xml_content": row[0], // 使用特殊键存储XML内容
			}
		} else if isJSONFormat {
			// JSON格式：直接使用JSON字符串
			testData = map[string]interface{}{
				"_json_content": row[0], // 使用特殊键存储JSON内容
			}
		} else {
			// 普通格式：构建测试数据
			testData = make(map[string]interface{})
			for j, value := range row {
				testData[headers[j]] = parseValue(value)
			}
		}

		testCase := models.TestCase{
			ID:          fmt.Sprintf("test_%d", i+1),
			Name:        fmt.Sprintf("测试用例_%d", i+1),
			Description: fmt.Sprintf("从CSV第%d行生成的测试用例", i+2),
			Type:        "auto",
			Data:        testData,
		}

		testCases = append(testCases, testCase)
	}

	return testCases, nil
}

// parseValue 解析字符串值为合适的类型
func parseValue(value string) interface{} {
	// 尝试解析为数字
	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal
	}

	// 尝试解析为浮点数
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// 尝试解析为布尔值
	if boolVal, err := strconv.ParseBool(value); err == nil {
		return boolVal
	}

	// 尝试解析为JSON
	var jsonVal interface{}
	if err := json.Unmarshal([]byte(value), &jsonVal); err == nil {
		return jsonVal
	}

	// 默认返回字符串
	return value
}

// buildHTTPRequests 构建HTTP请求
func buildHTTPRequests(testCases []models.TestCase, url, method string, timeout int, contentType string) []utils.HTTPRequest {
	requests := make([]utils.HTTPRequest, len(testCases))

	for i, testCase := range testCases {
		// 构建请求体
		body := ""
		headers := make(map[string]string)

		if strings.ToUpper(method) == "POST" {
			// POST请求，根据contentType格式化数据
			if strings.ToLower(contentType) == "xml" {
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
			} else {
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
			// GET请求，将JSON数据转换为查询参数
			var queryParams []string
			if jsonContent, exists := testCase.Data["_json_content"]; exists {
				// 解析JSON内容为查询参数
				var jsonData map[string]interface{}
				if err := json.Unmarshal([]byte(fmt.Sprintf("%v", jsonContent)), &jsonData); err == nil {
					for key, value := range jsonData {
						queryParams = append(queryParams, fmt.Sprintf("%s=%v", key, value))
					}
				}
			} else {
				// 从字段数据构建查询参数
				for key, value := range testCase.Data {
					queryParams = append(queryParams, fmt.Sprintf("%s=%v", key, value))
				}
			}
			
			// 将查询参数添加到URL
			if len(queryParams) > 0 {
				separator := "?"
				if strings.Contains(url, "?") {
					separator = "&"
				}
				requests[i].URL = url + separator + strings.Join(queryParams, "&")
			}
			headers["Accept"] = "application/json"
		} else {
			// 其他请求方法
			headers["Accept"] = "application/json"
		}

		requests[i] = utils.HTTPRequest{
			URL:     url,
			Method:  strings.ToUpper(method),
			Headers: headers,
			Body:    body,
			Timeout: timeout,
		}
	}

	return requests
}

// processResponses 处理响应结果
func processResponses(testCases []models.TestCase, responses []utils.HTTPResponse) []models.TestResult {
	results := make([]models.TestResult, len(testCases))

	for i, response := range responses {
		result := models.TestResult{
			TestCaseID:   testCases[i].ID,
			StatusCode:   response.StatusCode,
			ResponseBody: response.Body,
			Duration:     response.Duration.Milliseconds(),
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
		{"测试用例ID", "是否成功", "状态码", "响应体", "错误信息", "耗时(ms)"},
	}

	for _, result := range results {
		row := []string{
			result.TestCaseID,
			strconv.FormatBool(result.Success),
			strconv.Itoa(result.StatusCode),
			result.ResponseBody,
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

// convertToXML 将数据转换为XML格式
func convertToXML(data map[string]interface{}) (string, error) {
	// 创建一个包装结构来生成XML
	type XMLData struct {
		XMLName xml.Name               `xml:"data"`
		Fields  map[string]interface{} `xml:"-"`
	}

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
		var jsonData interface{}
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