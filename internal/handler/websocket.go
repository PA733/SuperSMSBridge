package handler

import (
	"log"
	"net/http"
	"sync"
	"time"

	"super-sms-bridge/internal/service"
	"super-sms-bridge/internal/telegram"
	"super-sms-bridge/pkg/utils"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境中应该更严格
	},
}

type WSHandler struct {
	tg      *telegram.Client
	secret  string
	clients sync.Map // 存储活跃的WebSocket连接，map[string]*websocket.Conn
}

type WSMessage struct {
	Action  string          `json:"action"`
	Payload service.Message `json:"payload"`
}

func NewWSHandler(tg *telegram.Client, secret string) *WSHandler {
	return &WSHandler{
		tg:     tg,
		secret: secret,
	}
}

func (h *WSHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	// 设置连接关闭处理
	defer func() {
		conn.Close()
	}()

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// 启动心跳检测
	go h.ping(conn)

	// 处理接收到的消息
	for {
		var wsMsg WSMessage
		err := conn.ReadJSON(&wsMsg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket读取错误: %v", err)
			}
			break
		}

		// 处理消息
		switch wsMsg.Action {
		case "send_message":
			h.handleSendMessage(conn, &wsMsg.Payload)
		default:
			log.Printf("未知的操作类型: %s", wsMsg.Action)
		}
	}
}

func (h *WSHandler) handleSendMessage(conn *websocket.Conn, msg *service.Message) {
	response := &service.Response{}

	// 验证签名
	if !utils.ValidateSign(msg.TimeStamp, msg.Sign, h.secret) {
		response.Code = http.StatusUnauthorized
		response.Message = "签名验证失败"
		conn.WriteJSON(response)
		return
	}

	// 发送消息到Telegram
	if err := h.tg.SendMessage(msg.Sender, msg.Text); err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = "发送消息失败: " + err.Error()
		conn.WriteJSON(response)
		return
	}

	// 存储连接，用于后续推送消息
	h.clients.Store(msg.Sender, conn)

	response.Code = 0
	response.Message = "发送成功"
	conn.WriteJSON(response)
}

// 心跳检测
func (h *WSHandler) ping(conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
			log.Printf("发送ping失败: %v", err)
			return
		}
	}
}

// PushMessage 推送消息给指定的sender
func (h *WSHandler) PushMessage(target string, text string) {
	if conn, ok := h.clients.Load(target); ok {
		wsConn := conn.(*websocket.Conn)
		pushMsg := struct {
			Action  string `json:"action"`
			Payload struct {
				Target string `json:"target"`
				Text   string `json:"text"`
			} `json:"payload"`
		}{
			Action: "push_message",
			Payload: struct {
				Target string `json:"target"`
				Text   string `json:"text"`
			}{
				Target: target,
				Text:   text,
			},
		}

		if err := wsConn.WriteJSON(pushMsg); err != nil {
			log.Printf("推送消息失败: %v", err)
			h.clients.Delete(target)
		}
	}
}
