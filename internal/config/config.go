package config

// Config 应用级配置
// 包含应用运行所需的所有配置
type Config struct {
	// 应用基础配置
	DataDir string `yaml:"data_dir"` // e.g., "./data" - directory for data storage

	// Database 配置
	Database DatabaseConfig `yaml:"database"` // Database configuration

	// Transport 配置
	Transport TransportConfig `yaml:"transport"` // Transport configuration
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	DomainDBPath           string `yaml:"domain_db_path"`            // Domain 数据库路径
	MaxOpenConns           int    `yaml:"max_open_conns"`            // 最大打开连接数
	MaxIdleConns           int    `yaml:"max_idle_conns"`            // 最大空闲连接数
	ConnMaxLifetimeSeconds int    `yaml:"conn_max_lifetime_seconds"` // 连接最大生存时间（秒）
}

// TransportConfig Transport 配置
type TransportConfig struct {
	HTTP HTTPConfig `yaml:"http"` // HTTP server configuration
}

// HTTPConfig HTTP 服务器配置
type HTTPConfig struct {
	Port int `yaml:"port"` // e.g., 8080 - HTTP server port
}
