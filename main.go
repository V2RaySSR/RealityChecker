package main

import (
	"fmt"
	"os"

	"RealityChecker/internal/cmd"
	"RealityChecker/internal/data"
)

// 版本信息，由构建时注入
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

// 导出版本信息供其他包使用
func GetVersion() string   { return Version }
func GetCommit() string    { return Commit }
func GetBuildTime() string { return BuildTime }

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


