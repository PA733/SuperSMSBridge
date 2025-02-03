package service

// Message 表示一个消息请求
type Message struct {
	Sender    string `json:"sender"`
	Text      string `json:"text"`
	TimeStamp string `json:"timestamp"`
	Sign      string `json:"sign"`
}

// Response 表示通用的响应格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
