package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// LoadConfig 从文件加载配置并应用默认值
func LoadConfig(file string) (*Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// 应用默认值
	ApplyDefaults(cfg)

	return cfg, nil
}

// ApplyDefaults 为配置项设置默认值
func ApplyDefaults(cfg *Config) {
	// 数据目录默认值
	if cfg.DataDir == "" {
		cfg.DataDir = "./data"
	}

	// HTTP 配置默认值
	if cfg.Transport.HTTP.Port == 0 {
		cfg.Transport.HTTP.Port = 8080 // 默认 HTTP 端口
	}

	// RPC 配置默认值
	if cfg.Transport.RPC.Registry.Port == 0 {
		cfg.Transport.RPC.Registry.Port = 50010 // 默认 Registry RPC 端口
	}
}
