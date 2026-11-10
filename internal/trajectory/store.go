package trajectory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Store 轨迹文件存储（JSON 落盘，目录默认 data/routes）
type Store struct {
	dir string
}

// NewStore 创建存储（目录不存在时在保存时自动创建，构造不失败）
func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

// SanitizeName 清洗轨迹名称：仅保留中文、字母、数字、下划线、连字符，
// 其余字符替换为下划线；长度限制 1-64。返回空视为非法。
func SanitizeName(name string) (string, error) {
	var b strings.Builder
	for _, r := range strings.TrimSpace(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
			b.WriteRune(r)
		case r >= 0x4e00 && r <= 0x9fff: // CJK 基本区
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	s := b.String()
	// 去除首尾下划线与内部的连续下划线
	s = strings.Trim(s, "_")
	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}
	if s == "" {
		return "", fmt.Errorf("轨迹名称不能为空（仅支持中文、字母、数字、下划线、连字符）")
	}
	runes := []rune(s)
	if len(runes) > 64 {
		s = string(runes[:64])
	}
	// 防御：不允许形如 "." 或 ".." 的目录语义
	if s == "." || s == ".." {
		return "", fmt.Errorf("轨迹名称非法")
	}
	return s, nil
}

func (s *Store) fileOf(name string) (string, error) {
	clean, err := SanitizeName(name)
	if err != nil {
		return "", err
	}
	return filepath.Join(s.dir, clean+".json"), nil
}

// Save 保存轨迹（原子写：临时文件 + 重命名）。名称会被清洗并回写到 route.Name。
func (s *Store) Save(r *Route) error {
	clean, err := SanitizeName(r.Name)
	if err != nil {
		return err
	}
	r.Name = clean
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("创建轨迹目录失败: %w", err)
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化轨迹失败: %w", err)
	}
	target := filepath.Join(s.dir, clean+".json")
	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("写入轨迹文件失败: %w", err)
	}
	if err := os.Rename(tmp, target); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("保存轨迹文件失败: %w", err)
	}
	return nil
}

// Get 按名称加载轨迹（含重算累计距离）
func (s *Store) Get(name string) (*Route, error) {
	path, err := s.fileOf(name)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("轨迹 %q 不存在", name)
		}
		return nil, fmt.Errorf("读取轨迹失败: %w", err)
	}
	var r Route
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("解析轨迹文件失败: %w", err)
	}
	if len(r.Points) < 2 {
		return nil, fmt.Errorf("轨迹 %q 数据不完整", r.Name)
	}
	r.recompute()
	return &r, nil
}

// List 列出所有轨迹（按名称排序）
func (s *Store) List() ([]*Route, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Route{}, nil
		}
		return nil, err
	}
	routes := make([]*Route, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(s.dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var r Route
		if err := json.Unmarshal(data, &r); err != nil || len(r.Points) < 2 {
			continue // 跳过损坏文件
		}
		r.recompute()
		routes = append(routes, &r)
	}
	sort.Slice(routes, func(i, j int) bool { return routes[i].Name < routes[j].Name })
	return routes, nil
}

// Delete 删除轨迹文件
func (s *Store) Delete(name string) error {
	path, err := s.fileOf(name)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("轨迹 %q 不存在", name)
		}
		return fmt.Errorf("删除轨迹失败: %w", err)
	}
	return nil
}
