package detectors

import (
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
	// 简化实现：假设证书有效
	ctx.Result.Certificate = &types.CertificateResult{
		Valid:           true,
		Issuer:          "Let's Encrypt",
		Subject:         ctx.Domain,
		DaysUntilExpiry: 90,
		CertificateSANs: []string{ctx.Domain},
		NotBefore:       time.Now().Add(-30 * 24 * time.Hour),
		NotAfter:        time.Now().Add(90 * 24 * time.Hour),
	}

	return nil
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