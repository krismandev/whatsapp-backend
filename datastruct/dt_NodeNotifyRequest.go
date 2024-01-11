package datastruct

type NodeNotifyRequest struct {
	BotType          string   `json:"bot_type"`
	PodName          string   `json:"podname"`
	Port             string   `json:"port"`
	Host             string   `json:"host"`
	FirstTime        string   `json:"first_time"`
	CurrentActiveBot []string `json:"current_active_bot"`
}
