package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/suoten/jt-simulate/internal/simulator/base"
)

// Storage 设备配置持久化存储
type Storage struct {
	mu   sync.RWMutex
	path string
}

// New 创建存储
func New(path string) *Storage {
	if path == "" {
		path = "data/devices.json"
	}
	// 确保目录存在
	dir := filepath.Dir(path)
	if dir != "" {
		os.MkdirAll(dir, 0755)
	}
	return &Storage{path: path}
}

// SaveDevices 保存设备列表
func (s *Storage) SaveDevices(devices []*base.DeviceConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal devices: %w", err)
	}
	return os.WriteFile(s.path, data, 0644)
}

// LoadDevices 加载设备列表
func (s *Storage) LoadDevices() ([]*base.DeviceConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 文件不存在不是错误
		}
		return nil, fmt.Errorf("read devices: %w", err)
	}

	var devices []*base.DeviceConfig
	if err := json.Unmarshal(data, &devices); err != nil {
		return nil, fmt.Errorf("unmarshal devices: %w", err)
	}
	return devices, nil
}

// Clear 清空存储
func (s *Storage) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return os.Remove(s.path)
}
