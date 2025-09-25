package detectors

import (
	"RealityChecker/internal/types"
)

// LocationCheckStage 地理位置检查阶段
// 检查IP地理位置是否符合Reality要求
type LocationCheckStage struct{}

// NewLocationCheckStage 创建地理位置检查阶段
func NewLocationCheckStage() *LocationCheckStage {
	return &LocationCheckStage{}
}

// Execute 执行地理位置检查
func (lcs *LocationCheckStage) Execute(ctx *types.PipelineContext) error {
	// 检查地理位置结果是否存在
	if ctx.Result.Location == nil {
		return nil
	}

	// 检查是否为中国
	if ctx.Result.Location.IsDomestic {
		// 标记为不适合，但不立即终止，让Pipeline继续执行
		// 最终结果会在Batch Manager中处理
	}

	return nil
}

// CanEarlyExit 是否可以早期退出
func (lcs *LocationCheckStage) CanEarlyExit() bool {
	return true // 地理位置检查可以早期退出
}

// Priority 优先级
func (lcs *LocationCheckStage) Priority() int {
	return 5 // 地理位置检查第五优先级（在IP解析之后）
}

// Name 阶段名称
func (lcs *LocationCheckStage) Name() string {
	return "location_check"
}
