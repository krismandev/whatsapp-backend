package datastruct

import lib "skeleton/lib"

//RequestJSONRequest is use for
type RequestJSONRequest struct {
	CallbackID      string `json:"callbackid,omitempty"`
	OriginalRequest string
	IPAddr          string
	ID              string           `json:"id"`
	Phone           string           `json:"phone"`
	ChatID          string           `json:"chat_id"`
	Connection      bool             `json:"connection"`
	Endpoint        string           `json:"endpoint"`
	Mode            string           `json:"mode"`
	ClientInfo      ClientInfoStruct `json:"clientInfo"`

	Key  string               `json:"key"`
	Data RequestBackendStruct `json:"data"`
}

type ClientInfoStruct struct {
	AppID              string
	IsActive           bool
	State              string
	CanReconnect       bool
	WAVersion          string
	OSVersion          string
	DeviceManufacturer string
	DeviceModel        string
	OSBuildNumber      string
	BatteryPercentage  int
	BatteryPlugged     bool
	Platform           string
	PushName           string
}

// RequestBackendStruct is used for
type RequestBackendStruct struct {
	ReferenceID string
	RequestID   string
	WebUserID   string
	UserID      string
	LoginID     string
	ID          string
	AppID       string
	BotID       string
	TransID     string
	RequestTime string
	MsgType     string
	MediaPath   string
	Filename    string
	Mimetype    string
	IsGroup     string
	HasMedia    bool
	Source      string
	Destination string
	Message     string
	Direction   string
	Author      string
	ReplyTo     string
	DBIgnore    bool
	Buttons     []ButtonAction
	ChargeData  lib.ChargeResult
}

type ButtonAction struct {
	Label        string `json:"label"`
	URL          string `json:"url"`
	CallbackData string `json:"callback_data"`
}

// RequestDataStruct is the main structure for single instance of Request
type RequestDataStruct struct {
	QueueID string               `json:"queue_id"`
	Key     string               `json:"key"`
	Data    RequestBackendStruct `json:"data"`
}

// MsgStateRequestDataStruct is the msgstate main structure for single request of queue
type MsgStateRequestDataStruct struct {
	QueueID string          `json:"queue_id"`
	Key     string          `json:"key"`
	Data    MsgStateRequest `json:"data"`
}

//RequestJSONResponse is use for
type RequestJSONResponse struct {
	TransID      string `json:"trans_id"`
	ResponseCode int    `json:"response_code"`
	ResponseDesc string `json:"response_desc"`
	ErrorDetail  string `json:"error_detail"`
}

// MsgStateRequest is used for
type MsgStateRequest struct {
	AppID       string
	BotID       string
	Destination string
	ID          string //facebook messageid
	TransID     string
	RequestTime string
	State       string
	Direction   string
	From        string
	To          string
	HasMedia    bool
	Body        string
}
