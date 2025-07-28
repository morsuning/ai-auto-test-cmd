package utils

import (
	"net"
	"strings"
	"testing"
)

// TestGenerateUserID 测试基于IP地址的用户ID生成功能
func TestGenerateUserID(t *testing.T) {
	userID := generateUserID()

	// 验证用户ID不为空
	if userID == "" {
		t.Error("用户ID不应为空")
	}

	// 验证用户ID格式
	if !strings.HasPrefix(userID, "ip_") && !strings.HasPrefix(userID, "fallback_") {
		t.Errorf("用户ID格式不正确，期望以'ip_'或'fallback_'开头，实际: %s", userID)
	}

	// 如果是IP格式，验证IP地址的有效性
	if strings.HasPrefix(userID, "ip_") {
		ipPart := strings.TrimPrefix(userID, "ip_")
		// 将下划线替换回点号和冒号
		ipStr := strings.ReplaceAll(ipPart, "_", ".")

		// 尝试解析为IPv4地址
		if ip := net.ParseIP(ipStr); ip == nil {
			// 如果IPv4解析失败，尝试IPv6格式
			ipStr = strings.ReplaceAll(ipPart, "_", ":")
			if ip := net.ParseIP(ipStr); ip == nil {
				t.Errorf("无法解析IP地址: %s", ipStr)
			}
		}
	}

	t.Logf("生成的用户ID: %s", userID)
}

// TestGetLocalIP 测试本机IP地址获取功能
func TestGetLocalIP(t *testing.T) {
	ip := getLocalIP()

	// IP地址可能为空（在某些网络环境下）
	if ip == "" {
		t.Log("未能获取到本机IP地址（这在某些网络环境下是正常的）")
		return
	}

	// 验证IP地址格式
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		t.Errorf("获取的IP地址格式无效: %s", ip)
		return
	}

	// 验证不是回环地址
	if parsedIP.IsLoopback() {
		t.Errorf("获取的IP地址不应是回环地址: %s", ip)
	}

	t.Logf("获取的本机IP地址: %s", ip)

	// 验证IP地址类型
	if parsedIP.To4() != nil {
		t.Logf("IP地址类型: IPv4")
	} else {
		t.Logf("IP地址类型: IPv6")
	}
}

// TestGenerateUserIDConsistency 测试用户ID生成的一致性
func TestGenerateUserIDConsistency(t *testing.T) {
	// 在同一台机器上，基于IP的用户ID应该是一致的
	userID1 := generateUserID()
	userID2 := generateUserID()

	if userID1 != userID2 {
		t.Errorf("在同一台机器上，用户ID应该保持一致。第一次: %s, 第二次: %s", userID1, userID2)
	}

	t.Logf("用户ID一致性测试通过: %s", userID1)
}

// TestGenerateUserIDCrossPlatform 测试跨平台兼容性
func TestGenerateUserIDCrossPlatform(t *testing.T) {
	userID := generateUserID()

	// 验证用户ID中不包含不安全的字符
	unsafeChars := []string{" ", "\t", "\n", "\r", "/", "\\", "?", "*", "<", ">", "|", "\"", ":"}
	for _, char := range unsafeChars {
		if strings.Contains(userID, char) {
			t.Errorf("用户ID包含不安全字符 '%s': %s", char, userID)
		}
	}

	// 验证用户ID长度合理
	if len(userID) > 100 {
		t.Errorf("用户ID过长: %d 字符", len(userID))
	}

	if len(userID) < 3 {
		t.Errorf("用户ID过短: %d 字符", len(userID))
	}

	t.Logf("跨平台兼容性测试通过，用户ID: %s (长度: %d)", userID, len(userID))
}
