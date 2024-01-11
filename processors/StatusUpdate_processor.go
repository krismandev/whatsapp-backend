package processors

import (
	"skeleton/config"
	"skeleton/connections"
	"skeleton/datastruct"
	"skeleton/lib"
	"time"

	log "github.com/sirupsen/logrus"
)

func PushStatusUpdate(conn *connections.Connections, req datastruct.FrontendEventStruct) {
	msgid, ok := req.Data.ID["id"].(string)
	if !ok {
		log.Errorf("Failed to PushStatusUpdate, facebook ID not found")
		return
	}

	var status = "failed"
	if req.Data.ACK == 0 {
		status = "pending"
	} else if req.Data.ACK == 1 {
		status = "sent"
	} else if req.Data.ACK == 2 {
		status = "received"
	} else if req.Data.ACK >= 3 {
		status = "read"
	}

	wsmap := map[string]string{
		"transid": req.TransID,
		"id":      msgid,
		"status":  status,
	}
	wsresp, wscode := HitStatusUpdate(conn, wsmap)
	log.Infof("Hit websocket status update OUTGOING MESSAGE. (HTTP Code %d) : Request : %+v - Response : %s", wscode, wsmap, wsresp)
}

func HitStatusUpdate(conn *connections.Connections, passParam map[string]string) (wsresp string, wscode int) {
	wsresp, wscode = lib.CallRestAPIOnBehalf(passParam, config.Param.WebsocketEndpoint+"/update-status", "POST", "", 4*time.Second)
	return
}

func HitFailedStatusUpdate(conn *connections.Connections, transid string) (wsresp string, wscode int) {
	// -- DISABLE SEMENTARA. SEPERTINYA NGGAK PERLU UPDATE STATUS CHAT DI WEBSOCKET
	// wsresp, wscode = HitStatusUpdate(conn, map[string]string{
	// 	"transid": transid,
	// 	"status":  "failed",
	// })
	return
}
