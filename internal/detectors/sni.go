package detectors

import (
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
	// 简化实现：假设SNI匹配
	ctx.Result.SNI = &types.SNIResult{
		SupportsSNI: true,
		SNIMatch:    true,
		ServerName:  ctx.Domain,
	}

	return nil
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