package batch

import (
	"sync"
	"time"
)

// ProgressTracker 进度跟踪器
type ProgressTracker struct {
	completed int
	total     int
	startTime time.Time
	mu        sync.RWMutex
}

// NewProgressTracker 创建进度跟踪器
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		startTime: time.Now(),
	}
}

// Update 更新进度
func (pt *ProgressTracker) Update(completed, total int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.completed = completed
	pt.total = total
}

// GetProgress 获取进度
func (pt *ProgressTracker) GetProgress() (int, int, time.Duration) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	return pt.completed, pt.total, time.Since(pt.startTime)
}
