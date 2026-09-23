package engine

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Scenario 场景定义
type Scenario struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Protocol    string         `json:"protocol"`
	Steps       []ScenarioStep `json:"steps"`
}

// ScenarioStep 场景步骤
type ScenarioStep struct {
	Name     string                 `json:"name"`
	Action   string                 `json:"action"`   // send, wait, alarm, location
	Delay    int                    `json:"delay"`    // 延迟(毫秒)
	Params   map[string]interface{} `json:"params"`
}

// ScenarioEngine 场景引擎
type ScenarioEngine struct {
	mu           sync.RWMutex
	scenarios    map[string]*Scenario
	running      bool
	runningName  string
	cancel       context.CancelFunc
	currentStep  int
	totalSteps   int
	stepProgress chan string // 步骤名称
}

// NewScenarioEngine 创建场景引擎
func NewScenarioEngine() *ScenarioEngine {
	return &ScenarioEngine{
		scenarios:    make(map[string]*Scenario),
		stepProgress: make(chan string, 64),
	}
}

// Register 注册场景
func (e *ScenarioEngine) Register(s *Scenario) {
	e.mu.Lock()
	defer e.mu.Unlock()
	key := s.ID
	if key == "" {
		key = s.Name
	}
	e.scenarios[key] = s
}

// List 列出所有场景
func (e *ScenarioEngine) List() []*Scenario {
	e.mu.RLock()
	defer e.mu.RUnlock()
	list := make([]*Scenario, 0, len(e.scenarios))
	for _, s := range e.scenarios {
		list = append(list, s)
	}
	return list
}

// Get 获取场景
func (e *ScenarioEngine) Get(name string) (*Scenario, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	// 先按 ID 查找
	s, ok := e.scenarios[name]
	if ok {
		return s, nil
	}
	// 再按名称查找
	for _, sc := range e.scenarios {
		if sc.Name == name {
			return sc, nil
		}
	}
	return nil, fmt.Errorf("scenario %s not found", name)
}

// Run 执行场景
func (e *ScenarioEngine) Run(ctx context.Context, name string, stepHandler func(step *ScenarioStep) error) error {
	s, err := e.Get(name)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	e.mu.Lock()
	e.running = true
	e.runningName = s.Name
	e.cancel = cancel
	e.currentStep = 0
	e.totalSteps = len(s.Steps)
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		e.running = false
		e.runningName = ""
		e.currentStep = 0
		e.totalSteps = 0
		e.mu.Unlock()
	}()

	for i, step := range s.Steps {
		e.mu.Lock()
		e.currentStep = i
		e.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(step.Delay) * time.Millisecond):
		}

		if err := stepHandler(&step); err != nil {
			return fmt.Errorf("step %d (%s): %w", i, step.Name, err)
		}
	}

	return nil
}

// Stop 停止场景
func (e *ScenarioEngine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.cancel != nil {
		e.cancel()
	}
	e.running = false
}

// IsRunning 是否正在运行
func (e *ScenarioEngine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

// RunningName 获取当前运行的场景名称
func (e *ScenarioEngine) RunningName() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.runningName
}

// Progress 获取当前运行进度
func (e *ScenarioEngine) Progress() (current, total int) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.currentStep, e.totalSteps
}

// DefaultScenarios 预置场景
func DefaultScenarios() []*Scenario {
	return []*Scenario{
		{
			ID:          "vehicle_online",
			Name:        "终端上线流程",
			Description: "模拟终端注册→鉴权→心跳→位置上报完整流程",
			Protocol:    "jt808",
			Steps: []ScenarioStep{
				{Name: "注册", Action: "send", Delay: 0, Params: map[string]interface{}{"msg_id": "0x0100"}},
				{Name: "等待应答", Action: "wait", Delay: 500, Params: map[string]interface{}{"msg_id": "0x8100"}},
				{Name: "鉴权", Action: "send", Delay: 100, Params: map[string]interface{}{"msg_id": "0x0102"}},
				{Name: "心跳", Action: "send", Delay: 1000, Params: map[string]interface{}{"msg_id": "0x0002"}},
				{Name: "位置上报", Action: "location", Delay: 500, Params: map[string]interface{}{"lat": 39.9093, "lon": 116.3974, "speed": 60}},
			},
		},
		{
			ID:          "overspeed_alarm",
			Name:        "超速报警",
			Description: "模拟车辆从正常速度逐渐加速到超速并触发报警",
			Protocol:    "jt808",
			Steps: []ScenarioStep{
				{Name: "正常行驶", Action: "location", Delay: 0, Params: map[string]interface{}{"speed": 80}},
				{Name: "加速", Action: "location", Delay: 2000, Params: map[string]interface{}{"speed": 100}},
				{Name: "超速", Action: "location", Delay: 2000, Params: map[string]interface{}{"speed": 120}},
				{Name: "超速报警", Action: "alarm", Delay: 500, Params: map[string]interface{}{"type": "overspeed"}},
				{Name: "减速", Action: "location", Delay: 3000, Params: map[string]interface{}{"speed": 100}},
				{Name: "恢复正常", Action: "location", Delay: 2000, Params: map[string]interface{}{"speed": 60}},
			},
		},
		{
			ID:          "fatigue_drive",
			Name:        "疲劳驾驶",
			Description: "模拟连续驾驶超过4小时触发疲劳驾驶报警",
			Protocol:    "jt808",
			Steps: []ScenarioStep{
				{Name: "开始驾驶", Action: "location", Delay: 0, Params: map[string]interface{}{"speed": 80}},
				{Name: "持续驾驶1", Action: "location", Delay: 5000, Params: map[string]interface{}{"speed": 80}},
				{Name: "持续驾驶2", Action: "location", Delay: 5000, Params: map[string]interface{}{"speed": 80}},
				{Name: "疲劳报警", Action: "alarm", Delay: 500, Params: map[string]interface{}{"type": "fatigue"}},
			},
		},
		{
			ID:          "emergency_alarm",
			Name:        "紧急报警",
			Description: "模拟触发紧急按钮报警",
			Protocol:    "jt808",
			Steps: []ScenarioStep{
				{Name: "正常行驶", Action: "location", Delay: 0, Params: map[string]interface{}{"speed": 60}},
				{Name: "触发紧急报警", Action: "alarm", Delay: 1000, Params: map[string]interface{}{"type": "emergency"}},
				{Name: "继续行驶", Action: "location", Delay: 2000, Params: map[string]interface{}{"speed": 40}},
			},
		},
		{
			ID:          "offline_reconnect",
			Name:        "离线重连",
			Description: "模拟终端断线后重连并补报历史位置",
			Protocol:    "jt808",
			Steps: []ScenarioStep{
				{Name: "正常位置", Action: "location", Delay: 0, Params: map[string]interface{}{"speed": 60}},
				{Name: "断线", Action: "disconnect", Delay: 1000, Params: map[string]interface{}{}},
				{Name: "等待重连", Action: "wait", Delay: 5000, Params: map[string]interface{}{}},
				{Name: "重新注册", Action: "send", Delay: 0, Params: map[string]interface{}{"msg_id": "0x0100"}},
				{Name: "补报位置", Action: "send", Delay: 500, Params: map[string]interface{}{"msg_id": "0x0704"}},
			},
		},
	}
}
