# Reality协议目标网站检测工具 - Go版本技术规范

## 项目概述

将现有的Python版本Reality协议目标网站检测工具完全重写为Go语言版本，实现跨平台编译、单文件部署、无依赖运行的目标。

## 核心功能确认

### 1. Reality协议目标检测
- **目标**: 检测网站是否适合作为Reality协议的目标
- **硬性条件**: 必须是海外HTTPS网站，支持TLS 1.3和X25519，证书有效未过期
- **检测流程**: 被墙检测 → 重定向检测 → 地理位置检测 → TLS检测 → 证书检测 → SNI检测 → CDN检测
- **优化策略**: 单次TLS连接获取所有信息，减少重复连接

### 2. 被墙检测
- **数据源**: GFWList黑名单文件 (`gfwlist.conf`)
- **检测逻辑**: 
  - 优先检查本地缓存
  - 支持在线更新GFWList
  - 检查主域名和子域名匹配
  - 格式解析: `- '+.domain.com'` 和 `- '+domain.com'`
- **结果**: 被墙网站直接标记为不适合Reality目标

### 3. 重定向检测
- **HTTP重定向**: 检测301/302重定向
- **反爬机制**: 检测403状态码和反爬策略
- **最终域名**: 获取重定向后的最终域名用于后续检测
- **透明处理**: 自动跟踪重定向，用户无需关心重定向过程
- **建议**: 使用最终域名作为Reality目标

### 4. 地理位置检测
- **数据源**: MaxMind GeoIP数据库 (`Country.mmdb`)
- **智能判断**: 
  - 优先基于域名特征判断（.cn域名、国内知名网站）
  - 然后基于IP地址GeoIP数据库判断
- **国内网站列表**: baidu.com, qq.com, taobao.com等
- **结果**: 国内网站不适合作为Reality目标

### 5. TLS检测
- **TLS 1.3支持**: 必须支持TLS 1.3协议
- **X25519支持**: 必须支持X25519密钥交换
- **HTTP/2支持**: 检测ALPN协议支持
- **密码套件**: 检测使用的加密套件
- **握手时间**: 测量TLS握手延迟

### 6. 证书检测
- **证书主题**: 检查证书的Common Name (CN)
- **SAN扩展**: 检查Subject Alternative Names
- **有效期**: 检查证书是否过期（硬性要求）
- **剩余天数**: 计算证书剩余有效天数
- **证书链**: 验证证书链完整性
- **硬性条件**: 证书必须有效未过期

### 7. SNI匹配检测
- **SNI支持**: 检测Server Name Indication支持
- **域名匹配**: 检查SNI中的域名是否与目标域名匹配
- **证书匹配**: 验证证书主题与SNI域名匹配

### 8. CDN检测
- **第三方工具**: 集成cdncheck工具
- **HTTP特征**: 检测HTTP响应头中的CDN特征
- **CNAME记录**: 检查DNS CNAME记录中的CDN特征
- **关键词匹配**: 基于CDN关键词数据库检测
- **结果**: CDN使用可能影响Reality协议效果

## 硬性要求与非硬性要求

### 硬性要求（必须符合，否则不适合作为Reality目标）
- **地理位置**: 必须是海外网站（非国内IP地址）
- **被墙状态**: 不能是被墙的网站
- **TLS支持**: 必须支持TLS 1.3协议
- **密钥交换**: 必须支持X25519密钥交换
- **证书有效性**: 证书必须有效且未过期
- **SNI匹配**: 必须支持SNI且域名匹配
- **网络连通性**: 必须能够建立TLS连接

### 非硬性要求（不影响作为Reality目标）
- **HTTP状态码**: 404、403等错误页面不影响使用
- **页面内容**: 页面显示什么内容不重要
- **CDN使用**: 使用CDN不影响基本功能
- **重定向**: 重定向过程对用户透明

### 判断逻辑
- **适合**: 所有硬性要求都符合
- **不适合**: 任何一个硬性要求不符合

### 9. 热门网站检测
- **热门网站列表**: 基于预定义的热门网站列表
- **建议**: 热门网站不适合作为Reality目标
- **原因**: 流量特征明显，容易被识别

### 10. 网络质量检测
- **Ping测试**: 测量基础网络延迟
- **TLS握手时间**: 测量SSL/TLS握手延迟
- **连接成功率**: 测试连接稳定性
- **延迟波动**: 计算延迟方差

## 技术架构设计

### 1. 项目结构
```
reality-checker-go/
├── main.go                 # 主程序入口
├── go.mod                  # Go模块文件
├── go.sum                  # 依赖校验文件
├── config.yaml             # 配置文件
├── internal/               # 内部包
│   ├── checker/           # 检测模块
│   │   ├── network.go     # 网络检测
│   │   ├── tls.go         # TLS检测
│   │   ├── cert.go        # 证书检测
│   │   ├── cdn.go         # CDN检测
│   │   ├── sni.go         # SNI检测
│   │   └── blocked.go     # 被墙检测
│   ├── data/              # 数据文件处理
│   │   ├── downloader.go  # 数据文件下载器
│   │   ├── geoip.go       # GeoIP数据库
│   │   ├── gfwlist.go     # GFWList处理
│   │   ├── cdncheck.go    # CDN检测工具
│   │   └── manager.go     # 数据文件管理器
│   ├── utils/             # 工具函数
│   │   ├── color.go       # 颜色输出
│   │   ├── table.go       # 表格格式化
│   │   └── dns.go         # DNS解析
│   └── report/            # 报告生成
│       ├── generator.go   # 报告生成器
│       └── formatter.go   # 格式化器
├── data/                  # 本地数据文件目录（按需下载）
│   ├── Country.mmdb       # GeoIP数据库
│   ├── gfwlist.conf       # GFWList文件
│   └── cdncheck           # CDN检测工具
├── build/                 # 构建脚本
│   ├── build.sh           # 构建脚本
│   └── cross-compile.sh   # 交叉编译脚本
└── README.md              # 项目说明
```

### 2. 核心模块设计

#### 2.1 网络检测模块 (`internal/checker/network.go`)
```go
type NetworkChecker struct {
    geoipDB    *geoip2.Reader
    gfwlist    *GFWList
    dnsClient  *dns.Client
}

func (nc *NetworkChecker) CheckLocation(domain string) (*LocationResult, error)
func (nc *NetworkChecker) CheckBlocked(domain string) (*BlockedResult, error)
func (nc *NetworkChecker) CheckRedirect(domain string) (*RedirectResult, error)
func (nc *NetworkChecker) CheckAccessibility(domain string) bool
```

#### 2.2 TLS检测模块 (`internal/checker/tls.go`)
```go
type TLSChecker struct {
    config *tls.Config
}

func (tc *TLSChecker) CheckTLS13(domain string) (*TLSResult, error)
func (tc *TLSChecker) CheckX25519(domain string) (*X25519Result, error)
func (tc *TLSChecker) CheckHTTP2(domain string) (*HTTP2Result, error)
func (tc *TLSChecker) MeasureHandshakeTime(domain string) (time.Duration, error)
```

#### 2.3 证书检测模块 (`internal/checker/cert.go`)
```go
type CertChecker struct{}

func (cc *CertChecker) AnalyzeCert(domain string) (*CertResult, error)
func (cc *CertChecker) CheckExpiration(cert *x509.Certificate) (*ExpirationResult, error)
func (cc *CertChecker) CheckSubject(cert *x509.Certificate, domain string) (*SubjectResult, error)
func (cc *CertChecker) CheckSAN(cert *x509.Certificate, domain string) (*SANResult, error)
```

#### 2.4 CDN检测模块 (`internal/checker/cdn.go`)
```go
type CDNChecker struct {
    keywords map[string][]string
    hotSites map[string]bool
    cdncheck *CDNCheckTool
}

func (cc *CDNChecker) CheckCDN(domain string) (*CDNResult, error)
func (cc *CDNChecker) CheckHotWebsite(domain string) (*HotWebsiteResult, error)
func (cc *CDNChecker) CheckHTTPFeatures(domain string) (*HTTPFeaturesResult, error)
func (cc *CDNChecker) CheckCNAME(domain string) (*CNAMEResult, error)
```

#### 2.5 SNI检测模块 (`internal/checker/sni.go`)
```go
type SNIChecker struct{}

func (sc *SNIChecker) CheckSNIMatch(domain string, cert *x509.Certificate) (*SNIResult, error)
func (sc *SNIChecker) CheckSNISupport(domain string) (*SNISupportResult, error)
```

### 3. 数据文件处理

#### 3.1 GeoIP数据库 (`internal/data/geoip.go`)
```go
type GeoIPDB struct {
    reader *geoip2.Reader
}

func (gdb *GeoIPDB) LoadDatabase(path string) error
func (gdb *GeoIPDB) GetCountry(ip string) (string, error)
func (gdb *GeoIPDB) IsDomestic(ip string) (bool, error)
func (gdb *GeoIPDB) Close() error
```

#### 3.2 GFWList处理 (`internal/data/gfwlist.go`)
```go
type GFWList struct {
    domains map[string]bool
    cache   *cache.Cache
}

func (gfw *GFWList) LoadFromFile(path string) error
func (gfw *GFWList) LoadFromURL(url string) error
func (gfw *GFWList) IsBlocked(domain string) bool
func (gfw *GFWList) Update() error
```

#### 3.3 数据文件下载器 (`internal/data/downloader.go`)
```go
type DataDownloader struct {
    config *Config
    logger *Logger
}

// 数据文件配置
type DataFileConfig struct {
    Name        string `yaml:"name"`
    URL         string `yaml:"url"`
    LocalPath   string `yaml:"local_path"`
    Checksum    string `yaml:"checksum"`
    UpdateFreq  string `yaml:"update_freq"` // daily, weekly, monthly
    Required    bool   `yaml:"required"`
}

func (dd *DataDownloader) DownloadDataFile(config DataFileConfig) error
func (dd *DataDownloader) IsFileValid(path, checksum string) bool
func (dd *DataDownloader) VerifyChecksum(path, checksum string) bool
func (dd *DataDownloader) DownloadFile(url, path string) error
```

#### 3.4 数据文件管理器 (`internal/data/manager.go`)
```go
type DataManager struct {
    downloader *DataDownloader
    config     *Config
    cache      *DataCache
}

func (dm *DataManager) EnsureDataFiles() error
func (dm *DataManager) CheckFeatureRequirements(feature string) error
func (dm *DataManager) UpdateAll() error
func (dm *DataManager) UpdateSpecific(fileName string) error
func (dm *DataManager) GetDataFileStatus() map[string]DataFileStatus
```

#### 3.5 CDN检测工具 (`internal/data/cdncheck.go`)
```go
type CDNCheckTool struct {
    binaryPath string
    version    string
}

func (cct *CDNCheckTool) Download() error
func (cct *CDNCheckTool) Check(domain string) (*CDNCheckResult, error)
func (cct *CDNCheckTool) GetVersion() (string, error)
```

### 4. 工具函数

#### 4.1 颜色输出 (`internal/utils/color.go`)
```go
type ColorOutput struct {
    enabled bool
}

func (co *ColorOutput) Red(text string) string
func (co *ColorOutput) Green(text string) string
func (co *ColorOutput) Yellow(text string) string
func (co *ColorOutput) Blue(text string) string
func (co *ColorOutput) Magenta(text string) string
func (co *ColorOutput) Cyan(text string) string
func (co *ColorOutput) Bold(text string) string
```

#### 4.2 表格格式化 (`internal/utils/table.go`)
```go
type TableFormatter struct {
    color *ColorOutput
}

func (tf *TableFormatter) CreateTable() *Table
func (tf *TableFormatter) AddRow(table *Table, items ...string)
func (tf *TableFormatter) Render(table *Table) string
```

#### 4.3 DNS解析 (`internal/utils/dns.go`)
```go
type DNSResolver struct {
    client *dns.Client
}

func (dr *DNSResolver) Resolve(domain string) ([]string, error)
func (dr *DNSResolver) ResolveCNAME(domain string) (string, error)
func (dr *DNSResolver) ResolveWithServer(domain, server string) ([]string, error)
```

### 5. 报告生成

#### 5.1 报告生成器 (`internal/report/generator.go`)
```go
type ReportGenerator struct {
    color   *ColorOutput
    table   *TableFormatter
}

func (rg *ReportGenerator) GenerateHeader(domain string) string
func (rg *ReportGenerator) GenerateProgress(step, description, status string) string
func (rg *ReportGenerator) GenerateReport(domain string, results *DetectionResults) string
func (rg *ReportGenerator) GenerateSummary(results *DetectionResults) string
```

#### 5.2 格式化器 (`internal/report/formatter.go`)
```go
type Formatter struct {
    color *ColorOutput
}

func (f *Formatter) FormatStatus(status string) string
func (f *Formatter) FormatTime(duration time.Duration) string
func (f *Formatter) FormatIP(ip string) string
func (f *Formatter) FormatLocation(location string) string
```

## 关键实现细节

### 1. VPS环境下的CDN检测优化

#### 1.1 国内CDN地址检测
```go
// 优先使用国内DNS服务器解析
func (dr *DNSResolver) ResolveDomestic(domain string) ([]string, error) {
    // 使用114.114.114.114, 8.8.8.8等DNS服务器
    servers := []string{"114.114.114.114", "8.8.8.8", "1.1.1.1"}
    
    for _, server := range servers {
        ips, err := dr.ResolveWithServer(domain, server)
        if err == nil && len(ips) > 0 {
            return ips, nil
        }
    }
    return nil, errors.New("DNS resolution failed")
}
```

#### 1.2 CDN特征检测
```go
// 检测HTTP响应头中的CDN特征
func (cc *CDNChecker) CheckHTTPFeatures(domain string) (*HTTPFeaturesResult, error) {
    // 发送HTTP请求
    resp, err := http.Get("https://" + domain)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // 检查Server头
    server := resp.Header.Get("Server")
    if cc.isCDNServer(server) {
        return &HTTPFeaturesResult{
            LikelyCDN: true,
            Evidence:  "Server: " + server,
        }, nil
    }
    
    // 检查其他CDN特征头
    for header, value := range resp.Header {
        if cc.isCDNHeader(header, value) {
            return &HTTPFeaturesResult{
                LikelyCDN: true,
                Evidence:  header + ": " + value,
            }, nil
        }
    }
    
    return &HTTPFeaturesResult{LikelyCDN: false}, nil
}
```

### 2. 跨平台编译支持

#### 2.1 构建脚本 (`build/cross-compile.sh`)
```bash
#!/bin/bash

# 支持的平台
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "windows/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

# 构建目录
BUILD_DIR="build/bin"

# 创建构建目录
mkdir -p $BUILD_DIR

# 交叉编译
for platform in "${PLATFORMS[@]}"; do
    GOOS=${platform%/*}
    GOARCH=${platform#*/}
    
    echo "Building for $GOOS/$GOARCH..."
    
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o $BUILD_DIR/reality-checker-$GOOS-$GOARCH .
done

echo "Build completed!"
```

#### 2.2 数据文件按需下载
```go
// 数据文件按需下载和缓存
type DataFileManager struct {
    config   *Config
    downloader *DataDownloader
    cache    *DataCache
}

// 初始化数据文件管理器
func NewDataFileManager(config *Config) *DataFileManager {
    return &DataFileManager{
        config:     config,
        downloader: NewDataDownloader(config),
        cache:      NewDataCache("data"),
    }
}

// 确保数据文件可用
func (dfm *DataFileManager) EnsureDataFile(fileName string) error {
    dataFile, exists := dfm.config.DataFiles[fileName]
    if !exists {
        return fmt.Errorf("数据文件 %s 未配置", fileName)
    }
    
    // 检查本地文件是否存在且有效
    if dfm.downloader.IsFileValid(dataFile.LocalPath, dataFile.Checksum) {
        return nil
    }
    
    // 检查是否需要更新
    if !dfm.cache.ShouldUpdate(dataFile.LocalPath, dataFile.UpdateFreq) {
        return nil
    }
    
    // 下载数据文件
    return dfm.downloader.DownloadDataFile(dataFile)
}

// 批量确保数据文件
func (dfm *DataFileManager) EnsureAllRequiredFiles() error {
    for name, dataFile := range dfm.config.DataFiles {
        if dataFile.Required {
            if err := dfm.EnsureDataFile(name); err != nil {
                return fmt.Errorf("确保数据文件 %s 失败: %v", name, err)
            }
        }
    }
    return nil
}
```

### 3. 错误处理和日志

#### 3.1 错误处理
```go
type DetectionError struct {
    Type    string
    Message string
    Domain  string
    Err     error
}

func (de *DetectionError) Error() string {
    return fmt.Sprintf("[%s] %s: %s", de.Type, de.Domain, de.Message)
}

// 错误类型
const (
    ErrorTypeNetwork    = "NETWORK"
    ErrorTypeTLS        = "TLS"
    ErrorTypeCert       = "CERT"
    ErrorTypeCDN        = "CDN"
    ErrorTypeBlocked    = "BLOCKED"
    ErrorTypeLocation   = "LOCATION"
)
```

#### 3.2 日志系统
```go
type Logger struct {
    level  LogLevel
    output io.Writer
}

type LogLevel int

const (
    LogLevelDebug LogLevel = iota
    LogLevelInfo
    LogLevelWarn
    LogLevelError
)

func (l *Logger) Debug(format string, args ...interface{})
func (l *Logger) Info(format string, args ...interface{})
func (l *Logger) Warn(format string, args ...interface{})
func (l *Logger) Error(format string, args ...interface{})
```

### 4. 配置管理

#### 4.1 配置文件
```go
type Config struct {
    // 网络配置
    Network struct {
        Timeout     time.Duration `yaml:"timeout"`
        Retries     int           `yaml:"retries"`
        DNS         []string      `yaml:"dns_servers"`
    } `yaml:"network"`
    
    // TLS配置
    TLS struct {
        Timeout     time.Duration `yaml:"timeout"`
        MinVersion  uint16        `yaml:"min_version"`
        MaxVersion  uint16        `yaml:"max_version"`
    } `yaml:"tls"`
    
    // 数据文件配置
    DataFiles map[string]DataFileConfig `yaml:"data_files"`
    
    // 输出配置
    Output struct {
        Color       bool   `yaml:"color"`
        Verbose     bool   `yaml:"verbose"`
        Format      string `yaml:"format"`
    } `yaml:"output"`
    
    // 并发配置
    Concurrency struct {
        MaxConcurrent int           `yaml:"max_concurrent"`
        CheckTimeout  time.Duration `yaml:"check_timeout"`
        CacheTTL      time.Duration `yaml:"cache_ttl"`
        RetryCount    int           `yaml:"retry_count"`
        RetryDelay    time.Duration `yaml:"retry_delay"`
    } `yaml:"concurrency"`
}
```

#### 4.2 配置加载
```go
func LoadConfig(path string) (*Config, error) {
    config := &Config{}
    
    // 设置默认值
    config.Network.Timeout = 10 * time.Second
    config.Network.Retries = 3
    config.Network.DNS = []string{"8.8.8.8", "1.1.1.1"}
    
    config.TLS.Timeout = 10 * time.Second
    config.TLS.MinVersion = tls.VersionTLS12
    config.TLS.MaxVersion = tls.VersionTLS13
    
    // 设置默认数据文件配置
    config.DataFiles = map[string]DataFileConfig{
        "geoip": {
            Name:       "GeoIP数据库",
            URL:        "https://github.com/Loyalsoldier/geoip/releases/latest/download/Country.mmdb",
            LocalPath:  "data/Country.mmdb",
            UpdateFreq: "weekly",
            Required:   true,
        },
        "gfwlist": {
            Name:       "GFWList",
            URL:        "https://raw.githubusercontent.com/Loyalsoldier/clash-rules/release/gfw.txt",
            LocalPath:  "data/gfwlist.conf",
            UpdateFreq: "daily",
            Required:   true,
        },
        "cdncheck": {
            Name:       "CDN检测工具",
            URL:        "", // 根据架构动态设置
            LocalPath:  "data/cdncheck",
            UpdateFreq: "monthly",
            Required:   false,
        },
    }
    
    config.Output.Color = true
    config.Output.Verbose = false
    config.Output.Format = "table"
    
    // 设置默认并发配置
    config.Concurrency.MaxConcurrent = runtime.NumCPU() * 2
    config.Concurrency.CheckTimeout = 30 * time.Second
    config.Concurrency.CacheTTL = 5 * time.Minute
    config.Concurrency.RetryCount = 2
    config.Concurrency.RetryDelay = 1 * time.Second
    
    // 加载配置文件
    if path != "" {
        data, err := os.ReadFile(path)
        if err != nil {
            return nil, err
        }
        
        err = yaml.Unmarshal(data, config)
        if err != nil {
            return nil, err
        }
    }
    
    return config, nil
}
```

#### 4.3 配置文件示例
```yaml
# config.yaml
network:
  timeout: 10s
  retries: 3
  dns_servers:
    - "8.8.8.8"
    - "1.1.1.1"
    - "114.114.114.114"

tls:
  timeout: 10s
  min_version: 771  # TLS 1.2
  max_version: 772  # TLS 1.3

data_files:
  geoip:
    name: "GeoIP数据库"
    url: "https://github.com/Loyalsoldier/geoip/releases/latest/download/Country.mmdb"
    local_path: "data/Country.mmdb"
    checksum: ""  # 可选，用于验证文件完整性
    update_freq: "weekly"
    required: true
    
  gfwlist:
    name: "GFWList"
    url: "https://raw.githubusercontent.com/Loyalsoldier/clash-rules/release/gfw.txt"
    local_path: "data/gfwlist.conf"
    update_freq: "daily"
    required: true
    
  cdncheck:
    name: "CDN检测工具"
    url: ""  # 根据系统架构动态设置
    local_path: "data/cdncheck"
    update_freq: "monthly"
    required: false

output:
  color: true
  verbose: false
  format: "table"

concurrency:
  max_concurrent: 8    # 最大并发数 (0=自动)
  check_timeout: 30s   # 单个检测超时
  cache_ttl: 5m        # 结果缓存TTL
  retry_count: 2       # 重试次数
  retry_delay: 1s      # 重试延迟
```

## 依赖管理

### 1. Go模块依赖
```go
// go.mod
module reality-checker-go

go 1.21

require (
    github.com/oschwald/geoip2-golang v1.9.0
    github.com/miekg/dns v1.1.56
    github.com/fatih/color v1.15.0
    github.com/olekukonko/tablewriter v0.0.5
    github.com/spf13/cobra v1.7.0
    github.com/spf13/viper v1.16.0
    gopkg.in/yaml.v3 v3.0.1
)
```

### 2. 系统依赖
- **无系统依赖**: 所有功能通过Go标准库和第三方包实现
- **数据文件**: 按需从网络下载，支持本地缓存
- **外部工具**: cdncheck工具通过Go实现替代

### 3. 数据文件管理

#### 3.1 数据文件下载器实现
```go
// 数据文件下载和更新
type DataDownloader struct {
    config      *Config
    logger      *Logger
    client      *http.Client
    downloadMux sync.RWMutex           // 下载锁
    downloading map[string]chan error  // 正在下载的文件
    versionCache map[string]string     // 版本缓存
    versionMux  sync.RWMutex          // 版本缓存锁
}

// 初始化下载器
func NewDataDownloader(config *Config) *DataDownloader {
    return &DataDownloader{
        config:       config,
        logger:       NewLogger(),
        client:       &http.Client{Timeout: 30 * time.Second},
        downloading:  make(map[string]chan error),
        versionCache: make(map[string]string),
    }
}

// 下载数据文件
func (dd *DataDownloader) DownloadDataFile(config DataFileConfig) error {
    // 检查本地文件是否存在且有效
    if dd.IsFileValid(config.LocalPath, config.Checksum) {
        dd.logger.Info("数据文件 %s 已存在且有效", config.Name)
        return nil
    }
    
    // 创建目录
    dir := filepath.Dir(config.LocalPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("创建目录失败: %v", err)
    }
    
    // 获取下载URL（支持动态URL生成）
    downloadURL := config.URL
    if downloadURL == "" {
        downloadURL = dd.getDynamicURL(config.Name)
        if downloadURL == "" {
            return fmt.Errorf("无法获取 %s 的下载URL", config.Name)
        }
    }
    
    // 下载文件
    dd.logger.Info("开始下载 %s...", config.Name)
    if err := dd.DownloadFile(downloadURL, config.LocalPath); err != nil {
        return fmt.Errorf("下载 %s 失败: %v", config.Name, err)
    }
    
    // 验证文件
    if config.Checksum != "" {
        if !dd.VerifyChecksum(config.LocalPath, config.Checksum) {
            return fmt.Errorf("文件校验失败: %s", config.Name)
        }
    }
    
    dd.logger.Info("数据文件 %s 下载完成", config.Name)
    return nil
}

// 获取动态URL（根据系统架构）
func (dd *DataDownloader) getDynamicURL(fileName string) string {
    switch fileName {
    case "cdncheck":
        return dd.getCDNCheckURL()
    default:
        return ""
    }
}

// 获取CDN检测工具的下载URL
func (dd *DataDownloader) getCDNCheckURL() string {
    // 获取最新版本
    latestVersion, err := dd.getLatestVersion("projectdiscovery/cdncheck")
    if err != nil {
        dd.logger.Warn("获取CDN检测工具最新版本失败: %v", err)
        return ""
    }
    
    // 根据系统架构确定URL
    goos := runtime.GOOS
    goarch := runtime.GOARCH
    
    var arch string
    switch goarch {
    case "amd64", "x86_64":
        arch = "amd64"
    case "arm64", "aarch64":
        arch = "arm64"
    case "armv7":
        arch = "armv7"
    default:
        dd.logger.Warn("不支持的架构: %s", goarch)
        return ""
    }
    
    // 构建下载URL
    version := strings.TrimPrefix(latestVersion, "v")
    if goos == "linux" {
        return fmt.Sprintf("https://github.com/projectdiscovery/cdncheck/releases/download/%s/cdncheck_%s_linux_%s.zip", 
            latestVersion, version, arch)
    } else if goos == "windows" {
        return fmt.Sprintf("https://github.com/projectdiscovery/cdncheck/releases/download/%s/cdncheck_%s_windows_%s.zip", 
            latestVersion, version, arch)
    } else if goos == "darwin" {
        return fmt.Sprintf("https://github.com/projectdiscovery/cdncheck/releases/download/%s/cdncheck_%s_darwin_%s.zip", 
            latestVersion, version, arch)
    }
    
    return ""
}

// 获取GitHub仓库的最新版本
func (dd *DataDownloader) getLatestVersion(repo string) (string, error) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
    
    resp, err := dd.client.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        return "", fmt.Errorf("API请求失败: %d", resp.StatusCode)
    }
    
    var release struct {
        TagName string `json:"tag_name"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
        return "", err
    }
    
    return release.TagName, nil
}

// 下载文件
func (dd *DataDownloader) DownloadFile(url, path string) error {
    resp, err := dd.client.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        return fmt.Errorf("下载失败: %d", resp.StatusCode)
    }
    
    // 检查是否是ZIP文件
    if strings.HasSuffix(url, ".zip") {
        return dd.downloadAndExtractZip(resp.Body, path)
    }
    
    // 普通文件下载
    file, err := os.Create(path)
    if err != nil {
        return err
    }
    defer file.Close()
    
    _, err = io.Copy(file, resp.Body)
    return err
}

// 下载并解压ZIP文件
func (dd *DataDownloader) downloadAndExtractZip(body io.Reader, targetPath string) error {
    // 创建临时文件
    tempFile, err := os.CreateTemp("", "cdncheck-*.zip")
    if err != nil {
        return err
    }
    defer os.Remove(tempFile.Name())
    defer tempFile.Close()
    
    // 下载ZIP文件
    _, err = io.Copy(tempFile, body)
    if err != nil {
        return err
    }
    
    // 解压ZIP文件
    return dd.extractZip(tempFile.Name(), targetPath)
}

// 解压ZIP文件
func (dd *DataDownloader) extractZip(zipPath, targetPath string) error {
    reader, err := zip.OpenReader(zipPath)
    if err != nil {
        return err
    }
    defer reader.Close()
    
    // 查找可执行文件
    var executableFile *zip.File
    for _, file := range reader.File {
        if !file.FileInfo().IsDir() && (file.Name == "cdncheck" || strings.HasSuffix(file.Name, "/cdncheck")) {
            executableFile = file
            break
        }
    }
    
    if executableFile == nil {
        return fmt.Errorf("ZIP文件中未找到cdncheck可执行文件")
    }
    
    // 创建目标目录
    if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
        return err
    }
    
    // 解压可执行文件
    rc, err := executableFile.Open()
    if err != nil {
        return err
    }
    defer rc.Close()
    
    targetFile, err := os.Create(targetPath)
    if err != nil {
        return err
    }
    defer targetFile.Close()
    
    _, err = io.Copy(targetFile, rc)
    if err != nil {
        return err
    }
    
    // 设置可执行权限
    return os.Chmod(targetPath, 0755)
}

// 检查文件有效性
func (dd *DataDownloader) IsFileValid(path, checksum string) bool {
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return false
    }
    
    if checksum != "" {
        return dd.VerifyChecksum(path, checksum)
    }
    
    return true
}

// 验证文件校验和
func (dd *DataDownloader) VerifyChecksum(path, expectedChecksum string) bool {
    file, err := os.Open(path)
    if err != nil {
        return false
    }
    defer file.Close()
    
    hash := sha256.New()
    if _, err := io.Copy(hash, file); err != nil {
        return false
    }
    
    actualChecksum := hex.EncodeToString(hash.Sum(nil))
    return actualChecksum == expectedChecksum
}
```

#### 3.2 数据文件管理器实现
```go
type DataManager struct {
    downloader *DataDownloader
    config     *Config
    cache      *DataCache
}

// 确保必需的数据文件存在
func (dm *DataManager) EnsureDataFiles() error {
    for name, dataFile := range dm.config.DataFiles {
        if dataFile.Required {
            if err := dm.downloader.DownloadDataFile(dataFile); err != nil {
                return fmt.Errorf("确保数据文件 %s 失败: %v", name, err)
            }
        }
    }
    return nil
}

// 检查特定功能需要的数据文件
func (dm *DataManager) CheckFeatureRequirements(feature string) error {
    switch feature {
    case "location":
        return dm.downloader.DownloadDataFile(dm.config.DataFiles["geoip"])
    case "blocked":
        return dm.downloader.DownloadDataFile(dm.config.DataFiles["gfwlist"])
    case "cdn":
        return dm.downloader.DownloadDataFile(dm.config.DataFiles["cdncheck"])
    }
    return nil
}

// 更新所有数据文件
func (dm *DataManager) UpdateAll() error {
    for name, dataFile := range dm.config.DataFiles {
        if err := dm.downloader.DownloadDataFile(dataFile); err != nil {
            return fmt.Errorf("更新数据文件 %s 失败: %v", name, err)
        }
    }
    return nil
}

// 更新特定数据文件
func (dm *DataManager) UpdateSpecific(fileName string) error {
    dataFile, exists := dm.config.DataFiles[fileName]
    if !exists {
        return fmt.Errorf("数据文件 %s 不存在", fileName)
    }
    
    return dm.downloader.DownloadDataFile(dataFile)
}
```

#### 3.3 数据文件缓存管理
```go
// 数据文件缓存管理
type DataCache struct {
    cacheDir string
    ttl      time.Duration
}

// 检查文件是否需要更新
func (dc *DataCache) ShouldUpdate(filePath string, updateFreq string) bool {
    info, err := os.Stat(filePath)
    if err != nil {
        return true
    }
    
    age := time.Since(info.ModTime())
    
    switch updateFreq {
    case "daily":
        return age > 24*time.Hour
    case "weekly":
        return age > 7*24*time.Hour
    case "monthly":
        return age > 30*24*time.Hour
    default:
        return false
    }
}
```

## 性能优化

### 1. 并发处理

#### 1.1 批量检测器设计
```go
// 批量检测器
type BatchChecker struct {
    config        *Config
    checker       *RealityChecker
    workerPool    *WorkerPool
    resultCache   *ResultCache
    progressChan  chan ProgressUpdate
    maxConcurrent int
    timeout       time.Duration
}

// 检测任务
type CheckTask struct {
    Index  int
    Domain string
    Result chan *DetectionResult
}

// 进度更新
type ProgressUpdate struct {
    Completed int
    Total     int
    Domain    string
    Status    string
}

// 初始化批量检测器
func NewBatchChecker(config *Config) *BatchChecker {
    return &BatchChecker{
        config:        config,
        checker:       NewRealityChecker(config),
        workerPool:    NewWorkerPool(config.MaxConcurrent),
        resultCache:   NewResultCache(config.CacheTTL),
        progressChan:  make(chan ProgressUpdate, 100),
        maxConcurrent: config.MaxConcurrent,
        timeout:       config.CheckTimeout,
    }
}
```

#### 1.2 Worker Pool实现
```go
// Worker Pool
type WorkerPool struct {
    workers    int
    jobQueue   chan CheckTask
    resultChan chan *DetectionResult
    wg         sync.WaitGroup
    checker    *RealityChecker
    ctx        context.Context
    cancel     context.CancelFunc
}

// 创建Worker Pool
func NewWorkerPool(workers int) *WorkerPool {
    if workers <= 0 {
        workers = runtime.NumCPU() * 2
    }
    
    ctx, cancel := context.WithCancel(context.Background())
    
    return &WorkerPool{
        workers:    workers,
        jobQueue:   make(chan CheckTask, workers*2),
        resultChan: make(chan *DetectionResult, workers*2),
        ctx:        ctx,
        cancel:     cancel,
    }
}

// 启动Worker Pool
func (wp *WorkerPool) Start(checker *RealityChecker) {
    wp.checker = checker
    
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go wp.worker(i)
    }
}

// Worker函数
func (wp *WorkerPool) worker(id int) {
    defer wp.wg.Done()
    
    for {
        select {
        case task := <-wp.jobQueue:
            // 执行检测任务
            result := wp.checkDomain(task.Domain)
            result.Index = task.Index
            result.Domain = task.Domain
            
            // 发送结果
            select {
            case task.Result <- result:
            case <-wp.ctx.Done():
                return
            }
            
        case <-wp.ctx.Done():
            return
        }
    }
}

// 检测单个域名
func (wp *WorkerPool) checkDomain(domain string) *DetectionResult {
    start := time.Now()
    
    result := &DetectionResult{
        Domain:    domain,
        StartTime: start,
    }
    
    // 执行检测逻辑
    ctx, cancel := context.WithTimeout(wp.ctx, 30*time.Second)
    defer cancel()
    
    // 使用context控制超时
    done := make(chan struct{})
    go func() {
        defer close(done)
        result = wp.checker.CheckDomainWithContext(ctx, domain)
    }()
    
    select {
    case <-done:
        result.Duration = time.Since(start)
    case <-ctx.Done():
        result.Error = fmt.Errorf("检测超时: %s", domain)
        result.Duration = time.Since(start)
    }
    
    return result
}

// 停止Worker Pool
func (wp *WorkerPool) Stop() {
    wp.cancel()
    close(wp.jobQueue)
    wp.wg.Wait()
}
```

#### 1.3 批量检测实现
```go
// 批量检测域名
func (bc *BatchChecker) CheckDomains(domains []string) ([]*DetectionResult, error) {
    if len(domains) == 0 {
        return nil, fmt.Errorf("域名列表为空")
    }
    
    // 启动Worker Pool
    bc.workerPool.Start(bc.checker)
    defer bc.workerPool.Stop()
    
    // 创建结果通道
    results := make([]*DetectionResult, len(domains))
    resultChans := make([]chan *DetectionResult, len(domains))
    
    for i := range resultChans {
        resultChans[i] = make(chan *DetectionResult, 1)
    }
    
    // 启动进度监控
    progressDone := make(chan struct{})
    go bc.monitorProgress(domains, results, progressDone)
    
    // 分发任务
    for i, domain := range domains {
        // 检查缓存
        if cached, exists := bc.resultCache.Get(domain); exists {
            results[i] = cached
            continue
        }
        
        task := CheckTask{
            Index:  i,
            Domain: domain,
            Result: resultChans[i],
        }
        
        select {
        case bc.workerPool.jobQueue <- task:
        case <-bc.workerPool.ctx.Done():
            return nil, fmt.Errorf("worker pool已停止")
        }
    }
    
    // 收集结果
    for i, resultChan := range resultChans {
        if results[i] != nil {
            continue // 已从缓存获取
        }
        
        select {
        case result := <-resultChan:
            results[i] = result
            // 缓存结果
            bc.resultCache.Set(result.Domain, result)
        case <-time.After(bc.timeout):
            results[i] = &DetectionResult{
                Domain: domains[i],
                Error:  fmt.Errorf("检测超时"),
            }
        }
    }
    
    close(progressDone)
    return results, nil
}

// 监控进度
func (bc *BatchChecker) monitorProgress(domains []string, results []*DetectionResult, done <-chan struct{}) {
    completed := 0
    total := len(domains)
    
    for {
        select {
        case <-done:
            return
        case <-time.After(100 * time.Millisecond):
            // 统计已完成的任务
            current := 0
            for _, result := range results {
                if result != nil {
                    current++
                }
            }
            
            if current > completed {
                completed = current
                bc.progressChan <- ProgressUpdate{
                    Completed: completed,
                    Total:     total,
                    Status:    fmt.Sprintf("%d/%d", completed, total),
                }
            }
        }
    }
}
```

#### 1.4 流式检测实现
```go
// 流式检测（实时输出结果）
func (bc *BatchChecker) CheckDomainsStream(domains []string, outputChan chan<- *DetectionResult) error {
    if len(domains) == 0 {
        return fmt.Errorf("域名列表为空")
    }
    
    // 启动Worker Pool
    bc.workerPool.Start(bc.checker)
    defer bc.workerPool.Stop()
    
    // 创建结果通道
    resultChans := make([]chan *DetectionResult, len(domains))
    for i := range resultChans {
        resultChans[i] = make(chan *DetectionResult, 1)
    }
    
    // 分发任务
    for i, domain := range domains {
        // 检查缓存
        if cached, exists := bc.resultCache.Get(domain); exists {
            outputChan <- cached
            continue
        }
        
        task := CheckTask{
            Index:  i,
            Domain: domain,
            Result: resultChans[i],
        }
        
        select {
        case bc.workerPool.jobQueue <- task:
        case <-bc.workerPool.ctx.Done():
            return fmt.Errorf("worker pool已停止")
        }
    }
    
    // 流式收集结果
    go func() {
        defer close(outputChan)
        
        for i, resultChan := range resultChans {
            if results[i] != nil {
                continue // 已从缓存获取
            }
            
            select {
            case result := <-resultChan:
                outputChan <- result
                // 缓存结果
                bc.resultCache.Set(result.Domain, result)
            case <-time.After(bc.timeout):
                outputChan <- &DetectionResult{
                    Domain: domains[i],
                    Error:  fmt.Errorf("检测超时"),
                }
            }
        }
    }()
    
    return nil
}
```

#### 1.5 并发控制配置
```go
// 并发控制配置
type ConcurrencyConfig struct {
    MaxConcurrent int           `yaml:"max_concurrent"`  // 最大并发数
    CheckTimeout  time.Duration `yaml:"check_timeout"`   // 单个检测超时
    CacheTTL      time.Duration `yaml:"cache_ttl"`       // 结果缓存TTL
    RetryCount    int           `yaml:"retry_count"`     // 重试次数
    RetryDelay    time.Duration `yaml:"retry_delay"`     // 重试延迟
}

// 默认配置
func DefaultConcurrencyConfig() *ConcurrencyConfig {
    return &ConcurrencyConfig{
        MaxConcurrent: runtime.NumCPU() * 2,
        CheckTimeout:  30 * time.Second,
        CacheTTL:      5 * time.Minute,
        RetryCount:    2,
        RetryDelay:    1 * time.Second,
    }
}
```

### 2. 连接复用
```go
// TLS连接复用
type ConnectionPool struct {
    connections map[string]*tls.Conn
    mutex       sync.RWMutex
}

func (cp *ConnectionPool) GetConnection(domain string) (*tls.Conn, error) {
    cp.mutex.RLock()
    conn, exists := cp.connections[domain]
    cp.mutex.RUnlock()
    
    if exists && conn != nil {
        return conn, nil
    }
    
    // 创建新连接
    newConn, err := cp.createConnection(domain)
    if err != nil {
        return nil, err
    }
    
    cp.mutex.Lock()
    cp.connections[domain] = newConn
    cp.mutex.Unlock()
    
    return newConn, nil
}
```

### 3. 缓存机制
```go
// 结果缓存
type ResultCache struct {
    cache map[string]*CachedResult
    mutex sync.RWMutex
    ttl   time.Duration
}

type CachedResult struct {
    Result    *DetectionResult
    Timestamp time.Time
}

func (rc *ResultCache) Get(domain string) (*DetectionResult, bool) {
    rc.mutex.RLock()
    defer rc.mutex.RUnlock()
    
    cached, exists := rc.cache[domain]
    if !exists {
        return nil, false
    }
    
    // 检查TTL
    if time.Since(cached.Timestamp) > rc.ttl {
        return nil, false
    }
    
    return cached.Result, true
}
```

## 部署和使用

### 1. 命令行接口设计

#### 1.1 主命令结构
```go
// 主命令
var rootCmd = &cobra.Command{
    Use:   "reality-checker",
    Short: "Reality协议目标网站检测工具",
    Long:  "一个用于检测网站是否适合作为Reality协议目标的Go语言工具",
}

// 检测命令
var checkCmd = &cobra.Command{
    Use:   "check [domain]",
    Short: "检测域名是否适合作为Reality目标",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        runCheckCommand(cmd, args)
    },
}

// 更新命令
var updateCmd = &cobra.Command{
    Use:   "update",
    Short: "更新数据文件",
    Run: func(cmd *cobra.Command, args []string) {
        runUpdateCommand(cmd, args)
    },
}

// 状态命令
var statusCmd = &cobra.Command{
    Use:   "status",
    Short: "查看数据文件状态",
    Run: func(cmd *cobra.Command, args []string) {
        runStatusCommand(cmd, args)
    },
}

// 清理命令
var cleanCmd = &cobra.Command{
    Use:   "clean",
    Short: "清理数据文件缓存",
    Run: func(cmd *cobra.Command, args []string) {
        runCleanCommand(cmd, args)
    },
}
```

#### 1.2 命令实现
```go
// 检测命令实现
func runCheckCommand(cmd *cobra.Command, args []string) {
    // 加载配置
    configPath, _ := cmd.Flags().GetString("config")
    config, err := LoadConfig(configPath)
    if err != nil {
        log.Fatal("加载配置失败:", err)
    }
    
    // 初始化数据文件管理器
    dataManager := NewDataFileManager(config)
    
    // 确保必需的数据文件存在
    if err := dataManager.EnsureAllRequiredFiles(); err != nil {
        log.Fatal("确保数据文件失败:", err)
    }
    
    // 获取参数
    filePath, _ := cmd.Flags().GetString("file")
    concurrent, _ := cmd.Flags().GetInt("concurrent")
    stream, _ := cmd.Flags().GetBool("stream")
    progress, _ := cmd.Flags().GetBool("progress")
    timeout, _ := cmd.Flags().GetDuration("timeout")
    
    // 设置并发配置
    if concurrent > 0 {
        config.MaxConcurrent = concurrent
    }
    config.CheckTimeout = timeout
    
    // 初始化批量检测器
    batchChecker := NewBatchChecker(config)
    
    var domains []string
    
    // 确定检测的域名列表
    if filePath != "" {
        // 从文件读取域名列表
        domains, err = loadDomainsFromFile(filePath)
        if err != nil {
            log.Fatal("读取域名文件失败:", err)
        }
    } else if len(args) > 0 {
        // 单个域名检测
        domains = args
    } else {
        log.Fatal("请指定要检测的域名或域名文件")
    }
    
    // 执行检测
    if stream {
        // 流式检测
        runStreamCheck(batchChecker, domains, cmd)
    } else {
        // 批量检测
        runBatchCheck(batchChecker, domains, cmd, progress)
    }
}

// 批量检测实现
func runBatchCheck(batchChecker *BatchChecker, domains []string, cmd *cobra.Command, showProgress bool) {
    // 启动进度监控
    if showProgress {
        go func() {
            for progress := range batchChecker.progressChan {
                fmt.Printf("\r进度: %s", progress.Status)
            }
            fmt.Println()
        }()
    }
    
    // 执行批量检测
    results, err := batchChecker.CheckDomains(domains)
    if err != nil {
        log.Fatal("批量检测失败:", err)
    }
    
    // 输出结果
    formatter := NewReportFormatter(batchChecker.config)
    output := formatter.FormatBatchResults(results)
    fmt.Println(output)
    
    // 输出到文件
    if outputFile, _ := cmd.Flags().GetString("output"); outputFile != "" {
        err := os.WriteFile(outputFile, []byte(output), 0644)
        if err != nil {
            log.Fatal("写入输出文件失败:", err)
        }
    }
}

// 流式检测实现
func runStreamCheck(batchChecker *BatchChecker, domains []string, cmd *cobra.Command) {
    resultChan := make(chan *DetectionResult, 100)
    
    // 启动流式检测
    err := batchChecker.CheckDomainsStream(domains, resultChan)
    if err != nil {
        log.Fatal("启动流式检测失败:", err)
    }
    
    // 实时输出结果
    formatter := NewReportFormatter(batchChecker.config)
    for result := range resultChan {
        output := formatter.FormatResult(result)
        fmt.Println(output)
    }
}

// 从文件加载域名列表
func loadDomainsFromFile(filePath string) ([]string, error) {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, err
    }
    
    lines := strings.Split(string(content), "\n")
    var domains []string
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line != "" && !strings.HasPrefix(line, "#") {
            domains = append(domains, line)
        }
    }
    
    return domains, nil
}

// 更新命令实现
func runUpdateCommand(cmd *cobra.Command, args []string) {
    configPath, _ := cmd.Flags().GetString("config")
    config, err := LoadConfig(configPath)
    if err != nil {
        log.Fatal("加载配置失败:", err)
    }
    
    dataManager := NewDataFileManager(config)
    
    // 检查是否指定了特定文件
    geoip, _ := cmd.Flags().GetBool("geoip")
    gfwlist, _ := cmd.Flags().GetBool("gfwlist")
    cdncheck, _ := cmd.Flags().GetBool("cdncheck")
    
    if geoip {
        err = dataManager.UpdateSpecific("geoip")
    } else if gfwlist {
        err = dataManager.UpdateSpecific("gfwlist")
    } else if cdncheck {
        err = dataManager.UpdateSpecific("cdncheck")
    } else {
        err = dataManager.UpdateAll()
    }
    
    if err != nil {
        log.Fatal("更新失败:", err)
    }
    
    fmt.Println("数据文件更新完成")
}

// 状态命令实现
func runStatusCommand(cmd *cobra.Command, args []string) {
    configPath, _ := cmd.Flags().GetString("config")
    config, err := LoadConfig(configPath)
    if err != nil {
        log.Fatal("加载配置失败:", err)
    }
    
    dataManager := NewDataFileManager(config)
    status := dataManager.GetDataFileStatus()
    
    // 输出状态表格
    table := tablewriter.NewWriter(os.Stdout)
    table.SetHeader([]string{"文件名", "状态", "大小", "修改时间", "需要更新"})
    
    for name, fileStatus := range status {
        table.Append([]string{
            name,
            fileStatus.Status,
            fileStatus.Size,
            fileStatus.ModTime,
            fileStatus.NeedsUpdate,
        })
    }
    
    table.Render()
}
```

#### 1.3 命令行参数
```go
func init() {
    // 全局参数
    rootCmd.PersistentFlags().StringP("config", "c", "", "配置文件路径")
    rootCmd.PersistentFlags().BoolP("verbose", "v", false, "详细输出")
    rootCmd.PersistentFlags().BoolP("color", "", true, "彩色输出")
    
    // 检测命令参数
    checkCmd.Flags().StringP("output", "o", "", "输出到文件")
    checkCmd.Flags().StringP("format", "f", "table", "输出格式 (table, json, yaml)")
    checkCmd.Flags().StringP("file", "", "", "批量检测文件")
    checkCmd.Flags().IntP("concurrent", "c", 0, "并发数 (0=自动)")
    checkCmd.Flags().Bool("stream", false, "流式输出结果")
    checkCmd.Flags().Bool("progress", false, "显示进度")
    checkCmd.Flags().Duration("timeout", 30*time.Second, "检测超时时间")
    checkCmd.Flags().StringP("geoip-path", "", "", "指定GeoIP数据库路径")
    checkCmd.Flags().StringP("gfwlist-path", "", "", "指定GFWList文件路径")
    checkCmd.Flags().StringP("cdncheck-path", "", "", "指定CDN检测工具路径")
    
    // 更新命令参数
    updateCmd.Flags().Bool("geoip", false, "只更新GeoIP数据库")
    updateCmd.Flags().Bool("gfwlist", false, "只更新GFWList")
    updateCmd.Flags().Bool("cdncheck", false, "只更新CDN检测工具")
    updateCmd.Flags().Bool("force", false, "强制更新所有文件")
    
    // 添加子命令
    rootCmd.AddCommand(checkCmd)
    rootCmd.AddCommand(updateCmd)
    rootCmd.AddCommand(statusCmd)
    rootCmd.AddCommand(cleanCmd)
}
```

### 2. 单文件部署
```bash
# 下载预编译的二进制文件
wget https://github.com/V2RaySSR/reality-checker-go/releases/latest/download/reality-checker-linux-amd64
chmod +x reality-checker-linux-amd64

# 直接运行（会自动下载必需的数据文件）
./reality-checker-linux-amd64 check apple.com

# 或者先更新数据文件再运行
./reality-checker-linux-amd64 update
./reality-checker-linux-amd64 check apple.com
```

### 3. 从源码编译
```bash
# 克隆仓库
git clone https://github.com/V2RaySSR/reality-checker-go.git
cd reality-checker-go

# 编译
go build -o reality-checker .

# 运行
./reality-checker check apple.com
```

### 4. 交叉编译
```bash
# 编译Linux版本
GOOS=linux GOARCH=amd64 go build -o reality-checker-linux-amd64 .

# 编译Windows版本
GOOS=windows GOARCH=amd64 go build -o reality-checker-windows-amd64.exe .

# 编译ARM64版本
GOOS=linux GOARCH=arm64 go build -o reality-checker-linux-arm64 .
```

### 5. 使用示例
```bash
# 检测单个域名（自动下载必需的数据文件）
./reality-checker check apple.com

# 详细输出
./reality-checker check -v apple.com

# 批量检测
./reality-checker check -f domains.txt

# 批量检测（指定并发数）
./reality-checker check -f domains.txt --concurrent 10

# 流式批量检测（实时输出）
./reality-checker check -f domains.txt --stream

# 批量检测（显示进度）
./reality-checker check -f domains.txt --progress

# 输出到文件
./reality-checker check apple.com -o result.txt

# 指定配置文件
./reality-checker check -c config.yaml apple.com

# 数据文件管理
./reality-checker update                    # 更新所有数据文件
./reality-checker update --geoip           # 只更新GeoIP数据库
./reality-checker update --gfwlist         # 只更新GFWList
./reality-checker status                   # 查看数据文件状态
./reality-checker clean                    # 清理数据文件缓存

# 指定数据文件路径
./reality-checker check apple.com --geoip-path /path/to/Country.mmdb

# 查看数据文件状态
./reality-checker status

# 强制更新所有数据文件
./reality-checker update --force

# 只更新CDN检测工具（会自动检测系统架构）
./reality-checker update --cdncheck
```

### 6. 数据文件说明

#### 6.1 GeoIP数据库
- **来源**: Loyalsoldier/geoip (更准确的GeoIP数据)
- **URL**: https://github.com/Loyalsoldier/geoip/releases/latest/download/Country.mmdb
- **更新频率**: 每周
- **用途**: 地理位置检测

#### 6.2 GFWList
- **来源**: Loyalsoldier/clash-rules (更完整的GFW规则)
- **URL**: https://raw.githubusercontent.com/Loyalsoldier/clash-rules/release/gfw.txt
- **更新频率**: 每天
- **用途**: 被墙检测

#### 6.3 CDN检测工具
- **来源**: projectdiscovery/cdncheck (专业CDN检测工具)
- **支持架构**: 
  - Linux: amd64, arm64, armv7
  - Windows: amd64, arm64
  - macOS: amd64, arm64
- **更新频率**: 每月
- **用途**: CDN特征检测

## 测试和验证

### 1. 单元测试
```go
func TestNetworkChecker_CheckLocation(t *testing.T) {
    checker := &NetworkChecker{}
    
    // 测试国内域名
    result, err := checker.CheckLocation("baidu.com")
    assert.NoError(t, err)
    assert.Equal(t, "国内", result.Location)
    
    // 测试海外域名
    result, err = checker.CheckLocation("apple.com")
    assert.NoError(t, err)
    assert.Equal(t, "海外", result.Location)
}
```

### 2. 集成测试
```go
func TestRealityChecker_CheckDomain(t *testing.T) {
    checker := NewRealityChecker()
    
    // 测试适合的域名
    result, err := checker.CheckDomain("apple.com")
    assert.NoError(t, err)
    assert.True(t, result.Suitable)
    
    // 测试不适合的域名
    result, err = checker.CheckDomain("baidu.com")
    assert.NoError(t, err)
    assert.False(t, result.Suitable)
}
```

### 3. 性能测试
```go
func BenchmarkRealityChecker_CheckDomain(b *testing.B) {
    checker := NewRealityChecker()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = checker.CheckDomain("apple.com")
    }
}
```

## 总结

这个Go版本的技术规范涵盖了从Python版本转换到Go版本的所有关键细节，包括：

1. **完整的功能映射**: 所有Python版本的功能都有对应的Go实现
2. **跨平台支持**: 支持Linux、Windows、macOS的x86_64和ARM64架构
3. **按需数据文件下载**: 数据文件按需从网络下载，支持本地缓存和自动更新
4. **单文件部署**: 核心程序无依赖部署，数据文件独立管理
5. **性能优化**: 并发处理、连接复用、结果缓存
6. **VPS优化**: 特别针对VPS环境下的CDN检测优化
7. **错误处理**: 完善的错误处理和日志系统
8. **配置管理**: 灵活的配置系统，支持数据文件源配置
9. **命令行接口**: 完整的数据文件管理命令
10. **测试覆盖**: 完整的测试体系

### 主要改进

相比原始的embed方案，新的按需下载方案具有以下优势：

1. **更小的二进制文件**: 核心程序更轻量，下载更快
2. **更灵活的数据管理**: 数据文件可以独立更新，无需重新编译
3. **更好的维护性**: 数据文件可以随时更新到最新版本
4. **更灵活的部署**: 支持多种数据源和本地文件覆盖
5. **更好的用户体验**: 用户可以选择性下载需要的数据文件

这个规范确保了Go版本能够完全替代Python版本，同时提供更好的性能、部署便利性和数据文件管理灵活性。
