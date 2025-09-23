// Package models 测试用例模型
package models

// TestCase 表示一个测试用例
type TestCase struct {
	ID          string         `json:"id"`          // 测试用例ID
	Name        string         `json:"name"`        // 测试用例名称
	Description string         `json:"description"` // 测试用例描述
	Type        string         `json:"type"`        // 测试用例类型（正例/反例）
	Data        map[string]any `json:"data"`        // 测试数据
	Expected    map[string]any `json:"expected"`    // 预期结果
}

// TestResult 表示一个测试结果
type TestResult struct {
	TestCaseID   string `json:"test_case_id"`    // 测试用例ID
	Success      bool   `json:"success"`         // 是否成功
	StatusCode   int    `json:"status_code"`     // HTTP状态码
	ResponseBody string `json:"response_body"`   // 响应体
	RequestBody  string `json:"request_body"`    // 原始请求报文
	Error        string `json:"error,omitempty"` // 错误信息（如果有）
	Duration     int64  `json:"duration"`        // 执行时间（毫秒）
}

// TestSuite 表示一组测试用例
type TestSuite struct {
	ID          string     `json:"id"`          // 测试套件ID
	Name        string     `json:"name"`        // 测试套件名称
	Description string     `json:"description"` // 测试套件描述
	TestCases   []TestCase `json:"test_cases"`  // 测试用例列表
}

// TestReport 表示测试报告
type TestReport struct {
	ID        string       `json:"id"`        // 报告ID
	Name      string       `json:"name"`      // 报告名称
	Timestamp int64        `json:"timestamp"` // 时间戳
	Results   []TestResult `json:"results"`   // 测试结果列表
	Summary   struct {
		Total   int `json:"total"`   // 总数
		Success int `json:"success"` // 成功数
		Failed  int `json:"failed"`  // 失败数
	} `json:"summary"` // 摘要
}
