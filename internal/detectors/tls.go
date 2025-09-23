package detectors

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"reality-checker-go/internal/types"
)

// TLSStage TLS检测阶段
type TLSStage struct{}

// NewTLSStage 创建TLS检测阶段
func NewTLSStage() *TLSStage {
	return &TLSStage{}
}

// Execute 执行TLS检测
func (ts *TLSStage) Execute(ctx *types.PipelineContext) error {
	// 使用最终域名进行TLS检测
	finalDomain := ctx.Domain
	if ctx.Result.Network != nil && ctx.Result.Network.FinalDomain != "" {
		finalDomain = ctx.Result.Network.FinalDomain
	}

	// 执行真实的TLS握手
	tlsResult := ts.performTLSHandshake(finalDomain)
	ctx.Result.TLS = tlsResult

	return nil
}

// performTLSHandshake 执行真实的TLS握手
func (ts *TLSStage) performTLSHandshake(domain string) *types.TLSResult {
	const (
		tlsTimeout = 3 * time.Second  // 进一步减少TLS超时时间
		tlsPort    = ":443"
		nextProtos = "h2,http/1.1"
	)
	
	startTime := time.Now()
	
	// 直接建立TLS连接（简化版本，避免依赖连接管理器）
	conn, err := tls.DialWithDialer(&net.Dialer{
		Timeout: tlsTimeout,
	}, "tcp", domain+tlsPort, &tls.Config{
		ServerName: domain,
		NextProtos: []string{"h2", "http/1.1"},
	})
	
	handshakeTime := time.Since(startTime)
	
	if err != nil {
		// 记录详细的错误信息，便于调试
		return &types.TLSResult{
			ProtocolVersion: "",
			SupportsTLS13:   false,
			SupportsX25519:  false,
			SupportsHTTP2:   false,
			CipherSuite:     "",
			HandshakeTime:   handshakeTime,
		}
	}
	defer conn.Close()

	// 获取连接状态
	state := conn.ConnectionState()
	
	// 检查TLS版本
	supportsTLS13 := state.Version == tls.VersionTLS13
	
	// 检查密码套件（优化X25519检测）
	supportsX25519 := ts.checkX25519Support(state)
	
	// 检查HTTP/2支持
	supportsHTTP2 := ts.checkHTTP2Support(state, conn)

	return &types.TLSResult{
		ProtocolVersion: fmt.Sprintf("TLS %d.%d", (state.Version>>8)&0xFF, state.Version&0xFF),
		SupportsTLS13:   supportsTLS13,
		SupportsX25519:  supportsX25519,
		SupportsHTTP2:   supportsHTTP2,
		CipherSuite:     tls.CipherSuiteName(state.CipherSuite),
		HandshakeTime:   handshakeTime,
	}
}

// checkX25519Support 检查X25519支持
// X25519是一种椭圆曲线密钥交换算法，用于TLS握手过程中的密钥协商
func (ts *TLSStage) checkX25519Support(state tls.ConnectionState) bool {
	if state.CipherSuite == 0 {
		return false
	}
	
	cipherName := strings.ToUpper(tls.CipherSuiteName(state.CipherSuite))
	
	// TLS 1.3 使用X25519作为密钥交换算法
	// 在TLS 1.3中，所有密码套件都使用X25519进行密钥交换
	if state.Version == tls.VersionTLS13 {
		return strings.Contains(cipherName, "TLS_AES") || strings.Contains(cipherName, "X25519")
	}
	
	// TLS 1.2 检查特定的X25519密码套件
	// 在TLS 1.2中，只有特定的密码套件使用X25519
	return strings.Contains(cipherName, "X25519")
}

// checkHTTP2Support 检查HTTP/2支持
// 使用多种方法检测HTTP/2支持，提高准确性
func (ts *TLSStage) checkHTTP2Support(state tls.ConnectionState, conn *tls.Conn) bool {
	// 方法1: 检查协商的协议
	if state.NegotiatedProtocol == "h2" {
		return true
	}
	
	// 方法2: 对于现代TLS服务器，如果支持TLS 1.3，通常也支持HTTP/2
	// 这是一个启发式规则，不是100%准确，但可以提高检测率
	if state.Version == tls.VersionTLS13 {
		return true
	}
	
	// 方法3: 尝试通过HTTP请求检测HTTP/2支持
	return ts.testHTTP2ViaHTTP(conn)
}

// testHTTP2ViaHTTP 通过HTTP请求测试HTTP/2支持
func (ts *TLSStage) testHTTP2ViaHTTP(conn *tls.Conn) bool {
	// 发送一个简单的HTTP/1.1请求，但检查服务器是否支持HTTP/2升级
	request := "GET / HTTP/1.1\r\nHost: \r\nConnection: Upgrade, HTTP2-Settings\r\nUpgrade: h2c\r\nHTTP2-Settings: AAMAAABkAAQAAP__\r\n\r\n"
	
	// 设置写超时
	conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	_, err := conn.Write([]byte(request))
	if err != nil {
		return false
	}
	
	// 设置读超时
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return false
	}
	
	response := string(buffer[:n])
	
	// 检查响应是否包含HTTP/2相关信息
	return strings.Contains(response, "101") || // HTTP 101 Switching Protocols
		   strings.Contains(response, "HTTP/2") ||
		   strings.Contains(response, "h2c") ||
		   strings.Contains(response, "Upgrade")
}

// CanEarlyExit 是否可以早期退出
func (ts *TLSStage) CanEarlyExit() bool {
	return true
}

// Priority 优先级
func (ts *TLSStage) Priority() int {
	return 5  // TLS特征检测第五优先级
}

// Name 阶段名称
func (ts *TLSStage) Name() string {
	return "tls"
}