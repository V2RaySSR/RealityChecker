package version

// 版本信息，由构建时注入
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

// GetVersion 获取版本号
func GetVersion() string {
	return Version
}

// GetCommit 获取提交哈希
func GetCommit() string {
	return Commit
}

// GetBuildTime 获取构建时间
func GetBuildTime() string {
	return BuildTime
}
