package datastruct

type PassMediaRequest struct {
	ID       string `json:"id"`
	BotID    string `json:"bot_id"`
	MediaKey string `json:"media_key"`
	Data     string `json:"data"`
	Filename string `json:"filename"`
	Mimetype string `json:"mimetype"`
}
