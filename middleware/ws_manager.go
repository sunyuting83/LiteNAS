package middleware

import (
	// 用于调用之前的系统采集逻辑
	"LiteNAS/controller/System"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lxzan/gws"
)

// Message 统一的消息结构体
type Message struct {
	Type string `json:"type"`
	UUID string `json:"uuid"`
	Data string `json:"data"`
}

// GopherInfo 存储连接的上下文信息
type GopherInfo struct {
	ID string
}

type WebSocketManager struct {
	sync.RWMutex
	// sessions: ID -> *gws.Conn
	sessions sync.Map
	// infoMap: *gws.Conn -> *GopherInfo
	infoMap sync.Map
}

// Manager 全局单例
var Manager = &WebSocketManager{}
var WSUpgrader *gws.Upgrader

func startGlobalCollector() {

	ticker := time.NewTicker(2 * time.Second)
	for range ticker.C {
		// 只有当有人在线时才采集
		if Manager.GetOnlineCount() > 0 {
			stats := System.CollectStats() // 调用你之前的 gopsutil 逻辑
			data, _ := json.Marshal(stats)

			// 包装成 Message 协议格式发送
			msg := Message{
				Type: "system_stats",
				Data: string(data),
			}
			finalData, _ := json.Marshal(msg)
			Manager.Broadcast(finalData)
		}
	}
}

// InitWS 初始化 Upgrader
func InitWS() {
	WSUpgrader = gws.NewUpgrader(Manager, &gws.ServerOption{
		CheckUtf8Enabled: true,
		Recovery:         gws.Recovery, // 增加容错
	})

	// 启动之前讨论的后台采集协程
	go startGlobalCollector()
}

// OnOpen 实现握手后的逻辑
func (m *WebSocketManager) OnOpen(c *gws.Conn) {
	id := uuid.NewString()
	info := &GopherInfo{ID: id}

	// 双向绑定
	m.sessions.Store(id, c)
	m.infoMap.Store(c, info)

	// 发送新连接自己的 ID 消息 (首连确认)
	m.sendToConn(c, "init", id, "connected")
}

// OnMessage 处理接收到的消息
func (m *WebSocketManager) OnMessage(c *gws.Conn, msg *gws.Message) {
	defer msg.Close()

	val, ok := m.infoMap.Load(c)
	if !ok {
		return
	}
	info := val.(*GopherInfo)

	var message Message
	if err := json.Unmarshal(msg.Bytes(), &message); err != nil {
		return
	}

	// 业务逻辑分发
	if message.UUID == info.ID {
		switch message.Type {
		case "active": // 心跳
			break
		case "kill_process":
			// 这里可以调用 System.KillProcess 的逻辑
		default:
			// 默认广播
			m.Broadcast(msg.Bytes())
		}
	}
}

// Broadcast 广播给所有人
func (m *WebSocketManager) Broadcast(data []byte) {
	m.sessions.Range(func(key, value any) bool {
		conn := value.(*gws.Conn)
		_ = conn.WriteMessage(gws.OpcodeText, data)
		return true
	})
}

// SendToID 定向发送 (供外部调用)
func (m *WebSocketManager) SendToID(id string, t, d string) {
	if value, ok := m.sessions.Load(id); ok {
		conn := value.(*gws.Conn)
		m.sendToConn(conn, t, id, d)
	}
}

// 内部辅助发送函数
func (m *WebSocketManager) sendToConn(c *gws.Conn, t, id, d string) {
	resp := Message{
		Type: t,
		UUID: id,
		Data: d,
	}
	payload, _ := json.Marshal(resp)
	_ = c.WriteMessage(gws.OpcodeText, payload)
}

// OnClose 清理
func (m *WebSocketManager) OnClose(c *gws.Conn, err error) {
	if val, ok := m.infoMap.Load(c); ok {
		info := val.(*GopherInfo)
		m.sessions.Delete(info.ID)
		m.infoMap.Delete(c)
	}
}

// 获取在线人数
func (m *WebSocketManager) GetOnlineCount() int {
	count := 0
	m.sessions.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

// 必须实现的接口
func (m *WebSocketManager) OnPing(c *gws.Conn, p []byte) { _ = c.WritePong(p) }
func (m *WebSocketManager) OnPong(c *gws.Conn, p []byte) {}
