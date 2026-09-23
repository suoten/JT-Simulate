package websocket

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Hub WebSocket Hub，管理所有WebSocket连接并广播消息
type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]bool
}

// New 创建Hub
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]bool),
	}
}

// HandleWS 处理WebSocket连接
func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
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
	clients := make([]*websocket.Conn, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		if err := client.WriteJSON(msg); err != nil {
			h.mu.Lock()
			delete(h.clients, client)
			h.mu.Unlock()
			client.Close()
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
