// Package utils 提供了一系列用于数据处理和测试用例生成的工具函数。
// 包含XML解析、JSON解析以及基于原始数据生成测试用例的功能。
package utils

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DifyChatflowRequest 表示发送给Dify Chatflow API的请求
type DifyChatflowRequest struct {
	Inputs         map[string]any `json:"inputs"`          // 输入参数
	Query          string         `json:"query"`           // 用户输入/提问内容
	ResponseMode   string         `json:"response_mode"`   // 响应模式：streaming 或 blocking
	User           string         `json:"user"`            // 用户标识
	ConversationID string         `json:"conversation_id"` // 会话ID（可选）
}

// DifyStreamEvent 表示Dify流式响应事件
type DifyStreamEvent struct {
	Event          string `json:"event"`
	TaskID         string `json:"task_id"`
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	Answer         string `json:"answer"`          // message事件的文本内容
	CreatedAt      int64  `json:"created_at"`      // 创建时间戳
	Data           any    `json:"data"`            // 其他事件的数据
	Metadata       any    `json:"metadata"`        // message_end事件的元数据
	WorkflowRunID  string `json:"workflow_run_id"` // workflow相关事件的ID
	Audio          string `json:"audio"`           // TTS音频数据
	Status         int    `json:"status"`          // error事件的状态码
	Code           string `json:"code"`            // error事件的错误码
	Message        string `json:"message"`         // error事件的错误消息
}

// GenerateTestCasesWithDify 使用Dify Chatflow API生成测试用例
func GenerateTestCasesWithDify(apiKey, baseURL, query string, inputs map[string]any, format, outputFile string, debug bool) error {
	// 构建请求URL - 使用新的chat-messages端点
	chatflowURL := fmt.Sprintf("%s/chat-messages", strings.TrimSuffix(baseURL, "/"))

	// 构建请求体
	reqBody := DifyChatflowRequest{
		Query:        query,
		Inputs:       inputs,
		ResponseMode: "streaming",
		User:         generateUserID(),
	}

	// 序列化请求体
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", chatflowURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	// Debug模式：显示实际的请求信息
	if debug {
		fmt.Println("\n🔍 ==================== DEBUG: HTTP请求详情 ====================")
		fmt.Printf("📍 请求URL: %s\n", req.URL.String())
		fmt.Printf("🔧 请求方法: %s\n", req.Method)
		fmt.Println("\n📋 请求头 (Headers):")
		for key, values := range req.Header {
			for _, value := range values {
				if key == "Authorization" {
					// 安全遮盖API密钥
					fmt.Printf("   %s: Bearer %s\n", key, maskAPIKey(apiKey))
				} else {
					fmt.Printf("   %s: %s\n", key, value)
				}
			}
		}
		fmt.Println("\n📦 请求体 (Request Body):")
		// 格式化JSON输出，使其更易读
		var prettyJSON bytes.Buffer
		if indentErr := json.Indent(&prettyJSON, jsonData, "", "  "); indentErr == nil {
			fmt.Printf("%s\n", prettyJSON.String())
		} else {
			fmt.Printf("%s\n", string(jsonData))
		}
		fmt.Println("🔍 ============================================================")
	}

	// 发送请求
	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// Debug模式：显示响应信息
	if debug {
		fmt.Println("\n🔍 ==================== DEBUG: HTTP响应详情 ====================")
		fmt.Printf("📊 响应状态码: %d %s\n", resp.StatusCode, resp.Status)
		fmt.Println("\n📋 响应头 (Response Headers):")
		for key, values := range resp.Header {
			for _, value := range values {
				fmt.Printf("   %s: %s\n", key, value)
			}
		}
		fmt.Println("🔍 ============================================================")
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API请求失败，状态码: %d，响应: %s", resp.StatusCode, string(body))
	}

	// 处理流式响应
	return processStreamingResponse(resp.Body, format, outputFile, debug)
}

// processStreamingResponse 处理Dify API的流式响应
func processStreamingResponse(body io.Reader, format, outputFile string, debug bool) error {
	scanner := bufio.NewScanner(body)
	var collectedText strings.Builder
	var errorMsg string

	fmt.Println("📡 开始接收流式数据...")

	for scanner.Scan() {
		line := scanner.Text()
		// 跳过空行和非data行
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		// 提取JSON数据
		jsonData := strings.TrimPrefix(line, "data: ")
		if jsonData == "" {
			continue
		}
		// 解析事件
		var event DifyStreamEvent
		if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
			fmt.Printf("⚠️  解析事件失败: %v\n", err)
			continue
		}
		// Debug模式：显示原始响应数据
		if debug {
			fmt.Printf("\n🔍 [DEBUG] 收到事件: %s\n", event.Event)
			fmt.Printf("原始数据: %s\n", jsonData)
			fmt.Println("----------------------------------------")
		}
		// 处理不同类型的事件
		switch event.Event {
		case "message":
			// LLM返回文本块事件，仅实时输出，不收集文本（避免重复收集）
			if event.Answer != "" {
				fmt.Print(event.Answer) // 实时流式输出文本片段
			}
			if debug {
				fmt.Printf("\n🔍 [DEBUG] Message ID: %s, Conversation ID: %s\n", event.MessageID, event.ConversationID)
			}
		case "message_file":
			// 文件事件，表示有新文件需要展示
			if debug {
				fmt.Printf("\n🔍 [DEBUG] 收到文件事件\n")
				if event.Data != nil {
					if dataBytes, err := json.Marshal(event.Data); err == nil {
						fmt.Printf("文件数据: %s\n", string(dataBytes))
					}
				}
			}
		case "message_end":
			// 消息结束事件，收到此事件则代表流式返回结束
			fmt.Println("\n\n✅ 消息接收完成!")
			if debug {
				fmt.Printf("🔍 [DEBUG] Message ID: %s, Conversation ID: %s\n", event.MessageID, event.ConversationID)
				if event.Metadata != nil {
					if metadataBytes, err := json.Marshal(event.Metadata); err == nil {
						fmt.Printf("元数据: %s\n", string(metadataBytes))
					}
				}
			}
		case "message_replace":
			// 消息内容替换事件
			if event.Answer != "" {
				fmt.Printf("\n🔄 消息内容被替换: %s\n", event.Answer)
				collectedText.Reset() // 清空之前收集的文本
				collectedText.WriteString(event.Answer)
			}
		case "workflow_started":
			// Workflow开始执行
			fmt.Println("🚀 Workflow开始执行...")
			if debug {
				fmt.Printf("🔍 [DEBUG] Workflow Run ID: %s\n", event.WorkflowRunID)
				if event.Data != nil {
					if dataBytes, err := json.Marshal(event.Data); err == nil {
						fmt.Printf("Workflow数据: %s\n", string(dataBytes))
					}
				}
			}
		case "node_started":
			// 节点开始执行
			if nodeData, ok := event.Data.(map[string]any); ok {
				if debug {
					if title, exists := nodeData["title"]; exists {
						fmt.Printf("🔧 节点开始: %v\n", title)
					}
					if inputs, exists := nodeData["inputs"]; exists {
						fmt.Printf("🔧 节点输入: %v\n", inputs)
					}
				}
			}
		case "node_finished":
			// 节点执行完成
			if nodeData, ok := event.Data.(map[string]any); ok {
				if debug {
					if status, statusExists := nodeData["status"]; statusExists {
						if status == "succeeded" {
							fmt.Printf("✅ 节点完成: %v\n", nodeData["title"])
						} else {
							fmt.Printf("❌ 节点失败: %v\n", nodeData["error"])
						}
					}
					if outputs, exists := nodeData["outputs"]; exists {
						fmt.Printf("🔧 节点输出: %v\n", outputs)
					}
				}
			}
		case "workflow_finished":
			// Workflow执行结束
			if workflowData, ok := event.Data.(map[string]any); ok {
				if debug {
					if dataBytes, err := json.Marshal(workflowData); err == nil {
						fmt.Printf("\n🔍 [DEBUG] workflow_finished完整数据: %s\n", string(dataBytes))
					}
				}

				if status, exists := workflowData["status"]; exists {
					if status == "succeeded" {
						fmt.Println("\n🎉 Workflow执行成功!")
					} else {
						fmt.Printf("\n❌ Workflow执行失败: %s\n", status)
						if errorField, exists := workflowData["error"]; exists && errorField != nil {
							errorMsg = errorField.(string)
							fmt.Printf("错误信息: %s\n", errorMsg)
						}
					}
				}

				// 检查outputs字段是否包含结果
				if outputs, exists := workflowData["outputs"]; exists {
					if outputsMap, ok := outputs.(map[string]any); ok {
						// 尝试从outputs中提取文本内容
						for _, value := range outputsMap {
							if valueStr, ok := value.(string); ok {
								// fmt.Printf("\n[输出] %s: \n---\n%s\n---", "最终结果", valueStr)
								// 只有当collectedText为空时才收集文本，避免重复收集
								if collectedText.Len() == 0 {
									collectedText.WriteString(valueStr)
								}
							}
						}
					}
				}
			}
		case "tts_message":
			// TTS音频流事件
			if debug {
				fmt.Printf("\n🔍 [DEBUG] 收到TTS音频数据，长度: %d\n", len(event.Audio))
				fmt.Printf("Message ID: %s, Task ID: %s\n", event.MessageID, event.TaskID)
			}
		case "tts_message_end":
			// TTS音频流结束事件
			if debug {
				fmt.Printf("\n🔍 [DEBUG] TTS音频流结束\n")
				fmt.Printf("Message ID: %s, Task ID: %s\n", event.MessageID, event.TaskID)
			}
		case "error":
			// 流式输出过程中出现的异常
			fmt.Printf("\n❌ 流式响应错误: [%d] %s - %s\n", event.Status, event.Code, event.Message)
			return fmt.Errorf("API返回错误: [%d] %s - %s", event.Status, event.Code, event.Message)
		case "ping":
			// 每10s一次的ping事件，保持连接存活
			if debug {
				fmt.Printf("\n🔍 [DEBUG] 收到ping事件，保持连接存活\n")
			}
		default:
			// Debug模式：输出未处理的事件类型
			if debug {
				fmt.Printf("\n🔍 [DEBUG] 未处理的事件类型: %s\n", event.Event)
				if event.Data != nil {
					if dataBytes, err := json.Marshal(event.Data); err == nil {
						fmt.Printf("事件数据: %s\n", string(dataBytes))
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取流式响应失败: %v", err)
	}

	// 检查是否有错误
	if errorMsg != "" {
		return fmt.Errorf("处理过程中出现错误: %s", errorMsg)
	}

	// 解析生成的测试用例
	generatedText := collectedText.String()
	if generatedText == "" {
		fmt.Println("\n⚠️  当前API未返回测试用例数据，不生成文件")
		return nil
	}

	fmt.Println("\n📝 正在解析生成的测试用例...")

	// 保存测试用例到CSV文件
	return saveTestCasesToCSV(generatedText, format, outputFile)
}

// saveTestCasesToCSV 将生成的测试用例保存为CSV文件
func saveTestCasesToCSV(generatedText, format, outputFile string) error {
	// 确保输出目录存在
	dir := filepath.Dir(outputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 创建CSV文件
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入CSV头部
	if format == "xml" {
		if err := writer.Write([]string{"XML"}); err != nil {
			return fmt.Errorf("写入CSV头部失败: %v", err)
		}
	} else {
		if err := writer.Write([]string{"JSON"}); err != nil {
			return fmt.Errorf("写入CSV头部失败: %v", err)
		}
	}

	// 解析生成的测试用例
	testCases := parseGeneratedTestCases(generatedText, format)
	if len(testCases) == 0 {
		return fmt.Errorf("未能解析出有效的测试用例")
	}

	// 写入测试用例
	for i, testCase := range testCases {
		if err := writer.Write([]string{testCase}); err != nil {
			return fmt.Errorf("写入测试用例 %d 失败: %v", i+1, err)
		}
	}

	fmt.Printf("📊 成功解析并保存 %d 个测试用例\n", len(testCases))
	return nil
}

// maskAPIKey 安全地遮盖API密钥，只显示前4位和后4位
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		// 如果API密钥太短，全部用星号遮盖
		return strings.Repeat("*", len(apiKey))
	}
	// 显示前4位和后4位，中间用星号遮盖
	return apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
}

// generateUserID 生成动态用户标识：当前日期时间+8位随机字符串
func generateUserID() string {
	// 获取当前时间，格式为 YYYYMMDDHHMMSS
	now := time.Now()
	timeStr := now.Format("20060102150405")

	// 生成8位随机字符串
	randomStr := generateRandomString(8)

	return timeStr + randomStr
}

// generateRandomString 生成指定长度的随机字符串（包含字母和数字）
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)

	// 使用crypto/rand生成安全的随机数
	if _, err := rand.Read(b); err != nil {
		// 如果crypto/rand失败，使用时间戳作为后备方案
		fallbackStr := fmt.Sprintf("%d", time.Now().UnixNano())
		if len(fallbackStr) >= length {
			return fallbackStr[:length]
		}
		return fallbackStr + strings.Repeat("0", length-len(fallbackStr))
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b)
}

// parseGeneratedTestCases 解析生成的测试用例文本
// 支持多种格式：
// JSON格式：
//  1. JSON Array格式：[{"data_field": value1}, {"data_field": value2}]
//  2. 连续JSON对象格式：{...} {...} {...}
//
// XML格式：
//  1. 连续XML对象格式，XML对象间通过4个$$$$符号分隔：<root>...</root> <root>...</root> <root>...</root>
//
// 通用格式：
//  1. 传统的逐行格式（向后兼容）
func parseGeneratedTestCases(text, format string) []string {
	var testCases []string
	seenTestCases := make(map[string]bool) // 用于去重

	// JSON格式的智能解析
	if format == "json" {
		// 首先尝试解析JSON Array格式
		if arrayTestCases := parseJSONArrayTestCases(text); len(arrayTestCases) > 0 {
			// 成功解析JSON Array格式，进行去重处理
			for _, testCase := range arrayTestCases {
				if !seenTestCases[testCase] {
					testCases = append(testCases, testCase)
					seenTestCases[testCase] = true
				}
			}
			return testCases
		}

		// 如果JSON Array解析失败，尝试解析连续JSON对象格式
		if consecutiveTestCases := parseConsecutiveJSONObjects(text); len(consecutiveTestCases) > 0 {
			// 成功解析连续JSON对象格式，进行去重处理
			for _, testCase := range consecutiveTestCases {
				if !seenTestCases[testCase] {
					testCases = append(testCases, testCase)
					seenTestCases[testCase] = true
				}
			}
			return testCases
		}
	}

	// XML格式的智能解析
	if format == "xml" {
		// 尝试解析连续XML对象格式
		if consecutiveTestCases := parseConsecutiveXMLObjects(text); len(consecutiveTestCases) > 0 {
			// 成功解析连续XML对象格式，进行去重处理
			for _, testCase := range consecutiveTestCases {
				if !seenTestCases[testCase] {
					testCases = append(testCases, testCase)
					seenTestCases[testCase] = true
				}
			}
			return testCases
		}
	}

	// 如果智能解析都失败，使用传统的逐行解析方式（向后兼容）
	lines := strings.Split(text, "\n")
	inCodeBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过空行
		if line == "" {
			continue
		}

		// 处理markdown代码块标记
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		// 跳过代码块外的内容
		if !inCodeBlock {
			continue
		}

		// 验证格式，只有通过验证的行才添加到结果中
		var isValid bool
		if format == "xml" {
			isValid = ValidateXMLFormat(line) == nil
		} else {
			isValid = ValidateJSONFormat(line) == nil
		}

		// 只有格式验证通过且未重复的行才添加到测试用例列表
		if isValid {
			if !seenTestCases[line] {
				testCases = append(testCases, line)
				seenTestCases[line] = true
			}
		}
	}

	return testCases
}

// parseJSONArrayTestCases 解析JSON Array格式的测试用例
// 输入格式：[{"data_field": value1}, {"data_field": value2}, ...]
// 输出：每个JSON对象的字符串表示
func parseJSONArrayTestCases(text string) []string {
	var testCases []string

	// 提取JSON Array内容（支持markdown代码块包装）
	jsonArrayText := extractJSONArrayFromText(text)
	if jsonArrayText == "" {
		return testCases
	}

	// 解析JSON Array
	var jsonArray []map[string]any
	if err := json.Unmarshal([]byte(jsonArrayText), &jsonArray); err != nil {
		// JSON Array解析失败，返回空结果
		return testCases
	}

	// 将每个JSON对象转换为字符串
	for i, jsonObj := range jsonArray {
		if jsonBytes, err := json.Marshal(jsonObj); err == nil {
			// 验证生成的JSON格式
			jsonStr := string(jsonBytes)
			if ValidateJSONFormat(jsonStr) == nil {
				testCases = append(testCases, jsonStr)
			} else {
				fmt.Printf("⚠️  跳过无效的JSON对象 %d: %s\n", i+1, jsonStr)
			}
		} else {
			fmt.Printf("⚠️  序列化JSON对象 %d 失败: %v\n", i+1, err)
		}
	}

	return testCases
}

// extractJSONArrayFromText 从文本中提取JSON Array内容
// 支持从markdown代码块中提取，也支持直接的JSON Array文本
func extractJSONArrayFromText(text string) string {
	lines := strings.Split(text, "\n")
	inCodeBlock := false
	var jsonLines []string

	// 首先尝试从markdown代码块中提取
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 处理markdown代码块标记
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		// 收集代码块内的内容
		if inCodeBlock && line != "" {
			jsonLines = append(jsonLines, line)
		}
	}

	// 如果从代码块中提取到内容，尝试解析
	if len(jsonLines) > 0 {
		jsonText := strings.Join(jsonLines, "")
		if isValidJSONArray(jsonText) {
			return jsonText
		}
	}

	// 如果代码块解析失败，尝试直接从整个文本中查找JSON Array
	cleanText := strings.TrimSpace(text)
	if isValidJSONArray(cleanText) {
		return cleanText
	}

	// 尝试查找文本中的JSON Array片段
	startIdx := strings.Index(cleanText, "[")
	endIdx := strings.LastIndex(cleanText, "]")
	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		jsonCandidate := cleanText[startIdx : endIdx+1]
		if isValidJSONArray(jsonCandidate) {
			return jsonCandidate
		}
	}

	return ""
}

// isValidJSONArray 检查字符串是否为有效的JSON Array
func isValidJSONArray(text string) bool {
	var jsonArray []any
	return json.Unmarshal([]byte(text), &jsonArray) == nil
}

// parseConsecutiveJSONObjects 解析连续JSON对象格式的测试用例
// 输入格式：{"name": "张三"} {"name": "李四"} ...
// 输出：每个JSON对象的字符串表示
func parseConsecutiveJSONObjects(text string) []string {
	var testCases []string

	// 提取连续JSON对象内容（支持markdown代码块包装）
	jsonObjectsText := extractConsecutiveJSONFromText(text)
	if jsonObjectsText == "" {
		return testCases
	}

	// 解析连续的JSON对象
	jsonObjects := splitConsecutiveJSONObjects(jsonObjectsText)

	// 验证并添加每个JSON对象
	for i, jsonStr := range jsonObjects {
		jsonStr = strings.TrimSpace(jsonStr)
		if jsonStr == "" {
			continue
		}

		// 验证JSON格式
		if ValidateJSONFormat(jsonStr) == nil {
			testCases = append(testCases, jsonStr)
		} else {
			fmt.Printf("⚠️  跳过无效的JSON对象 %d: %s\n", i+1, jsonStr)
		}
	}

	return testCases
}

// extractConsecutiveJSONFromText 从文本中提取连续JSON对象内容
// 支持从markdown代码块中提取，也支持直接的连续JSON对象文本
func extractConsecutiveJSONFromText(text string) string {
	lines := strings.Split(text, "\n")
	inCodeBlock := false
	var jsonLines []string

	// 首先尝试从markdown代码块中提取
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 处理markdown代码块标记
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		// 收集代码块内的内容
		if inCodeBlock && line != "" {
			jsonLines = append(jsonLines, line)
		}
	}

	// 如果从代码块中提取到内容，返回合并的文本
	if len(jsonLines) > 0 {
		jsonText := strings.Join(jsonLines, " ")
		if containsConsecutiveJSONObjects(jsonText) {
			return jsonText
		}
	}

	// 如果代码块解析失败，尝试直接从整个文本中查找连续JSON对象
	cleanText := strings.TrimSpace(text)
	if containsConsecutiveJSONObjects(cleanText) {
		return cleanText
	}

	return ""
}

// splitConsecutiveJSONObjects 分割连续的JSON对象
// 使用简单的大括号匹配来分割JSON对象
func splitConsecutiveJSONObjects(text string) []string {
	var jsonObjects []string
	var currentObject strings.Builder
	braceCount := 0
	inString := false
	escaped := false

	for _, char := range text {
		if escaped {
			escaped = false
			currentObject.WriteRune(char)
			continue
		}

		if char == '\\' {
			escaped = true
			currentObject.WriteRune(char)
			continue
		}

		if char == '"' {
			inString = !inString
			currentObject.WriteRune(char)
			continue
		}

		if !inString {
			if char == '{' {
				braceCount++
				currentObject.WriteRune(char)
			} else if char == '}' {
				braceCount--
				currentObject.WriteRune(char)

				// 当大括号匹配完成时，表示一个JSON对象结束
				if braceCount == 0 {
					jsonObj := strings.TrimSpace(currentObject.String())
					if jsonObj != "" {
						jsonObjects = append(jsonObjects, jsonObj)
					}
					currentObject.Reset()
				}
			} else if braceCount > 0 {
				// 只有在JSON对象内部时才添加字符
				currentObject.WriteRune(char)
			}
			// 忽略JSON对象外部的空白字符
		} else {
			// 在字符串内部，添加所有字符
			currentObject.WriteRune(char)
		}
	}

	return jsonObjects
}

// containsConsecutiveJSONObjects 检查文本是否包含连续的JSON对象
// 简单检查：至少包含两个独立的JSON对象（以}开头的{结尾）
func containsConsecutiveJSONObjects(text string) bool {
	// 移除所有空白字符进行简单检查
	cleanText := strings.ReplaceAll(text, " ", "")
	cleanText = strings.ReplaceAll(cleanText, "\n", "")
	cleanText = strings.ReplaceAll(cleanText, "\t", "")

	// 检查是否包含至少一个完整的JSON对象模式
	// 简单模式：}{ 表示两个连续的JSON对象
	return strings.Contains(cleanText, "}{")
}

// parseConsecutiveXMLObjects 解析连续XML对象格式的测试用例
// 输入格式：<user><name>张三</name></user>$$$$<user><name>李四</name></user>$$$$
// 输出：每个XML对象的字符串表示
func parseConsecutiveXMLObjects(text string) []string {
	var testCases []string

	// 提取连续XML对象内容（支持markdown代码块包装）
	xmlObjectsText := extractConsecutiveXMLFromText(text)
	if xmlObjectsText == "" {
		return testCases
	}

	// 解析使用$$$$分隔的XML对象
	xmlObjects := splitXMLObjectsByDelimiter(xmlObjectsText)

	// 验证并添加每个XML对象
	for i, xmlStr := range xmlObjects {
		xmlStr = strings.TrimSpace(xmlStr)
		if xmlStr == "" {
			continue
		}

		// 验证XML格式
		if ValidateXMLFormat(xmlStr) == nil {
			testCases = append(testCases, xmlStr)
		} else {
			fmt.Printf("⚠️  跳过无效的XML对象 %d: %s\n", i+1, xmlStr)
		}
	}

	return testCases
}

// extractConsecutiveXMLFromText 从文本中提取连续XML对象内容
// 支持从markdown代码块中提取，也支持直接的连续XML对象文本
func extractConsecutiveXMLFromText(text string) string {
	lines := strings.Split(text, "\n")
	inCodeBlock := false
	var xmlLines []string

	// 首先尝试从markdown代码块中提取
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 处理markdown代码块标记
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		// 收集代码块内的内容
		if inCodeBlock && line != "" {
			xmlLines = append(xmlLines, line)
		}
	}

	// 如果从代码块中提取到内容，返回合并的文本
	if len(xmlLines) > 0 {
		xmlText := strings.Join(xmlLines, " ")
		if containsXMLWithDelimiter(xmlText) {
			return xmlText
		}
	}

	// 如果代码块解析失败，尝试直接从整个文本中查找连续XML对象
	cleanText := strings.TrimSpace(text)
	if containsXMLWithDelimiter(cleanText) {
		return cleanText
	}

	return ""
}

// splitXMLObjectsByDelimiter 使用$$$$分隔符分割XML对象
// 输入格式：<xml1>...</xml1>$$$$<xml2>...</xml2>$$$$
// 输出：每个XML对象的字符串数组
func splitXMLObjectsByDelimiter(text string) []string {
	var xmlObjects []string

	// 使用$$$$作为分隔符分割文本
	delimiter := "$$$$"
	// 使用SplitAfter更高效地分割文本
	parts := strings.SplitAfter(text, delimiter)

	// 处理每个分割后的部分
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			xmlObjects = append(xmlObjects, part)
		}
	}

	// 如果没有找到分隔符，但文本包含XML内容，则将整个文本作为单个XML对象
	if len(xmlObjects) == 1 && xmlObjects[0] == strings.TrimSpace(text) {
		// 检查是否真的是单个XML对象
		if strings.Contains(text, "<") && strings.Contains(text, ">") {
			return xmlObjects
		}
		return []string{}
	}

	return xmlObjects
}

// containsXMLWithDelimiter 检查文本是否包含使用$$$$分隔的XML对象
// 或者包含单个XML对象
func containsXMLWithDelimiter(text string) bool {
	// 检查是否包含$$$$分隔符
	if strings.Contains(text, "$$$$") {
		return true
	}

	// 检查是否包含基本的XML结构
	if strings.Contains(text, "<") && strings.Contains(text, ">") {
		return true
	}

	return false
}
