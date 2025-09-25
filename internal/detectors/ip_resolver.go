package detectors

import (
	"context"
	"fmt"
	"net"
	"time"

	"RealityChecker/internal/types"
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

	// 快速连通性测试
	if !irs.quickConnectivityTest(ip) {
		return fmt.Errorf("网络不可达")
	}

	// 设置IP地址到Location结果中
	if ctx.Result.Location == nil {
		ctx.Result.Location = &types.LocationResult{}
	}
	ctx.Result.Location.IPAddress = ip

	return nil
}

// quickConnectivityTest 快速连通性测试
func (irs *IPResolverStage) quickConnectivityTest(ip string) bool {
	// 测试HTTPS端口443的连通性
	conn, err := net.DialTimeout("tcp", ip+":443", 2*time.Second)
	if err != nil {
		// 如果HTTPS不可达，尝试HTTP端口80
		conn, err = net.DialTimeout("tcp", ip+":80", 2*time.Second)
		if err != nil {
			return false
		}
	}
	conn.Close()
	return true
}

// resolveIP 解析IP地址
func (irs *IPResolverStage) resolveIP(domain string) (string, error) {
	// 检查是否已经是IP地址
	if net.ParseIP(domain) != nil {
		return domain, nil
	}

	// 使用自定义DNS解析器，设置更短的超时
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 2 * time.Second, // DNS查询超时2秒
			}
			return d.DialContext(ctx, network, address)
		},
	}

	// 解析域名
	ips, err := resolver.LookupIPAddr(context.Background(), domain)
	if err != nil {
		return "", err
	}

	if len(ips) == 0 {
		return "", fmt.Errorf("未找到IP地址")
	}

	// 优先选择IPv4地址
	for _, ipAddr := range ips {
		if ipAddr.IP.To4() != nil {
			return ipAddr.IP.String(), nil
		}
	}

	// 如果没有IPv4，使用IPv6
	return ips[0].IP.String(), nil
}

// CanEarlyExit 是否可以早期退出
func (irs *IPResolverStage) CanEarlyExit() bool {
	return true
}

// Priority 优先级
func (irs *IPResolverStage) Priority() int {
	return 3 // IP解析第三优先级
}

// Name 阶段名称
func (irs *IPResolverStage) Name() string {
	return "ip_resolver"
}
