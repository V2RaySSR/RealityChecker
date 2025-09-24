# Reality协议目标网站检测工具

一个专业的Reality协议目标网站检测工具，用于评估网站是否适合作为Reality协议的目标域名。

[V2RaySSR综合网](https://v2rayssr.com)

## ✨ 功能特性

* **被墙检测** - 基于GFWList检测网站是否被墙
* **地理位置检测** - 检测IP地理位置，国内网站直接终止
* **TLS协议检测** - 检测TLS 1.3和X25519支持
* **证书检测** - 检测证书有效性和SNI匹配
* **CDN检测** - 智能检测CDN使用情况
* **热门网站检测** - 检测是否为热门网站
* **重定向检测** - 检测域名重定向
* **批量检测** - 支持多域名并发检测，可与RealiTLScanner配合使用
* **智能报告** - 生成详细的检测分析报告

## 🚀 快速开始

### 系统要求

* **Linux VPS** - 主要针对VPS环境使用
* **Go 1.21+** - 用于本地编译（可选）

### 安装步骤

**方法1：直接下载（推荐）**

从 [Releases](https://github.com/V2RaySSR/RealityChecker/releases) 页面下载对应架构的zip文件：

```bash
# Linux x86_64
wget https://github.com/V2RaySSR/RealityChecker/releases/latest/download/reality-checker-linux-amd64.zip

# Linux ARM64
wget https://github.com/V2RaySSR/RealityChecker/releases/latest/download/reality-checker-linux-arm64.zip
```

解压后直接使用：
```bash
# 解压
unzip reality-checker-linux-amd64.zip

# 添加执行权限
chmod +x reality-checker

# 开始检测
./reality-checker check <域名>
```

**方法2：本地编译**

```bash
# 克隆项目
git clone https://github.com/V2RaySSR/RealityChecker.git
cd RealityChecker

# 编译程序
go build -o reality-checker

# 开始检测
./reality-checker check <域名>
```

## 🔍 使用示例

### 单域名检测

```bash
# 基础检测
./reality-checker check apple.com
```

### 批量检测

```bash
# 批量检测多个域名（逗号分隔）
./reality-checker batch "apple.com,tesla.com,microsoft.com"
```

### CSV文件检测

```bash
# 从CSV文件批量检测域名
./reality-checker csv domains.csv
```

### 推荐工作流程

对于大量域名检测，建议配合使用 [RealiTLScanner](https://github.com/XTLS/RealiTLScanner) 工具：

**1. 使用RealiTLScanner扫描VPS IP：**
```bash
./RealiTLScanner -addr <VPS IP> -port 443 -thread 50 -timeout 5 -out file.csv
```

**2. 使用本工具检测生成的CSV文件：**
```bash
./reality-checker csv file.csv
```

### 查看帮助

```bash
# 显示使用说明
./reality-checker

# 查看版本信息
./reality-checker version
```

## ⚡ 性能特性

* **多线程架构** - Worker Pool模式，高效任务分发
* **连接池管理** - 复用TLS和HTTP连接
* **DNS缓存** - 缓存DNS解析结果
* **自适应速率限制** - 根据服务器响应动态调整
* **内存监控** - 实时监控内存使用
* **自适应并发控制** - 根据系统性能动态调整并发数

## 🔧 故障排除

### 常见问题

**1. 数据文件下载失败**

程序启动时会自动下载必要的数据文件，如果下载失败，请检查网络连接：

```bash
# 检查网络连接
curl -I https://github.com/Loyalsoldier/geoip/releases/latest/download/Country.mmdb
curl -I https://raw.githubusercontent.com/Loyalsoldier/clash-rules/release/gfw.txt
curl -I https://raw.githubusercontent.com/V2RaySSR/RealityChecker/main/data/cdn_keywords.txt
curl -I https://raw.githubusercontent.com/V2RaySSR/RealityChecker/main/data/hot_websites.txt
```

**手动下载数据文件**

如果自动下载失败，请手动下载以下文件到 `data/` 目录：

- [Country.mmdb](https://github.com/Loyalsoldier/geoip/releases/latest/download/Country.mmdb)
- [gfwlist.conf](https://raw.githubusercontent.com/Loyalsoldier/clash-rules/release/gfw.txt)
- [cdn_keywords.txt](https://raw.githubusercontent.com/V2RaySSR/RealityChecker/main/data/cdn_keywords.txt)
- [hot_websites.txt](https://raw.githubusercontent.com/V2RaySSR/RealityChecker/main/data/hot_websites.txt)

## 📝 检测标准

### 推荐使用的网站特征

* ✅ 海外网站（非国内IP）
* ✅ 支持TLS 1.3协议
* ✅ 支持X25519加密算法
* ✅ 证书SNI匹配正确
* ✅ 未使用CDN
* ✅ 非热门网站
* ✅ 未被墙

### 不推荐使用的网站特征

* ❌ 国内网站
* ❌ 不支持TLS 1.3
* ❌ 不支持X25519
* ❌ 证书SNI不匹配
* ❌ 使用CDN
* ❌ 热门网站
* ❌ 被墙网站

## 🤝 贡献指南

欢迎提交Issue和Pull Request！

### 贡献方式

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📞 支持与反馈

* **GitHub Issues**: [提交问题](https://github.com/V2RaySSR/RealityChecker/issues)
* **讨论区**: [GitHub Discussions](https://github.com/V2RaySSR/RealityChecker/discussions)

## 🏆 致谢

感谢以下开源项目：

* [Loyalsoldier/geoip](https://github.com/Loyalsoldier/geoip) - GeoIP数据库
* [Loyalsoldier/clash-rules](https://github.com/Loyalsoldier/clash-rules) - GFW规则

---

**注意**: 本工具仅用于技术研究和学习目的，请遵守当地法律法规，合理使用网络资源。