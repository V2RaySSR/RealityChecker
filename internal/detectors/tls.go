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
	startTime := time.Now()
	
	// 直接建立TLS连接（简化版本，避免依赖连接管理器）
	conn, err := tls.DialWithDialer(&net.Dialer{
		Timeout: 10 * time.Second,
	}, "tcp", domain+":443", &tls.Config{
		ServerName: domain,
		NextProtos: []string{"h2", "http/1.1"},
	})
	
	handshakeTime := time.Since(startTime)
	
	if err != nil {
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
	supportsHTTP2 := state.NegotiatedProtocol == "h2"

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
func (ts *TLSStage) checkX25519Support(state tls.ConnectionState) bool {
	if state.CipherSuite == 0 {
		return false
	}
	
	cipherName := strings.ToUpper(tls.CipherSuiteName(state.CipherSuite))
	
	// TLS 1.3 使用X25519作为密钥交换
	if state.Version == tls.VersionTLS13 {
		return strings.Contains(cipherName, "TLS_AES") || strings.Contains(cipherName, "X25519")
	}
	
	// TLS 1.2 检查特定的X25519密码套件
	return strings.Contains(cipherName, "X25519")
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