# Reality协议目标网站检测工具

一个专业的Reality协议目标网站检测工具，用于评估网站是否适合作为Reality协议的目标域名。

**版本**: v2.1.0 | V2RaySSR综合网

## ✨ 功能特性

* **被墙检测** - 基于GFWList检测网站是否被墙
* **地理位置检测** - 检测IP地理位置，国内网站直接终止
* **TLS协议检测** - 检测TLS 1.3和X25519支持
* **证书检测** - 检测证书有效性和SNI匹配
* **CDN检测** - 智能检测CDN使用情况
* **热门网站检测** - 检测是否为热门网站
* **重定向检测** - 检测域名重定向
* **批量检测** - 支持多域名并发检测
* **智能报告** - 生成详细的检测分析报告

## 🚀 快速开始

### 系统要求

* **Go 1.21+** - 用于编译和运行
* **Linux/macOS/Windows** - 跨平台支持

### 安装步骤

**1. 克隆项目：**

```bash
git clone https://github.com/V2RaySSR/RealityChecker.git
cd RealityChecker
```

**2. 编译程序：**

```bash
go build -o reality-checker
```

**3. 开始检测：**

```bash
# 单域名检测
./reality-checker check <域名>

# 批量检测
./reality-checker batch "域名1,域名2,域名3"

# CSV文件检测
./reality-checker csv <csv文件>

# 查看帮助
./reality-checker
```

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

### 查看帮助

```bash
# 显示使用说明
./reality-checker
```

## 📊 检测报告

批量检测完成后会生成详细的统计分析报告，包括：

* **执行摘要** - 成功率、适合性率、被墙率统计
* **CDN分析** - CDN提供商分布和检测类型统计
* **地理分布** - 域名分布国家统计
* **TLS分析** - TLS 1.3、X25519、HTTP/2支持情况
* **证书分析** - 证书有效性和签发者分布
* **性能分析** - 检测时间和响应时间统计
* **智能建议** - 基于检测结果的建议和警告

## ⚡ 性能特性

* **多线程架构** - Worker Pool模式，高效任务分发
* **连接池管理** - 复用TLS和HTTP连接
* **DNS缓存** - 缓存DNS解析结果
* **自适应速率限制** - 根据服务器响应动态调整
* **内存监控** - 实时监控内存使用
* **自适应并发控制** - 根据系统性能动态调整并发数

## 🛠️ 开发指南

### 项目构建

```bash
# 本地开发构建
go build -o reality-checker

# 交叉编译 (多平台支持)
GOOS=linux GOARCH=amd64 go build -o reality-checker-linux-amd64
GOOS=windows GOARCH=amd64 go build -o reality-checker-windows-amd64.exe
GOOS=darwin GOARCH=amd64 go build -o reality-checker-darwin-amd64
GOOS=linux GOARCH=arm64 go build -o reality-checker-linux-arm64
```

### 核心模块

* **Engine** - 核心检测引擎
* **Batch Manager** - 批量检测管理器
* **Detectors** - 各种检测器（被墙、CDN、证书、TLS等）
* **Report Formatter** - 报告格式化系统
* **UI** - 用户界面（横幅、时间戳消息等）

## 🔧 故障排除

### 常见问题

**1. 数据文件下载失败**

```bash
# 检查网络连接
curl -I https://github.com/Loyalsoldier/geoip/releases/latest/download/Country.mmdb
```

**2. 检测结果不准确**

```bash
# 检查数据文件是否存在
ls -la data/
```

**3. 性能问题**

```bash
# 减少批量检测的域名数量
./reality-checker batch "domain1,domain2"
```

## ⚠️ 免责声明

本工具仅用于技术研究和学习目的，请遵守相关法律法规，不得用于非法用途。

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

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

## 🔧 技术实现

* **CDN检测** - 自研多维度检测算法，包括CNAME记录、HTTP响应头、ASN查询等多种检测方法
* **TLS检测** - 原生Go TLS库实现，支持TLS 1.3和X25519检测
* **地理位置检测** - 基于MaxMind GeoIP2数据库的IP地理位置判断

---

**注意**: 本工具仅用于技术研究和学习目的，请遵守当地法律法规，合理使用网络资源。