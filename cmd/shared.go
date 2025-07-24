// Package cmd 提供API自动化测试命令行工具的命令实现
package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
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
	IsXML         bool     // 使用XML格式
	IsJSON        bool     // 使用JSON格式
}

// addRequestFlags 为命令添加request相关的参数
func addRequestFlags(cmd *cobra.Command) {
	// 必填参数组（当使用exec时）
	cmd.Flags().String("request-url", "", "执行测试时的目标URL（使用-e参数时必需）")

	// 必填参数组 - 请求体格式（必须选择其一）
	cmd.Flags().Bool("request-xml", false, "执行测试时使用XML格式发送请求体")
	cmd.Flags().Bool("request-json", false, "执行测试时使用JSON格式发送请求体")

	// 请求控制参数组
	cmd.Flags().String("request-method", "post", "执行测试时的请求方法（get/post，默认post）")
	cmd.Flags().Int("request-timeout", 30, "执行测试时的请求超时时间（秒，默认30）")
	cmd.Flags().Int("request-concurrent", 1, "执行测试时的并发请求数（默认1）")

	// 结果保存参数组
	cmd.Flags().Bool("request-save", false, "执行测试时是否保存结果")
	cmd.Flags().String("request-save-path", "", "执行测试时的结果保存路径（默认为当前目录下的result.csv）")

	// 鉴权参数组
	cmd.Flags().String("request-auth-bearer", "", "执行测试时的Bearer Token认证")
	cmd.Flags().String("request-auth-basic", "", "执行测试时的Basic Auth认证，格式：\"username:password\"")
	cmd.Flags().String("request-auth-api-key", "", "执行测试时的API Key认证（通过X-API-Key头）")

	// 自定义HTTP头参数组
	cmd.Flags().StringSlice("request-header", []string{}, "执行测试时的自定义HTTP头，格式：\"Key: Value\"，可多次使用")

	// 调试参数组
	cmd.Flags().Bool("request-debug", false, "执行测试时启用调试模式，输出详细的请求信息")
}

// getRequestParams 从命令行参数中获取request相关参数
func getRequestParams(cmd *cobra.Command) (RequestParams, error) {
	url, _ := cmd.Flags().GetString("request-url")
	method, _ := cmd.Flags().GetString("request-method")
	save, _ := cmd.Flags().GetBool("request-save")
	savePath, _ := cmd.Flags().GetString("request-save-path")
	timeout, _ := cmd.Flags().GetInt("request-timeout")
	concurrent, _ := cmd.Flags().GetInt("request-concurrent")
	debug, _ := cmd.Flags().GetBool("request-debug")
	authBearer, _ := cmd.Flags().GetString("request-auth-bearer")
	authBasic, _ := cmd.Flags().GetString("request-auth-basic")
	authAPIKey, _ := cmd.Flags().GetString("request-auth-api-key")
	customHeaders, _ := cmd.Flags().GetStringSlice("request-header")
	isXML, _ := cmd.Flags().GetBool("request-xml")
	isJSON, _ := cmd.Flags().GetBool("request-json")

	return RequestParams{
		URL:           url,
		Method:        method,
		Save:          save,
		SavePath:      savePath,
		Timeout:       timeout,
		Concurrent:    concurrent,
		Debug:         debug,
		AuthBearer:    authBearer,
		AuthBasic:     authBasic,
		AuthAPIKey:    authAPIKey,
		CustomHeaders: customHeaders,
		IsXML:         isXML,
		IsJSON:        isJSON,
	}, nil
}

// validateRequestParams 验证request参数
func validateRequestParams(params RequestParams) error {
	// 验证URL
	if params.URL == "" {
		return fmt.Errorf("使用 -e 参数时必须指定 --request-url")
	}

	// 验证请求体格式参数
	if !params.IsXML && !params.IsJSON {
		return fmt.Errorf("使用 -e 参数时必须指定请求体格式，使用 --request-xml 或 --request-json 参数")
	}

	if params.IsXML && params.IsJSON {
		return fmt.Errorf("不能同时指定 --request-xml 和 --request-json 参数，请只选择一种格式")
	}

	// 验证GET请求的格式约束
	if strings.ToUpper(params.Method) == "GET" && params.IsXML {
		return fmt.Errorf("GET请求只支持JSON格式，请使用 --request-json 参数")
	}

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
	if err := executeBatchRequestsWithAuth(params.URL, params.Method, outputFile, params.Save, params.SavePath, params.Timeout, params.Concurrent, contentType, params.Debug, authConfig); err != nil {
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
	requests, err := buildHTTPRequestsWithAuth(testCases, params.URL, params.Method, params.Timeout, useJSON, useXML, authConfig)
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