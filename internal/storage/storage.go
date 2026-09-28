package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/suoten/jt-simulate/internal/logger"
	"github.com/suoten/jt-simulate/internal/simulator/base"
)

// Storage 设备配置持久化存储接口
type Storage interface {
	SaveDevices(devices []*base.DeviceConfig) error
	LoadDevices() ([]*base.DeviceConfig, error)
	Clear() error
}

// New 根据类型创建存储
func New(path string) *JSONStorage {
	if path == "" {
		path = "data/devices.json"
	}
	// 确保目录存在
	dir := filepath.Dir(path)
	if dir != "" {
		os.MkdirAll(dir, 0755)
	}
	return &JSONStorage{path: path}
}

// JSONStorage JSON 文件持久化（原子写入 + 自动备份）
type JSONStorage struct {
	mu   sync.RWMutex
	path string
}

// SaveDevices 保存设备列表（原子写入：先写临时文件再 rename，避免写入中途崩溃导致数据损坏）
func (s *JSONStorage) SaveDevices(devices []*base.DeviceConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal devices: %w", err)
	}

	// 原子写入：先写到临时文件，然后 rename
	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("write tmp file: %w", err)
	}

	// 在 Windows 上 rename 目标文件已存在时会覆盖，但需要先确保目标目录存在
	if err := os.Rename(tmpPath, s.path); err != nil {
		// rename 失败时尝试直接写入
		if err2 := os.WriteFile(s.path, data, 0644); err2 != nil {
			return fmt.Errorf("rename failed: %w, direct write also failed: %v", err, err2)
		}
	}

	return nil
}

// LoadDevices 加载设备列表（如果主文件损坏，自动尝试从备份恢复）
func (s *JSONStorage) LoadDevices() ([]*base.DeviceConfig, error) {
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
		// 主文件损坏，尝试从备份恢复
		logger.Warn("设备文件解析失败，尝试从备份恢复", "path", s.path, "error", err)
		backupPath := s.path + ".bak"
		backupData, backupErr := os.ReadFile(backupPath)
		if backupErr == nil {
			if json.Unmarshal(backupData, &devices) == nil {
				logger.Info("从备份恢复设备列表成功", "count", len(devices))
				// 恢复后异步写回主文件（不持有锁，避免锁升级死锁）
				go s.writeBackup(data)
				go s.writeBackupToMain(backupData)
				return devices, nil
			}
		}
		return nil, fmt.Errorf("unmarshal devices: %w", err)
	}

	// 加载成功后异步创建备份（不持有锁）
	go s.writeBackup(data)

	return devices, nil
}

// writeBackup 异步创建备份文件（独立获取锁）
func (s *JSONStorage) writeBackup(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = os.WriteFile(s.path+".bak", data, 0644)
}

// writeBackupToMain 异步将备份数据写回主文件（独立获取锁）
func (s *JSONStorage) writeBackupToMain(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = os.WriteFile(s.path, data, 0644)
}

// Clear 清空存储
func (s *JSONStorage) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return err
	}
	// 同时清理备份
	_ = os.Remove(s.path + ".bak")
	return nil
}

// SaveDebounced 防抖保存（短时间内多次调用只保存最后一次）
type debouncedStorage struct {
	inner    Storage
	mu       sync.Mutex
	timer    *time.Timer
	pending  []*base.DeviceConfig
	delay    time.Duration
	flushing bool
}

// NewDebounced 包装一个 Storage，添加防抖功能
func NewDebounced(inner Storage, delay time.Duration) Storage {
	if delay <= 0 {
		delay = 2 * time.Second
	}
	return &debouncedStorage{
		inner: inner,
		delay: delay,
	}
}

// SaveDevices 延迟保存（防抖）
func (d *debouncedStorage) SaveDevices(devices []*base.DeviceConfig) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.pending = devices

	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.delay, func() {
		d.flush()
	})
	return nil
}

// flush 执行实际保存
func (d *debouncedStorage) flush() {
	d.mu.Lock()
	if d.flushing || d.pending == nil {
		d.mu.Unlock()
		return
	}
	devices := d.pending
	d.pending = nil
	d.flushing = true
	d.mu.Unlock()

	if err := d.inner.SaveDevices(devices); err != nil {
		logger.Error("防抖保存设备失败", "error", err)
	}

	d.mu.Lock()
	d.flushing = false
	d.mu.Unlock()
}

// Flush 立即执行挂起的保存
func (d *debouncedStorage) Flush() error {
	d.mu.Lock()
	if d.pending == nil {
		d.mu.Unlock()
		return nil
	}
	devices := d.pending
	d.pending = nil
	d.mu.Unlock()

	return d.inner.SaveDevices(devices)
}

// LoadDevices 委托给内部存储
func (d *debouncedStorage) LoadDevices() ([]*base.DeviceConfig, error) {
	return d.inner.LoadDevices()
}

// Clear 委托给内部存储
func (d *debouncedStorage) Clear() error {
	d.mu.Lock()
	if d.timer != nil {
		d.timer.Stop()
	}
	d.pending = nil
	d.mu.Unlock()
	return d.inner.Clear()
}

// noopStorage 空实现（不持久化）
type noopStorage struct{}

func (noopStorage) SaveDevices([]*base.DeviceConfig) error { return nil }
func (noopStorage) LoadDevices() ([]*base.DeviceConfig, error) { return nil, nil }
func (noopStorage) Clear() error { return nil }

// NoopStorage 返回一个空操作的 Storage
func NoopStorage() Storage {
	return noopStorage{}
}
