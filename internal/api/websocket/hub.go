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

const (
	// clientSendQueue 单个客户端的发送队列长度；积满视为慢客户端并踢除，
	// 避免一个停止读取的浏览器页签阻塞（乃至停摆）所有设备的消息广播
	clientSendQueue = 64
	// wsWriteTimeout 单条消息写超时：超过即判定客户端失活
	wsWriteTimeout = 10 * time.Second
	// wsPingInterval 服务端 ping 周期（配合读空闲超时检测半开连接）
	wsPingInterval = 30 * time.Second
	// wsReadIdleTimeout 读空闲超时：超过未收到任何数据/pong 即断开
	wsReadIdleTimeout = 90 * time.Second
)

// clientEntry 单个 WebSocket 客户端。
// 写操作全部由独立的写 goroutine 串行执行（经 send 队列），
// 因此无需写锁；done 保证退出信号只广播一次。
type clientEntry struct {
	conn *websocket.Conn
	send chan interface{}
	done chan struct{}
	once sync.Once
}

func (c *clientEntry) shutdown() {
	c.once.Do(func() {
		close(c.done)
		c.conn.Close()
	})
}

// Hub WebSocket Hub，管理所有WebSocket连接并广播消息
type Hub struct {
	mu      sync.RWMutex
	clients map[*clientEntry]bool
	closed  bool
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

	entry := &clientEntry{
		conn: conn,
		send: make(chan interface{}, clientSendQueue),
		done: make(chan struct{}),
	}

	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		conn.Close()
		return
	}
	h.clients[entry] = true
	h.mu.Unlock()

	// 写 goroutine：串行消费发送队列，带写超时和 ping 保活
	go h.writePump(entry)

	// 读循环：驱动 pong/保活检测；读到错误（客户端关闭/超时）时清理
	defer h.removeClient(entry)
	conn.SetReadDeadline(time.Now().Add(wsReadIdleTimeout))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(wsReadIdleTimeout))
		return nil
	})
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}
	}
}

// writePump 单客户端写循环
func (h *Hub) writePump(c *clientEntry) {
	ticker := time.NewTicker(wsPingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.done:
			return
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
			if err := c.conn.WriteJSON(msg); err != nil {
				h.removeClient(c)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				h.removeClient(c)
				return
			}
		}
	}
}

// removeClient 移除并关闭客户端（幂等）
func (h *Hub) removeClient(c *clientEntry) {
	h.mu.Lock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
	} else {
		h.mu.Unlock()
		c.shutdown()
		return
	}
	h.mu.Unlock()
	c.shutdown()
}

// Broadcast 广播消息给所有连接的客户端。
// 非阻塞投递：队列积满的慢客户端被直接踢除，绝不阻塞调用方
// （调用方是设备心跳/位置上报 goroutine，阻塞会停摆设备）。
func (h *Hub) Broadcast(msg interface{}) {
	h.mu.RLock()
	if h.closed {
		h.mu.RUnlock()
		return
	}
	clients := make([]*clientEntry, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.send <- msg:
		default:
			// 发送队列已满：客户端消费过慢，踢除
			h.removeClient(client)
		}
	}
}

// Close 关闭 Hub（进程关闭时调用）：踢除全部客户端，
// 否则被 hijack 的 WebSocket 连接不受 http.Server.Shutdown 影响
func (h *Hub) Close() {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}
	h.closed = true
	clients := make([]*clientEntry, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.clients = make(map[*clientEntry]bool)
	h.mu.Unlock()

	for _, c := range clients {
		c.shutdown()
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
