package detectors

import (
	"RealityChecker/internal/types"
)

// StatusCheckStage 状态码检查阶段
// 检查HTTP状态码是否符合Reality要求
type StatusCheckStage struct{}

// NewStatusCheckStage 创建状态码检查阶段
func NewStatusCheckStage() *StatusCheckStage {
	return &StatusCheckStage{}
}

// Execute 执行状态码检查
func (scs *StatusCheckStage) Execute(ctx *types.PipelineContext) error {
	// 检查网络结果是否存在
	if ctx.Result.Network == nil {
		return nil
	}

	// 获取状态码
	statusCode := ctx.Result.Network.StatusCode
	accessible := ctx.Result.Network.Accessible

	// 分类状态码
	category := types.ClassifyStatusCode(statusCode, accessible)
	ctx.Result.StatusCodeCategory = category

	return nil
}

// CanEarlyExit 是否可以早期退出
func (scs *StatusCheckStage) CanEarlyExit() bool {
	return true // 状态码检查可以早期退出
}

// Priority 优先级
func (scs *StatusCheckStage) Priority() int {
	return 3 // 状态码检查第三优先级（在重定向检测之后）
}

// Name 阶段名称
func (scs *StatusCheckStage) Name() string {
	return "status_check"
}
