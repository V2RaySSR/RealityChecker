package batch

import (
	"fmt"
	"sync"

	"reality-checker-go/internal/types"
)

// Scheduler 任务调度器
type Scheduler struct {
	config    *types.Config
	mu        sync.RWMutex
	running   bool
}

// NewScheduler 创建任务调度器
func NewScheduler(config *types.Config) *Scheduler {
	return &Scheduler{
		config: config,
	}
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("调度器已在运行")
	}

	s.running = true
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.running = false
	return nil
}
