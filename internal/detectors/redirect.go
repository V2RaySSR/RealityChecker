package detectors

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"reality-checker-go/internal/types"
)

// RedirectStage 重定向检测阶段
type RedirectStage struct{}

// NewRedirectStage 创建重定向检测阶段
func NewRedirectStage() *RedirectStage {
	return &RedirectStage{}
}

// Execute 执行重定向检测
func (rs *RedirectStage) Execute(ctx *types.PipelineContext) error {
	// 创建HTTP客户端，禁用自动重定向
	client := &http.Client{
		Timeout: 5 * time.Second,  // 减少HTTP客户端超时时间
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// 跟踪重定向
	result := rs.followRedirects(client, ctx.Domain)
	
	// 设置网络结果
	ctx.Result.Network = &types.NetworkResult{
		Accessible:     result.Accessible,
		StatusCode:     result.StatusCode,
		FinalDomain:    result.FinalDomain,
		RedirectChain:  result.RedirectChain,
		IsRedirected:   result.IsRedirected,
		RedirectCount:  result.RedirectCount,
		URL:           result.URL,
		ResponseTime:  time.Since(ctx.StartTime),
		Headers:       result.Headers, // 保存HTTP响应头
	}

	// 更新域名
	if result.IsRedirected {
		ctx.Domain = result.FinalDomain
	}

	return nil
}

// RedirectResult 重定向结果
type RedirectResult struct {
	Accessible     bool
	StatusCode     int
	FinalDomain    string
	RedirectChain  []string
	IsRedirected   bool
	RedirectCount  int
	URL            string
	Headers        map[string]string // HTTP响应头
}

// followRedirects 跟踪重定向
func (rs *RedirectStage) followRedirects(client *http.Client, domain string) *RedirectResult {
	const (
		maxRedirects = 5
		httpsScheme  = "https://"
	)
	
	result := &RedirectResult{
		Accessible:    false,
		StatusCode:    0,
		FinalDomain:   domain,
		RedirectChain: []string{domain},
		IsRedirected:  false,
		RedirectCount: 0,
		URL:          httpsScheme + domain,
	}

	currentURL := httpsScheme + domain

	for i := 0; i < maxRedirects; i++ {
		req, err := http.NewRequest("GET", currentURL, nil)
		if err != nil {
			break
		}

		// 添加浏览器头
		const (
			userAgent      = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
			acceptHeader   = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
			acceptLanguage = "en-US,en;q=0.9"
		)
		
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept", acceptHeader)
		req.Header.Set("Accept-Language", acceptLanguage)

		resp, err := client.Do(req)
		if err != nil {
			break
		}

		result.Accessible = true
		result.StatusCode = resp.StatusCode
		result.URL = currentURL
		
		// 保存HTTP响应头
		result.Headers = make(map[string]string)
		for name, values := range resp.Header {
			if len(values) > 0 {
				result.Headers[name] = values[0] // 取第一个值
			}
		}

		// 检查是否有重定向
		const (
			redirectMin = 300
			redirectMax = 400
		)
		if resp.StatusCode >= redirectMin && resp.StatusCode < redirectMax {
			location := resp.Header.Get("Location")
			if location != "" {
				// 处理相对URL
				const (
					rootPathPrefix = "/"
					httpPrefix     = "http"
				)
				
				if strings.HasPrefix(location, rootPathPrefix) {
					parsedURL, _ := url.Parse(currentURL)
					location = parsedURL.Scheme + "://" + parsedURL.Host + location
				} else if !strings.HasPrefix(location, httpPrefix) {
					location = httpsScheme + location
				}

				parsedLocation, err := url.Parse(location)
				if err == nil {
					newDomain := parsedLocation.Hostname()
					if newDomain != domain && newDomain != "" {
						result.RedirectChain = append(result.RedirectChain, newDomain)
						result.IsRedirected = true
						result.RedirectCount++
						currentURL = location
						domain = newDomain
						resp.Body.Close()
						continue
					}
				}
			}
		}

		// 没有重定向或重定向结束
		parsedURL, _ := url.Parse(currentURL)
		result.FinalDomain = parsedURL.Hostname()
		resp.Body.Close()
		break
	}

	return result
}

// CanEarlyExit 是否可以早期退出
func (rs *RedirectStage) CanEarlyExit() bool {
	return false
}

// Priority 优先级
func (rs *RedirectStage) Priority() int {
	return 2  // 重定向检测第二优先级
}

// Name 阶段名称
func (rs *RedirectStage) Name() string {
	return "redirect"
}