package report

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"reality-checker-go/internal/types"
)

// TableFormatter 表格格式化器
type TableFormatter struct {
	config *types.Config
}

// NewTableFormatter 创建表格格式化器
func NewTableFormatter(config *types.Config) *TableFormatter {
	return &TableFormatter{
		config: config,
	}
}

// FormatSuitableTable 格式化适合域名的表格
func (tf *TableFormatter) FormatSuitableTable(results []*types.DetectionResult) string {
	var buf strings.Builder
	
	// 创建表
	t := table.NewWriter()
	t.SetOutputMirror(&buf)
	
	// 设置表头
	t.AppendHeader(table.Row{
		"最终域名", "TLS1.3", "X25519", "H2", "SNI匹配", "握手时间", "证书时间", "CDN", "热门", "推荐",
	})
	
	// 设置表格样式 - 正常边框
	t.SetStyle(table.StyleDefault)
	t.Style().Options.SeparateRows = true
	t.Style().Options.SeparateColumns = true
	t.Style().Options.DrawBorder = true
	t.Style().Options.SeparateHeader = true
	
	// 自定义颜色方案，适配深色背景
	t.Style().Color.Header = []text.Color{text.FgHiWhite, text.Bold}
	t.Style().Color.Row = []text.Color{text.FgWhite}
	t.Style().Color.Border = []text.Color{text.FgWhite}
	
	// 设置列对齐方式
	t.SetColumnConfigs([]table.ColumnConfig{
		{Name: "最终域名", Align: text.AlignLeft},
		{Name: "TLS1.3", Align: text.AlignCenter},
		{Name: "X25519", Align: text.AlignCenter},
		{Name: "H2", Align: text.AlignCenter},
		{Name: "SNI匹配", Align: text.AlignCenter},
		{Name: "握手时间", Align: text.AlignCenter},
		{Name: "证书时间", Align: text.AlignCenter},
		{Name: "CDN", Align: text.AlignCenter},
		{Name: "热门", Align: text.AlignCenter},
		{Name: "推荐", Align: text.AlignLeft},
	})
	
	// 添加数据行
	for _, result := range results {
		// 最终域名
		finalDomain := result.Domain
		if result.Network != nil && result.Network.FinalDomain != "" {
			finalDomain = result.Network.FinalDomain
		}
		
		// TLS1.3
		var tls13Text string
		if result.TLS != nil && result.TLS.SupportsTLS13 {
			tls13Text = text.FgGreen.Sprint("✓")
		} else {
			tls13Text = text.FgRed.Sprint("✗")
		}
		
		// X25519
		var x25519Text string
		if result.TLS != nil && result.TLS.SupportsX25519 {
			x25519Text = text.FgGreen.Sprint("✓")
		} else {
			x25519Text = text.FgRed.Sprint("✗")
		}
		
		// HTTP/2
		var h2Text string
		if result.TLS != nil && result.TLS.SupportsHTTP2 {
			h2Text = text.FgGreen.Sprint("✓")
		} else {
			h2Text = text.FgRed.Sprint("✗")
		}
		
		// SNI匹配
		var sniText string
		if result.SNI != nil && result.SNI.SNIMatch {
			sniText = text.FgGreen.Sprint("✓")
		} else {
			sniText = text.FgRed.Sprint("✗")
		}
		
		// 握手时间
		var handshakeText string
		if result.TLS != nil && result.TLS.HandshakeTime > 0 {
			handshakeMs := int(result.TLS.HandshakeTime.Milliseconds())
			handshakeText = fmt.Sprintf("%dms", handshakeMs)
			
			// 根据时间设置颜色
			if handshakeMs <= 200 {
				handshakeText = text.FgGreen.Sprint(handshakeText)
			} else if handshakeMs <= 500 {
				handshakeText = text.FgYellow.Sprint(handshakeText)
			} else {
				handshakeText = text.FgRed.Sprint(handshakeText)
			}
		} else {
			handshakeText = text.FgRed.Sprint("N/A")
		}
		
		// 证书时间
		var certText string
		if result.Certificate != nil && result.Certificate.Valid {
			days := result.Certificate.DaysUntilExpiry
			certText = fmt.Sprintf("%d天", days)
			
			// 根据剩余天数设置颜色
			if days >= 60 {
				certText = text.FgGreen.Sprint(certText)
			} else if days >= 30 {
				certText = text.FgYellow.Sprint(certText)
			} else {
				certText = text.FgRed.Sprint(certText)
			}
		} else {
			certText = text.FgRed.Sprint("无效")
		}
		
		// CDN
		var cdnText string
		if result.CDN != nil && result.CDN.IsCDN {
			confidence := result.CDN.Confidence
			cdnText = text.FgCyan.Sprint(confidence)
		} else {
			cdnText = text.FgRed.Sprint("无")
		}
		
		// 热门
		var hotText string
		if result.CDN != nil && result.CDN.IsHotWebsite {
			hotText = text.FgRed.Sprint("✓")
		} else {
			hotText = "-"
		}
		
		// 推荐星级计算
		recommendText := tf.calculateRecommendationStars(result)
		
		// 添加行数据
		t.AppendRow(table.Row{
			finalDomain,
			tls13Text,
			x25519Text,
			h2Text,
			sniText,
			handshakeText,
			certText,
			cdnText,
			hotText,
			recommendText,
		})
	}
	
	// 渲染输出
	t.Render()
	return buf.String()
}

// FormatUnsuitableSummary 格式化不适合域名的汇总
func (tf *TableFormatter) FormatUnsuitableSummary(results []*types.DetectionResult) string {
	if len(results) == 0 {
		return ""
	}
	
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("不适合的域名 (%d个):\n", len(results)))
	
	// 统计各种不适合的原因
	blockedCount := 0
	domesticCount := 0
	otherCount := 0
	
	for _, result := range results {
		if result.Error != nil {
			reason := result.Error.Error()
			if strings.Contains(reason, "域名被墙") {
				blockedCount++
			} else if strings.Contains(reason, "国内网站") {
				domesticCount++
			} else {
				otherCount++
			}
		}
	}
	
	// 显示统计信息
	if blockedCount > 0 {
		buf.WriteString(fmt.Sprintf("   - %d个域名被墙\n", blockedCount))
	}
	if domesticCount > 0 {
		buf.WriteString(fmt.Sprintf("   - %d个国内网站\n", domesticCount))
	}
	if otherCount > 0 {
		buf.WriteString(fmt.Sprintf("   - %d个其他原因\n", otherCount))
	}
	
ga	// 添加空行，与后续输出拉开距离
	buf.WriteString("\n")
	
	return buf.String()
}

// calculateRecommendationStars 计算推荐星级
func (tf *TableFormatter) calculateRecommendationStars(result *types.DetectionResult) string {
	stars := 0
	
	// 1. TLS硬性条件检查 (TLS1.3 + X25519 + H2 + SNI匹配)
	if result.TLS != nil && result.TLS.SupportsTLS13 && 
	   result.TLS.SupportsX25519 && result.TLS.SupportsHTTP2 &&
	   result.SNI != nil && result.SNI.SNIMatch {
		stars++
	}
	
	// 2. 握手时间延迟小 (<= 200ms)
	if result.TLS != nil && result.TLS.HandshakeTime > 0 {
		handshakeMs := int(result.TLS.HandshakeTime.Milliseconds())
		if handshakeMs <= 200 {
			stars++
		}
	}
	
	// 3. 没有CDN (不使用CDN更安全)
	if result.CDN == nil || !result.CDN.IsCDN {
		stars++
	}
	
	// 4. 不是热门网站 (热门网站不推荐作为Reality目标)
	if result.CDN != nil && !result.CDN.IsHotWebsite {
		stars++
	}
	
	// 5. 证书时间长 (>= 60天)
	if result.Certificate != nil && result.Certificate.Valid {
		if result.Certificate.DaysUntilExpiry >= 60 {
			stars++
		}
	}
	
	// 生成星级显示 - 只显示实际获得的星级
	var starsText string
	for i := 0; i < stars; i++ {
		starsText += text.FgYellow.Sprint("*")
	}
	
	return starsText
}