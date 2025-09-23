package detectors

import (
	"crypto/tls"
	"net"
	"time"

	"reality-checker-go/internal/types"
)

// SNIStage SNI检测阶段
type SNIStage struct{}

// NewSNIStage 创建SNI检测阶段
func NewSNIStage() *SNIStage {
	return &SNIStage{}
}

// Execute 执行SNI检测
func (ss *SNIStage) Execute(ctx *types.PipelineContext) error {
	// 使用最终域名进行SNI检测
	finalDomain := ctx.Domain
	if ctx.Result.Network != nil && ctx.Result.Network.FinalDomain != "" {
		finalDomain = ctx.Result.Network.FinalDomain
	}

	// 检测SNI支持
	supportsSNI, sniMatch := ss.checkSNI(finalDomain)
	
	ctx.Result.SNI = &types.SNIResult{
		SupportsSNI: supportsSNI,
		SNIMatch:    sniMatch,
		ServerName:  finalDomain,
	}

	return nil
}

// checkSNI 检查SNI支持
func (ss *SNIStage) checkSNI(domain string) (bool, bool) {
	// 建立TLS连接测试SNI
	conn, err := tls.DialWithDialer(&net.Dialer{
		Timeout: 5 * time.Second,
	}, "tcp", domain+":443", &tls.Config{
		ServerName: domain,
	})
	
	if err != nil {
		// 如果连接失败，假设不支持SNI
		return false, false
	}
	defer conn.Close()

	// 检查连接状态
	state := conn.ConnectionState()
	
	// 如果成功建立连接，说明支持SNI
	// 检查服务器名称是否匹配
	sniMatch := state.ServerName == domain
	
	return true, sniMatch
}

// CanEarlyExit 是否可以早期退出
func (ss *SNIStage) CanEarlyExit() bool {
	return true
}

// Priority 优先级
func (ss *SNIStage) Priority() int {
	return 6  // SNI检测第六优先级
}

// Name 阶段名称
func (ss *SNIStage) Name() string {
	return "sni"
}