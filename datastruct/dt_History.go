package datastruct

type HistoryChatRequest struct {
	BotID     string                      `json:"bot_id"`
	Recipient string                      `json:"recipient"`
	Chats     []map[string]interface{}    `json:"chats"`
	DR        map[string]map[string]int64 `json:"dr"` //msgid -> key : timestamp
}

type WebsocketHistoryResponse struct {
	TransID string `json:"transid"`
	Data    struct {
		List []map[string]interface{} `json:"list"`
	} `json:"data"`
}

type FrontendHistoryRequest struct {
	BotID   string                   `json:"bot_id"`
	History []map[string]interface{} `json:"-"`
}
