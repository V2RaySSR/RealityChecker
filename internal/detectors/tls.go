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
	
	// 建立TLS连接
	conn, err := tls.DialWithDialer(&net.Dialer{
		Timeout: 10 * time.Second,
	}, "tcp", domain+":443", &tls.Config{
		ServerName: domain,
		NextProtos: []string{"h2", "http/1.1"}, // 支持HTTP/2
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
	
	// 检查密码套件
	supportsX25519 := false
	if state.CipherSuite != 0 {
		cipherName := tls.CipherSuiteName(state.CipherSuite)
		supportsX25519 = strings.Contains(strings.ToUpper(cipherName), "X25519") ||
			strings.Contains(strings.ToUpper(cipherName), "TLS_AES")
	}
	
	// 检查HTTP/2支持
	supportsHTTP2 := false
	if state.NegotiatedProtocol != "" && state.NegotiatedProtocol == "h2" {
		supportsHTTP2 = true
	}

	return &types.TLSResult{
		ProtocolVersion: fmt.Sprintf("TLS %d.%d", (state.Version>>8)&0xFF, state.Version&0xFF),
		SupportsTLS13:   supportsTLS13,
		SupportsX25519:  supportsX25519,
		SupportsHTTP2:   supportsHTTP2,
		CipherSuite:     tls.CipherSuiteName(state.CipherSuite),
		HandshakeTime:   handshakeTime,
	}
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