package detectors

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"reality-checker-go/internal/types"
)

// CertificateStage 证书检测阶段
type CertificateStage struct{}

// NewCertificateStage 创建证书检测阶段
func NewCertificateStage() *CertificateStage {
	return &CertificateStage{}
}

// Execute 执行证书检测
func (cs *CertificateStage) Execute(ctx *types.PipelineContext) error {
	// 使用最终域名进行证书检测
	finalDomain := ctx.Domain
	if ctx.Result.Network != nil && ctx.Result.Network.FinalDomain != "" {
		finalDomain = ctx.Result.Network.FinalDomain
	}

	// 检测证书
	certResult := cs.checkCertificate(finalDomain)
	ctx.Result.Certificate = certResult

	return nil
}

// checkCertificate 检查证书
func (cs *CertificateStage) checkCertificate(domain string) *types.CertificateResult {
	// 建立TLS连接获取证书
	conn, err := tls.DialWithDialer(&net.Dialer{
		Timeout: 5 * time.Second,
	}, "tcp", domain+":443", &tls.Config{
		ServerName: domain,
	})
	
	if err != nil {
		return &types.CertificateResult{
			Valid: false,
			Error: fmt.Sprintf("证书检测失败: %v", err),
		}
	}
	defer conn.Close()

	// 获取证书
	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return &types.CertificateResult{
			Valid: false,
			Error: "未找到证书",
		}
	}

	cert := state.PeerCertificates[0]
	now := time.Now()
	
	// 计算证书有效性
	valid := now.After(cert.NotBefore) && now.Before(cert.NotAfter)
	daysUntilExpiry := int(cert.NotAfter.Sub(now).Hours() / 24)
	
	// 获取证书SANs
	var sans []string
	for _, san := range cert.DNSNames {
		sans = append(sans, san)
	}
	
	// 如果没有SANs，使用CommonName
	if len(sans) == 0 && cert.Subject.CommonName != "" {
		sans = []string{cert.Subject.CommonName}
	}

	return &types.CertificateResult{
		Valid:           valid,
		Issuer:          cert.Issuer.String(),
		Subject:         cert.Subject.String(),
		DaysUntilExpiry: daysUntilExpiry,
		CertificateSANs: sans,
		NotBefore:       cert.NotBefore,
		NotAfter:        cert.NotAfter,
	}
}

// CanEarlyExit 是否可以早期退出
func (cs *CertificateStage) CanEarlyExit() bool {
	return true
}

// Priority 优先级
func (cs *CertificateStage) Priority() int {
	return 7  // 证书检测第七优先级
}

// Name 阶段名称
func (cs *CertificateStage) Name() string {
	return "certificate"
}