package batch

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"reality-checker-go/internal/core"
	"reality-checker-go/internal/report"
	"reality-checker-go/internal/types"
)

// Manager 批量管理器
type Manager struct {
	engine         *core.Engine
	scheduler      *Scheduler
	resultCache    *ResultCache
	progressTracker *ProgressTracker
	formatter      *report.Formatter
	config         *types.Config
	mu             sync.RWMutex
	running        bool
}

// NewManager 创建批量管理器
func NewManager(config *types.Config) *Manager {
	return &Manager{
		config:    config,
		formatter: report.NewFormatter(config),
	}
}

// NewManagerWithEngine 使用现有引擎创建批量管理器
func NewManagerWithEngine(engine *core.Engine, config *types.Config) *Manager {
	return &Manager{
		engine:    engine,
		config:    config,
		formatter: report.NewFormatter(config),
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

	// 创建调度器
	bm.scheduler = NewScheduler(bm.config)
	if err := bm.scheduler.Start(); err != nil {
		return fmt.Errorf("启动调度器失败: %v", err)
	}

	// 创建结果缓存
	bm.resultCache = NewResultCache(bm.config.Cache.TTL)

	// 创建进度跟踪器
	bm.progressTracker = NewProgressTracker()

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

	// 停止调度器
	if bm.scheduler != nil {
		bm.scheduler.Stop()
	}

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
		
		// 动态计算合适的并发数
		concurrency := bm.calculateOptimalConcurrency(len(domains))
		semaphore := make(chan struct{}, concurrency)
		
		// 并发控制已就绪
		
		for i, domain := range domains {
			wg.Add(1)
			go func(index int, domain string) {
				defer wg.Done()
				
				// 获取信号量
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
				
				// 检测域名
				result, err := bm.engine.CheckDomain(ctx, domain)
				
				// 发送结果
				resultChan <- &ProgressResult{
					Index:  index,
					Domain: domain,
					Result: result,
					Error:  err,
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
			fmt.Printf("正在检测 [%d/%d]: %s... ", completed, len(domains), progressResult.Domain)
			
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

详细结果:
`,
		formatDuration(report.TotalDuration),
		report.Statistics.TotalDomains,
		report.Summary.SuccessRate*100,
		report.Summary.SuitabilityRate*100,
	))
	
	// 详细结果
	for i, domainResult := range report.Results {
		result.WriteString(fmt.Sprintf("%d. %s: ", i+1, domainResult.Domain))
		
		if domainResult.Error != nil {
			result.WriteString(fmt.Sprintf("检测失败 - %v\n", domainResult.Error))
			continue
		}
		
		if domainResult.Suitable {
			result.WriteString("适合=true")
		} else {
			result.WriteString("适合=false")
		}
		
		// 网络信息
		if domainResult.Network != nil {
			if domainResult.Network.IsRedirected {
				result.WriteString(fmt.Sprintf(", 重定向: %s->%s", domainResult.Domain, domainResult.Network.FinalDomain))
			}
			result.WriteString(fmt.Sprintf(", 状态码=%d", domainResult.Network.StatusCode))
		}
		
		// TLS信息
		if domainResult.TLS != nil {
			result.WriteString(fmt.Sprintf(", TLS1.3=%t, X25519=%t, HTTP2=%t", 
				domainResult.TLS.SupportsTLS13, 
				domainResult.TLS.SupportsX25519, 
				domainResult.TLS.SupportsHTTP2))
			
			// 添加TLS握手时间
			if domainResult.TLS.HandshakeTime > 0 {
				handshakeMs := int(domainResult.TLS.HandshakeTime.Milliseconds())
				result.WriteString(fmt.Sprintf(", 握手时间=%dms", handshakeMs))
			}
		}
		
		// 地理位置
		if domainResult.Location != nil {
			result.WriteString(fmt.Sprintf(", 位置=%s", domainResult.Location.Country))
		}
		
		// CDN信息
		if domainResult.CDN != nil && domainResult.CDN.IsCDN {
			result.WriteString(fmt.Sprintf(", CDN=%s(%s)", domainResult.CDN.CDNProvider, domainResult.CDN.Confidence))
		}
		
		// 热门网站
		if domainResult.CDN != nil && domainResult.CDN.IsHotWebsite {
			result.WriteString(", 热门网站")
		}
		
		// 被墙信息
		if domainResult.Blocked != nil && domainResult.Blocked.IsBlocked {
			result.WriteString(", 被墙")
		}
		
		result.WriteString("\n")
	}
	
	return result.String()
}

// calculateOptimalConcurrency 计算最优并发数
func (bm *Manager) calculateOptimalConcurrency(domainCount int) int {
	// 保守的并发策略，适合大批量检测
	if domainCount <= 5 {
		return domainCount // 小批量：每个域名一个并发
	} else if domainCount <= 20 {
		return 3 // 中小批量：3个并发
	} else if domainCount <= 50 {
		return 4 // 中批量：4个并发
	} else if domainCount <= 100 {
		return 5 // 大批量：5个并发
	} else {
		return 6 // 超大批量：最多6个并发
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
