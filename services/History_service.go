package services

import (
	"context"
	"encoding/json"
	"skeleton/config"
	connections "skeleton/connections"
	dt "skeleton/datastruct"
	"skeleton/lib"
	"skeleton/processors"
	"time"

	"github.com/sirupsen/logrus"
)

// HistoryServices provides operations for endpoint
type HistoryServices interface {
	CallHistoryGeneration(context.Context, dt.HistoryChatRequest, *connections.Connections) dt.GlobalJSONResponse
	StoreHistoryGeneration(context.Context, dt.HistoryChatRequest, *connections.Connections) dt.GlobalJSONResponse
}

// HistoryService is use for
type HistoryService struct{}

func (HistoryService) CallHistoryGeneration(ctx context.Context, req dt.HistoryChatRequest, conn *connections.Connections) dt.GlobalJSONResponse {
	var err error

	if len(req.BotID) == 0 {
		return dt.ResponseHandlerHelper(ctx, dt.ErrInvalidParameter, "Parameter not completed", err)
	}

	// call websocket to get the last stored timestamp by bot id
	wurl := config.Param.WebsocketEndpoint + "/chat-history"
	body, httpstatus := lib.CallRestAPI(wurl, "GET", "{\"bot_id\":\""+req.BotID+"\"}", time.Duration(5)*time.Second)
	logrus.Infof("Call websocket %s for bot %s (HTTP %d) - Response : %d", wurl, req.BotID, httpstatus, body)

	var contactStateResponse dt.WebsocketHistoryResponse
	err = json.Unmarshal([]byte(body), &contactStateResponse)
	if err != nil {
		return dt.ResponseHandlerHelper(ctx, dt.ErrFailed, "Server Error. Failed to decode the response", err)
	}

	var pass dt.FrontendHistoryRequest
	pass.BotID = req.BotID
	pass.History = contactStateResponse.Data.List

	// call the whatsapp FE to trigger the chat generation
	podaddr, err := processors.NodeSelector(conn, req.BotID, false)
	if err != nil {
		logrus.Infof("Cannot select node pod available in generate-history-chat : %+v", err)
	}
	port, _ := conn.DBRedis.Get("NODE-POD-" + podaddr + "-PORT").Result()
	if len(port) == 0 {
		port = config.Param.WAClientPort
	}
	finalURL := "http://" + podaddr + config.Param.WAClientNamespace + port

	wadata, err := json.Marshal(pass)
	furl := finalURL + "/generate-history-chat"
	febody, fehttp := lib.CallRestAPI(furl, "GET", string(wadata), time.Duration(5)*time.Second)
	logrus.Infof("Call frontend %s for bot %s (HTTP %d) - Response : %d", furl, req.BotID, fehttp, febody)

	if fehttp >= 200 && fehttp < 400 {
		return dt.ResponseHandlerHelper(ctx, dt.ErrSuccess, "Success", err)
	}

	return dt.ResponseHandlerHelper(ctx, dt.ErrFailed, "Failed to connect to whatsapp service", err)
}

func (HistoryService) StoreHistoryGeneration(ctx context.Context, req dt.HistoryChatRequest, conn *connections.Connections) dt.GlobalJSONResponse {
	var err error

	// pass request to websocket.. wkwkw
	passjsondata, err := json.Marshal(req)
	if err != nil {
		logrus.Errorf("Failed to marshal the request " + err.Error())
	}

	body, httpresp := lib.CallRestAPI(config.Param.WebsocketEndpoint+"/store-history", "POST", string(passjsondata), time.Duration(10)*time.Second)
	logrus.Infof("Call websocket store-history (HTTP %d) Response : %s", httpresp, body)

	if httpresp >= 200 && httpresp < 400 {
		return dt.ResponseHandlerHelper(ctx, dt.ErrSuccess, "Request received", err)
	}

	return dt.ResponseHandlerHelper(ctx, dt.ErrFailed, "Failed to store old chat data to websocket", err)
}
