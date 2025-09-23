package detectors

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"reality-checker-go/internal/types"
)

// CDNStage CDN检测阶段
type CDNStage struct {
	cnameStrongSuffix    map[string]bool
	httpStrongHeader     map[string]bool
	httpMediumHeader     map[string]bool
	httpValueCdnDomains  map[string]bool
	asnStrongExact       map[string]bool
	nsHintSuffix         map[string]bool
	certIssuerHint       map[string]bool
	excludeServerTokens  map[string]bool
	excludeKeywordsGeneric map[string]bool
}

// NewCDNStage 创建CDN检测阶段
func NewCDNStage() *CDNStage {
	stage := &CDNStage{
		cnameStrongSuffix:    make(map[string]bool),
		httpStrongHeader:     make(map[string]bool),
		httpMediumHeader:     make(map[string]bool),
		httpValueCdnDomains:  make(map[string]bool),
		asnStrongExact:       make(map[string]bool),
		nsHintSuffix:         make(map[string]bool),
		certIssuerHint:       make(map[string]bool),
		excludeServerTokens:  make(map[string]bool),
		excludeKeywordsGeneric: make(map[string]bool),
	}
	stage.loadCDNKeywords()
	return stage
}

// Execute 执行CDN检测
func (cs *CDNStage) Execute(ctx *types.PipelineContext) error {
	// 使用最终域名进行CDN检测
	finalDomain := ctx.Domain
	if ctx.Result.Network != nil && ctx.Result.Network.FinalDomain != "" {
		finalDomain = ctx.Result.Network.FinalDomain
	}
	
	// 检测CDN
	isCDN, provider, confidence, evidence := cs.detectCDN(finalDomain, ctx.Result.Network)
	
	ctx.Result.CDN = &types.CDNResult{
		IsCDN:       isCDN,
		CDNProvider: provider,
		Confidence:  confidence,
		Evidence:    evidence,
	}

	return nil
}

// detectCDN 检测CDN
// 使用多种检测方法，按置信度从高到低进行检测
// 高置信度方法：CNAME记录、HTTP响应头、ASN查询等
// 中等置信度方法：NS记录、通用HTTP头等  
// 低置信度方法：证书签发者等
func (cs *CDNStage) detectCDN(domain string, networkResult *types.NetworkResult) (bool, string, string, string) {
	// 高置信度检测方法（优先级顺序）
	highConfidenceChecks := []func() (string, string){
		func() (string, string) { return cs.checkCNAMEStrongSuffix(domain) },
		func() (string, string) { return cs.checkHTTPStrongHeader(networkResult) },
		func() (string, string) { return cs.checkHTTPValueCdnDomains(networkResult) },
		func() (string, string) { return cs.checkASNStrongExact(domain) },
	}
	
	// 中等置信度检测方法
	mediumConfidenceChecks := []func() (string, string){
		func() (string, string) { return cs.checkNSHintSuffix(domain) },
		func() (string, string) { return cs.checkHTTPMediumHeader(networkResult) },
	}
	
	// 低置信度检测方法
	lowConfidenceChecks := []func() (string, string){
		func() (string, string) { return cs.checkCertIssuerHint(domain) },
	}
	
	// 按置信度顺序检测
	for _, check := range highConfidenceChecks {
		if provider, evidence := check(); provider != "" {
			return true, provider, "高", evidence
		}
	}
	
	for _, check := range mediumConfidenceChecks {
		if provider, evidence := check(); provider != "" {
			return true, provider, "中", evidence
		}
	}
	
	for _, check := range lowConfidenceChecks {
		if provider, evidence := check(); provider != "" {
			return true, provider, "低", evidence
		}
	}
	
	return false, "", "", ""
}

// checkCNAMEStrongSuffix 检查CNAME强后缀特征
func (cs *CDNStage) checkCNAMEStrongSuffix(domain string) (string, string) {
	// 使用Go原生DNS解析器查询CNAME记录
	cname, err := net.LookupCNAME(domain)
	if err != nil {
		return "", ""
	}
	
	cnameLower := strings.ToLower(cname)
	
	// 检查CNAME记录是否包含CDN后缀
	cnameClean := strings.TrimSuffix(cnameLower, ".")
	
	for suffix := range cs.cnameStrongSuffix {
		// 移除注释部分
		cleanSuffix := strings.Split(suffix, "#")[0]
		cleanSuffix = strings.TrimSpace(cleanSuffix)
		suffixLower := strings.ToLower(cleanSuffix)
		if strings.Contains(cnameClean, suffixLower) {
			provider := cs.getProviderFromSuffix(suffix)
			return provider, fmt.Sprintf("CNAME记录特征: CNAME记录包含%s", cleanSuffix)
		}
	}
	
	return "", ""
}

// checkHTTPStrongHeader 检查HTTP强响应头特征
func (cs *CDNStage) checkHTTPStrongHeader(networkResult *types.NetworkResult) (string, string) {
	if networkResult == nil || networkResult.Headers == nil {
		return "", ""
	}
	
	// 检查HTTP强响应头
	for headerName := range cs.httpStrongHeader {
		// 检查header名称
		for respHeaderName, respHeaderValue := range networkResult.Headers {
			if strings.EqualFold(respHeaderName, headerName) {
				provider := cs.getProviderFromHeader(headerName)
				return provider, fmt.Sprintf("HTTP强响应头特征: %s=%s", respHeaderName, respHeaderValue)
			}
		}
		
		// 检查server头（特殊处理）
		if strings.HasPrefix(headerName, "server: ") {
			serverValue := strings.TrimPrefix(headerName, "server: ")
			if serverHeader, exists := networkResult.Headers["Server"]; exists {
				if strings.Contains(strings.ToLower(serverHeader), strings.ToLower(serverValue)) {
					provider := cs.getProviderFromHeader(headerName)
					return provider, fmt.Sprintf("HTTP强响应头特征: Server=%s", serverHeader)
				}
			}
		}
	}
	
	return "", ""
}

// checkASNStrongExact 检查ASN强特征
func (cs *CDNStage) checkASNStrongExact(domain string) (string, string) {
	// TODO: 需要ASN查询功能
	// 实际实现需要集成ASN查询库或API来查询真实的ASN信息
	// 当前使用关键字库中的ASN列表，但需要真实的ASN查询功能
	
	// 解析IP地址
	ips, err := net.LookupIP(domain)
	if err != nil || len(ips) == 0 {
		return "", ""
	}
	
	// 获取第一个IPv4地址
	var ip net.IP
	for _, i := range ips {
		if i.To4() != nil {
			ip = i
			break
		}
	}
	if ip == nil {
		ip = ips[0]
	}
	
	// 这里需要真实的ASN查询，暂时返回空
	// 未来需要集成ASN查询库，然后与cs.asnStrongExact中的关键字比较
	
	return "", ""
}

// checkHTTPValueCdnDomains 检查HTTP头值中的CDN域名
func (cs *CDNStage) checkHTTPValueCdnDomains(networkResult *types.NetworkResult) (string, string) {
	if networkResult == nil || networkResult.Headers == nil {
		return "", ""
	}
	
	// 检查HTTP响应头值中是否包含CDN域名
	for headerName, respHeaderValue := range networkResult.Headers {
		respHeaderValueLower := strings.ToLower(respHeaderValue)
		
		for cdnDomain := range cs.httpValueCdnDomains {
			// 移除注释部分
			cleanDomain := strings.Split(cdnDomain, "#")[0]
			cleanDomain = strings.TrimSpace(cleanDomain)
			
			if strings.Contains(respHeaderValueLower, strings.ToLower(cleanDomain)) {
				provider := cs.getProviderFromHeader(cdnDomain)
				return provider, fmt.Sprintf("HTTP头值CDN域名特征: %s包含%s", headerName, cleanDomain)
			}
		}
	}
	
	return "", ""
}

// checkNSHintSuffix 检查NS提示
func (cs *CDNStage) checkNSHintSuffix(domain string) (string, string) {
	// 查询NS记录
	nsRecords, err := net.LookupNS(domain)
	if err != nil {
		return "", ""
	}
	
	for _, ns := range nsRecords {
		nsHost := strings.ToLower(ns.Host)
		for hint := range cs.nsHintSuffix {
			if strings.Contains(nsHost, strings.ToLower(hint)) {
				provider := cs.getProviderFromNShint(hint)
				return provider, fmt.Sprintf("NS记录: %s", ns.Host)
			}
		}
	}
	
	return "", ""
}

// checkCertIssuerHint 检查证书签发者提示
func (cs *CDNStage) checkCertIssuerHint(domain string) (string, string) {
	const (
		certPort = ":443"
		certTimeout = 5 * time.Second  // 增加CDN证书检测超时时间，减少误判
	)
	
	// 建立TLS连接获取证书
	conn, err := tls.DialWithDialer(&net.Dialer{
		Timeout: certTimeout,
	}, "tcp", domain+certPort, &tls.Config{
		ServerName: domain,
	})
	if err != nil {
		return "", ""
	}
	defer conn.Close()
	
	// 获取证书
	cert := conn.ConnectionState().PeerCertificates[0]
	
	// 检查证书签发者
	issuer := cert.Issuer.String()
	issuerLower := strings.ToLower(issuer)
	
	// 使用关键字库中的证书签发者列表
	for certIssuer := range cs.certIssuerHint {
		// 移除注释部分
		cleanIssuer := strings.Split(certIssuer, "#")[0]
		cleanIssuer = strings.TrimSpace(cleanIssuer)
		
		if strings.Contains(issuerLower, strings.ToLower(cleanIssuer)) {
			provider := cs.getProviderFromHeader(certIssuer)
			return provider, fmt.Sprintf("证书签发者提示: %s", issuer)
		}
	}
	
	return "", ""
}

// checkHTTPMediumHeader 检查HTTP中等响应头
func (cs *CDNStage) checkHTTPMediumHeader(networkResult *types.NetworkResult) (string, string) {
	if networkResult == nil || networkResult.Headers == nil {
		return "", ""
	}
	
	// 检查HTTP中等响应头
	for headerName := range cs.httpMediumHeader {
		for respHeaderName, respHeaderValue := range networkResult.Headers {
			if strings.EqualFold(respHeaderName, headerName) {
				provider := cs.getProviderFromHeader(headerName)
				return provider, fmt.Sprintf("HTTP中等响应头特征: %s=%s", respHeaderName, respHeaderValue)
			}
		}
	}
	
	return "", ""
}

// getProviderFromSuffix 根据后缀获取CDN提供商
func (cs *CDNStage) getProviderFromSuffix(suffix string) string {
	// 不使用硬编码，直接返回检测到的后缀作为证据
	// 具体的CDN提供商信息由关键字库提供
	return "CDN"
}

// getProviderFromNShint 根据NS提示获取CDN提供商
func (cs *CDNStage) getProviderFromNShint(hint string) string {
	// 不使用硬编码，直接返回检测到的NS提示作为证据
	// 具体的CDN提供商信息由关键字库提供
	return "CDN"
}

// getProviderFromHeader 根据HTTP头获取CDN提供商
func (cs *CDNStage) getProviderFromHeader(header string) string {
	// 不使用硬编码，直接返回检测到的头信息作为证据
	// 具体的CDN提供商信息由关键字库提供
	return "CDN"
}

// loadCDNKeywords 加载CDN关键词
func (cs *CDNStage) loadCDNKeywords() {
	file, err := os.Open("data/cdn_keywords.txt")
	if err != nil {
		fmt.Printf("[WARNING] 无法加载CDN关键字文件: %v\n", err)
		return
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	currentSection := ""
	loadedCount := 0
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// 检查是否是节标题
		if strings.HasSuffix(line, ":") {
			currentSection = line
			continue
		}
		
		// 根据节标题分类加载
		switch currentSection {
		case "cname_strong_suffix:":
			cs.cnameStrongSuffix[line] = true
			loadedCount++
		case "http_strong_header:":
			cs.httpStrongHeader[line] = true
			loadedCount++
		case "http_medium_header:":
			cs.httpMediumHeader[line] = true
			loadedCount++
		case "http_value_cdn_domains:":
			cs.httpValueCdnDomains[line] = true
			loadedCount++
		case "asn_strong_exact:":
			cs.asnStrongExact[line] = true
			loadedCount++
		case "ns_hint_suffix:":
			cs.nsHintSuffix[line] = true
			loadedCount++
		case "cert_issuer_hint:":
			cs.certIssuerHint[line] = true
			loadedCount++
		case "exclude_server_tokens:":
			cs.excludeServerTokens[line] = true
		case "exclude_keywords_generic:":
			cs.excludeKeywordsGeneric[line] = true
		}
	}
	
	if err := scanner.Err(); err != nil {
		fmt.Printf("[WARNING] 读取CDN关键字文件时出错: %v\n", err)
		return
	}
	
	// 成功加载关键字（静默模式）
}

// CanEarlyExit 是否可以早期退出
func (cs *CDNStage) CanEarlyExit() bool {
	return false
}

// Priority 优先级
func (cs *CDNStage) Priority() int {
	return 8  // CDN检测第八优先级 - 信息性检测
}

// Name 阶段名称
func (cs *CDNStage) Name() string {
	return "cdn"
}