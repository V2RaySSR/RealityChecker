package core

import (
	"context"
	"fmt"
	"sync"

	"RealityChecker/internal/network"
	"RealityChecker/internal/types"
)

// Engine 主检测引擎
type Engine struct {
	config       *types.Config
	pipeline     *Pipeline
	coordinator  *Coordinator
	connections  *network.ConnectionManager
	cache        *network.CacheManager
	mu           sync.RWMutex
	running      bool
}

// NewEngine 创建新的检测引擎
func NewEngine(config *types.Config) *Engine {
	engine := &Engine{
		config: config,
	}

	// 初始化组件
	engine.connections = network.NewConnectionManager(config)
	engine.cache = network.NewCacheManager(config)
	engine.pipeline = NewPipeline(engine.connections, engine.cache, config)
	engine.coordinator = NewCoordinator(engine.pipeline, config)

	return engine
}

// Start 启动引擎
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

	// 启动缓存管理器
	if err := e.cache.Start(); err != nil {
		return fmt.Errorf("启动缓存管理器失败: %v", err)
	}

	// 启动协调器
	if err := e.coordinator.Start(); err != nil {
		return fmt.Errorf("启动协调器失败: %v", err)
	}

	e.running = true
	return nil
}

// Stop 停止引擎
func (e *Engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return nil
	}

	// 停止协调器
	if err := e.coordinator.Stop(); err != nil {
		fmt.Printf("停止协调器失败: %v\n", err)
	}

	// 停止缓存管理器
	if err := e.cache.Stop(); err != nil {
		fmt.Printf("停止缓存管理器失败: %v\n", err)
	}

	// 停止连接管理器
	if err := e.connections.Stop(); err != nil {
		fmt.Printf("停止连接管理器失败: %v\n", err)
	}

	e.running = false
	return nil
}

// CheckDomain 检测单个域名
func (e *Engine) CheckDomain(ctx context.Context, domain string) (*types.DetectionResult, error) {
	if !e.running {
		return nil, fmt.Errorf("引擎未运行")
	}

	return e.coordinator.CheckDomain(ctx, domain)
}

// CheckDomains 批量检测域名
func (e *Engine) CheckDomains(ctx context.Context, domains []string) ([]*types.DetectionResult, error) {
	if !e.running {
		return nil, fmt.Errorf("引擎未运行")
	}

	return e.coordinator.CheckDomains(ctx, domains)
}

// CheckDomainsStream 流式批量检测域名
func (e *Engine) CheckDomainsStream(ctx context.Context, domains []string) (<-chan *types.DetectionResult, error) {
	if !e.running {
		return nil, fmt.Errorf("引擎未运行")
	}

	return e.coordinator.CheckDomainsStream(ctx, domains)
}

// GetStats 获取引擎统计信息
func (e *Engine) GetStats() *EngineStats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return &EngineStats{
		Running:      e.running,
		Connections:  e.connections.GetStats(),
		Cache:        e.cache.GetStats(),
		Coordinator:  e.coordinator.GetStats(),
	}
}

// EngineStats 引擎统计信息
type EngineStats struct {
	Running     bool                    `json:"running"`
	Connections *types.ConnectionStats  `json:"connections"`
	Cache       *types.CacheStats       `json:"cache"`
	Coordinator *CoordinatorStats       `json:"coordinator"`
}

// CoordinatorStats 协调器统计
type CoordinatorStats struct {
	TotalTasks      int `json:"total_tasks"`
	CompletedTasks  int `json:"completed_tasks"`
	FailedTasks     int `json:"failed_tasks"`
	ActiveWorkers   int `json:"active_workers"`
}
