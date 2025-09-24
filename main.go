package main

import (
	"fmt"
	"os"

	"RealityChecker/internal/cmd"
	"RealityChecker/internal/data"
)

func main() {
	// 检查并下载必要的数据文件
	downloader := data.NewDownloader()
	if err := downloader.EnsureDataFiles(); err != nil {
		fmt.Printf("数据文件检查失败: %v\n", err)
		os.Exit(1)
	}

	// 创建根命令
	rootCmd, err := cmd.NewRootCmd()
	if err != nil {
		fmt.Printf("初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 执行命令
	rootCmd.Execute()
}


