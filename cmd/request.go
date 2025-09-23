// Package cmd 提供API自动化测试命令行工具的命令实现
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
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

注意：如果URL未指定协议（http://或https://），系统将自动添加http://前缀。
例如：localhost:8080/user 将被处理为 http://localhost:8080/user

基本示例：
  # 根据测试用例文件xxx.csv,批量使用POST方法请求目标系统http接口，发送JSON格式数据
  atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --json

  # 使用本地服务器（自动添加http://协议）
  atc request -u localhost:8080/api/test -m post -f xxx.csv --json

  # 根据测试用例文件xxx.csv,批量使用POST方法请求目标系统http接口，发送XML格式数据
  atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --xml

  # 根据测试用例文件xxx.csv,批量使用GET方法请求目标系统http接口，结果默认保存至当前目录
  atc request -u https://xxx.system.com/xxx/xxx -m get -f xxx.csv --json -s

  # GET请求支持在body中放置JSON/XML数据，同时可以添加URL查询参数
  atc request -u https://xxx.system.com/xxx/xxx -m get -f xxx.csv --json --query "version=v1" --query "debug=true"

  # 启用调试模式，详细输出每个请求的URL、HTTP头和请求体信息，以及响应详情
  atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --json --debug

配置文件示例：
  # 使用配置文件中的参数
  atc request -c config.toml --json

  # 使用配置文件，命令行参数覆盖配置文件中的设置
  atc request -c config.toml -u https://api.example.com/test --json

鉴权示例：
  # 使用Bearer Token鉴权发送请求
  atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --json --auth-bearer "your_token_here"

  # 使用Basic Auth鉴权发送请求
  atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --json --auth-basic "username:password"

  # 使用API Key鉴权发送请求
  atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --json --auth-api-key "your_api_key"

自定义HTTP头示例：
  # 添加自定义HTTP头发送请求
  atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --json --header "X-API-Key: your_api_key" --header "X-Client-Version: 1.0"

  # 组合使用鉴权和自定义头
  atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --json --auth-bearer "token" --header "X-Request-ID: 12345"

URL查询参数示例：
  # 添加URL查询参数（适用于任何请求方法）
  atc request -u https://xxx.system.com/xxx/xxx -m get -f xxx.csv --json --query "version=v1" --query "debug=true"

  # 组合使用查询参数、鉴权和自定义头
  atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --json --query "api_version=2.0" --auth-bearer "token" --header "X-Request-ID: 12345"`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取配置文件参数
		configFile, _ := cmd.Flags().GetString("config")

		// 获取命令行参数
		url, _ := cmd.Flags().GetString("url")
		method, _ := cmd.Flags().GetString("method")
		filePath, _ := cmd.Flags().GetString("file")
		save, _ := cmd.Flags().GetBool("save")
		savePath, _ := cmd.Flags().GetString("save-path")
		timeout, _ := cmd.Flags().GetInt("timeout")
		concurrent, _ := cmd.Flags().GetInt("concurrent")
		debug, _ := cmd.Flags().GetBool("debug")

		// 获取鉴权参数
		authBearer, _ := cmd.Flags().GetString("auth-bearer")
		authBasic, _ := cmd.Flags().GetString("auth-basic")
		authAPIKey, _ := cmd.Flags().GetString("auth-api-key")
		customHeaders, _ := cmd.Flags().GetStringSlice("header")

		// 获取请求体格式参数
		isXML, _ := cmd.Flags().GetBool("xml")
		isJSON, _ := cmd.Flags().GetBool("json")

		// 获取查询参数
		queryParams, _ := cmd.Flags().GetStringSlice("query")

		// 从配置文件读取参数（如果指定了配置文件）
		if configFile != "" {
			config, err := utils.LoadConfig(configFile)
			if err != nil {
				fmt.Printf("❌ 加载配置文件失败: %v\n", err)
				os.Exit(1)
			}

			// 从配置文件补充缺失的参数
			if url == "" && config.Request.URL != "" {
				url = config.Request.URL
			}
			if method == "get" && config.Request.Method != "" { // 只有当method是默认值时才从配置文件读取
				method = config.Request.Method
			}
			if filePath == "" && config.Request.File != "" {
				filePath = config.Request.File
			}
			if savePath == "" && config.Request.SavePath != "" {
				savePath = config.Request.SavePath
			}
			if timeout == 5 && config.Request.Timeout != 0 { // 只有当timeout是默认值时才从配置文件读取
				timeout = config.Request.Timeout
			}
			if concurrent == 3 && config.Request.Concurrent != 0 { // 只有当concurrent是默认值时才从配置文件读取
				concurrent = config.Request.Concurrent
			}
			if authBearer == "" && config.Request.AuthBearer != "" {
				authBearer = config.Request.AuthBearer
			}
			if authBasic == "" && config.Request.AuthBasic != "" {
				authBasic = config.Request.AuthBasic
			}
			if authAPIKey == "" && config.Request.AuthAPIKey != "" {
				authAPIKey = config.Request.AuthAPIKey
			}
			if len(customHeaders) == 0 && len(config.Request.Headers) > 0 {
				customHeaders = config.Request.Headers
			}
			if len(queryParams) == 0 && len(config.Request.Query) > 0 {
				queryParams = config.Request.Query
			}
		}

		// 验证必需参数
		if url == "" {
			fmt.Println("❌ 错误: 必须指定目标URL（通过 -u 参数或配置文件）")
			os.Exit(1)
		}
		if filePath == "" {
			fmt.Println("❌ 错误: 必须指定测试用例文件路径（通过 -f 参数或配置文件）")
			os.Exit(1)
		}

		// 验证请求体格式参数
		if !isXML && !isJSON {
			fmt.Println("❌ 错误: 必须指定请求体格式，使用 --xml 或 --json 参数")
			os.Exit(1)
		}

		if isXML && isJSON {
			fmt.Println("❌ 错误: 不能同时指定 --xml 和 --json 参数，请只选择一种格式")
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

		// 构建鉴权配置
		authConfig := AuthConfig{
			BearerToken:   authBearer,
			BasicAuth:     authBasic,
			APIKey:        authAPIKey,
			CustomHeaders: customHeaders,
		}

		// 执行批量请求
		if err := executeBatchRequestsWithAuth(url, method, filePath, save, savePath, timeout, concurrent, contentType, debug, authConfig, queryParams); err != nil {
			fmt.Printf("❌ 执行失败: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(requestCmd)

	// 配置文件参数组
	requestCmd.Flags().StringP("config", "c", "", "配置文件路径（可选，默认为config.toml）")

	// 必填参数组
	requestCmd.Flags().StringP("url", "u", "", "目标URL（可选，可从配置文件读取）")
	requestCmd.Flags().StringP("file", "f", "", "测试用例文件路径（可选，可从配置文件读取）")

	// 必填参数组 - 请求体格式（必须选择其一）
	requestCmd.Flags().BoolP("xml", "x", false, "使用XML格式发送请求体")
	requestCmd.Flags().BoolP("json", "j", false, "使用JSON格式发送请求体")

	// 请求控制参数组
	requestCmd.Flags().StringP("method", "m", "get", "请求方法（get/post，默认get，可从配置文件读取）")
	requestCmd.Flags().IntP("timeout", "t", 30, "请求超时时间（秒，默认30，可从配置文件读取）")
	requestCmd.Flags().IntP("concurrent", "C", 3, "并发请求数（默认3，可从配置文件读取）")

	// 结果保存参数组
	requestCmd.Flags().BoolP("save", "s", false, "是否保存结果")
	requestCmd.Flags().String("save-path", "", "结果保存路径（默认为当前目录下的result.csv，可从配置文件读取）")

	// 鉴权参数组
	requestCmd.Flags().String("auth-bearer", "", "Bearer Token认证（可选，可从配置文件读取）")
	requestCmd.Flags().String("auth-basic", "", "Basic Auth认证，格式：\"username:password\"（可选，可从配置文件读取）")
	requestCmd.Flags().String("auth-api-key", "", "API Key认证（通过X-API-Key头）（可选，可从配置文件读取）")

	// 自定义HTTP头参数组
	requestCmd.Flags().StringSlice("header", []string{}, "自定义HTTP头，格式：\"Key: Value\"，可多次使用（可选，可从配置文件读取）")

	// 查询参数组
	requestCmd.Flags().StringSliceP("query", "q", []string{}, "URL查询参数，格式：\"key=value\"，可多次使用（可选，可从配置文件读取）")

	// 调试参数组
	requestCmd.Flags().Bool("debug", false, "启用调试模式，输出详细的请求信息")

	// 自定义参数显示顺序
	requestCmd.Flags().SortFlags = false
}

// executeBatchRequestsWithAuth 执行批量请求（支持鉴权）
func executeBatchRequestsWithAuth(url, method, filePath string, save bool, savePath string, timeout, concurrent int, contentType string, debug bool, authConfig AuthConfig, queryParams []string) error {
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
	useJSON := strings.ToLower(contentType) == "json"
	useXML := strings.ToLower(contentType) == "xml"
	requests, err := buildHTTPRequestsWithAuth(testCases, url, method, timeout, useJSON, useXML, authConfig, queryParams)
	if err != nil {
		return err
	}

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
	results := processResponses(testCases, responses, requests)

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

		var testData map[string]any

		if isXMLFormat {
			// XML格式：直接使用XML字符串
			testData = map[string]any{
				"_xml_content": row[0], // 使用特殊键存储XML内容
			}
		} else if isJSONFormat {
			// JSON格式：直接使用JSON字符串
			testData = map[string]any{
				"_json_content": row[0], // 使用特殊键存储JSON内容
			}
		} else {
			// 普通格式：构建测试数据
			testData = make(map[string]any)
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
func parseValue(value string) any {
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
	var jsonVal any
	if err := json.Unmarshal([]byte(value), &jsonVal); err == nil {
		return jsonVal
	}

	// 默认返回字符串
	return value
}
