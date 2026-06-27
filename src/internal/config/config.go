package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	DNS       DNSConfig       `yaml:"dns"`
	Scanner   ScannerConfig   `yaml:"scanner"`
	Logger    LoggerConfig    `yaml:"logger"`
	Domains   []string        `yaml:"domains"`
	Upstream  []string        `yaml:"upstream"`
}

// DNSConfig DNS服务器配置
type DNSConfig struct {
	ListenAddr string `yaml:"listen_addr"`
	ListenPort int    `yaml:"listen_port"`
}

// ScannerConfig IP扫描器配置
type ScannerConfig struct {
	Interval      time.Duration `yaml:"interval"`
	Timeout       time.Duration `yaml:"timeout"`
	Concurrency   int           `yaml:"concurrency"`
	TestPort      int           `yaml:"test_port"`
	MaxLatency    time.Duration `yaml:"max_latency"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	OutputPath string `yaml:"output_path"`
	MaxSize    int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age_days"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		DNS: DNSConfig{
			ListenAddr: "127.0.0.1",
			ListenPort: 53,
		},
		Scanner: ScannerConfig{
			Interval:    5 * time.Minute,
			Timeout:     3 * time.Second,
			Concurrency: 100,
			TestPort:    443,
			MaxLatency:  5 * time.Second,
		},
		Logger: LoggerConfig{
			Level:      "info",
			Format:     "json",
			OutputPath: "tmp/logs/app.log",
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
		},
		Domains: []string{
			"github.com",
			"www.github.com",
			"api.github.com",
			"raw.githubusercontent.com",
			"gist.github.com",
			"github.io",
			"avatars.githubusercontent.com",
			"camo.githubusercontent.com",
			"cloud.githubusercontent.com",
			"gist.githubusercontent.com",
			"user-images.githubusercontent.com",
			"docs.github.com",
			"githubusercontent.com",
			"objects.githubusercontent.com",
			"media.githubusercontent.com",
			"codeload.github.com",
			"github.githubassets.com",
			"githubstatus.com",
			"support.github.com",
			"github.dev",
		},
		Upstream: []string{
			"8.8.8.8:53",
			"8.8.4.4:53",
			"114.114.114.114:53",
		},
	}
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return cfg, nil
}

// Save 保存配置到文件
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}
