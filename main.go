package main

import (
	"context"
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

// printUsage 打印使用说明
func printUsage() {
	fmt.Println("Reality协议目标网站检测器 v2.0")
	fmt.Println("")
	fmt.Println("用法:")
	fmt.Println("  reality-checker check <domain>          检测单个域名")
	fmt.Println("  reality-checker batch <domain1,domain2,...>  批量检测域名")
	fmt.Println("")
	fmt.Println("示例:")
	fmt.Println("  reality-checker check apple.com")
	fmt.Println("  reality-checker batch apple.com,tesla.com,microsoft.com")
}
