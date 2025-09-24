package cmd

import (
	"fmt"

	"RealityChecker/internal/config"
	"RealityChecker/internal/report"
	"RealityChecker/internal/ui"
)

// executeCheck 执行单域名检测
func (r *RootCmd) executeCheck(domain string) {
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
