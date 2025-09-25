package detectors

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"RealityChecker/internal/types"
)

// ComprehensiveTLSStage 综合TLS检测阶段
// 在一个TLS连接中完成所有TLS相关检测：TLS1.3、X25519、HTTP/2、SNI、证书
type ComprehensiveTLSStage struct{}

// NewComprehensiveTLSStage 创建综合TLS检测阶段
func NewComprehensiveTLSStage() *ComprehensiveTLSStage {
	return &ComprehensiveTLSStage{}
}

// Execute 执行综合TLS检测
func (cts *ComprehensiveTLSStage) Execute(ctx *types.PipelineContext) error {
	// 使用最终域名进行TLS检测
	finalDomain := ctx.Domain
	if ctx.Result.Network != nil && ctx.Result.Network.FinalDomain != "" {
		finalDomain = ctx.Result.Network.FinalDomain
	}

	// 执行综合TLS检测
	tlsResult := cts.performComprehensiveTLSDetection(ctx, finalDomain)

	// 设置所有TLS相关结果
	ctx.Result.TLS = tlsResult.TLS
	ctx.Result.SNI = tlsResult.SNI
	ctx.Result.Certificate = tlsResult.Certificate

	// 在TLS检测完成后，检查是否需要CDN检测
	if ctx.Result.CDN == nil {
		cdnResult := cts.performCDNDetection(ctx, finalDomain)
		ctx.Result.CDN = cdnResult
	}

	return nil
}

// ComprehensiveTLSResult 综合TLS检测结果
type ComprehensiveTLSResult struct {
	TLS         *types.TLSResult
	SNI         *types.SNIResult
	Certificate *types.CertificateResult
}

// performComprehensiveTLSDetection 执行综合TLS检测
func (cts *ComprehensiveTLSStage) performComprehensiveTLSDetection(ctx *types.PipelineContext, domain string) *ComprehensiveTLSResult {
	// 获取连接管理器
	connMgr, ok := ctx.Connections.(interface {
		GetTLSConnection(context.Context, string) (*tls.Conn, error)
		GetX25519TLSConnection(context.Context, string) (*tls.Conn, error)
		CloseTLSConnection(*tls.Conn)
	})
	if !ok {
		// 如果连接管理器不可用，回退到直接连接
		return cts.performDirectTLSDetection(domain)
	}

	// 第一次握手：正常TLS握手，检测TLS1.3、HTTP/2、SNI、证书
	startTime := time.Now()
	normalConn, err := connMgr.GetTLSConnection(ctx.Context, domain)
	if err != nil {
		// 连接失败时，normalConn可能为nil，不需要关闭
		return cts.createFailedResult(startTime)
	}

	// 获取第一次握手的结果
	normalState := normalConn.ConnectionState()
	handshakeTime := time.Since(startTime)

	// 分析第一次握手结果
	firstResult := cts.analyzeTLSState(normalState, domain, handshakeTime)

	// 关闭第一次握手的连接
	connMgr.CloseTLSConnection(normalConn)

	// 检查第一次握手的关键要求
	if !cts.checkCriticalRequirements(firstResult) {
		return firstResult
	}

	// 第二次握手：强制X25519握手，检测X25519支持
	supportsX25519 := cts.checkX25519Support(domain, 3*time.Second)

	// 更新TLS结果中的X25519支持
	firstResult.TLS.SupportsX25519 = supportsX25519

	return firstResult
}

// analyzeTLSState 分析TLS连接状态
func (cts *ComprehensiveTLSStage) analyzeTLSState(state tls.ConnectionState, domain string, handshakeTime time.Duration) *ComprehensiveTLSResult {
	// TLS检测
	supportsTLS13 := state.Version == tls.VersionTLS13
	supportsHTTP2 := state.NegotiatedProtocol == "h2" && state.NegotiatedProtocolIsMutual

	// SNI检测
	supportsSNI := true // 成功建立连接说明支持SNI
	sniMatch := false
	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]
		sniMatch = cert.VerifyHostname(domain) == nil
	}

	// 证书检测（真正的有效性检测）
	var certResult *types.CertificateResult
	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]
		now := time.Now()

		// 检查证书是否在有效期内
		inValidityPeriod := now.After(cert.NotBefore) && now.Before(cert.NotAfter)

		// 检查证书链是否可信
		chainTrusted := len(state.VerifiedChains) > 0

		// 检查主机名验证（已在上面完成）
		hostnameValid := sniMatch

		// 只有三者都满足才认为证书有效
		certValid := inValidityPeriod && chainTrusted && hostnameValid

		var daysUntilExpiry int
		if certValid {
			daysUntilExpiry = int(time.Until(cert.NotAfter).Hours() / 24)
		}

		certResult = &types.CertificateResult{
			Valid:           certValid,
			Issuer:          cert.Issuer.String(),
			Subject:         cert.Subject.String(),
			NotBefore:       cert.NotBefore,
			NotAfter:        cert.NotAfter,
			DaysUntilExpiry: daysUntilExpiry,
		}
	}

	return &ComprehensiveTLSResult{
		TLS: &types.TLSResult{
			ProtocolVersion: fmt.Sprintf("TLS %d.%d", (state.Version>>8)&0xFF, state.Version&0xFF),
			SupportsTLS13:   supportsTLS13,
			SupportsX25519:  false, // 将在第二次握手后更新
			SupportsHTTP2:   supportsHTTP2,
			CipherSuite:     tls.CipherSuiteName(state.CipherSuite),
			HandshakeTime:   handshakeTime,
		},
		SNI: &types.SNIResult{
			SupportsSNI: supportsSNI,
			SNIMatch:    sniMatch,
			ServerName:  domain,
		},
		Certificate: certResult,
	}
}

// performDirectTLSDetection 直接TLS检测（回退方案）
func (cts *ComprehensiveTLSStage) performDirectTLSDetection(domain string) *ComprehensiveTLSResult {
	// 简化的直接检测实现
	return &ComprehensiveTLSResult{
		TLS: &types.TLSResult{
			ProtocolVersion: "",
			SupportsTLS13:   false,
			SupportsX25519:  false,
			SupportsHTTP2:   false,
			CipherSuite:     "",
			HandshakeTime:   0,
		},
		SNI: &types.SNIResult{
			SupportsSNI: false,
			SNIMatch:    false,
			ServerName:  domain,
		},
		Certificate: nil,
	}
}

// createFailedResult 创建失败结果
func (cts *ComprehensiveTLSStage) createFailedResult(startTime time.Time) *ComprehensiveTLSResult {
	return &ComprehensiveTLSResult{
		TLS: &types.TLSResult{
			ProtocolVersion: "",
			SupportsTLS13:   false,
			SupportsX25519:  false,
			SupportsHTTP2:   false,
			CipherSuite:     "",
			HandshakeTime:   time.Since(startTime),
		},
		SNI: &types.SNIResult{
			SupportsSNI: false,
			SNIMatch:    false,
			ServerName:  "",
		},
		Certificate: nil,
	}
}

// checkX25519Support 检查X25519支持（正确的检测方法）
func (cts *ComprehensiveTLSStage) checkX25519Support(domain string, timeout time.Duration) bool {
	const port = ":443"

	// 专门做一次"仅X25519"的握手
	x25519Config := &tls.Config{
		ServerName:       domain,
		CurvePreferences: []tls.CurveID{tls.X25519}, // 强制仅使用X25519
		NextProtos:       []string{"h2", "http/1.1"},
		MinVersion:       tls.VersionTLS13, // 确保使用TLS1.3
		MaxVersion:       tls.VersionTLS13,
	}

	conn, err := tls.DialWithDialer(&net.Dialer{
		Timeout: timeout,
	}, "tcp", domain+port, x25519Config)

	if err != nil {
		// X25519握手失败，说明不支持X25519
		return false
	}
	defer conn.Close()

	// 检查连接状态
	state := conn.ConnectionState()

	// 握手成功且使用TLS1.3，说明支持X25519
	return state.Version == tls.VersionTLS13
}

// CanEarlyExit 是否可以早期退出
func (cts *ComprehensiveTLSStage) CanEarlyExit() bool {
	return false // TLS检测需要网络连接，不能早期退出
}

// Priority 优先级
func (cts *ComprehensiveTLSStage) Priority() int {
	return 4 // 综合TLS检测第四优先级
}

// Name 阶段名称
func (cts *ComprehensiveTLSStage) Name() string {
	return "comprehensive_tls"
}

// checkCriticalRequirements 检查关键要求
func (cts *ComprehensiveTLSStage) checkCriticalRequirements(result *ComprehensiveTLSResult) bool {
	// 检查TLS1.3支持
	if !result.TLS.SupportsTLS13 {
		return false
	}

	// 检查HTTP/2支持
	if !result.TLS.SupportsHTTP2 {
		return false
	}

	// 检查SNI匹配
	if !result.SNI.SNIMatch {
		return false
	}

	// 检查证书有效性
	if result.Certificate == nil || !result.Certificate.Valid {
		return false
	}

	return true
}

// performCDNDetection 执行证书CDN检测
func (cts *ComprehensiveTLSStage) performCDNDetection(ctx *types.PipelineContext, domain string) *types.CDNResult {
	// 只执行证书相关的CDN检测（低置信度）
	// 使用已有的证书信息，避免重复TLS连接
	if ctx.Result.Certificate != nil {
		// 检查证书签发者
		issuer := ctx.Result.Certificate.Issuer
		issuerLower := strings.ToLower(issuer)

		// 检查是否包含CDN特征
		cdnKeywords := []string{"cloudflare", "amazon", "google", "akamai"}
		for _, keyword := range cdnKeywords {
			if strings.Contains(issuerLower, keyword) {
				return &types.CDNResult{
					IsCDN:       true,
					CDNProvider: "CDN",
					Confidence:  "低",
					Evidence:    fmt.Sprintf("证书签发者提示: %s", issuer),
				}
			}
		}
	}

	return nil
}
