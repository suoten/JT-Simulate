package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config 全局配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Log      LogConfig      `mapstructure:"log"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Simulator SimulatorConfig `mapstructure:"simulator"`
	Targets  []TargetConfig `mapstructure:"targets"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // json, console
	Output string `mapstructure:"output"` // stdout, file
	Dir    string `mapstructure:"dir"`
}

type StorageConfig struct {
	Type string `mapstructure:"type"` // sqlite, memory
	Path string `mapstructure:"path"`
}

type SimulatorConfig struct {
	DefaultProtocol string `mapstructure:"default_protocol"`
	HeartbeatInterval int  `mapstructure:"heartbeat_interval"`
	LocationInterval  int  `mapstructure:"location_interval"`
	ReconnectInterval int  `mapstructure:"reconnect_interval"`
}

type TargetConfig struct {
	Name     string `mapstructure:"name"`
	Protocol string `mapstructure:"protocol"`
	Address  string `mapstructure:"address"`
	Active   bool   `mapstructure:"active"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".jt-simulate")

	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8095,
			Mode: "debug",
		},
		Log: LogConfig{
			Level:  "info",
			Format: "console",
			Output: "stdout",
			Dir:    filepath.Join(dataDir, "logs"),
		},
		Storage: StorageConfig{
			Type: "sqlite",
			Path: filepath.Join(dataDir, "jt-simulate.db"),
		},
		Simulator: SimulatorConfig{
			DefaultProtocol: "jt808",
			HeartbeatInterval: 60,
			LocationInterval:  30,
			ReconnectInterval: 15,
		},
		Targets: []TargetConfig{
			{
				Name:     "default",
				Protocol: "jt808",
				Address:  "127.0.0.1:7611",
				Active:   true,
			},
		},
	}
}

// Load 加载配置
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	v := viper.New()
	v.SetConfigType("yaml")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
		v.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".jt-simulate"))
	}

	// 设置默认值
	v.SetDefault("server.host", cfg.Server.Host)
	v.SetDefault("server.port", cfg.Server.Port)
	v.SetDefault("server.mode", cfg.Server.Mode)
	v.SetDefault("log.level", cfg.Log.Level)
	v.SetDefault("log.format", cfg.Log.Format)
	v.SetDefault("log.output", cfg.Log.Output)
	v.SetDefault("storage.type", cfg.Storage.Type)
	v.SetDefault("storage.path", cfg.Storage.Path)
	v.SetDefault("simulator.default_protocol", cfg.Simulator.DefaultProtocol)
	v.SetDefault("simulator.heartbeat_interval", cfg.Simulator.HeartbeatInterval)
	v.SetDefault("simulator.location_interval", cfg.Simulator.LocationInterval)
	v.SetDefault("simulator.reconnect_interval", cfg.Simulator.ReconnectInterval)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
		// 配置文件不存在时使用默认配置
		return cfg, nil
	}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}

// Save 保存配置
func Save(cfg *Config, configPath string) error {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(configPath)

	v.Set("server", cfg.Server)
	v.Set("log", cfg.Log)
	v.Set("storage", cfg.Storage)
	v.Set("simulator", cfg.Simulator)
	v.Set("targets", cfg.Targets)

	return v.WriteConfig()
}
