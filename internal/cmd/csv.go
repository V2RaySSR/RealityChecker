package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"RealityChecker/internal/ui"
)

// executeCSV 从CSV文件批量检测域名
func (r *RootCmd) executeCSV(csvFile string) {
	// 检查文件是否存在
	if _, err := os.Stat(csvFile); os.IsNotExist(err) {
		fmt.Println()
		fmt.Printf("错误：CSV文件不存在 '%s'\n", csvFile)
		fmt.Println("请使用 RealiTLScanner 工具扫描，得到 CSV 文件")
		fmt.Println("命令：./RealiTLScanner -addr <VPS IP> -port 443 -thread 50 -timeout 5 -out file.csv")
		fmt.Println("（提示：RealiTLScanner 不要在VPS上面运行）")
		fmt.Println()
		return
	}
	
	// 读取CSV文件
	file, err := os.Open(csvFile)
	if err != nil {
		fmt.Println()
		fmt.Printf("错误：无法打开CSV文件 '%s': %v\n", csvFile, err)
		fmt.Println()
		return
	}
	defer file.Close()

	// 解析CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println()
		fmt.Printf("错误：解析CSV文件失败: %v\n", err)
		fmt.Println("请使用 RealiTLScanner 工具扫描，得到 CSV 文件")
		fmt.Println("命令：./RealiTLScanner -addr <VPS IP> -port 443 -thread 50 -timeout 5 -out file.csv")
		fmt.Println()
		return
	}

	if len(records) < 2 {
		fmt.Println()
		fmt.Println("错误：CSV文件格式错误或为空")
		fmt.Println("请使用 RealiTLScanner 工具扫描，得到 CSV 文件")
		fmt.Println("命令：./RealiTLScanner -addr <VPS IP> -port 443 -thread 50 -timeout 5 -out file.csv")
		fmt.Println()
		return
	}

	// 提取域名（从CERT_DOMAIN列）
	domains := extractDomainsFromCSV(records)
	if len(domains) == 0 {
		fmt.Println()
		fmt.Println("错误：未找到有效的域名")
		fmt.Println("请使用 RealiTLScanner 工具扫描，得到 CSV 文件")
		fmt.Println("命令：./RealiTLScanner -addr <VPS IP> -port 443 -thread 50 -timeout 5 -out file.csv")
		fmt.Println()
		return
	}

	fmt.Printf("从CSV文件提取到 %d 个域名:\n", len(domains))
	// 显示前10个域名作为预览
	previewCount := 10
	if len(domains) < previewCount {
		previewCount = len(domains)
	}
	for i := 0; i < previewCount; i++ {
		fmt.Printf("  %d. %s\n", i+1, domains[i])
	}
	if len(domains) > previewCount {
		fmt.Printf("  ... 还有 %d 个域名\n", len(domains)-previewCount)
	}
	fmt.Println("")
	ui.PrintTimestampedMessage("开始批量检测...")
	
	_, err = r.batchManager.CheckDomains(r.ctx, domains)
	if err != nil {
		fmt.Printf("批量检测失败: %v\n", err)
		return
	}
}

// extractDomainsFromCSV 从CSV记录中提取域名
func extractDomainsFromCSV(records [][]string) []string {
	var domains []string
	domainSet := make(map[string]bool) // 用于去重
	
	// 跳过标题行，从第二行开始处理
	for i := 1; i < len(records); i++ {
		if len(records[i]) < 3 {
			continue
		}
		
		certDomain := strings.TrimSpace(records[i][2]) // CERT_DOMAIN列
		if certDomain == "" {
			continue
		}
		
		// 清理域名（移除引号等）
		certDomain = strings.Trim(certDomain, "\"")
		
		// 排除一些不需要的域名
		if shouldExcludeDomain(certDomain) {
			continue
		}
		
		// 去重
		if !domainSet[certDomain] {
			domains = append(domains, certDomain)
			domainSet[certDomain] = true
		}
	}
	
	return domains
}

// shouldExcludeDomain 判断是否应该排除某个域名
func shouldExcludeDomain(domain string) bool {
	// 1. 排除包含通配符(*)的域名
	if strings.Contains(domain, "*") {
		return true
	}
	
	// 2. 排除列表
	excludePatterns := []string{
		"localhost",
		"server.domain.com",
		"johnnasmalley.hostname",
		"Kubernetes Ingress Controller Fake Certificate",
		"CloudFlare Origin Certificate",
		"FortiGate",
		"Unspecified",
	}
	
	domainLower := strings.ToLower(domain)
	
	for _, pattern := range excludePatterns {
		if strings.Contains(domainLower, strings.ToLower(pattern)) {
			return true
		}
	}
	
	// 3. 排除IP地址格式
	if strings.Contains(domain, ".") && !strings.Contains(domain, "..") {
		parts := strings.Split(domain, ".")
		if len(parts) == 4 {
			// 可能是IP地址，简单检查
			isIP := true
			for _, part := range parts {
				if len(part) > 3 {
					isIP = false
					break
				}
			}
			if isIP {
				return true // 是IP地址，排除
			}
		}
	}
	
	// 4. 排除无效域名（太短或包含特殊字符）
	if len(domain) < 3 {
		return true
	}
	
	// 5. 排除包含多个连续点的域名
	if strings.Contains(domain, "..") {
		return true
	}
	
	return false
}
