package data

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// DataFile 数据文件配置
type DataFile struct {
	Name      string
	URL       string
	LocalPath string
}

// Downloader 数据文件下载器
type Downloader struct {
	timeout time.Duration
	retries int
	retryDelay time.Duration
}

// NewDownloader 创建下载器
func NewDownloader() *Downloader {
	return &Downloader{
		timeout:    30 * time.Second,
		retries:    3,
		retryDelay: 2 * time.Second,
	}
}

// EnsureDataFiles 确保所有数据文件存在且最新
func (d *Downloader) EnsureDataFiles() error {
	// 定义需要下载的文件
	files := []DataFile{
		{
			Name:      "cdn_keywords.txt",
			URL:       "https://raw.githubusercontent.com/V2RaySSR/RealityChecker/main/data/cdn_keywords.txt",
			LocalPath: "data/cdn_keywords.txt",
		},
		{
			Name:      "hot_websites.txt",
			URL:       "https://raw.githubusercontent.com/V2RaySSR/RealityChecker/main/data/hot_websites.txt",
			LocalPath: "data/hot_websites.txt",
		},
		{
			Name:      "gfwlist.conf",
			URL:       "https://raw.githubusercontent.com/Loyalsoldier/clash-rules/release/gfw.txt",
			LocalPath: "data/gfwlist.conf",
		},
		{
			Name:      "Country.mmdb",
			URL:       "https://github.com/Loyalsoldier/geoip/releases/latest/download/Country.mmdb",
			LocalPath: "data/Country.mmdb",
		},
	}

	// 确保data目录存在
	if err := os.MkdirAll("data", 0755); err != nil {
		return fmt.Errorf("创建data目录失败: %v", err)
	}

	// 检查并下载每个文件
	for _, file := range files {
		if err := d.ensureFile(file); err != nil {
			return err
		}
	}

	return nil
}

// ensureFile 确保单个文件存在且最新
func (d *Downloader) ensureFile(file DataFile) error {
	// 检查文件是否存在
	exists, err := d.fileExists(file.LocalPath)
	if err != nil {
		return fmt.Errorf("检查文件 %s 失败: %v", file.Name, err)
	}

	// 如果文件不存在，直接下载
	if !exists {
		return d.downloadWithRetry(file)
	}

	// 检查文件是否需要更新（3天）
	needsUpdate, err := d.needsUpdate(file.LocalPath)
	if err != nil {
		return fmt.Errorf("检查文件 %s 更新时间失败: %v", file.Name, err)
	}

	// 如果需要更新，下载新文件
	if needsUpdate {
		return d.downloadWithRetry(file)
	}

	return nil
}

// fileExists 检查文件是否存在
func (d *Downloader) fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// needsUpdate 检查文件是否需要更新（超过3天）
func (d *Downloader) needsUpdate(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	// 检查文件修改时间是否超过3天
	threeDaysAgo := time.Now().Add(-3 * 24 * time.Hour)
	return info.ModTime().Before(threeDaysAgo), nil
}

// downloadWithRetry 带重试的下载
func (d *Downloader) downloadWithRetry(file DataFile) error {
	for i := 0; i < d.retries; i++ {
		if i > 0 {
			fmt.Printf("重试中... (%d/%d)\n", i, d.retries)
			time.Sleep(d.retryDelay)
		}

		err := d.downloadFile(file)
		if err == nil {
			return nil // 成功
		}

		fmt.Printf("错误：下载 %s 失败 - %s %v\n", file.Name, file.URL, err)
	}

	// 所有重试都失败了，显示手动下载说明
	d.showManualDownloadInstructions()
	return fmt.Errorf("下载失败，已重试 %d 次", d.retries)
}

// downloadFile 下载单个文件
func (d *Downloader) downloadFile(file DataFile) error {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: d.timeout,
	}

	// 发送请求
	resp, err := client.Get(file.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// 创建临时文件
	tmpFile := file.LocalPath + ".tmp"
	out, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer out.Close()

	// 复制数据
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(tmpFile) // 清理临时文件
		return err
	}

	// 关闭文件
	out.Close()

	// 原子性替换原文件
	if err := os.Rename(tmpFile, file.LocalPath); err != nil {
		os.Remove(tmpFile) // 清理临时文件
		return err
	}

	return nil
}

// showManualDownloadInstructions 显示手动下载说明
func (d *Downloader) showManualDownloadInstructions() {
	fmt.Println("程序终止：缺少必要的数据文件")
	fmt.Println()
	fmt.Println("请手动下载以下文件到 data/ 目录：")
	fmt.Println("1. cdn_keywords.txt: https://raw.githubusercontent.com/V2RaySSR/RealityChecker/main/data/cdn_keywords.txt")
	fmt.Println("2. hot_websites.txt: https://raw.githubusercontent.com/V2RaySSR/RealityChecker/main/data/hot_websites.txt")
	fmt.Println("3. gfwlist.conf: https://raw.githubusercontent.com/Loyalsoldier/clash-rules/release/gfw.txt")
	fmt.Println("4. Country.mmdb: https://github.com/Loyalsoldier/geoip/releases/latest/download/Country.mmdb")
	fmt.Println()
	fmt.Println("下载完成后重新运行程序即可。")
}
