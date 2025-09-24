package cmd

import (
	"fmt"
	"strings"

	"RealityChecker/internal/ui"
)

// executeBatch 执行批量检测
func (r *RootCmd) executeBatch(domainsStr string) {
	// 解析域名列表
	domains := parseDomains(domainsStr)
	if len(domains) == 0 {
		fmt.Println("没有有效的域名")
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

// parseDomains 解析域名列表
func parseDomains(domainsStr string) []string {
	var domains []string
	for _, domain := range strings.Split(domainsStr, ",") {
		domain = strings.TrimSpace(domain)
		if domain != "" {
			domains = append(domains, domain)
		}
	}
	return domains
}
