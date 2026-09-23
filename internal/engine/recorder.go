package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// RecordedMessage 录制的消息
type RecordedMessage struct {
	Timestamp int64  `json:"timestamp"` // Unix毫秒
	Direction string `json:"direction"`  // send, recv
	Protocol  string `json:"protocol"`
	Hex       string `json:"hex"`
	MsgID     string `json:"msg_id"`
	Phone     string `json:"phone"`
}

// Recording 录制会话
type Recording struct {
	Name      string            `json:"name"`
	Protocol  string            `json:"protocol"`
	StartTime int64             `json:"start_time"`
	EndTime   int64             `json:"end_time"`
	Messages  []RecordedMessage `json:"messages"`
}

// Recorder 消息录制器
type Recorder struct {
	mu        sync.Mutex
	recording *Recording
	active    bool
}

// NewRecorder 创建录制器
func NewRecorder() *Recorder {
	return &Recorder{}
}

// Start 开始录制
func (r *Recorder) Start(name, protocol string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recording = &Recording{
		Name:      name,
		Protocol:  protocol,
		StartTime: time.Now().UnixMilli(),
	}
	r.active = true
}

// Record 记录一条消息
func (r *Recorder) Record(direction, protocol, hex, msgID, phone string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.active || r.recording == nil {
		return
	}
	r.recording.Messages = append(r.recording.Messages, RecordedMessage{
		Timestamp: time.Now().UnixMilli(),
		Direction: direction,
		Protocol:  protocol,
		Hex:       hex,
		MsgID:     msgID,
		Phone:     phone,
	})
}

// Stop 停止录制
func (r *Recorder) Stop() *Recording {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.active = false
	if r.recording != nil {
		r.recording.EndTime = time.Now().UnixMilli()
	}
	return r.recording
}

// IsActive 是否正在录制
func (r *Recorder) IsActive() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.active
}

// Save 保存到文件
func (r *Recorder) Save(path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.recording == nil {
		return fmt.Errorf("no recording to save")
	}
	data, err := json.MarshalIndent(r.recording, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadRecording 从文件加载录制
func LoadRecording(path string) (*Recording, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rec Recording
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// Player 回放播放器
type Player struct {
	recording *Recording
	mu        sync.Mutex
	playing   bool
	cancel    chan struct{}
}

// NewPlayer 创建播放器
func NewPlayer(rec *Recording) *Player {
	return &Player{
		recording: rec,
		cancel:    make(chan struct{}),
	}
}

// Play 回放消息
func (p *Player) Play(handler func(msg *RecordedMessage)) error {
	p.mu.Lock()
	if p.playing {
		p.mu.Unlock()
		return fmt.Errorf("already playing")
	}
	p.playing = true
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		p.playing = false
		p.mu.Unlock()
	}()

	if len(p.recording.Messages) == 0 {
		return nil
	}

	baseTime := p.recording.Messages[0].Timestamp

	for _, msg := range p.recording.Messages {
		select {
		case <-p.cancel:
			return nil
		default:
		}

		// 等待到消息的时间
		delay := time.Duration(msg.Timestamp-baseTime) * time.Millisecond
		if delay > 0 {
			time.Sleep(delay)
		}

		handler(&msg)
		baseTime = msg.Timestamp
	}

	return nil
}

// Stop 停止回放
func (p *Player) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.playing {
		close(p.cancel)
	}
}

// IsPlaying 是否正在播放
func (p *Player) IsPlaying() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playing
}
