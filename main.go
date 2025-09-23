package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"reality-checker-go/internal/batch"
	"reality-checker-go/internal/config"
	"reality-checker-go/internal/core"
	"reality-checker-go/internal/report"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 创建引擎
	engine := core.NewEngine(cfg)
	if err := engine.Start(); err != nil {
		fmt.Printf("启动引擎失败: %v\n", err)
		os.Exit(1)
	}
	defer engine.Stop()

	// 创建批量管理器（共享引擎）
	batchManager := batch.NewManagerWithEngine(engine, cfg)
	if err := batchManager.Start(); err != nil {
		fmt.Printf("启动批量管理器失败: %v\n", err)
		os.Exit(1)
	}
	defer batchManager.Stop()

	// 处理命令行参数
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// 设置信号处理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		cancel()
	}()

	// 执行命令
	switch os.Args[1] {
	case "check":
		if len(os.Args) < 3 {
			fmt.Println("用法: reality-checker check <domain>")
			os.Exit(1)
		}
		checkSingle(ctx, engine, os.Args[2])
	case "batch":
		if len(os.Args) < 3 {
			fmt.Println("用法: reality-checker batch <domain1,domain2,...>")
			os.Exit(1)
		}
		checkBatch(ctx, batchManager, os.Args[2])
	case "csv":
		if len(os.Args) < 3 {
			fmt.Println("用法: reality-checker csv <csv_file>")
			os.Exit(1)
		}
		checkCSV(ctx, batchManager, os.Args[2])
	default:
		printUsage()
		os.Exit(1)
	}
}

// checkSingle 检测单个域名
func checkSingle(ctx context.Context, engine *core.Engine, domain string) {
	fmt.Printf("开始检测域名: %s\n", domain)
	
	result, err := engine.CheckDomain(ctx, domain)
	if err != nil {
		fmt.Printf("检测失败: %v\n", err)
		return
	}

	// 使用格式化器输出结果
	cfg, _ := config.LoadConfig("")
	formatter := report.NewFormatter(cfg)
	fmt.Printf("\n%s", formatter.FormatSingleResult(result))
}

// checkBatch 批量检测域名
func checkBatch(ctx context.Context, batchManager *batch.Manager, domainsStr string) {
	// 解析域名列表
	domains := parseDomains(domainsStr)
	if len(domains) == 0 {
		fmt.Println("没有有效的域名")
		return
	}

	fmt.Printf("开始批量检测 %d 个域名...\n", len(domains))
	
	_, err := batchManager.CheckDomains(ctx, domains)
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

// checkCSV 从CSV文件批量检测域名
func checkCSV(ctx context.Context, batchManager *batch.Manager, csvFile string) {
	// 读取CSV文件
	file, err := os.Open(csvFile)
	if err != nil {
		fmt.Printf("无法打开CSV文件: %v\n", err)
		return
	}
	defer file.Close()

	// 解析CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("解析CSV文件失败: %v\n", err)
		return
	}

	if len(records) < 2 {
		fmt.Println("CSV文件格式错误或为空")
		return
	}

	// 提取域名（从CERT_DOMAIN列）
	domains := extractDomainsFromCSV(records)
	if len(domains) == 0 {
		fmt.Println("未找到有效的域名")
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
	fmt.Println("开始批量检测...")
	
	_, err = batchManager.CheckDomains(ctx, domains)
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

// printUsage 打印使用说明
func printUsage() {
	fmt.Println("Reality协议目标网站检测器 v2.0")
	fmt.Println("")
	fmt.Println("用法:")
	fmt.Println("  reality-checker check <domain>          检测单个域名")
	fmt.Println("  reality-checker batch <domain1,domain2,...>  批量检测域名")
	fmt.Println("  reality-checker csv <csv_file>          从CSV文件批量检测域名")
	fmt.Println("")
	fmt.Println("示例:")
	fmt.Println("  reality-checker check apple.com")
	fmt.Println("  reality-checker batch apple.com,tesla.com,microsoft.com")
	fmt.Println("  reality-checker csv file.csv")
}
