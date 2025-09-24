package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"RealityChecker/internal/batch"
	"RealityChecker/internal/config"
	"RealityChecker/internal/core"
	"RealityChecker/internal/ui"
	"RealityChecker/internal/version"
)

// RootCmd 根命令结构
type RootCmd struct {
	engine       *core.Engine
	batchManager *batch.Manager
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewRootCmd 创建根命令
func NewRootCmd() (*RootCmd, error) {
	// 加载配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %v", err)
	}

	// 创建引擎
	engine := core.NewEngine(cfg)
	if err := engine.Start(); err != nil {
		return nil, fmt.Errorf("启动引擎失败: %v", err)
	}

	// 创建批量管理器（共享引擎）
	batchManager := batch.NewManagerWithEngine(engine, cfg)
	if err := batchManager.Start(); err != nil {
		engine.Stop()
		return nil, fmt.Errorf("启动批量管理器失败: %v", err)
	}

	// 设置信号处理
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		cancel()
	}()

	return &RootCmd{
		engine:       engine,
		batchManager: batchManager,
		ctx:          ctx,
		cancel:       cancel,
	}, nil
}

// Execute 执行命令
func (r *RootCmd) Execute() {
	defer r.cleanup()

	if len(os.Args) < 2 {
		ui.PrintUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "check":
		if len(os.Args) < 3 {
			fmt.Println("错误：缺少域名参数")
			fmt.Println("用法: reality-checker check <domain>")
			fmt.Println("示例: reality-checker check apple.com")
			os.Exit(1)
		}
		r.executeCheck(os.Args[2])
	case "batch":
		if len(os.Args) < 3 {
			fmt.Println("错误：缺少域名参数")
			fmt.Println("用法: reality-checker batch <domain1> <domain2> <domain3> ...")
			fmt.Println("示例: reality-checker batch apple.com google.com microsoft.com")
			os.Exit(1)
		}
		// 将所有参数（除了命令名）合并为空格分隔的字符串
		domainsStr := strings.Join(os.Args[2:], " ")
		r.executeBatch(domainsStr)
	case "csv":
		if len(os.Args) < 3 {
			fmt.Println("错误：缺少CSV文件参数")
			fmt.Println("用法: reality-checker csv <csv_file>")
			fmt.Println("示例: reality-checker csv domains.csv")
			os.Exit(1)
		}
		r.executeCSV(os.Args[2])
	case "version", "-v", "--version":
		r.showVersion()
	default:
		fmt.Printf("错误：未知命令 '%s'\n", os.Args[1])
		fmt.Println("可用命令: check, batch, csv, version")
		fmt.Println()
		ui.PrintUsage()
		os.Exit(1)
	}
}

// showVersion 显示版本信息
func (r *RootCmd) showVersion() {
	fmt.Printf("Reality协议目标网站检测工具\n")
	fmt.Printf("版本: %s\n", version.GetVersion())
	fmt.Printf("提交: %s\n", version.GetCommit())
	fmt.Printf("构建时间: %s\n", version.GetBuildTime())
	fmt.Printf("GitHub: https://github.com/V2RaySSR/RealityChecker\n")
}

// cleanup 清理资源
func (r *RootCmd) cleanup() {
	if r.batchManager != nil {
		r.batchManager.Stop()
	}
	if r.engine != nil {
		r.engine.Stop()
	}
	if r.cancel != nil {
		r.cancel()
	}
}
