package detectors

import (
	"bufio"
	"os"
	"strings"

	"RealityChecker/internal/types"
)

// HotWebsiteStage 热门网站检测阶段
type HotWebsiteStage struct {
	hotWebsites map[string]bool
}

// NewHotWebsiteStage 创建热门网站检测阶段
func NewHotWebsiteStage() *HotWebsiteStage {
	stage := &HotWebsiteStage{
		hotWebsites: make(map[string]bool),
	}
	stage.loadHotWebsites()
	return stage
}

// Execute 执行热门网站检测
func (hws *HotWebsiteStage) Execute(ctx *types.PipelineContext) error {
	// 检测是否为热门网站
	isHotWebsite := hws.detectHotWebsite(ctx.Domain)
	
	// 更新CDN结果
	if ctx.Result.CDN != nil {
		ctx.Result.CDN.IsHotWebsite = isHotWebsite
	} else {
		ctx.Result.CDN = &types.CDNResult{
			IsHotWebsite: isHotWebsite,
		}
	}

	// 热门网站检测只是信息性的，不影响适合性判断
	// 热门网站只是建议不推荐，但不是硬性要求

	return nil
}

// detectHotWebsite 检测热门网站
func (hws *HotWebsiteStage) detectHotWebsite(domain string) bool {
	domain = strings.ToLower(domain)
	
	// 检查是否为热门网站
	if hws.hotWebsites[domain] {
		return true
	}
	
	// 如果域名以www.开头，也检查去掉www.的版本
	if strings.HasPrefix(domain, "www.") {
		domainWithoutWWW := domain[4:]
		return hws.hotWebsites[domainWithoutWWW]
	}
	
	// 如果域名不以www.开头，也检查加上www.的版本
	domainWithWWW := "www." + domain
	return hws.hotWebsites[domainWithWWW]
}

// loadHotWebsites 加载热门网站列表
func (hws *HotWebsiteStage) loadHotWebsites() {
	file, err := os.Open("data/hot_websites.txt")
	if err != nil {
		return
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			hws.hotWebsites[strings.ToLower(line)] = true
		}
	}
}

// CanEarlyExit 是否可以早期退出
func (hws *HotWebsiteStage) CanEarlyExit() bool {
	return false
}

// Priority 优先级
func (hws *HotWebsiteStage) Priority() int {
	return 15  // 热门网站检测最后优先级 - 负面检测
}

// Name 阶段名称
func (hws *HotWebsiteStage) Name() string {
	return "hot_website"
}
