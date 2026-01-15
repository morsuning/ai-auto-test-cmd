// Package cmd 提供API自动化测试命令行工具的命令实现
package cmd

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/morsuning/ai-auto-test-cmd/models"
	"github.com/morsuning/ai-auto-test-cmd/utils"
)

// parseDuration 解析时长字符串（支持s/m/h单位）
// 示例: "30s" -> 30秒, "5m" -> 5分钟, "1h" -> 1小时
func parseDuration(durationStr string) (time.Duration, error) {
	if durationStr == "" {
		return 0, fmt.Errorf("时长不能为空")
	}

	// 使用正则表达式解析时长字符串
	// 格式: 数字 + 单位(s/m/h)
	re := regexp.MustCompile(`^(\d+)([smh])$`)
	matches := re.FindStringSubmatch(durationStr)

	if len(matches) != 3 {
		return 0, fmt.Errorf("无效的时长格式: %s，支持的格式示例: 30s, 5m, 1h", durationStr)
	}

	// 解析数字部分
	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("无效的数字: %s", matches[1])
	}

	if value <= 0 {
		return 0, fmt.Errorf("时长必须大于0")
	}

	// 根据单位转换为time.Duration
	unit := matches[2]
	switch unit {
	case "s":
		return time.Duration(value) * time.Second, nil
	case "m":
		return time.Duration(value) * time.Minute, nil
	case "h":
		return time.Duration(value) * time.Hour, nil
	default:
		return 0, fmt.Errorf("不支持的时间单位: %s，支持的单位: s(秒), m(分钟), h(小时)", unit)
	}
}

// BenchmarkStats 压测统计信息
type BenchmarkStats struct {
	totalRequests   int64      // 总请求数
	successRequests int64      // 成功请求数
	failedRequests  int64      // 失败请求数
	responseTimes   []int64    // 响应时间列表(毫秒)
	mu              sync.Mutex // 保护响应时间列表
	startTime       time.Time  // 压测开始时间
}

// recordRequest 记录一次请求结果
func (s *BenchmarkStats) recordRequest(success bool, duration int64) {
	atomic.AddInt64(&s.totalRequests, 1)
	if success {
		atomic.AddInt64(&s.successRequests, 1)
	} else {
		atomic.AddInt64(&s.failedRequests, 1)
	}

	s.mu.Lock()
	s.responseTimes = append(s.responseTimes, duration)
	s.mu.Unlock()
}

// getStats 获取统计结果
func (s *BenchmarkStats) getStats() (total, success, failed int64, avgRT, minRT, maxRT, p90, p95, p99 int64) {
	total = atomic.LoadInt64(&s.totalRequests)
	success = atomic.LoadInt64(&s.successRequests)
	failed = atomic.LoadInt64(&s.failedRequests)

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.responseTimes) == 0 {
		return
	}

	// 复制并排序响应时间
	times := make([]int64, len(s.responseTimes))
	copy(times, s.responseTimes)
	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })

	// 计算最小、最大响应时间
	minRT = times[0]
	maxRT = times[len(times)-1]

	// 计算平均响应时间
	var sum int64
	for _, t := range times {
		sum += t
	}
	avgRT = sum / int64(len(times))

	// 计算百分位数
	p90 = percentile(times, 90)
	p95 = percentile(times, 95)
	p99 = percentile(times, 99)

	return
}

// percentile 计算百分位数
func percentile(sortedTimes []int64, p float64) int64 {
	if len(sortedTimes) == 0 {
		return 0
	}
	index := int(math.Ceil(float64(len(sortedTimes))*p/100.0)) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sortedTimes) {
		index = len(sortedTimes) - 1
	}
	return sortedTimes[index]
}

// executeBenchmark 执行压测
func executeBenchmark(testCases []models.TestCase, url, method string, duration time.Duration, concurrent int, contentType string, authConfig AuthConfig, queryParams []string, ignoreTLS bool) error {
	fmt.Println("\n🎯 === 压测模式 ===")
	fmt.Printf("压测时长: %v\n", duration)
	fmt.Printf("并发数: %d\n", concurrent)
	fmt.Printf("测试用例数: %d\n", len(testCases))
	fmt.Printf("目标URL: %s\n", url)
	fmt.Printf("请求方法: %s\n\n", method)

	// 构建HTTP请求列表
	useJSON := contentType == "json"
	useXML := contentType == "xml"
	requests, err := buildHTTPRequestsWithAuth(testCases, url, method, 30, useJSON, useXML, authConfig, queryParams, ignoreTLS)
	if err != nil {
		return fmt.Errorf("构建请求失败: %v", err)
	}

	// 初始化统计信息
	stats := &BenchmarkStats{
		startTime: time.Now(),
	}

	// 创建上下文，用于控制压测时长
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// 启动压测工作协程
	fmt.Println("🚀 开始压测...")
	startBenchmarkWorkers(ctx, requests, concurrent, stats)

	// 等待压测完成
	<-ctx.Done()

	// 输出压测报告
	printBenchmarkReport(stats, duration)

	return nil
}

// startBenchmarkWorkers 启动压测工作协程
func startBenchmarkWorkers(ctx context.Context, requests []utils.HTTPRequest, concurrent int, stats *BenchmarkStats) {
	var wg sync.WaitGroup

	// 创建请求通道
	requestChan := make(chan utils.HTTPRequest, concurrent*2)

	// 启动工作协程
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case req, ok := <-requestChan:
					if !ok {
						return
					}
					// 发送请求并记录结果
					resp := utils.SendRequest(req)
					success := resp.Error == nil && resp.StatusCode >= 200 && resp.StatusCode < 300
					stats.recordRequest(success, resp.Duration.Milliseconds())
				}
			}
		}()
	}

	// 生产请求
	go func() {
		defer close(requestChan)
		requestIndex := 0
		for {
			select {
			case <-ctx.Done():
				return
			case requestChan <- requests[requestIndex]:
				requestIndex = (requestIndex + 1) % len(requests)
			}
		}
	}()

	// 等待所有工作协程结束
	wg.Wait()
}

// printBenchmarkReport 打印压测报告
func printBenchmarkReport(stats *BenchmarkStats, duration time.Duration) {
	total, success, failed, avgRT, minRT, maxRT, p90, p95, p99 := stats.getStats()

	fmt.Println("\n📊 === 压测报告 ===")
	fmt.Println(strings.Repeat("=", 65))

	// 基本统计
	fmt.Println("\n【请求统计】")
	fmt.Printf("  总请求数:   %d\n", total)
	fmt.Printf("  成功请求:   %d\n", success)
	fmt.Printf("  失败请求:   %d\n", failed)
	if total > 0 {
		successRate := float64(success) / float64(total) * 100
		fmt.Printf("  成功率:     %.2f%%\n", successRate)
	}

	// 性能指标
	fmt.Println("\n【性能指标】")
	actualDuration := time.Since(stats.startTime)
	qps := float64(total) / actualDuration.Seconds()
	fmt.Printf("  QPS:        %.2f 请求/秒\n", qps)
	fmt.Printf("  总耗时:     %v\n", actualDuration.Round(time.Millisecond))
	fmt.Printf("  预期时长:   %v\n", duration)

	// 响应时间统计
	if total > 0 {
		fmt.Println("\n【响应时间】")
		fmt.Printf("  平均:       %d ms\n", avgRT)
		fmt.Printf("  最小:       %d ms\n", minRT)
		fmt.Printf("  最大:       %d ms\n", maxRT)
		fmt.Printf("  P90:        %d ms\n", p90)
		fmt.Printf("  P95:        %d ms\n", p95)
		fmt.Printf("  P99:        %d ms\n", p99)
	}

	fmt.Println("\n" + strings.Repeat("=", 65))
}
