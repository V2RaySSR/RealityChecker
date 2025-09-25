package batch

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"RealityChecker/internal/core"
	"RealityChecker/internal/report"
	"RealityChecker/internal/types"
)

// Manager 批量管理器
type Manager struct {
	engine         *core.Engine
	formatter      *report.Formatter
	tableFormatter *report.TableFormatter
	config         *types.Config
	mu             sync.RWMutex
	running        bool
}

// NewManager 创建批量管理器
func NewManager(config *types.Config) *Manager {
	return &Manager{
		config:         config,
		formatter:      report.NewFormatter(config),
		tableFormatter: report.NewTableFormatter(config),
	}
}

// NewManagerWithEngine 使用现有引擎创建批量管理器
func NewManagerWithEngine(engine *core.Engine, config *types.Config) *Manager {
	return &Manager{
		engine:         engine,
		config:         config,
		formatter:      report.NewFormatter(config),
		tableFormatter: report.NewTableFormatter(config),
	}
}

// Start 启动批量管理器
func (bm *Manager) Start() error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if bm.running {
		return fmt.Errorf("批量管理器已在运行")
	}

	// 如果没有引擎，创建新引擎
	if bm.engine == nil {
		bm.engine = core.NewEngine(bm.config)
		if err := bm.engine.Start(); err != nil {
			return fmt.Errorf("启动引擎失败: %v", err)
		}
	}

	// 批量管理器简化：直接使用引擎，无需额外的调度器和缓存

	bm.running = true
	return nil
}

// Stop 停止批量管理器
func (bm *Manager) Stop() error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if !bm.running {
		return nil
	}

	// 批量管理器简化：无需停止额外的组件

	// 停止引擎
	if bm.engine != nil {
		bm.engine.Stop()
	}

	bm.running = false
	return nil
}

// CheckDomains 批量检测域名
func (bm *Manager) CheckDomains(ctx context.Context, domains []string) ([]*types.DetectionResult, error) {
	if !bm.running {
		return nil, fmt.Errorf("批量管理器未运行")
	}

	if len(domains) == 0 {
		return []*types.DetectionResult{}, nil
	}

	startTime := time.Now()

	// 使用流式检测显示实时进度
	results, err := bm.CheckDomainsWithProgress(ctx, domains)
	if err != nil {
		return nil, err
	}

	// 生成批量报告
	batchReport := bm.generateBatchReport(results, startTime, time.Now())

	// 打印报告
	fmt.Println(bm.formatBatchReport(batchReport))

	return results, nil
}

// CheckDomainsWithProgress 带进度显示的并发批量检测
func (bm *Manager) CheckDomainsWithProgress(ctx context.Context, domains []string) ([]*types.DetectionResult, error) {
	results := make([]*types.DetectionResult, len(domains))
	resultChan := make(chan *ProgressResult, len(domains))

	// 启动并发检测
	go func() {
		defer close(resultChan)

		// 使用WaitGroup控制并发
		var wg sync.WaitGroup

		// 使用配置的最大并发数，提高检测效率
		concurrency := int(bm.config.Concurrency.MaxConcurrent) // 使用配置的并发数（默认8个）
		semaphore := make(chan struct{}, concurrency)

		for i, domain := range domains {
			wg.Add(1)
			go func(index int, domain string) {
				defer wg.Done()

				// 获取信号量
				select {
				case semaphore <- struct{}{}:
					defer func() {
						<-semaphore
					}()
				case <-ctx.Done():
					return
				}

				// 检测域名
				result, err := bm.engine.CheckDomain(ctx, domain)

				// 发送结果
				select {
				case resultChan <- &ProgressResult{
					Index:  index,
					Domain: domain,
					Result: result,
					Error:  err,
				}:
				case <-ctx.Done():
					return
				}
			}(i, domain)
		}

		wg.Wait()
	}()

	// 收集结果并显示进度
	completed := 0
	for completed < len(domains) {
		select {
		case progressResult := <-resultChan:
			results[progressResult.Index] = progressResult.Result
			completed++

			// 显示进度
			fmt.Printf("[%s] 正在检测 [%d/%d]: %s... ", time.Now().Format("15:04:05"), completed, len(domains), progressResult.Domain)

			if progressResult.Error != nil {
				fmt.Printf("失败 - %v\n", progressResult.Error)
			} else if progressResult.Result.Suitable {
				fmt.Printf("适合\n")
			} else {
				// 获取不适合的原因
				reason := "未知原因"
				if progressResult.Result.Error != nil {
					reason = progressResult.Result.Error.Error()
				}
				fmt.Printf("不适合 - %s\n", reason)
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return results, nil
}

// ProgressResult 进度结果
type ProgressResult struct {
	Index  int
	Domain string
	Result *types.DetectionResult
	Error  error
}

// CheckDomainsStream 流式批量检测域名
func (bm *Manager) CheckDomainsStream(ctx context.Context, domains []string) (<-chan *types.DetectionResult, error) {
	if !bm.running {
		return nil, fmt.Errorf("批量管理器未运行")
	}

	return bm.engine.CheckDomainsStream(ctx, domains)
}

// generateBatchReport 生成批量报告
func (bm *Manager) generateBatchReport(results []*types.DetectionResult, startTime, endTime time.Time) *types.BatchReport {
	stats := &types.Statistics{
		TotalDomains: len(results),
	}

	for _, result := range results {
		// 区分技术错误和正常的检测结果
		if result.Error == nil {
			// 没有技术错误，检测成功
			stats.SuccessfulChecks++
		} else {
			// 检查是否是正常的检测结果（被墙、国内等）
			errorMsg := result.Error.Error()
			if strings.Contains(errorMsg, "域名被墙") || strings.Contains(errorMsg, "国内网站") {
				// 被墙和国内网站是正常的检测结果，不算失败
				stats.SuccessfulChecks++
			} else {
				// 真正的技术错误
				stats.FailedChecks++
			}
		}

		if result.Suitable {
			stats.SuitableDomains++
		}

		if result.Blocked != nil && result.Blocked.IsBlocked {
			stats.BlockedDomains++
		}
	}

	return &types.BatchReport{
		StartTime:     startTime,
		EndTime:       endTime,
		TotalDuration: endTime.Sub(startTime),
		Results:       results,
		Statistics:    stats,
		Summary: &types.BatchSummary{
			SuccessRate:     float64(stats.SuccessfulChecks) / float64(stats.TotalDomains),
			SuitabilityRate: float64(stats.SuitableDomains) / float64(stats.TotalDomains),
			BlockingRate:    float64(stats.BlockedDomains) / float64(stats.TotalDomains),
		},
	}
}

// formatBatchReport 格式化批量报告
func (bm *Manager) formatBatchReport(report *types.BatchReport) string {
	var result strings.Builder

	// 报告头部
	result.WriteString(fmt.Sprintf(`
批量检测报告
总耗时: %s
检测域名: %d 个
成功率: %.1f%%
适合性率: %.1f%%

`,
		formatDuration(report.TotalDuration),
		report.Statistics.TotalDomains,
		report.Summary.SuccessRate*100,
		report.Summary.SuitabilityRate*100,
	))

	// 分离适合和不适合的域名
	var suitableResults []*types.DetectionResult
	var unsuitableResults []*types.DetectionResult
	var excludedResults []*types.DetectionResult // 状态码不自然的域名

	for _, domainResult := range report.Results {
		if domainResult.Suitable && domainResult.Error == nil {
			suitableResults = append(suitableResults, domainResult)
		} else {
			// 检查是否因为状态码不自然而被排除
			if domainResult.StatusCodeCategory == types.StatusCodeCategoryExcluded {
				excludedResults = append(excludedResults, domainResult)
			} else {
				unsuitableResults = append(unsuitableResults, domainResult)
			}
		}
	}

	// 显示适合的域名表格
	if len(suitableResults) > 0 {
		// 按星级排序：1星在最上面，5星在最下面
		bm.sortByRecommendationStars(suitableResults)

		result.WriteString("适合的域名:\n\n")
		result.WriteString(bm.tableFormatter.FormatSuitableTable(suitableResults))
		result.WriteString("\n")
	}

	// 显示不适合的域名统计
	if len(unsuitableResults) > 0 {
		result.WriteString(bm.tableFormatter.FormatUnsuitableSummary(unsuitableResults))
	}

	// 显示被排除的域名（状态码不自然）
	if len(excludedResults) > 0 {
		result.WriteString("\n")
		result.WriteString(bm.formatExcludedDomains(excludedResults))
	}

	return result.String()
}

// formatExcludedDomains 格式化被排除的域名（状态码不自然）
func (bm *Manager) formatExcludedDomains(excludedResults []*types.DetectionResult) string {
	if len(excludedResults) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("状态码不自然的域名 (%d个):\n", len(excludedResults)))

	// 统计各种状态码
	statusCodeCounts := make(map[int]int)
	for _, domainResult := range excludedResults {
		if domainResult.Network != nil {
			statusCodeCounts[domainResult.Network.StatusCode]++
		}
	}

	// 按状态码排序显示
	var statusCodes []int
	for statusCode := range statusCodeCounts {
		statusCodes = append(statusCodes, statusCode)
	}
	sort.Ints(statusCodes)

	for _, statusCode := range statusCodes {
		count := statusCodeCounts[statusCode]
		result.WriteString(fmt.Sprintf("   - %d个状态码 %d\n", count, statusCode))
	}

	return result.String()
}

// calculateOptimalConcurrency 计算最优并发数
func (bm *Manager) calculateOptimalConcurrency(domainCount int) int {
	// 更激进的并发策略，提高检测效率
	if domainCount <= 5 {
		return domainCount // 小批量：每个域名一个并发
	} else if domainCount <= 20 {
		return 6 // 中小批量：6个并发
	} else if domainCount <= 50 {
		return 8 // 中批量：8个并发
	} else if domainCount <= 100 {
		return 10 // 大批量：10个并发
	} else {
		return 12 // 超大批量：最多12个并发
	}
}

// formatDuration 格式化时间显示
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.0fµs", float64(d.Nanoseconds())/1000)
	} else if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Nanoseconds())/1000000)
	} else if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	} else {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
}

// sortByRecommendationStars 按推荐星级排序，1星在最上面，5星在最下面
func (bm *Manager) sortByRecommendationStars(results []*types.DetectionResult) {
	// 使用sort.Slice进行排序
	sort.Slice(results, func(i, j int) bool {
		starsI := bm.calculateStars(results[i])
		starsJ := bm.calculateStars(results[j])
		return starsI < starsJ // 升序排列：1星在前，5星在后
	})
}

// calculateStars 计算域名的推荐星级数量
func (bm *Manager) calculateStars(result *types.DetectionResult) int {
	stars := 0

	// 1. TLS硬性条件检查 (TLS1.3 + X25519 + H2 + SNI匹配)
	if result.TLS != nil && result.TLS.SupportsTLS13 &&
		result.TLS.SupportsX25519 && result.TLS.SupportsHTTP2 &&
		result.SNI != nil && result.SNI.SNIMatch {
		stars++
	}

	// 2. 握手时间延迟小 (<= 200ms)
	if result.TLS != nil && result.TLS.HandshakeTime > 0 {
		handshakeMs := int(result.TLS.HandshakeTime.Milliseconds())
		if handshakeMs <= 200 {
			stars++
		}
	}

	// 3. 没有CDN (不使用CDN更安全)
	if result.CDN == nil || !result.CDN.IsCDN {
		stars++
	}

	// 4. 不是热门网站 (热门网站不推荐作为Reality目标)
	if result.CDN != nil && !result.CDN.IsHotWebsite {
		stars++
	}

	// 5. 证书时间长 (>= 60天)
	if result.Certificate != nil && result.Certificate.Valid {
		if result.Certificate.DaysUntilExpiry >= 60 {
			stars++
		}
	}

	return stars
}
