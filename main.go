package main

import (
	"fmt"
	"os"

	"RealityChecker/internal/cmd"
)

func main() {
	// 创建根命令
	rootCmd, err := cmd.NewRootCmd()
	if err != nil {
		fmt.Printf("初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 执行命令
	rootCmd.Execute()
}


