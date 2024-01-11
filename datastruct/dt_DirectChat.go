package datastruct

type DirectChatRequest struct {
	RequestID string              `json:"requestid"`
	BotID     string              `json:"botid"`
	ChatData  []map[string]string `json:"chat_data"`
}

type DirectChatResponse struct {
	ResponseCode   int                 `json:"response_code"`
	ResponseDesc   string              `json:"response_desc"`
	TotalRequest   int                 `json:"total_request"`
	SuccessRequest int                 `json:"success_request"`
	FailedRequest  int                 `json:"failed_request"`
	Detail         []map[string]string `json:"detail"`
}
