package service

import (
	"blog/model"
	"blog/utils"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// writeWait 写入超时时间
	writeWait = 10 * time.Second
	// pongWait 等待客户端 pong 的最大时间
	pongWait = 60 * time.Second
	// pingPeriod 服务端发送 ping 的周期，必须小于 pongWait
	pingPeriod = (pongWait * 9) / 10
	// sendBufferSize 每个客户端发送通道缓冲大小
	sendBufferSize = 256
)

// websocketMessage 通过 WebSocket 下发的统一消息格式。
type websocketMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// Client 表示一个已连接的 WebSocket 客户端。
type Client struct {
	UserID uint
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
}

// newClient 创建客户端实例。
func newClient(userID uint, hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		UserID: userID,
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan []byte, sendBufferSize),
	}
}

// readPump 负责从 WebSocket 读取消息并处理心跳。
// 当连接断开或出错时，向 Hub 注销客户端。
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				utils.Logger.Errorf("websocket 异常关闭: %v", err)
			}
			break
		}
	}
}

// writePump 负责把 Send 通道中的消息写入 WebSocket，并周期性发送 ping。
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 尝试把通道中积压的消息一并发送
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Hub 维护所有在线客户端，并按用户 ID 分组。
type Hub struct {
	clients    map[uint]map[*Client]struct{}
	register   chan *Client
	unregister chan *Client
	notify     chan *notificationTarget
	stop       chan struct{}
	wg         sync.WaitGroup
}

type notificationTarget struct {
	UserID       uint
	Notification *model.Notification
}

// NewHub 创建一个新的 Hub 实例（未启动）。
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uint]map[*Client]struct{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		notify:     make(chan *notificationTarget),
		stop:       make(chan struct{}),
	}
}

// Start 启动 Hub 的事件循环。
func (h *Hub) Start() {
	h.wg.Add(1)
	go h.run()
}

// Stop 停止 Hub 事件循环，等待循环退出。
func (h *Hub) Stop() {
	close(h.stop)
	h.wg.Wait()
}

func (h *Hub) run() {
	defer h.wg.Done()
	for {
		select {
		case client := <-h.register:
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]struct{})
			}
			h.clients[client.UserID][client] = struct{}{}

		case client := <-h.unregister:
			if userClients, ok := h.clients[client.UserID]; ok {
				delete(userClients, client)
				close(client.Send)
				if len(userClients) == 0 {
					delete(h.clients, client.UserID)
				}
			}

		case target := <-h.notify:
			h.broadcastToUser(target.UserID, target.Notification)

		case <-h.stop:
			return
		}
	}
}

func (h *Hub) broadcastToUser(userID uint, notification *model.Notification) {
	userClients, ok := h.clients[userID]
	if !ok {
		return
	}

	msg := websocketMessage{
		Event: "notification",
		Data:  notification,
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		utils.Logger.Errorf("序列化 WebSocket 消息失败: %v", err)
		return
	}

	for client := range userClients {
		select {
		case client.Send <- payload:
		default:
			utils.Logger.Warnf("用户 %d 的 WebSocket 发送通道已满，丢弃一条通知", userID)
		}
	}
}

// globalHub 是包级单例，默认未启动。
var globalHub *Hub
var globalHubOnce sync.Once

// StartNotificationHub 启动全局 WebSocket Hub，重复调用无副作用。
func StartNotificationHub() {
	globalHubOnce.Do(func() {
		globalHub = NewHub()
		globalHub.Start()
	})
}

// StopNotificationHub 停止全局 WebSocket Hub。
func StopNotificationHub() {
	if globalHub != nil {
		globalHub.Stop()
	}
}

// NotifyUserRealtime 向指定用户的所有在线客户端推送通知。
// Hub 未启动或用户不在线时无操作。
func NotifyUserRealtime(userID uint, notification *model.Notification) {
	if globalHub == nil || userID == 0 || notification == nil {
		return
	}
	select {
	case globalHub.notify <- &notificationTarget{UserID: userID, Notification: notification}:
	default:
		// Hub 忙碌时不阻塞业务主流程
	}
}

// RegisterWSClient 将客户端注册到全局 Hub，通常由 handler 调用。
func RegisterWSClient(userID uint, conn *websocket.Conn) {
	if globalHub == nil {
		conn.Close()
		return
	}
	client := newClient(userID, globalHub, conn)
	globalHub.register <- client
	go client.writePump()
	go client.readPump()
}
