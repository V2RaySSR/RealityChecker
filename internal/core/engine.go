package core

import (
	"context"
	"fmt"
	"sync"

	"RealityChecker/internal/network"
	"RealityChecker/internal/types"
)

// Engine 主检测引擎（简化版本）
type Engine struct {
	config      *types.Config
	pipeline    *Pipeline
	connections *network.ConnectionManager
	mu          sync.RWMutex
	running     bool
}

// NewEngine 创建新的检测引擎（简化版本）
func NewEngine(config *types.Config) *Engine {
	engine := &Engine{
		config: config,
	}

	// 初始化组件
	engine.connections = network.NewConnectionManager(config)
	engine.pipeline = NewPipeline(engine.connections, config)

	return engine
}

// Start 启动引擎（简化版本）
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return fmt.Errorf("引擎已在运行")
	}

	// 启动连接管理器
	if err := e.connections.Start(); err != nil {
		return fmt.Errorf("启动连接管理器失败: %v", err)
	}

	// 缓存管理器已移除

	e.running = true
	return nil
}

// Stop 停止引擎（简化版本）
func (e *Engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return nil
	}

	// 缓存管理器已移除

	// 停止连接管理器
	if err := e.connections.Stop(); err != nil {
		fmt.Printf("停止连接管理器失败: %v\n", err)
	}

	e.running = false
	return nil
}

// CheckDomain 检测单个域名（直接使用pipeline，简化架构）
func (e *Engine) CheckDomain(ctx context.Context, domain string) (*types.DetectionResult, error) {
	if !e.running {
		return nil, fmt.Errorf("引擎未运行")
	}

	return e.pipeline.Execute(ctx, domain)
}

// CheckDomains 批量检测域名（移除并发控制，由调用方管理）
func (e *Engine) CheckDomains(ctx context.Context, domains []string) ([]*types.DetectionResult, error) {
	if !e.running {
		return nil, fmt.Errorf("引擎未运行")
	}

	if len(domains) == 0 {
		return []*types.DetectionResult{}, nil
	}

	// 移除Engine层的并发控制，由Batch Manager统一管理
	results := make([]*types.DetectionResult, len(domains))

	for i, domain := range domains {
		// 直接执行检测，不进行并发控制
		result, err := e.pipeline.Execute(ctx, domain)
		if err != nil {
			result = &types.DetectionResult{
				Domain: domain,
				Error:  err,
			}
		}
		results[i] = result
	}

	return results, nil
}

// CheckDomainsStream 流式批量检测域名
func (e *Engine) CheckDomainsStream(ctx context.Context, domains []string) (<-chan *types.DetectionResult, error) {
	if !e.running {
		return nil, fmt.Errorf("引擎未运行")
	}

	resultChan := make(chan *types.DetectionResult, len(domains))

	go func() {
		defer close(resultChan)

		semaphore := make(chan struct{}, int(e.config.Concurrency.MaxConcurrent))
		var wg sync.WaitGroup

		for _, domain := range domains {
			wg.Add(1)
			go func(domain string) {
				defer wg.Done()

				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				result, err := e.pipeline.Execute(ctx, domain)
				if err != nil {
					result = &types.DetectionResult{
						Domain: domain,
						Error:  err,
					}
				}

				select {
				case resultChan <- result:
				case <-ctx.Done():
					return
				}
			}(domain)
		}

		wg.Wait()
	}()

	return resultChan, nil
}

// GetStats 获取引擎统计信息（简化版本）
func (e *Engine) GetStats() *EngineStats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return &EngineStats{
		Running:     e.running,
		Connections: e.connections.GetStats(),
		Cache:       nil, // 缓存管理器已移除
	}
}

// EngineStats 引擎统计信息（简化版本）
type EngineStats struct {
	Running     bool                   `json:"running"`
	Connections *types.ConnectionStats `json:"connections"`
	Cache       *types.CacheStats      `json:"cache"`
}

// CoordinatorStats 协调器统计
type CoordinatorStats struct {
	TotalTasks     int `json:"total_tasks"`
	CompletedTasks int `json:"completed_tasks"`
	FailedTasks    int `json:"failed_tasks"`
	ActiveWorkers  int `json:"active_workers"`
}
