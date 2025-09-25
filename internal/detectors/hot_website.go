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

	// 使用最终域名进行热门网站检测
	finalDomain := ctx.Domain
	if ctx.Result.Network != nil && ctx.Result.Network.FinalDomain != "" {
		finalDomain = ctx.Result.Network.FinalDomain
	}

	// 检测是否为热门网站
	isHotWebsite := hws.detectHotWebsite(finalDomain)

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

	// 1. 精确匹配
	if hws.hotWebsites[domain] {
		return true
	}

	// 2. 通配符匹配
	if hws.matchWildcard(domain) {
		return true
	}

	// 3. www.前缀处理
	if strings.HasPrefix(domain, "www.") {
		domainWithoutWWW := domain[4:]
		// 精确匹配
		if hws.hotWebsites[domainWithoutWWW] {
			return true
		}
		// 通配符匹配
		if hws.matchWildcard(domainWithoutWWW) {
			return true
		}
	} else {
		domainWithWWW := "www." + domain
		// 精确匹配
		if hws.hotWebsites[domainWithWWW] {
			return true
		}
		// 通配符匹配
		if hws.matchWildcard(domainWithWWW) {
			return true
		}
	}

	return false
}

// matchWildcard 通配符匹配
func (hws *HotWebsiteStage) matchWildcard(domain string) bool {
	// 遍历所有热门网站模式
	for pattern := range hws.hotWebsites {
		if strings.HasPrefix(pattern, "*.") {
			// 提取通配符的基础域名
			baseDomain := pattern[2:] // 去掉 "*."

			// 检查域名是否匹配通配符模式
			if hws.isSubdomain(domain, baseDomain) {
				return true
			}
		}
	}
	return false
}

// isSubdomain 检查是否为子域名
func (hws *HotWebsiteStage) isSubdomain(domain, baseDomain string) bool {
	// 1. 直接匹配：domain == baseDomain
	if domain == baseDomain {
		return true
	}

	// 2. 子域名匹配：domain 以 "." + baseDomain 结尾
	suffix := "." + baseDomain
	return strings.HasSuffix(domain, suffix)
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
	return 9 // 热门网站检测在CDN检测之后 - 需要CDN结果
}

// Name 阶段名称
func (hws *HotWebsiteStage) Name() string {
	return "hot_website"
}
