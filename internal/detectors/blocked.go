package detectors

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"reality-checker-go/internal/types"
)

// BlockedStage 被墙检测阶段
type BlockedStage struct {
	gfwlist map[string]bool
}

// NewBlockedStage 创建被墙检测阶段
func NewBlockedStage() *BlockedStage {
	stage := &BlockedStage{
		gfwlist: make(map[string]bool),
	}
	stage.loadGFWList()
	return stage
}

// Execute 执行被墙检测
func (bs *BlockedStage) Execute(ctx *types.PipelineContext) error {
	// 检查是否被墙
	isBlocked, reason := bs.checkBlocked(ctx.Domain)
	
	ctx.Result.Blocked = &types.BlockedResult{
		IsBlocked:      isBlocked,
		BlockedReasons: []string{reason},
		MatchType:      "gfwlist",
	}

	if isBlocked {
		ctx.EarlyExit = true
		return fmt.Errorf("域名被墙（%s）", reason)
	}

	return nil
}

// checkBlocked 检查是否被墙
func (bs *BlockedStage) checkBlocked(domain string) (bool, string) {
	domain = strings.ToLower(domain)
	
	// 检查GFWList
	if bs.gfwlist[domain] {
		return true, "仅参考黑名单"
	}
	
	// 检查通配符匹配
	parts := strings.Split(domain, ".")
	for i := 0; i < len(parts); i++ {
		wildcard := "*." + strings.Join(parts[i:], ".")
		if bs.gfwlist[wildcard] {
			return true, fmt.Sprintf("通配符匹配: %s", wildcard)
		}
	}
	
	return false, ""
}

// loadGFWList 加载GFWList
func (bs *BlockedStage) loadGFWList() {
	file, err := os.Open("data/gfwlist.conf")
	if err != nil {
		return
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	inPayload := false
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if line == "payload:" {
			inPayload = true
			continue
		}
		
		if !inPayload {
			continue
		}
		
		if strings.HasPrefix(line, "- '") && strings.HasSuffix(line, "'") {
			domain := strings.TrimPrefix(line, "- '")
			domain = strings.TrimSuffix(domain, "'")
			
			if strings.HasPrefix(domain, "+.") {
				domain = strings.TrimPrefix(domain, "+.")
			}
			
			if domain != "" {
				bs.gfwlist[domain] = true
			}
		}
	}
}

// CanEarlyExit 是否可以早期退出
func (bs *BlockedStage) CanEarlyExit() bool {
	return true
}

// Priority 优先级
func (bs *BlockedStage) Priority() int {
	return 1  // 被墙检测优先级最高
}

// Name 阶段名称
func (bs *BlockedStage) Name() string {
	return "blocked"
}
