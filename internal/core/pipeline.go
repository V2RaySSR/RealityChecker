package core

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"RealityChecker/internal/detectors"
	"RealityChecker/internal/network"
	"RealityChecker/internal/types"
)

// Pipeline 检测流水线
type Pipeline struct {
	stages      []types.DetectionStage
	config      *types.Config
	earlyExit   bool
	connections *network.ConnectionManager
}

// NewPipeline 创建新的检测流水线
func NewPipeline(connections *network.ConnectionManager, config *types.Config) *Pipeline {
	pipeline := &Pipeline{
		config:      config,
		earlyExit:   true,
		connections: connections,
	}

	// 初始化检测阶段
	pipeline.initializeStages()

	return pipeline
}

// initializeStages 初始化检测阶段
func (p *Pipeline) initializeStages() {
	p.stages = []types.DetectionStage{
		detectors.NewBlockedStage(),          // 1. 被墙检测 (最高优先级)
		detectors.NewRedirectStage(),         // 2. 重定向检测
		detectors.NewStatusCheckStage(),      // 3. 状态码检查
		detectors.NewIPResolverStage(),       // 4. IP解析
		detectors.NewLocationStage(),         // 5. 地理位置检测
		detectors.NewLocationCheckStage(),    // 6. 地理位置检查
		detectors.NewComprehensiveTLSStage(), // 7. 综合TLS检测 (TLS1.3、X25519、H2、SNI、证书、CDN)
		detectors.NewHotWebsiteStage(),       // 8. 热门网站检测
	}

	// 按优先级排序
	sort.Slice(p.stages, func(i, j int) bool {
		return p.stages[i].Priority() < p.stages[j].Priority()
	})
}

// Execute 执行检测流水线
func (p *Pipeline) Execute(ctx context.Context, domain string) (*types.DetectionResult, error) {
	startTime := time.Now()

	// 创建流水线上下文
	pipelineCtx := &types.PipelineContext{
		Domain:      domain,
		StartTime:   startTime,
		Result:      &types.DetectionResult{Domain: domain, StartTime: startTime},
		Connections: p.connections, // 传递连接管理器给检测器
		Cache:       nil,           // 缓存管理器已移除
		Config:      p.config,
		Context:     ctx, // 传递原始context
		EarlyExit:   false,
	}

	// 并发执行检测阶段，提高检测效率
	p.executeStagesConcurrently(ctx, pipelineCtx)

	// 计算总耗时
	pipelineCtx.Result.Duration = time.Since(startTime)

	// 评估适合性
	p.evaluateSuitability(pipelineCtx.Result)

	return pipelineCtx.Result, nil
}

// executeStagesConcurrently 并发执行检测阶段
func (p *Pipeline) executeStagesConcurrently(ctx context.Context, pipelineCtx *types.PipelineContext) {
	// 将检测阶段分为两组：阻塞检测和网络检测
	var blockingStages []types.DetectionStage
	var networkStages []types.DetectionStage

	for _, stage := range p.stages {
		if stage.CanEarlyExit() {
			blockingStages = append(blockingStages, stage)
		} else {
			networkStages = append(networkStages, stage)
		}
	}

	// 先执行阻塞检测（被墙检测、地理位置检测等）
	for _, stage := range blockingStages {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := stage.Execute(pipelineCtx); err != nil {
			pipelineCtx.Result.Error = err
			if stage.CanEarlyExit() {
				pipelineCtx.EarlyExit = true
				return
			}
		}

		// 检查是否需要早期退出
		if p.earlyExit && stage.CanEarlyExit() && pipelineCtx.EarlyExit {
			pipelineCtx.Result.EarlyExit = true
			return
		}
	}

	// 如果被阻塞，直接返回
	if pipelineCtx.EarlyExit {
		return
	}

	// 并发执行网络检测阶段
	if len(networkStages) > 0 {
		p.executeNetworkStagesConcurrently(ctx, pipelineCtx, networkStages)
	}
}

// executeNetworkStagesConcurrently 并发执行网络检测阶段
func (p *Pipeline) executeNetworkStagesConcurrently(ctx context.Context, pipelineCtx *types.PipelineContext, stages []types.DetectionStage) {
	// 使用信号量控制网络检测的并发数
	networkConcurrency := 4 // 网络检测使用4个并发
	semaphore := make(chan struct{}, networkConcurrency)

	var wg sync.WaitGroup
	for i, stage := range stages {
		wg.Add(1)
		go func(index int, s types.DetectionStage) {
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

			// 执行检测阶段
			func() {
				defer func() {
					if r := recover(); r != nil {
						pipelineCtx.Result.Error = fmt.Errorf("检测阶段 %s panic: %v", s.Name(), r)
					}
				}()

				if err := s.Execute(pipelineCtx); err != nil {
					pipelineCtx.Result.Error = err
				}
			}()
		}(i, stage)
	}

	wg.Wait()
}

// evaluateSuitability 评估适合性
func (p *Pipeline) evaluateSuitability(result *types.DetectionResult) {
	// 检查硬性条件
	if result.Blocked != nil && result.Blocked.IsBlocked {
		result.Suitable = false
		result.Error = fmt.Errorf("域名被墙")
		return
	}

	if result.Location != nil && result.Location.IsDomestic {
		result.Suitable = false
		result.Error = fmt.Errorf("国内网站")
		return
	}

	if result.Network != nil && !result.Network.Accessible {
		result.Suitable = false
		result.Error = fmt.Errorf("网络不可达")
		result.StatusCodeCategory = types.StatusCodeCategoryNetwork
		return
	}

	// 检查状态码是否安全
	if result.Network != nil && result.Network.Accessible {
		statusCodeCategory := types.ClassifyStatusCode(result.Network.StatusCode, true)
		result.StatusCodeCategory = statusCodeCategory

		// 如果状态码不安全，标记为不适合
		if statusCodeCategory == types.StatusCodeCategoryExcluded {
			result.Suitable = false
			result.Error = fmt.Errorf("状态码不自然: %d", result.Network.StatusCode)
			return
		}
	}

	if result.TLS != nil {
		if !result.TLS.SupportsTLS13 {
			result.Suitable = false
			result.Error = fmt.Errorf("不支持TLS 1.3")
			return
		}
		if !result.TLS.SupportsX25519 {
			result.Suitable = false
			result.Error = fmt.Errorf("不支持X25519密钥交换")
			return
		}
		if !result.TLS.SupportsHTTP2 {
			result.Suitable = false
			result.Error = fmt.Errorf("不支持HTTP/2")
			return
		}
	}

	if result.Certificate != nil {
		if !result.Certificate.Valid {
			result.Suitable = false
			result.Error = fmt.Errorf("证书无效")
			return
		}
		// 只有真正过期的证书才标记为不适合（天数小于等于0）
		if result.Certificate.DaysUntilExpiry <= 0 {
			result.Suitable = false
			result.Error = fmt.Errorf("证书已过期（%d天）", result.Certificate.DaysUntilExpiry)
			return
		}
	}

	if result.SNI != nil && (!result.SNI.SupportsSNI || !result.SNI.SNIMatch) {
		result.Suitable = false
		result.Error = fmt.Errorf("SNI不匹配")
		return
	}

	// 所有硬性条件都符合
	result.Suitable = true
	result.HardRequirementsMet = true
}

// SetEarlyExit 设置是否早期退出
func (p *Pipeline) SetEarlyExit(earlyExit bool) {
	p.earlyExit = earlyExit
}

// GetStages 获取检测阶段
func (p *Pipeline) GetStages() []types.DetectionStage {
	return p.stages
}

// AddStage 添加检测阶段
func (p *Pipeline) AddStage(stage types.DetectionStage) {
	p.stages = append(p.stages, stage)
	sort.Slice(p.stages, func(i, j int) bool {
		return p.stages[i].Priority() < p.stages[j].Priority()
	})
}

// RemoveStage 移除检测阶段
func (p *Pipeline) RemoveStage(name string) {
	var newStages []types.DetectionStage
	for _, stage := range p.stages {
		if stage.Name() != name {
			newStages = append(newStages, stage)
		}
	}
	p.stages = newStages
}
