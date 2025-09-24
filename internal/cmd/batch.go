package cmd

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"RealityChecker/internal/ui"
)

// executeBatch 执行批量检测
func (r *RootCmd) executeBatch(domainsStr string) {
	// 解析域名列表
	domains, invalidDomains := parseDomains(domainsStr)
	
	// 显示无效域名警告
	if len(invalidDomains) > 0 {
		fmt.Printf("警告：发现 %d 个无效域名，已跳过：\n", len(invalidDomains))
		for _, domain := range invalidDomains {
			fmt.Printf("   - %s\n", domain)
		}
		fmt.Println()
	}
	
	if len(domains) == 0 {
		fmt.Println("错误：没有有效的域名可以检测")
		fmt.Println("提示：请检查域名格式，例如：apple.com, google.com")
		return
	}

	ui.PrintBanner()
	ui.PrintTimestampedMessage("开始批量检测 %d 个域名...", len(domains))
	
	_, err := r.batchManager.CheckDomains(r.ctx, domains)
	if err != nil {
		fmt.Printf("批量检测失败: %v\n", err)
		return
	}

	// 详细结果已在batch manager中打印，无需重复
}

// parseDomains 解析域名列表，返回有效域名和无效域名
func parseDomains(domainsStr string) ([]string, []string) {
	var validDomains []string
	var invalidDomains []string
	
	// 支持空格分隔的域名列表
	fields := strings.Fields(domainsStr)
	for _, domain := range fields {
		domain = strings.TrimSpace(domain)
		if domain == "" {
			continue
		}
		
		if isValidDomain(domain) {
			validDomains = append(validDomains, domain)
		} else {
			invalidDomains = append(invalidDomains, domain)
		}
	}
	return validDomains, invalidDomains
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
