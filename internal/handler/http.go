package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"super-sms-bridge/internal/service"
	"super-sms-bridge/internal/telegram"
	"super-sms-bridge/pkg/utils"
)

type HTTPHandler struct {
	tg     *telegram.Client
	secret string // 用于验证签名的密钥
}

func NewHTTPHandler(tg *telegram.Client, secret string) *HTTPHandler {
	return &HTTPHandler{
		tg:     tg,
		secret: secret,
	}
}

func (h *HTTPHandler) HandleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, &service.Response{
			Code:    http.StatusMethodNotAllowed,
			Message: "仅支持POST请求",
		})
		return
	}

	// 输出请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	// log.Printf("Body: %s", string(body))

	var msg service.Message
	if err := json.Unmarshal(body, &msg); err != nil {
		writeJSON(w, &service.Response{
			Code:    http.StatusBadRequest,
			Message: "请求格式错误",
		})
		return
	}

	// 验证签名
	if !utils.ValidateSign(msg.TimeStamp, msg.Sign, h.secret) {
		writeJSON(w, &service.Response{
			Code:    http.StatusUnauthorized,
			Message: "签名验证失败",
		})
		return
	}

	// 发送消息到Telegram
	if err := h.tg.SendMessage(msg.Sender, msg.Text); err != nil {
		writeJSON(w, &service.Response{
			Code:    http.StatusInternalServerError,
			Message: "发送消息失败: " + err.Error(),
		})
		return
	}

	writeJSON(w, &service.Response{
		Code:    0,
		Message: "发送成功",
	})
}

func writeJSON(w http.ResponseWriter, resp *service.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
