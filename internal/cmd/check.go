package cmd

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"RealityChecker/internal/config"
	"RealityChecker/internal/report"
	"RealityChecker/internal/ui"
)

// executeCheck 执行单域名检测
func (r *RootCmd) executeCheck(domain string) {
	// 验证域名格式
	domain = strings.TrimSpace(domain)
	if !isValidDomain(domain) {
		fmt.Println()
		fmt.Printf("错误：域名格式无效 '%s'\n", domain)
		fmt.Println("提示：请检查域名格式，例如：apple.com, google.com")
		fmt.Println("域名要求：")
		fmt.Println("   - 只能包含字母、数字、连字符和点")
		fmt.Println("   - 不能以点开头或结尾")
		fmt.Println("   - 不能包含连续的点")
		fmt.Println("   - 长度不超过253个字符")
		fmt.Println()
		return
	}
	
	ui.PrintBanner()
	ui.PrintTimestampedMessage("开始检测域名: %s", domain)
	
	result, err := r.engine.CheckDomain(r.ctx, domain)
	if err != nil {
		fmt.Printf("检测失败: %v\n", err)
		return
	}

	// 使用格式化器输出结果
	cfg, _ := config.LoadConfig("")
	formatter := report.NewFormatter(cfg)
	fmt.Printf("\n%s", formatter.FormatSingleResult(result))
}

// isValidDomain 验证域名格式是否有效
func isValidDomain(domain string) bool {
	// 基本长度检查
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}
	
	// 检查是否包含非法字符
	if strings.ContainsAny(domain, " \t\n\r") {
		return false
	}
	
	// 检查是否以点开头或结尾
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}
	
	// 检查是否包含连续的点
	if strings.Contains(domain, "..") {
		return false
	}
	
	// 使用正则表达式验证域名格式
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if !domainRegex.MatchString(domain) {
		return false
	}
	
	// 尝试解析域名（不进行实际DNS查询）
	_, err := net.LookupHost(domain)
	if err != nil {
		// 即使DNS解析失败，只要格式正确就认为是有效的
		// 因为可能是网络问题或域名确实不存在
		return true
	}
	
	return true
}

