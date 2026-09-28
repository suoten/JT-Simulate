package websocket

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/suoten/jt-simulate/internal/logger"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// clientEntry 单个 WebSocket 客户端（带写锁保护并发写）
type clientEntry struct {
	conn   *websocket.Conn
	writeMu sync.Mutex // 保护 WriteJSON 的并发访问
}

// Hub WebSocket Hub，管理所有WebSocket连接并广播消息
type Hub struct {
	mu      sync.RWMutex
	clients map[*clientEntry]bool
}

// New 创建Hub
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*clientEntry]bool),
	}
}

// HandleWS 处理WebSocket连接
func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("WebSocket upgrade error", "error", err)
		return
	}

	entry := &clientEntry{conn: conn}

	h.mu.Lock()
	h.clients[entry] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, entry)
		h.mu.Unlock()
		conn.Close()
	}()

	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}
	}
}

// Broadcast 广播消息给所有连接的客户端
func (h *Hub) Broadcast(msg interface{}) {
	h.mu.RLock()
	clients := make([]*clientEntry, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		client.writeMu.Lock()
		err := client.conn.WriteJSON(msg)
		client.writeMu.Unlock()
		if err != nil {
			h.mu.Lock()
			delete(h.clients, client)
			h.mu.Unlock()
			client.conn.Close()
		}
	}
}

// BroadcastMessage 广播一条消息记录（用于实时监控）
func (h *Hub) BroadcastMessage(direction, phone, msgName, msgID string, raw []byte) {
	// 构造前端期望的消息格式
	now := time.Now()
	summary := msgName
	if msgID != "" {
		summary = msgName + " (" + msgID + ")"
	}
	h.Broadcast(map[string]interface{}{
		"direction": direction,
		"phone":     phone,
		"msg_name":  msgName,
		"msg_id":    msgID,
		"summary":   summary,
		"protocol":  "jt808",
		"hex":       hex.EncodeToString(raw),
		"time":      now.Format("15:04:05.000"),
		"raw_hex":   fmt.Sprintf("% X", raw),
	})
}
