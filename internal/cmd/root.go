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
			fmt.Println("用法: reality-checker check <domain>")
			os.Exit(1)
		}
		r.executeCheck(os.Args[2])
	case "batch":
		if len(os.Args) < 3 {
			fmt.Println("用法: reality-checker batch <domain1,domain2,...>")
			os.Exit(1)
		}
		r.executeBatch(os.Args[2])
	case "csv":
		if len(os.Args) < 3 {
			fmt.Println("用法: reality-checker csv <csv_file>")
			os.Exit(1)
		}
		r.executeCSV(os.Args[2])
	default:
		ui.PrintUsage()
		os.Exit(1)
	}
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
