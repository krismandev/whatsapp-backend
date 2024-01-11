package datastruct

// FrontendDeviceInfoResponse is used for handling device information response
type FrontendDeviceInfoResponse struct {
	Type    string                  `json:"type"`
	Code    int                     `json:"code"`
	Message string                  `json:"message"`
	Detail  DeviceInformationStruct `json:"detail"`
}

// DeviceInformationStruct is used for
type DeviceInformationStruct struct {
	IsActive          bool
	State             string
	CanReconnect      bool
	BatteryPercentage int
	BatteryPlugged    bool
	PushName          string
	WAVersion         string
	Platform          string
}

// FrontendSendMessageRequest is used for
type FrontendSendMessageRequest struct {
	Phone    string           `json:"phone"`
	Receiver string           `json:"receiver"`
	Mimetype string           `json:"mimetype"`
	Filename string           `json:"filename"`
	Data     string           `json:"data"`
	Message  string           `json:"message"`
	Image    string           `json:"image"`
	ReplyTo  string           `json:"replyTo"`
	TransID  string           `json:"transID"`
	Buttons  []WAButtonStruct `json:"buttons"`
}

// WAButtonStruct is new WA button interface, only the field "Body" that is required.
type WAButtonStruct struct {
	ID   string `json:"id"`
	Body string `json:"body"`
}

// FrontendMessageQueue is used to handle message queue from frontend (return raw json only)
type FrontendMessageQueue struct {
	Data string `json:"data"`
}

// FrontendRealMessageQueue is used to handle message queue from frontend
type FrontendRealMessageQueue struct {
	BotID string                       `json:"session_name"`
	Key   string                       `json:"key"`
	Data  FrontendMessageRequestStruct `json:"data"`
}

// FrontendMessageRequestStruct is used to handle single message queue from frontend
type FrontendMessageRequestStruct struct {
	ID          string
	Sender      string
	Destination string
	HasMedia    bool
	Filename    string
	Mimetype    string
	Body        string
	Type        string
	Author      string
	Timestamp   int64
	IsStatus    bool
	Broadcast   bool
	FromMe      bool
	ReplyTo     string
}

// EventSingleDataStruct Will handle event string from nodejs that contain string with encoded json
type EventSingleDataStruct struct {
	Data string `json:"data"`
}

// FrontendEventStruct is struct from frontend event
type FrontendEventStruct struct {
	BotID           string                 `json:"session_name"`
	Key             string                 `json:"key"`
	FailedIteration int                    `json:"failed_iteration"`
	QRString        string                 `json:"qrstring"`
	Data            FrontendSubEventStruct `json:"data"`
	Iteration       int                    `json:"iteration"`
	TransID         string                 `json:"transid"`
}

// FrontendSubEventStruct dibuat karena karena format dari nodejs cukup kacau,
type FrontendSubEventStruct struct {
	ID        map[string]interface{} `json:"id"`
	ACK       int                    `json:"ack"`
	QRString  string                 `json:"qrstring"`
	Iteration int                    `json:"iteration"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	HasMedia  bool                   `json:"hasMedia,omitempty"`
	Body      string                 `json:"body"`
}

// APIHandlerEventStruct is event struct that will be passed to API Handler
type APIHandlerEventStruct struct {
	Key   string `json:"key"`
	BotID string `json:"botID"`
	Data  string `json:"data"`
}
