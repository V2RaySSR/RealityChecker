package config

import (
	"fmt"
	"os"
	"time"

	"reality-checker-go/internal/types"
	"gopkg.in/yaml.v3"
)

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*types.Config, error) {
	if configPath == "" {
		configPath = "config.yaml"
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 如果配置文件不存在，创建默认配置
		defaultConfig := getDefaultConfig()
		if err := SaveConfig(defaultConfig, configPath); err != nil {
			return nil, fmt.Errorf("创建默认配置文件失败: %v", err)
		}
		return defaultConfig, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析YAML
	var config types.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证和设置默认值
	validateAndSetDefaults(&config)

	return &config, nil
}

// SaveConfig 保存配置文件
func SaveConfig(config *types.Config, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *types.Config {
	return &types.Config{
		Network: types.NetworkConfig{
			Timeout:    5 * time.Second,  // 与config.yaml一致
			Retries:    1,                // 与config.yaml一致
			DNSServers: []string{"114.114.114.114", "223.5.5.5"}, // 使用国内DNS
		},
		TLS: types.TLSConfig{
			MinVersion: 771, // TLS 1.2
			MaxVersion: 772, // TLS 1.3
		},
		Concurrency: types.ConcurrencyConfig{
			MaxConcurrent: 8,              // 与config.yaml一致，适合VPS
			CheckTimeout:  5 * time.Second, // 与config.yaml一致
			CacheTTL:      5 * time.Minute,
		},
		Output: types.OutputConfig{
			Color:   true,
			Verbose: false,
			Format:  "table",
		},
		Cache: types.CacheConfig{
			DNSEnabled:    true,
			ResultEnabled: true,
			TTL:           5 * time.Minute,
			MaxSize:       1000,
		},
		Batch: types.BatchConfig{
			StreamOutput: false,
			ProgressBar:  true,
			ReportFormat: "text",
			Timeout:      30 * time.Second, // 与config.yaml一致
		},
	}
}

// validateAndSetDefaults 验证配置并设置默认值
func validateAndSetDefaults(config *types.Config) {
	// 网络配置验证
	if config.Network.Timeout <= 0 {
		config.Network.Timeout = 30 * time.Second
	}
	if config.Network.Retries < 0 {
		config.Network.Retries = 3
	}
	if len(config.Network.DNSServers) == 0 {
		config.Network.DNSServers = []string{"8.8.8.8", "1.1.1.1"}
	}

	// TLS配置验证
	if config.TLS.MinVersion == 0 {
		config.TLS.MinVersion = 771 // TLS 1.2
	}
	if config.TLS.MaxVersion == 0 {
		config.TLS.MaxVersion = 772 // TLS 1.3
	}

	// 并发配置验证
	if config.Concurrency.MaxConcurrent <= 0 {
		config.Concurrency.MaxConcurrent = 8
	}
	if config.Concurrency.CheckTimeout <= 0 {
		config.Concurrency.CheckTimeout = 30 * time.Second
	}
	if config.Concurrency.CacheTTL <= 0 {
		config.Concurrency.CacheTTL = 5 * time.Minute
	}

	// 输出配置验证
	if config.Output.Format == "" {
		config.Output.Format = "table"
	}

	// 缓存配置验证
	if config.Cache.TTL <= 0 {
		config.Cache.TTL = 5 * time.Minute
	}
	if config.Cache.MaxSize <= 0 {
		config.Cache.MaxSize = 1000
	}

	// 批量配置验证
	if config.Batch.ReportFormat == "" {
		config.Batch.ReportFormat = "text"
	}
	if config.Batch.Timeout <= 0 {
		config.Batch.Timeout = 60 * time.Second
	}
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() string {
	// 按优先级查找配置文件
	paths := []string{
		"./config.yaml",
		"~/reality-checker/config.yaml",
		"/etc/reality-checker/config.yaml",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 返回默认路径
	return "./config.yaml"
}
