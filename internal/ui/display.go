package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"RealityChecker/internal/version"
)

// GitHubRelease GitHub发布信息结构
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
}

// PrintUsage 打印使用说明
func PrintUsage() {
	fmt.Printf("Reality协议目标网站检测器 %s\n\n", version.GetVersion())
	fmt.Println("用法:")
	fmt.Println("  reality-checker check <domain>          检测单个域名")
	fmt.Println("  reality-checker batch <domain1> <domain2> <domain3> ...  批量检测域名")
	fmt.Println("  reality-checker csv <csv_file>          从CSV文件批量检测域名")
	fmt.Println("")
	fmt.Println("示例:")
	fmt.Println("  reality-checker check apple.com")
	fmt.Println("  reality-checker batch apple.com tesla.com microsoft.com")
	fmt.Println("  reality-checker csv file.csv")
}

// PrintTimestampedMessage 打印带时间戳的消息
func PrintTimestampedMessage(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	fmt.Printf("[%s] %s\n", timestamp, message)
}

// PrintError 打印错误信息（带空行间距）
func PrintError(message string) {
	fmt.Println()
	fmt.Println(message)
	fmt.Println()
}

// PrintErrorWithDetails 打印错误信息和详细信息
func PrintErrorWithDetails(message string, details ...string) {
	fmt.Println()
	fmt.Println(message)
	for _, detail := range details {
		fmt.Println(detail)
	}
	fmt.Println()
}

// getLatestVersion 获取GitHub最新版本号
func getLatestVersion() string {
	// 设置超时时间
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 请求GitHub API
	resp, err := client.Get("https://api.github.com/repos/V2RaySSR/RealityChecker/releases/latest")
	if err != nil {
		return "" // 网络错误时返回空字符串
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "" // HTTP错误时返回空字符串
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "" // 读取错误时返回空字符串
	}

	// 解析JSON
	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "" // 解析错误时返回空字符串
	}

	// 返回版本号，如果没有tag_name则使用name
	if release.TagName != "" {
		return release.TagName
	}
	if release.Name != "" {
		return release.Name
	}

	return "" // 没有版本信息时返回空字符串
}

// getVersionInfo 获取版本信息字符串
func getVersionInfo() string {
	currentVersion := version.GetVersion() // 使用动态版本号

	latestVersion := getLatestVersion()

	if latestVersion == "" {
		// 无法获取最新版本，只显示当前版本
		return currentVersion
	}

	if latestVersion == currentVersion {
		// 版本相同，只显示当前版本
		return currentVersion
	}

	// 版本不同，显示当前版本和最新版本
	return fmt.Sprintf("%s (最新: %s)", currentVersion, latestVersion)
}

// getDisplayWidth 计算字符串的显示宽度（中文字符占2个位置）
func getDisplayWidth(s string) int {
	width := 0
	for _, r := range s {
		if r < 128 {
			// ASCII字符占1个位置
			width++
		} else {
			// 中文字符占2个位置
			width += 2
		}
	}
	return width
}

// PrintAdvertisement 打印广告信息
func PrintAdvertisement() {
	// 使用颜色代码
	blue := "\033[36m"   // 青色
	yellow := "\033[33m" // 黄色
	white := "\033[37m"  // 白色
	reset := "\033[0m"   // 重置颜色

	fmt.Println()
	fmt.Printf("%s-----------------------------------------------------%s\n", white, reset)
	fmt.Println()
	fmt.Printf("%s %s五年老机场%s %s%shttps://goii.cc/mn%s %s（牧牛云）%s\n",
		white,
		yellow, reset,
		blue, white, reset,
		yellow, reset)
	fmt.Println()
}
