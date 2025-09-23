package types

import (
	"time"
)

// DetectionResult 检测结果
type DetectionResult struct {
	Domain            string            `json:"domain"`
	Index             int               `json:"index"`
	StartTime         time.Time         `json:"start_time"`
	Duration          time.Duration     `json:"duration"`
	Suitable          bool              `json:"suitable"`
	Error             error             `json:"error,omitempty"`
	HardRequirementsMet bool            `json:"hard_requirements_met"`
	
	// 检测结果
	Network      *NetworkResult      `json:"network,omitempty"`
	TLS          *TLSResult          `json:"tls,omitempty"`
	Certificate  *CertificateResult  `json:"certificate,omitempty"`
	SNI          *SNIResult          `json:"sni,omitempty"`
	CDN          *CDNResult          `json:"cdn,omitempty"`
	Blocked      *BlockedResult      `json:"blocked,omitempty"`
	Location     *LocationResult     `json:"location,omitempty"`
	Summary      *DetectionSummary   `json:"summary,omitempty"`
}

// NetworkResult 网络检测结果
type NetworkResult struct {
	Accessible     bool          `json:"accessible"`
	ResponseTime   time.Duration `json:"response_time"`
	StatusCode     int           `json:"status_code"`
	FinalDomain    string        `json:"final_domain"`
	RedirectChain  []string      `json:"redirect_chain"`
	IsRedirected   bool          `json:"is_redirected"`
	RedirectCount  int           `json:"redirect_count"`
	URL            string        `json:"url"`
	HandshakeTime  time.Duration `json:"handshake_time"`
	Headers        map[string]string `json:"headers,omitempty"` // HTTP响应头
}

// TLSResult TLS检测结果
type TLSResult struct {
	ProtocolVersion string        `json:"protocol_version"`
	SupportsTLS13   bool          `json:"supports_tls13"`
	SupportsX25519  bool          `json:"supports_x25519"`
	SupportsHTTP2   bool          `json:"supports_http2"`
	CipherSuite     string        `json:"cipher_suite"`
	HandshakeTime   time.Duration `json:"handshake_time"`
}

// CertificateResult 证书检测结果
type CertificateResult struct {
	Valid           bool      `json:"valid"`
	Issuer          string    `json:"issuer"`
	Subject         string    `json:"subject"`
	DaysUntilExpiry int       `json:"days_until_expiry"`
	CertificateSANs []string  `json:"certificate_sans"`
	NotBefore       time.Time `json:"not_before"`
	NotAfter        time.Time `json:"not_after"`
	Error           string    `json:"error,omitempty"`
}

// SNIResult SNI检测结果
type SNIResult struct {
	SupportsSNI bool   `json:"supports_sni"`
	SNIMatch    bool   `json:"sni_match"`
	ServerName  string `json:"server_name"`
}

// CDNResult CDN检测结果
type CDNResult struct {
	IsCDN       bool    `json:"is_cdn"`
	CDNProvider string  `json:"cdn_provider"`
	Confidence  string  `json:"confidence"`
	Evidence    string  `json:"evidence"`
	IsHotWebsite bool   `json:"is_hot_website"`
	Error       error   `json:"error,omitempty"`
}

// BlockedResult 被墙检测结果
type BlockedResult struct {
	IsBlocked      bool     `json:"is_blocked"`
	BlockedReasons []string `json:"blocked_reasons"`
	MatchType      string   `json:"match_type"`
}

// LocationResult 地理位置检测结果
type LocationResult struct {
	Country    string `json:"country"`
	IsDomestic bool   `json:"is_domestic"`
	IPAddress  string `json:"ip_address"`
	ISP        string `json:"isp"`
	ASN        string `json:"asn"`
	City       string `json:"city"`
	Region     string `json:"region"`
}

// DetectionSummary 检测摘要
type DetectionSummary struct {
	TotalChecks     int      `json:"total_checks"`
	PassedChecks    int      `json:"passed_checks"`
	FailedChecks    int      `json:"failed_checks"`
	Warnings        []string `json:"warnings"`
	Recommendations []string `json:"recommendations"`
}

// BatchReport 批量检测报告
type BatchReport struct {
	StartTime        time.Time         `json:"start_time"`
	EndTime          time.Time         `json:"end_time"`
	TotalDuration    time.Duration     `json:"total_duration"`
	Results          []*DetectionResult `json:"results"`
	Statistics       *Statistics       `json:"statistics"`
	PerformanceStats *PerformanceStats `json:"performance_stats"`
	CDNStats         *CDNStats         `json:"cdn_stats"`
	GeographicStats  *GeographicStats  `json:"geographic_stats"`
	TLSStats         *TLSStats         `json:"tls_stats"`
	CertificateStats *CertificateStats `json:"certificate_stats"`
	Summary          *BatchSummary     `json:"summary"`
}

// Statistics 统计信息
type Statistics struct {
	TotalDomains    int `json:"total_domains"`
	SuccessfulChecks int `json:"successful_checks"`
	FailedChecks    int `json:"failed_checks"`
	SuitableDomains int `json:"suitable_domains"`
	BlockedDomains  int `json:"blocked_domains"`
	ErrorDomains    int `json:"error_domains"`
}

// PerformanceStats 性能统计
type PerformanceStats struct {
	TotalTime        time.Duration `json:"total_time"`
	AverageTime      time.Duration `json:"average_time"`
	MinTime          time.Duration `json:"min_time"`
	MaxTime          time.Duration `json:"max_time"`
	AverageHandshake time.Duration `json:"average_handshake"`
}

// CDNStats CDN统计
type CDNStats struct {
	CDNDomains       int                    `json:"cdn_domains"`
	CDNProviders     map[string]int         `json:"cdn_providers"`
	CDNTypes         map[string]int         `json:"cdn_types"`
	ConfidenceLevels map[string]int         `json:"confidence_levels"`
}

// GeographicStats 地理统计
type GeographicStats struct {
	Countries     map[string]int `json:"countries"`
	DomesticCount int            `json:"domestic_count"`
	ForeignCount  int            `json:"foreign_count"`
}

// TLSStats TLS统计
type TLSStats struct {
	TLS13Support    int `json:"tls13_support"`
	X25519Support   int `json:"x25519_support"`
	HTTP2Support    int `json:"http2_support"`
	AverageHandshake time.Duration `json:"average_handshake"`
}

// CertificateStats 证书统计
type CertificateStats struct {
	ValidCertificates int `json:"valid_certificates"`
	InvalidCertificates int `json:"invalid_certificates"`
	ExpiringSoon      int `json:"expiring_soon"`
	AverageExpiry     int `json:"average_expiry"`
}

// BatchSummary 批量检测摘要
type BatchSummary struct {
	SuccessRate      float64 `json:"success_rate"`
	SuitabilityRate  float64 `json:"suitability_rate"`
	BlockingRate     float64 `json:"blocking_rate"`
	CDNUsageRate     float64 `json:"cdn_usage_rate"`
	Recommendations  []string `json:"recommendations"`
	Warnings         []string `json:"warnings"`
}

// DetectionStage 检测阶段接口
type DetectionStage interface {
	Execute(ctx *PipelineContext) error
	CanEarlyExit() bool
	Priority() int
	Name() string
}

// PipelineContext 流水线上下文
type PipelineContext struct {
	Domain       string
	StartTime    time.Time
	Result       *DetectionResult
	Connections  *ConnectionManager
	Cache        *CacheManager
	Config       *Config
	EarlyExit    bool
	Error        error
}

// ConnectionManager 连接管理器
type ConnectionManager struct {
	HTTPClient *HTTPClient
	TLSClient  *TLSClient
	DNSResolver *DNSResolver
}

// CacheManager 缓存管理器
type CacheManager struct {
	DNSCache   *DNSCache
	ResultCache *ResultCache
	CDNCache   *CDNCache
}

// HTTPClient HTTP客户端
type HTTPClient struct {
	Client    interface{} // *http.Client
	Transport interface{} // *http.Transport
}

// TLSClient TLS客户端
type TLSClient struct {
	Config *TLSConfig
	Conn   interface{} // *tls.Conn
}

// DNSResolver DNS解析器
type DNSResolver struct {
	Cache map[string]*DNSEntry
	TTL   time.Duration
}

// DNSEntry DNS条目
type DNSEntry struct {
	IP        string
	Timestamp time.Time
	TTL       time.Duration
}

// DNSCache DNS缓存
type DNSCache struct {
	Cache map[string]*DNSEntry
	TTL   time.Duration
}

// ResultCache 结果缓存
type ResultCache struct {
	Cache map[string]*CachedResult
	TTL   time.Duration
}

// CachedResult 缓存结果
type CachedResult struct {
	Result    *DetectionResult
	Timestamp time.Time
}

// CDNCache CDN缓存
type CDNCache struct {
	Cache map[string]*CDNResult
	TTL   time.Duration
}

// TLSConfig TLS配置
type TLSConfig struct {
	MinVersion    uint16
	MaxVersion    uint16
	CipherSuites  []uint16
	ServerName    string
	NextProtos    []string
}

// Config 配置结构
type Config struct {
	Network      NetworkConfig      `yaml:"network"`
	TLS          TLSConfig          `yaml:"tls"`
	Concurrency  ConcurrencyConfig  `yaml:"concurrency"`
	Output       OutputConfig       `yaml:"output"`
	Cache        CacheConfig        `yaml:"cache"`
	Batch        BatchConfig        `yaml:"batch"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	Timeout    time.Duration `yaml:"timeout"`
	Retries    int           `yaml:"retries"`
	DNSServers []string      `yaml:"dns_servers"`
}

// ConcurrencyConfig 并发配置
type ConcurrencyConfig struct {
	MaxConcurrent int           `yaml:"max_concurrent"`
	CheckTimeout  time.Duration `yaml:"check_timeout"`
	CacheTTL      time.Duration `yaml:"cache_ttl"`
}

// OutputConfig 输出配置
type OutputConfig struct {
	Color   bool   `yaml:"color"`
	Verbose bool   `yaml:"verbose"`
	Format  string `yaml:"format"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	DNSEnabled    bool          `yaml:"dns_enabled"`
	ResultEnabled bool          `yaml:"result_enabled"`
	TTL           time.Duration `yaml:"ttl"`
	MaxSize       int           `yaml:"max_size"`
}

// BatchConfig 批量配置
type BatchConfig struct {
	StreamOutput bool          `yaml:"stream_output"`
	ProgressBar  bool          `yaml:"progress_bar"`
	ReportFormat string        `yaml:"report_format"`
	Timeout      time.Duration `yaml:"timeout"`
}

// ConnectionStats 连接统计
type ConnectionStats struct {
	ActiveConnections int `json:"active_connections"`
	TotalConnections  int `json:"total_connections"`
	FailedConnections int `json:"failed_connections"`
}

// CacheStats 缓存统计
type CacheStats struct {
	DNSCacheSize    int     `json:"dns_cache_size"`
	ResultCacheSize int     `json:"result_cache_size"`
	CDNCacheSize    int     `json:"cdn_cache_size"`
	HitRate         float64 `json:"hit_rate"`
}
