package service

import (
	"blog/model"
	"encoding/json"
	"testing"
	"time"
)

func TestHub_RegisterUnregisterAndNotify(t *testing.T) {
	hub := NewHub()
	hub.Start()
	defer hub.Stop()

	userID := uint(1)
	client := &Client{
		UserID: userID,
		Hub:    hub,
		Send:   make(chan []byte, sendBufferSize),
	}

	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	if _, ok := hub.clients[userID][client]; !ok {
		t.Fatal("客户端注册后未在 Hub 中找到")
	}

	notification := &model.Notification{
		ID:      100,
		UserID:  userID,
		Type:    model.NotificationTypeFollow,
		Title:   "新增关注",
		Content: "有人关注了你",
	}
	hub.notify <- &notificationTarget{UserID: userID, Notification: notification}

	select {
	case msg := <-client.Send:
		var payload websocketMessage
		if err := json.Unmarshal(msg, &payload); err != nil {
			t.Fatalf("收到的消息不是合法 JSON: %v", err)
		}
		if payload.Event != "notification" {
			t.Errorf("事件名期望 notification，得到 %s", payload.Event)
		}
	case <-time.After(time.Second):
		t.Fatal("未在超时时间内收到通知消息")
	}

	hub.unregister <- client
	time.Sleep(50 * time.Millisecond)

	if _, ok := hub.clients[userID]; ok {
		t.Fatal("用户所有客户端注销后，Hub 中仍保留该用户连接")
	}
}

func TestHub_MultipleClientsSameUser(t *testing.T) {
	hub := NewHub()
	hub.Start()
	defer hub.Stop()

	userID := uint(2)
	c1 := &Client{UserID: userID, Hub: hub, Send: make(chan []byte, sendBufferSize)}
	c2 := &Client{UserID: userID, Hub: hub, Send: make(chan []byte, sendBufferSize)}

	hub.register <- c1
	hub.register <- c2
	time.Sleep(50 * time.Millisecond)

	if len(hub.clients[userID]) != 2 {
		t.Fatalf("期望用户有 2 个客户端，实际 %d", len(hub.clients[userID]))
	}

	notification := &model.Notification{
		ID:      101,
		UserID:  userID,
		Type:    model.NotificationTypePostLike,
		Title:   "有人赞了你的文章",
		Content: "test",
	}
	hub.notify <- &notificationTarget{UserID: userID, Notification: notification}

	for _, client := range []*Client{c1, c2} {
		select {
		case <-client.Send:
		case <-time.After(time.Second):
			t.Fatal("多客户端场景下未全部收到通知")
		}
	}
}

func TestNotifyUserRealtime_NoHub(t *testing.T) {
	// 确保 Hub 未启动时不会 panic 或阻塞
	NotifyUserRealtime(1, &model.Notification{ID: 1, UserID: 1})
}
