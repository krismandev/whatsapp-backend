package processors

import (
	"skeleton/config"
	"skeleton/datastruct"
	"skeleton/lib"
	"time"
)

func InitTranslogFromDirectRequest(req datastruct.DirectChatRequest) lib.WATranslog {
	chatdata := make(map[string]string)
	if len(req.ChatData) > 0 {
		chatdata = req.ChatData[0]
	}
	chatdata = req.ChatData[0]
	translog := lib.WATranslog{
		SavePath: config.Param.Translog + "-translog.log",
		Data: map[string]string{
			"requestid":   req.RequestID,
			"bot_id":      req.BotID,
			"userapi_id":  chatdata["userapi_id"],
			"source":      "api",
			"referenceid": chatdata["referenceid"],
			"reply_to":    chatdata["reply_to"],
			"flow":        "out",
			"message":     chatdata["message"],
			"status":      "success",
			"created_at":  time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	return translog
}

func InitTranslogFromQueueRequest(req datastruct.RequestBackendStruct) lib.WATranslog {
	rid := req.RequestID
	if len(rid) == 0 {
		rid = req.TransID
	}
	translog := lib.WATranslog{
		SavePath: config.Param.Translog + "-translog.log",
		Data: map[string]string{
			"requestid":   rid,
			"bot_id":      req.BotID,
			"userapi_id":  req.UserID,
			"source":      "web",
			"referenceid": req.ReferenceID,
			"reply_to":    req.ReplyTo,
			"flow":        req.Direction,
			"message":     req.Message,
			"recipient":   req.Destination,
			"status":      "success",
			"created_at":  time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	return translog
}
