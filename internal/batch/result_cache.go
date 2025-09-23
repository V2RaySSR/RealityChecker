package batch

import (
	"sync"
	"time"

	"reality-checker-go/internal/types"
)

// ResultCache 结果缓存
type ResultCache struct {
	cache map[string]*CachedResult
	ttl   time.Duration
	mu    sync.RWMutex
}

// CachedResult 缓存结果
type CachedResult struct {
	Result    *types.DetectionResult
	Timestamp time.Time
}

// NewResultCache 创建结果缓存
func NewResultCache(ttl time.Duration) *ResultCache {
	return &ResultCache{
		cache: make(map[string]*CachedResult),
		ttl:   ttl,
	}
}

// Get 获取缓存结果
func (rc *ResultCache) Get(domain string) (*types.DetectionResult, bool) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	cached, exists := rc.cache[domain]
	if !exists {
		return nil, false
	}

	// 检查TTL
	if time.Since(cached.Timestamp) > rc.ttl {
		return nil, false
	}

	return cached.Result, true
}

// Set 设置缓存结果
func (rc *ResultCache) Set(domain string, result *types.DetectionResult) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.cache[domain] = &CachedResult{
		Result:    result,
		Timestamp: time.Now(),
	}
}
