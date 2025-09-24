package cmd

import (
	"fmt"
	"strings"

	"RealityChecker/internal/ui"
)

// executeBatch 执行批量检测
func (r *RootCmd) executeBatch(domainsStr string) {
	// 解析域名列表
	domains, invalidDomains := parseDomains(domainsStr)
	
	if len(domains) == 0 {
		fmt.Println()
		fmt.Println("错误：没有有效的域名可以检测")
		fmt.Println("提示：请检查域名格式，例如：apple.com, google.com")
		fmt.Println()
		return
	}

	ui.PrintBanner()
	
	// 显示无效域名警告（在横幅下面）
	if len(invalidDomains) > 0 {
		fmt.Printf("警告：发现 %d 个无效域名，已跳过：\n", len(invalidDomains))
		
		// 只显示前5个无效域名，避免显示过多
		displayCount := 5
		if len(invalidDomains) < displayCount {
			displayCount = len(invalidDomains)
		}
		
		for i := 0; i < displayCount; i++ {
			fmt.Printf("   - %s\n", invalidDomains[i])
		}
		
		// 如果还有更多无效域名，显示省略提示
		if len(invalidDomains) > displayCount {
			fmt.Printf("   ... 还有 %d 个无效域名\n", len(invalidDomains)-displayCount)
		}
		
		fmt.Println()
	}
	
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

