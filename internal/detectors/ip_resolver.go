package detectors

import (
	"fmt"
	"net"

	"reality-checker-go/internal/types"
)

// IPResolverStage IP解析阶段
type IPResolverStage struct{}

// NewIPResolverStage 创建IP解析阶段
func NewIPResolverStage() *IPResolverStage {
	return &IPResolverStage{}
}

// Execute 执行IP解析
func (irs *IPResolverStage) Execute(ctx *types.PipelineContext) error {
	// 解析IP地址
	ip, err := irs.resolveIP(ctx.Domain)
	if err != nil {
		return fmt.Errorf("IP解析失败: %v", err)
	}

	// 设置IP地址到Location结果中
	if ctx.Result.Location == nil {
		ctx.Result.Location = &types.LocationResult{}
	}
	ctx.Result.Location.IPAddress = ip

	return nil
}

// resolveIP 解析IP地址
func (irs *IPResolverStage) resolveIP(domain string) (string, error) {
	// 检查是否已经是IP地址
	if net.ParseIP(domain) != nil {
		return domain, nil
	}

	// 解析域名
	ips, err := net.LookupIP(domain)
	if err != nil {
		return "", err
	}
	
	if len(ips) == 0 {
		return "", fmt.Errorf("未找到IP地址")
	}
	
	// 优先选择IPv4地址
	for _, ip := range ips {
		if ip.To4() != nil {
			return ip.String(), nil
		}
	}
	
	// 如果没有IPv4，使用IPv6
	return ips[0].String(), nil
}

// CanEarlyExit 是否可以早期退出
func (irs *IPResolverStage) CanEarlyExit() bool {
	return true
}

// Priority 优先级
func (irs *IPResolverStage) Priority() int {
	return 3  // IP解析第三优先级
}

// Name 阶段名称
func (irs *IPResolverStage) Name() string {
	return "ip_resolver"
}
