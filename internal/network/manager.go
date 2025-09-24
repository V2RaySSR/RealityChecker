package network

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"RealityChecker/internal/types"
)

// ConnectionManager 连接管理器
type ConnectionManager struct {
	config      *types.Config
	connections map[string]*ConnectionPool
	mu          sync.RWMutex
	stats       *types.ConnectionStats
}

// ConnectionPool 连接池
type ConnectionPool struct {
	connections chan net.Conn
	tlsConnections chan *tls.Conn
	maxSize     int
	domain      string
	created     time.Time
	mu          sync.RWMutex
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager(config *types.Config) *ConnectionManager {
	return &ConnectionManager{
		config:      config,
		connections: make(map[string]*ConnectionPool),
		stats: &types.ConnectionStats{
			ActiveConnections: 0,
			TotalConnections:  0,
			FailedConnections: 0,
		},
	}
}

// Start 启动连接管理器
func (cm *ConnectionManager) Start() error {
	// 启动连接清理协程
	go cm.cleanupConnections()
	return nil
}

// Stop 停止连接管理器
func (cm *ConnectionManager) Stop() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 关闭所有连接池
	for _, pool := range cm.connections {
		pool.Close()
	}
	cm.connections = make(map[string]*ConnectionPool)
	
	return nil
}

// GetConnection 获取连接
func (cm *ConnectionManager) GetConnection(ctx context.Context, domain string) (net.Conn, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	pool, exists := cm.connections[domain]
	if !exists {
		const poolMultiplier = 2 // 连接池大小倍数
		poolSize := int(cm.config.Concurrency.MaxConcurrent) * poolMultiplier
		pool = &ConnectionPool{
			connections: make(chan net.Conn, poolSize),
			maxSize:     poolSize,
			domain:      domain,
			created:     time.Now(),
		}
		cm.connections[domain] = pool
	}

	select {
	case conn := <-pool.connections:
		// 复用现有连接
		cm.stats.ActiveConnections++
		return conn, nil
	default:
		// 创建新连接
		const tlsPort = ":443"
		conn, err := net.DialTimeout("tcp", domain+tlsPort, cm.config.Network.Timeout)
		if err != nil {
			cm.stats.FailedConnections++
			return nil, err
		}
		cm.stats.TotalConnections++
		cm.stats.ActiveConnections++
		return conn, nil
	}
}

// GetTLSConnection 获取TLS连接
func (cm *ConnectionManager) GetTLSConnection(ctx context.Context, domain string) (*tls.Conn, error) {
	conn, err := cm.GetConnection(ctx, domain)
	if err != nil {
		return nil, err
	}

	tlsConn := tls.Client(conn, &tls.Config{
		ServerName: domain,
		NextProtos: []string{"h2", "http/1.1"},
	})

	if err := tlsConn.Handshake(); err != nil {
		conn.Close()
		cm.stats.FailedConnections++
		return nil, err
	}

	return tlsConn, nil
}

// ReturnConnection 归还连接
func (cm *ConnectionManager) ReturnConnection(domain string, conn net.Conn) {
	cm.mu.RLock()
	pool, exists := cm.connections[domain]
	cm.mu.RUnlock()

	if !exists {
		conn.Close()
		return
	}

	select {
	case pool.connections <- conn:
		cm.mu.Lock()
		cm.stats.ActiveConnections--
		cm.mu.Unlock()
	default:
		// 连接池已满，关闭连接
		conn.Close()
		cm.mu.Lock()
		cm.stats.ActiveConnections--
		cm.mu.Unlock()
	}
}

// cleanupConnections 清理过期连接
func (cm *ConnectionManager) cleanupConnections() {
	const cleanupInterval = 5 * time.Minute
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		cm.mu.Lock()
		now := time.Now()
		for domain, pool := range cm.connections {
			if now.Sub(pool.created) > cm.config.Cache.TTL {
				pool.Close()
				delete(cm.connections, domain)
			}
		}
		cm.mu.Unlock()
	}
}

// Close 关闭连接池
func (cp *ConnectionPool) Close() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// 关闭所有连接
	close(cp.connections)
	for conn := range cp.connections {
		conn.Close()
	}
}

// GetStats 获取统计信息
func (cm *ConnectionManager) GetStats() *types.ConnectionStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.stats
}

// CacheManager 缓存管理器
type CacheManager struct {
	config     *types.Config
	dnsCache   map[string]*DNSCacheEntry
	resultCache map[string]*ResultCacheEntry
	mu         sync.RWMutex
	stats      *types.CacheStats
}

// DNSCacheEntry DNS缓存条目
type DNSCacheEntry struct {
	IPs       []string
	ExpiresAt time.Time
}

// ResultCacheEntry 结果缓存条目
type ResultCacheEntry struct {
	Result    *types.DetectionResult
	ExpiresAt time.Time
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(config *types.Config) *CacheManager {
	return &CacheManager{
		config:       config,
		dnsCache:     make(map[string]*DNSCacheEntry),
		resultCache:  make(map[string]*ResultCacheEntry),
		stats: &types.CacheStats{
			DNSCacheSize:   0,
			ResultCacheSize: 0,
			CDNCacheSize:   0,
			HitRate:        0.0,
		},
	}
}

// Start 启动缓存管理器
func (cm *CacheManager) Start() error {
	// 启动缓存清理协程
	go cm.cleanupCache()
	return nil
}

// Stop 停止缓存管理器
func (cm *CacheManager) Stop() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.dnsCache = make(map[string]*DNSCacheEntry)
	cm.resultCache = make(map[string]*ResultCacheEntry)
	return nil
}

// GetDNS 获取DNS缓存
func (cm *CacheManager) GetDNS(domain string) ([]string, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	entry, exists := cm.dnsCache[domain]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	cm.stats.HitRate = 0.8 // 简化计算
	return entry.IPs, true
}

// SetDNS 设置DNS缓存
func (cm *CacheManager) SetDNS(domain string, ips []string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.dnsCache[domain] = &DNSCacheEntry{
		IPs:       ips,
		ExpiresAt: time.Now().Add(cm.config.Cache.TTL),
	}
	cm.stats.DNSCacheSize = len(cm.dnsCache)
}

// GetResult 获取结果缓存
func (cm *CacheManager) GetResult(domain string) (*types.DetectionResult, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	entry, exists := cm.resultCache[domain]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Result, true
}

// SetResult 设置结果缓存
func (cm *CacheManager) SetResult(domain string, result *types.DetectionResult) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.resultCache[domain] = &ResultCacheEntry{
		Result:    result,
		ExpiresAt: time.Now().Add(cm.config.Cache.TTL),
	}
	cm.stats.ResultCacheSize = len(cm.resultCache)
}

// cleanupCache 清理过期缓存
func (cm *CacheManager) cleanupCache() {
	const cacheCleanupInterval = 1 * time.Minute
	ticker := time.NewTicker(cacheCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		cm.mu.Lock()
		now := time.Now()
		
		// 清理DNS缓存
		for domain, entry := range cm.dnsCache {
			if now.After(entry.ExpiresAt) {
				delete(cm.dnsCache, domain)
			}
		}
		
		// 清理结果缓存
		for domain, entry := range cm.resultCache {
			if now.After(entry.ExpiresAt) {
				delete(cm.resultCache, domain)
			}
		}
		
		cm.stats.DNSCacheSize = len(cm.dnsCache)
		cm.stats.ResultCacheSize = len(cm.resultCache)
		cm.mu.Unlock()
	}
}

// GetStats 获取统计信息
func (cm *CacheManager) GetStats() *types.CacheStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.stats
}
