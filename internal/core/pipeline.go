package core

import (
	"context"
	"fmt"
	"sort"
	"time"

	"RealityChecker/internal/detectors"
	"RealityChecker/internal/network"
	"RealityChecker/internal/types"
)

// Pipeline 检测流水线
type Pipeline struct {
	stages    []types.DetectionStage
	context   *types.PipelineContext
	config    *types.Config
	earlyExit bool
	connections *network.ConnectionManager
	cache       *network.CacheManager
}

// NewPipeline 创建新的检测流水线
func NewPipeline(connections *network.ConnectionManager, cache *network.CacheManager, config *types.Config) *Pipeline {
	pipeline := &Pipeline{
		config:      config,
		earlyExit:   true,
		connections: connections,
		cache:       cache,
	}

	// 初始化检测阶段
	pipeline.initializeStages()

	return pipeline
}

// initializeStages 初始化检测阶段
func (p *Pipeline) initializeStages() {
	p.stages = []types.DetectionStage{
		detectors.NewBlockedStage(),     // 1. 被墙检测 (最高优先级)
		detectors.NewRedirectStage(),    // 2. 重定向检测
		detectors.NewIPResolverStage(),  // 3. IP解析
		detectors.NewLocationStage(),    // 4. 地理位置检测
		detectors.NewTLSStage(),         // 5. TLS特征检测
		detectors.NewSNIStage(),         // 6. SNI检测
		detectors.NewCertificateStage(), // 7. 证书检测
		detectors.NewCDNStage(),         // 8. CDN检测
		detectors.NewHotWebsiteStage(),  // 9. 热门网站检测
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
		Connections: nil, // 当前检测器不需要ConnectionManager
		Cache:       nil, // 当前检测器不需要CacheManager
		Config:      p.config,
		EarlyExit:   false,
	}

	// 执行各个检测阶段
	for _, stage := range p.stages {
		select {
		case <-ctx.Done():
			return pipelineCtx.Result, ctx.Err()
		default:
		}

		// 为每个阶段设置超时
		stageCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
		done := make(chan error, 1)
		
		go func() {
			done <- stage.Execute(pipelineCtx)
		}()
		
		select {
		case err := <-done:
			cancel()
			if err != nil {
				pipelineCtx.Result.Error = err
				if stage.CanEarlyExit() {
					pipelineCtx.EarlyExit = true
					break
				}
			}
		case <-stageCtx.Done():
			cancel()
			pipelineCtx.Result.Error = fmt.Errorf("检测阶段 %s 超时", stage.Name())
			pipelineCtx.EarlyExit = true
			break
		}

		// 检查是否需要早期退出
		if p.earlyExit && stage.CanEarlyExit() && pipelineCtx.EarlyExit {
			pipelineCtx.Result.EarlyExit = true
			break
		}
	}

	// 计算总耗时
	pipelineCtx.Result.Duration = time.Since(startTime)

	// 评估适合性
	p.evaluateSuitability(pipelineCtx.Result)

	return pipelineCtx.Result, nil
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
		return
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
