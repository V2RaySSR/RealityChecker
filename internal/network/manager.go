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
	config          *types.Config
	httpConnections map[string]*HTTPConnectionPool // HTTP连接池
	tlsConnections  map[string]*TLSConnectionPool  // TLS连接池
	mu              sync.RWMutex
	stats           *types.ConnectionStats
}

// HTTPConnectionPool HTTP连接池
type HTTPConnectionPool struct {
	connections chan net.Conn
	maxSize     int
	domain      string
	created     time.Time
	mu          sync.RWMutex
}

// TLSConnectionPool TLS连接池
type TLSConnectionPool struct {
	connections chan *tls.Conn
	maxSize     int
	domain      string
	created     time.Time
	mu          sync.RWMutex
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager(config *types.Config) *ConnectionManager {
	return &ConnectionManager{
		config:          config,
		httpConnections: make(map[string]*HTTPConnectionPool),
		tlsConnections:  make(map[string]*TLSConnectionPool),
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

	// 关闭所有HTTP连接
	for _, pool := range cm.httpConnections {
		close(pool.connections)
		for conn := range pool.connections {
			conn.Close()
		}
	}
	cm.httpConnections = make(map[string]*HTTPConnectionPool)

	// 关闭所有TLS连接
	for _, pool := range cm.tlsConnections {
		close(pool.connections)
		for conn := range pool.connections {
			conn.Close()
		}
	}
	cm.tlsConnections = make(map[string]*TLSConnectionPool)

	return nil
}

// GetHTTPConnection 获取HTTP连接
func (cm *ConnectionManager) GetHTTPConnection(ctx context.Context, domain string) (net.Conn, error) {
	// 总是创建新的HTTP连接
	const httpPort = ":80"
	conn, err := net.DialTimeout("tcp", domain+httpPort, cm.config.Network.Timeout)
	if err != nil {
		cm.mu.Lock()
		cm.stats.FailedConnections++
		cm.mu.Unlock()
		return nil, err
	}
	cm.mu.Lock()
	cm.stats.TotalConnections++
	cm.stats.ActiveConnections++
	cm.mu.Unlock()
	return conn, nil
}

// GetTLSConnection 获取TLS连接
func (cm *ConnectionManager) GetTLSConnection(ctx context.Context, domain string) (*tls.Conn, error) {
	// 总是创建新的TLS连接，确保ALPN协商正确
	const tlsPort = ":443"
	tcpConn, err := net.DialTimeout("tcp", domain+tlsPort, cm.config.Network.Timeout)
	if err != nil {
		cm.mu.Lock()
		cm.stats.FailedConnections++
		cm.mu.Unlock()
		return nil, err
	}

	// 创建TLS连接
	tlsConn := tls.Client(tcpConn, &tls.Config{
		ServerName: domain,
		NextProtos: []string{"h2", "http/1.1"}, // h2优先
	})

	// 执行TLS握手
	if err := tlsConn.Handshake(); err != nil {
		tcpConn.Close()
		cm.mu.Lock()
		cm.stats.FailedConnections++
		cm.mu.Unlock()
		return nil, err
	}

	cm.mu.Lock()
	cm.stats.TotalConnections++
	cm.stats.ActiveConnections++
	cm.mu.Unlock()
	return tlsConn, nil
}

// GetX25519TLSConnection 获取强制X25519的TLS连接
func (cm *ConnectionManager) GetX25519TLSConnection(ctx context.Context, domain string) (*tls.Conn, error) {
	// 创建强制X25519的TLS连接
	const tlsPort = ":443"
	tcpConn, err := net.DialTimeout("tcp", domain+tlsPort, cm.config.Network.Timeout)
	if err != nil {
		cm.mu.Lock()
		cm.stats.FailedConnections++
		cm.mu.Unlock()
		return nil, err
	}

	// 创建强制X25519的TLS连接
	tlsConn := tls.Client(tcpConn, &tls.Config{
		ServerName:       domain,
		NextProtos:       []string{"h2", "http/1.1"},
		CurvePreferences: []tls.CurveID{tls.X25519}, // 强制X25519
	})

	// 执行TLS握手
	if err := tlsConn.Handshake(); err != nil {
		tcpConn.Close()
		cm.mu.Lock()
		cm.stats.FailedConnections++
		cm.mu.Unlock()
		return nil, err
	}

	cm.mu.Lock()
	cm.stats.TotalConnections++
	cm.stats.ActiveConnections++
	cm.mu.Unlock()
	return tlsConn, nil
}

// CloseConnection 关闭连接
func (cm *ConnectionManager) CloseConnection(conn net.Conn) {
	if conn != nil {
		conn.Close()
		cm.mu.Lock()
		cm.stats.ActiveConnections--
		cm.mu.Unlock()
	}
}

// CloseTLSConnection 关闭TLS连接
func (cm *ConnectionManager) CloseTLSConnection(conn *tls.Conn) {
	if conn != nil {
		conn.Close()
		cm.mu.Lock()
		cm.stats.ActiveConnections--
		cm.mu.Unlock()
	}
}

// GetStats 获取连接统计信息
func (cm *ConnectionManager) GetStats() *types.ConnectionStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return &types.ConnectionStats{
		ActiveConnections: cm.stats.ActiveConnections,
		TotalConnections:  cm.stats.TotalConnections,
		FailedConnections: cm.stats.FailedConnections,
	}
}

// cleanupConnections 清理过期连接
func (cm *ConnectionManager) cleanupConnections() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		cm.mu.Lock()
		now := time.Now()

		// 清理HTTP连接池
		for domain, pool := range cm.httpConnections {
			if now.Sub(pool.created) > 5*time.Minute {
				// 先关闭所有连接
				for {
					select {
					case conn := <-pool.connections:
						conn.Close()
					default:
						goto httpCleanupDone
					}
				}
			httpCleanupDone:
				close(pool.connections)
				delete(cm.httpConnections, domain)
			}
		}

		// 清理TLS连接池
		for domain, pool := range cm.tlsConnections {
			if now.Sub(pool.created) > 5*time.Minute {
				// 先关闭所有连接
				for {
					select {
					case conn := <-pool.connections:
						conn.Close()
					default:
						goto tlsCleanupDone
					}
				}
			tlsCleanupDone:
				close(pool.connections)
				delete(cm.tlsConnections, domain)
			}
		}

		cm.mu.Unlock()
	}
}
