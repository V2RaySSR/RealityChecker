package ui

import (
	"fmt"
	"strings"
)

// PrintBanner 打印程序横幅
func PrintBanner() {

	// 获取版本信息
	versionInfo := getVersionInfo()

	// 使用颜色代码
	white := "\033[37m"
	cyan := "\033[36m"  // 青色用于网站信息
	reset := "\033[0m"

	// 计算版本信息长度，确保居中对齐
	versionText := fmt.Sprintf("Reality协议目标网站检测工具 %s", versionInfo)
	websiteText := "V2RaySSR 综合网"

	// 横幅宽度（与表格保持一致）
	width := 95

	// 计算字符显示宽度（中文字符占2个位置）
	versionDisplayWidth := getDisplayWidth(versionText)
	websiteDisplayWidth := getDisplayWidth(websiteText)

	// 计算居中位置
	versionPadding := (width - 2 - versionDisplayWidth) / 2
	websitePadding := (width - 2 - websiteDisplayWidth) / 2

	// 确保padding不为负数
	if versionPadding < 0 { versionPadding = 0 }
	if websitePadding < 0 { websitePadding = 0 }

	// 计算右侧剩余空间
	versionRightSpace := width - 2 - versionPadding - versionDisplayWidth
	websiteRightSpace := width - 2 - websitePadding - websiteDisplayWidth

	if versionRightSpace < 0 { versionRightSpace = 0 }
	if websiteRightSpace < 0 { websiteRightSpace = 0 }

	fmt.Printf("%s╔%s╗%s\n", white, strings.Repeat("═", width-2), reset)
	fmt.Printf("%s║%s%s%s║%s\n", white, strings.Repeat(" ", versionPadding), versionText, strings.Repeat(" ", versionRightSpace), reset)
	fmt.Printf("%s║%s║%s\n", white, strings.Repeat(" ", width-2), reset)
	fmt.Printf("%s║%s%s%s%s%s║%s\n", white, strings.Repeat(" ", websitePadding), cyan, websiteText, reset, strings.Repeat(" ", websiteRightSpace), reset)
	fmt.Printf("%s╚%s╝%s\n", white, strings.Repeat("═", width-2), reset)
	fmt.Println("")
}
