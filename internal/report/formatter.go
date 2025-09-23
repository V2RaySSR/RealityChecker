package report

import (
	"fmt"
	"strings"
	"time"

	"reality-checker-go/internal/types"
)

// Formatter 报告格式化器
type Formatter struct {
	config *types.Config
}

// NewFormatter 创建格式化器
func NewFormatter(config *types.Config) *Formatter {
	return &Formatter{
		config: config,
	}
}

// FormatSingleResult 格式化单个检测结果
func (f *Formatter) FormatSingleResult(result *types.DetectionResult) string {
	var output strings.Builder
	
	// 基本信息
	output.WriteString(fmt.Sprintf("域名: %s\n", result.Domain))
	output.WriteString(fmt.Sprintf("适合性: %t\n", result.Suitable))
	
	// 耗时
	durationSeconds := int(result.Duration.Seconds())
	output.WriteString(fmt.Sprintf("耗时: %ds\n", durationSeconds))
	
	// 错误信息
	if result.Error != nil {
		output.WriteString(fmt.Sprintf("错误: %v\n", result.Error))
	}
	
	// 网络信息
	if result.Network != nil {
		output.WriteString(fmt.Sprintf("网络: 可达=%t, 状态码=%d\n", 
			result.Network.Accessible, result.Network.StatusCode))
		
		if result.Network.IsRedirected {
			output.WriteString(fmt.Sprintf("重定向: %s -> %s (跳转%d次)\n", 
				result.Domain, result.Network.FinalDomain, result.Network.RedirectCount))
			if len(result.Network.RedirectChain) > 1 {
				output.WriteString(fmt.Sprintf("重定向链: %s\n", 
					strings.Join(result.Network.RedirectChain, " -> ")))
			}
		} else {
			output.WriteString(fmt.Sprintf("最终域名: %s\n", result.Network.FinalDomain))
		}
	}
	
	// TLS信息
	if result.TLS != nil {
		tlsInfo := fmt.Sprintf("TLS: 1.3=%t, X25519=%t, HTTP2=%t", 
			result.TLS.SupportsTLS13, result.TLS.SupportsX25519, result.TLS.SupportsHTTP2)
		
		if result.TLS.HandshakeTime > 0 {
			handshakeMs := int(result.TLS.HandshakeTime.Milliseconds())
			tlsInfo += fmt.Sprintf(", 握手时间=%dms", handshakeMs)
		}
		
		output.WriteString(fmt.Sprintf("%s\n", tlsInfo))
	}
	
	// 地理位置
	if result.Location != nil {
		output.WriteString(fmt.Sprintf("地理位置: %s, 国内=%t\n", 
			result.Location.Country, result.Location.IsDomestic))
	}
	
	// 被墙检测
	if result.Blocked != nil {
		output.WriteString(fmt.Sprintf("被墙: %t\n", result.Blocked.IsBlocked))
	}
	
	// CDN信息
	if result.CDN != nil {
		if result.CDN.IsCDN {
			output.WriteString(fmt.Sprintf("CDN: 是, 提供商=%s, 置信度=%s, 证据=%s\n", 
				result.CDN.CDNProvider, result.CDN.Confidence, result.CDN.Evidence))
		} else {
			output.WriteString("CDN: 否\n")
		}
		if result.CDN.IsHotWebsite {
			output.WriteString("热门网站: 是\n")
		}
	}
	
	return output.String()
}

// FormatBatchResult 格式化批量检测结果
func (f *Formatter) FormatBatchResult(results []*types.DetectionResult, totalDuration time.Duration) string {
	var output strings.Builder
	
	// 统计信息
	suitableCount := 0
	successCount := 0
	
	for _, result := range results {
		if result.Error == nil {
			successCount++
		}
		if result.Suitable {
			suitableCount++
		}
	}
	
	totalCount := len(results)
	successRate := float64(successCount) / float64(totalCount) * 100
	suitableRate := float64(suitableCount) / float64(totalCount) * 100
	
	// 格式化耗时
	formattedDuration := f.formatDuration(totalDuration)
	
	output.WriteString(fmt.Sprintf("批量检测报告\n"))
	output.WriteString(fmt.Sprintf("总耗时: %s\n", formattedDuration))
	output.WriteString(fmt.Sprintf("检测域名: %d 个\n", totalCount))
	output.WriteString(fmt.Sprintf("成功率: %.1f%%\n", successRate))
	output.WriteString(fmt.Sprintf("适合性率: %.1f%%\n", suitableRate))
	output.WriteString("\n")
	
	// 详细结果
	output.WriteString("详细结果:\n")
	for i, result := range results {
		output.WriteString(fmt.Sprintf("%d. %s: 适合=%t", i+1, result.Domain, result.Suitable))
		
		// 网络信息
		if result.Network != nil {
			if result.Network.IsRedirected {
				output.WriteString(fmt.Sprintf(", 重定向: %s->%s, 状态码=%d", 
					result.Domain, result.Network.FinalDomain, result.Network.StatusCode))
			} else {
				output.WriteString(fmt.Sprintf(", 状态码=%d", result.Network.StatusCode))
			}
		}
		
		// TLS信息
		if result.TLS != nil {
			output.WriteString(fmt.Sprintf(", TLS1.3=%t, X25519=%t, HTTP2=%t", 
				result.TLS.SupportsTLS13, result.TLS.SupportsX25519, result.TLS.SupportsHTTP2))
			
			if result.TLS.HandshakeTime > 0 {
				handshakeMs := int(result.TLS.HandshakeTime.Milliseconds())
				output.WriteString(fmt.Sprintf(", 握手时间=%dms", handshakeMs))
			}
		}
		
		// 地理位置
		if result.Location != nil {
			output.WriteString(fmt.Sprintf(", 位置=%s", result.Location.Country))
		}
		
		// CDN信息
		if result.CDN != nil && result.CDN.IsCDN {
			output.WriteString(fmt.Sprintf(", CDN=%s(%s)", result.CDN.CDNProvider, result.CDN.Confidence))
		}
		
		// 错误信息
		if result.Error != nil {
			output.WriteString(fmt.Sprintf(", 错误=%v", result.Error))
		}
		
		output.WriteString("\n")
	}
	
	return output.String()
}

// formatDuration 格式化持续时间
func (f *Formatter) formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Nanoseconds())/1e6)
	} else if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	} else {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
}

// FormatProgress 格式化进度信息
func (f *Formatter) FormatProgress(current, total int, domain string, status string, reason string) string {
	progress := fmt.Sprintf("正在检测 [%d/%d]: %s... %s", current, total, domain, status)
	if reason != "" {
		progress += fmt.Sprintf(" - %s", reason)
	}
	return progress
}
