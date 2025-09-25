package report

import (
	"fmt"
	"strings"
	"time"

	"RealityChecker/internal/types"
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

	// 使用表格格式化器，和批量检测保持一致的格式
	tableFormatter := NewTableFormatter(f.config)

	// 显示域名检测结果表格（无论适合与否）
	output.WriteString("检测结果:\n\n")
	output.WriteString(tableFormatter.FormatSuitableTable([]*types.DetectionResult{result}))
	output.WriteString("\n")

	// 如果不适合，显示不适合的原因
	if !result.Suitable || result.Error != nil {
		var unsuitableResults []*types.DetectionResult
		unsuitableResults = append(unsuitableResults, result)
		output.WriteString(tableFormatter.FormatUnsuitableSummary(unsuitableResults))
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

	output.WriteString("批量检测报告\n")
	output.WriteString("总耗时: " + formattedDuration + "\n")
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

		// SNI信息
		if result.SNI != nil {
			output.WriteString(fmt.Sprintf(", SNI匹配=%t", result.SNI.SNIMatch))
		}

		// 证书信息
		if result.Certificate != nil {
			output.WriteString(fmt.Sprintf(", 证书有效=%t", result.Certificate.Valid))
		}

		// 地理位置
		if result.Location != nil {
			output.WriteString(fmt.Sprintf(", 位置=%s", result.Location.Country))
		}

		// CDN信息（批量检测中不显示详细特征）
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
