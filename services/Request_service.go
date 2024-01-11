package services

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"skeleton/config"
	connections "skeleton/connections"
	dt "skeleton/datastruct"
	"skeleton/lib"
	"skeleton/processors"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/wzshiming/ssdb"
)

// RequestServices provides operations for endpoint
type RequestServices interface {
	DirectChat(context.Context, dt.DirectChatRequest, *connections.Connections) dt.DirectChatResponse
	NodeMetrics(context.Context, dt.RequestJSONRequest, *connections.Connections) dt.GlobalJSONResponse
	NodeNotify(context.Context, dt.NodeNotifyRequest, *connections.Connections) dt.GlobalJSONResponse
	CallFrontend(context.Context, dt.RequestJSONRequest, *connections.Connections) dt.GlobalJSONResponse
	CreateRequest(context.Context, dt.RequestJSONRequest, *connections.Connections) dt.GlobalJSONResponse
	DeviceStatusRequest(context.Context, dt.RequestJSONRequest, *connections.Connections) dt.GlobalJSONResponse
	PassMedia(context.Context, dt.PassMediaRequest, *connections.Connections) dt.GlobalJSONResponse
	GetContacts(context.Context, dt.RequestJSONRequest, *connections.Connections) dt.GetContactResponse
	GetNameInformation(context.Context, dt.RequestJSONRequest, *connections.Connections) dt.ContactDetailResponse
	ManualStoreMessage(context.Context, dt.RequestDataStruct, *connections.Connections) dt.GlobalJSONResponse
}

// RequestService is use for
type RequestService struct{}

func (RequestService) DirectChat(ctx context.Context, req dt.DirectChatRequest, conn *connections.Connections) (resp dt.DirectChatResponse) {
	resp.ResponseCode = dt.ErrSuccess
	resp.ResponseDesc = dt.DescSuccess
	resp.TotalRequest = 0
	resp.SuccessRequest = 0
	resp.FailedRequest = 0

	translog := processors.InitTranslogFromDirectRequest(req)

	if len(req.BotID) < 3 || len(req.ChatData) == 0 {
		resp.ResponseCode = dt.ErrGetCustStatusIncompleteRequest
		resp.ResponseDesc = dt.DescGetCustStatusIncompleteRequest

		translog.SetDataField("status", "failed")
		translog.SetDataField("error_detail", "INCOMPLETE_PARAMETER")
		translog.Store()
		return
	}

	// try get direct access to node
	podaddr, err := processors.NodeSelector(conn, req.BotID, false)
	if err != nil {
		resp.ResponseCode = dt.ErrNoData
		resp.ResponseDesc = "Sorry bot is inactive right now"

		translog.SetDataField("status", "failed")
		translog.SetDataField("error_detail", "ERR_NODE_SELECTOR="+err.Error())
		translog.Store()
		return
	}
	port, _ := conn.DBRedis.Get("NODE-POD-" + podaddr + "-PORT").Result()
	if len(port) == 0 {
		port = config.Param.WAClientPort
	}

	// tetap hit ke nodejs satu per satu
	for _, list := range req.ChatData {
		resp.TotalRequest = resp.TotalRequest + 1

		var sendResp = make(map[string]string)
		sendResp["botid"] = req.BotID
		sendResp["destination"] = list["from"]
		sendResp["status"] = "sent"
		sendResp["transid"] = list["transid"]

		if len(list["from"]) == 0 || len(list["transid"]) == 0 || (len(list["message"]) == 0 && list["has_media"] != "1") {
			// incomplete request
			sendResp["status"] = "failed"
			sendResp["error"] = "Incomplete request"

			resp.Detail = append(resp.Detail, sendResp)
			resp.FailedRequest = resp.FailedRequest + 1
			processors.HitFailedStatusUpdate(conn, list["transid"])
			translog.SetDataField("status", "failed")
			translog.SetDataField("transid", list["transid"])
			translog.SetDataField("error_detail", "INCOMPLETE_PARAMETER")
			translog.Store()

			continue
		}

		//hit Frontend AP
		var pass dt.FrontendSendMessageRequest
		pass.Phone = req.BotID
		pass.Receiver = list["from"]
		pass.Message = list["message"]
		pass.ReplyTo = list["reply_to"]
		pass.TransID = list["transid"]

		var httpResponse dt.FrontendDeviceInfoResponse
		var bodyResp string
		var httpCode int

		var hitURL string
		if len(list["message_type"]) != 0 && list["message_type"] != "text" {
			imgPath := config.Param.PublicStoragePath + "/" + req.BotID + "/media-data/" + list["transid"] + ".data"
			storedStream, err := ioutil.ReadFile(imgPath)
			if err != nil {
				// cannot send image : failed to open image file
				log.Errorf("Error ioutil.Readfile() : " + err.Error())
				sendResp["status"] = "failed"
				sendResp["error"] = "Cannot process the document"
				resp.Detail = append(resp.Detail, sendResp)
				resp.FailedRequest = resp.FailedRequest + 1
				processors.HitFailedStatusUpdate(conn, list["transid"])

				translog.SetDataField("status", "failed")
				translog.SetDataField("transid", list["transid"])
				translog.SetDataField("error_detail", "ERR_DOCUMENT="+err.Error())
				translog.Store()

				continue
			}

			var passmedia dt.PassMediaRequest
			err = json.Unmarshal(storedStream, &passmedia)
			if err != nil {
				log.Errorf("Failed to unmarshal the stored document : " + err.Error())
				sendResp["status"] = "failed"
				sendResp["error"] = "Cannot parse the document"
				resp.Detail = append(resp.Detail, sendResp)
				resp.FailedRequest = resp.FailedRequest + 1
				processors.HitFailedStatusUpdate(conn, list["transid"])

				translog.SetDataField("status", "failed")
				translog.SetDataField("transid", list["transid"])
				translog.SetDataField("error_detail", "ERR_UNMARSHAL_DOCUMENT="+err.Error())
				translog.Store()
				continue
			}

			if list["message_type"] == "image" {
				pass.Image = passmedia.Data
				hitURL = "/app/send-image"
			} else if list["message_type"] == "audio" {
				pass.Filename = passmedia.Filename
				pass.Mimetype = passmedia.Mimetype
				pass.Data = passmedia.Data
				hitURL = "/app/send-audio"
			} else if list["message_type"] == "video" {
				pass.Filename = passmedia.Filename
				pass.Mimetype = passmedia.Mimetype
				pass.Data = passmedia.Data
				hitURL = "/app/send-video"
			} else {
				pass.Filename = passmedia.Filename
				pass.Mimetype = passmedia.Mimetype
				pass.Data = passmedia.Data
				hitURL = "/app/send-document"
			}
		} else {
			// default type : text masuk kesini
			hitURL = "/app/send-message"
		}

		if len(list["buttons"]) > 0 {
			// unmarshal as ButtonAction first
			var requestBtns []dt.ButtonAction
			err = json.Unmarshal([]byte(list["buttons"]), &requestBtns)
			if err != nil {
				logrus.Infof("Failed to unmarshal the button struct : %+v", err)
			} else {
				for _, btn := range requestBtns {
					if len(btn.Label) > 0 {
						var wabtn dt.WAButtonStruct
						wabtn.Body = btn.Label
						pass.Buttons = append(pass.Buttons, wabtn)
					}
				}
			}
		}

		//Check is registered on WA
		checkRegisterURL := "http://" + podaddr + config.Param.WAClientNamespace + port + "/app/check-registered"
		var checkRegistered bool = true
		var httpResponsecheck dt.FrontendDeviceInfoResponse
		msisdn := lib.NormalizeMsisdn(pass.Receiver)

		result, _ := conn.SSDB.Get(dt.SSDBCheckNumberKey + msisdn)
		if result.String() == dt.WARegistered {
			log.Infof("[cached] msisdn : %s is registered", msisdn)
			checkRegistered = false
		} else if result.String() == dt.WANotRegistered {
			log.Infof("[cached] msisdn : %s is not registered", msisdn)
			sendResp["status"] = "failed"
			sendResp["error"] = "ERR_PHONE_NOT_REGISTERED"
			list["status"] = "failed"
			list["error_result"] = "ERR_PHONE_NOT_REGISTERED"
			resp.Detail = append(resp.Detail, sendResp)
			resp.FailedRequest = resp.FailedRequest + 1
			translog.SetDataField("status", "failed")
			translog.SetDataField("transid", list["transid"])
			translog.SetDataField("error_detail", "ERR_PHONE_NOT_REGISTERED")
			translog.Store()
			wsresp, wscode := lib.CallRestAPIOnBehalf(list, config.Param.WebsocketEndpoint+"/push", "POST", "", 4*time.Second)
			log.Infof("Hit websocket endpoint DIRECT OUT MODE. (HTTP Code %d) : Request : %+v - Response : %s", wscode, list, wsresp)
			return
		}
		log.Infof("msisdn : %s , checkRegistered Value : %v", msisdn, checkRegistered)
		if checkRegistered {
			//hit check-register
			bodyRespCheck, httpCodeCheck := lib.CallRestAPIOnBehalf(pass, checkRegisterURL, "POST", "", 5*time.Second)
			if httpCodeCheck >= 200 && httpCodeCheck < 400 {

				json.Unmarshal([]byte(bodyRespCheck), &httpResponsecheck)
				log.Infof("Frontend Body Parsed Requests : %+v", httpResponsecheck)
				if httpResponsecheck.Message == "notregistered" {
					log.Infof("[api] msisdn : %s is not registered", msisdn)
					conn.SSDB.SetX(dt.SSDBCheckNumberKey+msisdn, ssdb.Value(dt.WANotRegistered), time.Duration(config.Param.NotRegisteredOnWAKeepDuration)*time.Hour)
					sendResp["status"] = "failed"
					sendResp["error"] = "ERR_PHONE_NOT_REGISTERED"
					list["status"] = "failed"
					list["error_result"] = "ERR_PHONE_NOT_REGISTERED"
					resp.Detail = append(resp.Detail, sendResp)
					resp.FailedRequest = resp.FailedRequest + 1
					translog.SetDataField("status", "failed")
					translog.SetDataField("transid", list["transid"])
					translog.SetDataField("error_detail", "ERR_PHONE_NOT_REGISTERED")
					translog.Store()
					wsresp, wscode := lib.CallRestAPIOnBehalf(list, config.Param.WebsocketEndpoint+"/push", "POST", "", 4*time.Second)
					log.Infof("Hit websocket endpoint DIRECT OUT MODE. (HTTP Code %d) : Request : %+v - Response : %s", wscode, list, wsresp)
					return
				} else if httpResponsecheck.Message == "registered" {
					log.Infof("[api] msisdn : %s is registered", msisdn)
					conn.SSDB.Set(dt.SSDBCheckNumberKey+msisdn, ssdb.Value(dt.WARegistered))
				}
			}
		}
		//end Check is registered on WA

		finalURL := "http://" + podaddr + config.Param.WAClientNamespace + port + hitURL
		bodyResp, httpCode = lib.CallRestAPIOnBehalf(pass, finalURL, "POST", "", 5*time.Second)
		log.Infof("Hit DIRECT Frontend : %s . HTTP Code %d - %s", finalURL, httpCode, bodyResp)

		if httpCode >= 200 && httpCode < 400 {
			//request success (setidaknya ada body)
			err := json.Unmarshal([]byte(bodyResp), &httpResponse)
			log.Infof("Frontend Body Parsed Requests : %+v", httpResponse)
			if err != nil {
				log.Infof("Err json.Unmarshal = %+v", err)
				sendResp["status"] = "failed"
				sendResp["error"] = "Cannot parse the bot response"
				resp.Detail = append(resp.Detail, sendResp)
				resp.FailedRequest = resp.FailedRequest + 1
				processors.HitFailedStatusUpdate(conn, list["transid"])

				translog.SetDataField("status", "failed")
				translog.SetDataField("transid", list["transid"])
				translog.SetDataField("error_detail", "FAIL_WARESULT="+err.Error())
				translog.Store()

				return
			}

			// GRAB NEW ID FROM frontendRequest.Message
			rawMessage := strings.Split(httpResponse.Message, "|")
			var newMsgID string
			if len(rawMessage) > 1 {
				newMsgID = rawMessage[len(rawMessage)-1]
				sendResp["messageid"] = newMsgID
				list["id"] = newMsgID

				translog.SetDataField("messageid", newMsgID)
				translog.SetDataField("transid", list["transid"])
			}

			if httpResponse.Type != "success" && len(httpResponse.Message) > 0 {
				sendResp["status"] = "failed"
				sendResp["error"] = httpResponse.Message
				list["status"] = "failed"
				list["error_result"] = "EXPECTED_WAERROR=" + httpResponse.Message
				resp.FailedRequest = resp.FailedRequest + 1
				translog.SetDataField("status", "failed")
				translog.SetDataField("transid", list["transid"])
				translog.SetDataField("error_detail", "EXPECTED_WAERROR="+httpResponse.Message)
				translog.Store()
			} else {
				sendResp["status"] = "success"
				list["status"] = "success"
				resp.SuccessRequest = resp.SuccessRequest + 1
				translog.Store()
			}

			resp.Detail = append(resp.Detail, sendResp)
		} else {
			// tetap coba parse sekalipun error non 200
			err := json.Unmarshal([]byte(bodyResp), &httpResponse)
			log.Infof("Frontend Body Parsed Requests : %+v", httpResponse)
			list["status"] = "failed"
			sendResp["status"] = "failed"
			sendResp["error"] = "BOT_ENDPOINT_CONNECT_FAILED"
			if err != nil {
				log.Infof("Err json.Unmarshal = %+v", err)
				if len(httpResponse.Message) > 0 {
					sendResp["error"] = httpResponse.Message
				}
			} else {
				sendResp["error"] = httpResponse.Message
			}

			// hardcode tweak : saat nodejs parsing throw error ke string, entah kenapa dikasi prefix "Error: xxxxxx"
			// hapus string "Error :" jika ada
			sendResp["error"] = strings.Replace(sendResp["error"], "Error: ", "", 1)
			list["error_result"] = "ERR_WANOT200=" + sendResp["error"]
			resp.Detail = append(resp.Detail, sendResp)
			resp.FailedRequest = resp.FailedRequest + 1

			translog.SetDataField("status", "failed")
			translog.SetDataField("transid", list["transid"])
			translog.SetDataField("error_detail", "ERR_WANOT200="+sendResp["error"])
			translog.Store()

		}

		// sebelum return response, hit websocket dulu utk push data outgoing message
		wsresp, wscode := lib.CallRestAPIOnBehalf(list, config.Param.WebsocketEndpoint+"/push", "POST", "", 4*time.Second)
		log.Infof("Hit websocket endpoint DIRECT OUT MODE. (HTTP Code %d) : Request : %+v - Response : %s", wscode, list, wsresp)

	}

	return
}

func (RequestService) NodeMetrics(ctx context.Context, req dt.RequestJSONRequest, conn *connections.Connections) dt.GlobalJSONResponse {
	// will return available pods
	nodePodLists, _ := conn.DBRedis.Get("NODE-POD-LISTS").Result()
	var nodeLists []string
	if len(nodePodLists) > 0 {
		nodeLists = strings.Split(nodePodLists, "|")
	}

	resp := dt.ResponseHandlerHelper(ctx, dt.ErrSuccess, "Node status", nil)
	for _, nl := range nodeLists {
		var stat = make(map[string]string)
		stat["node"] = nl
		stat["last_timestamp"], _ = conn.DBRedis.Get("NODE-POD-" + nl + "-TIMESTAMP").Result()
		stat["type"], _ = conn.DBRedis.Get("NODE-POD-" + nl + "-TYPE").Result()
		stat["port"], _ = conn.DBRedis.Get("NODE-POD-" + nl + "-PORT").Result()
		stat["bots_count"], _ = conn.DBRedis.Get("NODE-POD-" + nl + "-BOT-CONNECTED").Result()
		stat["active_bots"], _ = conn.DBRedis.Get("NODE-POD-" + nl + "-BOTS").Result()
		resp.Data = append(resp.Data, stat)
	}

	return resp
}

func (RequestService) NodeNotify(ctx context.Context, req dt.NodeNotifyRequest, conn *connections.Connections) dt.GlobalJSONResponse {
	var err error
	log.Infof("NodeNotify CALLED : %+v", req)
	processors.NodeUpdateConfig(conn, req)
	go processors.BotTypeMapping(conn, &lib.WaitG)
	return dt.ResponseHandlerHelper(ctx, dt.ErrSuccess, "Node notify called", err)
}

// this method will call frontend endpoint. so API handler can bypass the request to frontend
func (RequestService) CallFrontend(ctx context.Context, req dt.RequestJSONRequest, conn *connections.Connections) dt.GlobalJSONResponse {
	var err error

	if len(req.Phone) == 0 {
		return dt.ResponseHandlerHelper(ctx, dt.ErrNoData, "Please define the phone number first", err)
	}

	// logic to get the available pod
	var selectNew = true
	if req.Endpoint != "connect" {
		selectNew = false
	}
	podaddr, err := processors.NodeSelector(conn, req.Phone, selectNew)
	if err != nil {
		if req.Endpoint != "connect" {
			// force disconnect via msgconfig
			processors.StoreMsgConfig(conn, config.Param.RedisPrefix+"BotStatus-"+req.Phone, "0")
			// trigger refresh config
			processors.ReloadMsgConfig()

		}

		return dt.ResponseHandlerHelper(ctx, dt.ErrFailed, "No whatsapp server instance available right now", err)
	}
	port, _ := conn.DBRedis.Get("NODE-POD-" + podaddr + "-PORT").Result()
	if len(port) == 0 {
		port = config.Param.WAClientPort
	}
	finalURL := "http://" + podaddr + config.Param.WAClientNamespace + port
	logrus.Infof("TRY CALL DYNAMIC FRONTEND (%s) WITH FINAL URL : %s", req.Endpoint, finalURL)

	if req.Endpoint == "connect" {
		var pass = dt.RequestJSONRequest{
			Phone: req.Phone,
		}

		_, httpCode := lib.CallRestAPIOnBehalf(pass, finalURL+"/app/connect", "GET", "", 5*time.Second)
		conn.DBRedis.Set("BOT-RECONNECT-COUNTER-"+req.Phone, "0", 0) //reset reconnect counter supaya bisa autoconnect lagi
		if httpCode >= 200 && httpCode < 400 {
			// success called
			return dt.ResponseHandlerHelper(ctx, dt.ErrSuccess, "Frontend successfully called", err)
		}
	} else if req.Endpoint == "disconnect" {
		var pass = dt.RequestJSONRequest{
			Phone: req.Phone,
			Mode:  req.Mode,
		}

		_, httpCode := lib.CallRestAPIOnBehalf(pass, finalURL+"/app/disconnect", "GET", "", 5*time.Second)

		// langsung hardcode redis set disconnect :
		processors.StoreMsgConfig(conn, config.Param.RedisPrefix+"BotStatus-"+req.Phone, "0")
		// trigger refresh config
		processors.ReloadMsgConfig()

		if httpCode >= 200 && httpCode < 400 {
			// remove key node pod selector cache
			conn.DBRedis.Del("NODE-SCHEDULED-" + req.Phone)
			// success called
			return dt.ResponseHandlerHelper(ctx, dt.ErrSuccess, "Frontend successfully called", err)
		}
	} else if req.Endpoint == "generate-history-chat" {
		var pass = map[string]string{
			"bot_id": req.Phone,
		}
		_, httpCode := lib.CallRestAPIOnBehalf(pass, finalURL+"/generate-history-chat", "GET", "", 5*time.Second)
		if httpCode >= 200 && httpCode < 400 {
			// success called
			return dt.ResponseHandlerHelper(ctx, dt.ErrSuccess, "Frontend history generation successfully called", err)
		}
	} else if req.Endpoint == "retract-message" {
		var pass = map[string]string{
			"phone":     req.Phone,
			"chat_id":   req.ChatID,
			"messageid": req.ID,
		}
		_, httpCode := lib.CallRestAPIOnBehalf(pass, finalURL+"/retract-message", "POST", "", 5*time.Second)
		if httpCode >= 200 && httpCode < 400 {
			// success called
			return dt.ResponseHandlerHelper(ctx, dt.ErrSuccess, "Frontend retract message called", err)
		}
	}

	return dt.ResponseHandlerHelper(ctx, dt.ErrFailed, "Failed to connect to whatsapp service", err)
}

// CreateRequest is use for
func (RequestService) CreateRequest(ctx context.Context, req dt.RequestJSONRequest, conn *connections.Connections) dt.GlobalJSONResponse {
	logger := log.WithFields(dt.GetLogFieldValues(ctx, "CreateRequest"))
	logger.Info("Processing CreateRequest")
	var err error

	// check if all the required field is filled
	if len(req.Key) == 0 {
		return dt.ResponseHandlerHelper(ctx, dt.ErrData, dt.DescData, err)
	}

	// will return lastInsertID. just ignore if you dont need them
	insertError := processors.InsertRequest(req, conn)
	if insertError != nil {
		return dt.ResponseHandlerHelper(ctx, dt.ErrFailedQuery, dt.DescFailedQuery, insertError)
	}

	return dt.ResponseHandlerHelper(ctx, dt.ErrSuccess, "Data inserted successfully", err)
}

// DeviceStatusRequest is used to push the FE state to msgConfig
func (RequestService) DeviceStatusRequest(ctx context.Context, req dt.RequestJSONRequest, conn *connections.Connections) dt.GlobalJSONResponse {
	logger := log.WithFields(dt.GetLogFieldValues(ctx, "DeviceStatusRequest"))
	logger.Info("Processing DeviceStatusRequest")
	var err error

	if len(req.Phone) == 0 {
		return dt.ResponseHandlerHelper(ctx, dt.ErrData, "Please fill the phone parameter", err)
	}

	if len(req.Phone) > 0 && len(req.ClientInfo.State) > 0 {
		// log clientinfo state data
		body, stt := lib.CallRestAPIOnBehalf(req, config.Param.WhatsappEventLogURL+"/logstate", "POST", "", time.Duration(5)*time.Second)
		log.Info("Called WhatsappEventLogURL logstate response HTTP %d : %s", stt, body)
		return dt.ResponseHandlerHelper(ctx, dt.ErrSuccess, "Bot state stored successfully", err)
	}

	var value string
	if req.Connection {
		value = "1"
	} else {
		value = "0"
	}
	err = processors.StoreMsgConfig(conn, config.Param.RedisPrefix+"BotStatus-"+req.Phone, value)
	if err != nil {
		return dt.ResponseHandlerHelper(ctx, dt.ErrData, "Failed to store the device condition", err)
	}

	// store additional bot data & credentials
	processors.StoreMsgConfig(conn, config.Param.RedisPrefix+"BotEndpoint-"+req.Phone, config.Param.PublicEndpointURL)
	processors.StoreMsgConfig(conn, config.Param.RedisPrefix+"MsgConfigRedisUrl-"+req.Phone, config.Param.Redis.RedisURL)
	processors.StoreMsgConfig(conn, config.Param.RedisPrefix+"MsgConfigRedisPassword-"+req.Phone, config.Param.Redis.RedisPassword)
	processors.StoreMsgConfig(conn, config.Param.RedisPrefix+"MsgConfigRedisEventStream-"+req.Phone, config.Param.RedisPrefix+config.Param.Redis.EventStream)

	// hit APIHandlerEndpoint to reload this config to API Handler
	reloadURL := config.Param.APIHandlerEndpoint + "/reload"
	lib.CallRestAPI(reloadURL, "GET", "{}", 5*time.Second)
	log.Info("Called APIHandlerEndpoint to reload the configuration state")

	go processors.BotTypeMapping(conn, &lib.WaitG)
	log.Info("Call processors.NodeMapping to map the latest bot state")

	return dt.ResponseHandlerHelper(ctx, dt.ErrSuccess, "Bot status stored successfully", err)
}

func (RequestService) PassMedia(ctx context.Context, req dt.PassMediaRequest, conn *connections.Connections) dt.GlobalJSONResponse {
	var err error
	// all fields are required
	if len(req.ID) == 0 || len(req.Data) == 0 || len(req.MediaKey) == 0 || len(req.Mimetype) == 0 || len(req.BotID) == 0 {
		log.Errorf("PassMedia : incomplete parameters %+v", req)
		return dt.ResponseHandlerHelper(ctx, dt.ErrData, "Please fill all the parameters", err)
	}

	if err, ok := processors.StorePassMedia(req); !ok {
		log.Errorf("Failed to store passmedia : %+v", err)
		return dt.ResponseHandlerHelper(ctx, dt.ErrFailed, "Failed to store media data", err)
	}

	return dt.ResponseHandlerHelper(ctx, dt.ErrSuccess, "Media data has been stored successfully", err)
}

func (RequestService) GetContacts(ctx context.Context, req dt.RequestJSONRequest, conn *connections.Connections) (resp dt.GetContactResponse) {
	if len(req.Phone) == 0 {
		resp.Code = 404
		resp.Message = "Please define the bot phone number"
		return
	}

	// call FE
	var pass = dt.RequestJSONRequest{
		Phone: req.Phone,
	}
	httpBody, httpCode := lib.CallRestAPIOnBehalf(pass, config.Param.WAClientEndpoint+"/app/get-contacts", "GET", "", 25*time.Second)
	if httpCode >= 200 && httpCode < 400 {
		// success called
		err := json.Unmarshal([]byte(httpBody), &resp)
		if err != nil {
			resp.Code = 500
			resp.Message = "Failed to decode the data"
			return
		}
	}

	return
}

func (RequestService) GetNameInformation(ctx context.Context, req dt.RequestJSONRequest, conn *connections.Connections) (resp dt.ContactDetailResponse) {
	if len(req.Phone) == 0 {
		resp.Code = 404
		resp.Message = "Please define the bot phone number"
		return
	}
	if len(req.ChatID) == 0 {
		resp.Code = 404
		resp.Message = "Please define the chat ID"
		return
	}

	// call FE
	var pass = dt.RequestJSONRequest{
		Phone:  req.Phone,
		ChatID: req.ChatID,
	}
	httpBody, httpCode := lib.CallRestAPIOnBehalf(pass, config.Param.WAClientEndpoint+"/app/name-information", "GET", "", 5*time.Second)
	if httpCode >= 200 && httpCode < 400 {
		// success called
		err := json.Unmarshal([]byte(httpBody), &resp)
		if err != nil {
			resp.Code = 500
			resp.Message = "Failed to decode the data"
			return
		}
	}

	return
}

func (RequestService) ManualStoreMessage(ctx context.Context, req dt.RequestDataStruct, conn *connections.Connections) dt.GlobalJSONResponse {
	var err error
	req.Data.AppID = config.Param.AppID
	log.Infof("ManualStoreMessage CALLED : %+v", req)
	processors.PushChatToWebsocketAfterSend(req, "success", "")
	return dt.ResponseHandlerHelper(ctx, dt.ErrSuccess, "ManualStoreMessage called", err)
}
