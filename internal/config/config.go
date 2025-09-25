package config

import (
	"fmt"
	"os"
	"time"

	"RealityChecker/internal/types"

	"gopkg.in/yaml.v3"
)

// LoadConfig 加载配置
func LoadConfig(configPath string) (*types.Config, error) {
	// 获取默认配置
	config := getDefaultConfig()

	// 如果提供了配置文件路径，尝试加载
	if configPath != "" {
		if err := loadConfigFromFile(config, configPath); err != nil {
			return nil, fmt.Errorf("加载配置文件失败: %v", err)
		}
	} else {
		// 尝试从默认位置加载配置文件
		defaultPaths := []string{
			"config.yaml",
			"config.yml",
			"./config.yaml",
			"./config.yml",
		}

		for _, path := range defaultPaths {
			if _, err := os.Stat(path); err == nil {
				if err := loadConfigFromFile(config, path); err == nil {
					break // 成功加载，跳出循环
				}
			}
		}
	}

	// 验证并设置默认值
	validateAndSetDefaults(config)
	return config, nil
}

// loadConfigFromFile 从文件加载配置
func loadConfigFromFile(config *types.Config, filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", filePath)
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析YAML
	var fileConfig types.Config
	if err := yaml.Unmarshal(data, &fileConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 合并配置（文件配置覆盖默认配置）
	mergeConfig(config, &fileConfig)

	return nil
}

// mergeConfig 合并配置
func mergeConfig(defaultConfig *types.Config, fileConfig *types.Config) {
	// 网络配置
	if fileConfig.Network.Timeout > 0 {
		defaultConfig.Network.Timeout = fileConfig.Network.Timeout
	}
	if fileConfig.Network.Retries >= 0 {
		defaultConfig.Network.Retries = fileConfig.Network.Retries
	}
	if len(fileConfig.Network.DNSServers) > 0 {
		defaultConfig.Network.DNSServers = fileConfig.Network.DNSServers
	}

	// TLS配置
	if fileConfig.TLS.MinVersion > 0 {
		defaultConfig.TLS.MinVersion = fileConfig.TLS.MinVersion
	}
	if fileConfig.TLS.MaxVersion > 0 {
		defaultConfig.TLS.MaxVersion = fileConfig.TLS.MaxVersion
	}

	// 并发配置
	if fileConfig.Concurrency.MaxConcurrent > 0 {
		defaultConfig.Concurrency.MaxConcurrent = fileConfig.Concurrency.MaxConcurrent
	}
	if fileConfig.Concurrency.CheckTimeout > 0 {
		defaultConfig.Concurrency.CheckTimeout = fileConfig.Concurrency.CheckTimeout
	}
	if fileConfig.Concurrency.CacheTTL > 0 {
		defaultConfig.Concurrency.CacheTTL = fileConfig.Concurrency.CacheTTL
	}

	// 输出配置
	if fileConfig.Output.Format != "" {
		defaultConfig.Output.Format = fileConfig.Output.Format
	}
	defaultConfig.Output.Color = fileConfig.Output.Color
	defaultConfig.Output.Verbose = fileConfig.Output.Verbose

	// 缓存配置
	defaultConfig.Cache.DNSEnabled = fileConfig.Cache.DNSEnabled
	defaultConfig.Cache.ResultEnabled = fileConfig.Cache.ResultEnabled
	if fileConfig.Cache.TTL > 0 {
		defaultConfig.Cache.TTL = fileConfig.Cache.TTL
	}
	if fileConfig.Cache.MaxSize > 0 {
		defaultConfig.Cache.MaxSize = fileConfig.Cache.MaxSize
	}

	// 批量配置
	defaultConfig.Batch.StreamOutput = fileConfig.Batch.StreamOutput
	defaultConfig.Batch.ProgressBar = fileConfig.Batch.ProgressBar
	if fileConfig.Batch.ReportFormat != "" {
		defaultConfig.Batch.ReportFormat = fileConfig.Batch.ReportFormat
	}
	if fileConfig.Batch.Timeout > 0 {
		defaultConfig.Batch.Timeout = fileConfig.Batch.Timeout
	}
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *types.Config {
	return &types.Config{
		Network: types.NetworkConfig{
			Timeout:    3 * time.Second, // 减少到3秒
			Retries:    1,
			DNSServers: []string{"8.8.8.8", "1.1.1.1"},
		},
		TLS: types.TLSConfig{
			MinVersion: 771, // TLS 1.2
			MaxVersion: 772, // TLS 1.3
		},
		Concurrency: types.ConcurrencyConfig{
			MaxConcurrent: 8,
			CheckTimeout:  3 * time.Second, // 减少到3秒
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
			Timeout:      30 * time.Second,
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
