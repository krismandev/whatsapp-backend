package processors

import (
	"encoding/json"
	"skeleton/config"
	"skeleton/connections"
	dt "skeleton/datastruct"
	"skeleton/lib"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// IncomingMessageListener will listen from redis queue if there is new message from FE
func IncomingMessageListener(conn *connections.Connections, wg *sync.WaitGroup, closeFlag <-chan struct{}) {
	wg.Add(1)
	defer wg.Done()
	// create group & consumer reader
	err := lib.RedisCreateStreamReader(conn.FrontendMsg, config.Param.RedisPrefix+config.Param.FrontendMsg.EventStream, config.Param.RedisPrefix+config.Param.FrontendMsg.GroupName)
	if err != nil {
		log.Error("[IncomingMessageListener] Failed to create stream reader : " + err.Error())
	}
	readNew := false
	for {
		select {
		case <-closeFlag:
			return
		default:
			que := lib.RedisQueue{
				Db:           conn.FrontendMsg,
				Name:         config.Param.RedisPrefix + config.Param.FrontendMsg.EventStream,
				GroupName:    config.Param.RedisPrefix + config.Param.FrontendMsg.GroupName,
				ConsumerName: config.Param.RedisPrefix + config.Param.FrontendMsg.ConsumerName,
				ReadTimeout:  5000,
			}
			res, err := que.ReadQueueAsJson(1, readNew)
			if err != nil {
				// if readNew == false {
				// 	log.Error("Error ReadQueueAsJson : " + err.Error())
				// }
				if !strings.Contains(err.Error(), "is empty") && !strings.Contains(err.Error(), "timeout") {
					log.Error("Error ReadQueueAsJson : " + err.Error())
				}
			}
			if len(res) == 0 && readNew == false {
				log.Info("Frontend Redis queue is empty in old mode. Change frontend redis stream read mode from OLD to NEW mode")
				readNew = true
			}

			// format yang dikirim dari nodejs agak kacau.. jadi perlu double decode json -_-
			for _, item := range res {
				var objmap dt.FrontendMessageQueue
				jsonslice := []byte(item.Json)
				if err := json.Unmarshal(jsonslice, &objmap); err != nil {
					log.Fatal(err)
					DeleteFrontendQueue(conn, item.ID)
					continue
				}

				if len(objmap.Data) == 0 {
					// skipped because the data is invalid
					log.Error("SKIPPED : Invalid queue data format from frontend")
					DeleteFrontendQueue(conn, item.ID)
					continue
				}

				// second decode
				var realobj dt.FrontendRealMessageQueue
				realslice := []byte(objmap.Data)
				if err := json.Unmarshal(realslice, &realobj); err != nil {
					log.Fatal(err)
					DeleteFrontendQueue(conn, item.ID)
					continue
				}

				// incoming msg store condition : has sender, receiver, not a status, and not from me
				// last update : msg now can be empty body (to support file upload)
				if len(realobj.Data.Sender) > 0 && len(realobj.Data.Destination) > 0 && !realobj.Data.IsStatus && !realobj.Data.FromMe {

					splitSender := strings.Split(realobj.Data.Sender, "@")
					splitDestination := strings.Split(realobj.Data.Destination, "@")
					destination := splitDestination[0]
					// check if the destination = bot
					if !CheckBotConnected(conn, destination) {
						// nomor bot tidak aktif. saat ini belum ada pengecekan untuk membedakan apakah request ini valid atau tidak
						// request dianggap auto valid. will be updated later
					}

					// check if the sender is single user or group.
					splitSenderPhone := strings.Split(splitSender[0], "-")
					if len(splitSenderPhone) > 1 || strings.Contains(splitSender[1], "g.us") {
						log.Infof("Sender %s is from group. Incoming message %s now pushed with additional parameter", splitSender[0], item.ID)
						err := StoreIncomingRequest(conn, realobj.BotID, realobj.Data.Sender, destination, realobj, "1")
						if err != nil {
							log.Errorf("Error handle incoming request : %+v", err)
						}
						log.Info("Successfully store new incoming group message to queue")

						DeleteFrontendQueue(conn, item.ID)
						continue
					}

					// store incoming request to queue
					err := StoreIncomingRequest(conn, realobj.BotID, splitSenderPhone[0], destination, realobj, "")
					if err != nil {
						log.Errorf("Error handle incoming request : %+v", err)
					}
					log.Info("Successfully store new incoming message to queue")
					DeleteFrontendQueue(conn, item.ID)
					continue
				}

				// data skipped
				DeleteFrontendQueue(conn, item.ID)
				continue
			}
		}
	}
}

// DeleteFrontendQueue is used to delete single queue based by ID
func DeleteFrontendQueue(conn *connections.Connections, queueID string) {
	que := lib.RedisQueue{
		Db:           conn.FrontendMsg,
		Name:         config.Param.RedisPrefix + config.Param.FrontendMsg.EventStream,
		GroupName:    config.Param.RedisPrefix + config.Param.FrontendMsg.GroupName,
		ConsumerName: config.Param.RedisPrefix + config.Param.FrontendMsg.ConsumerName,
		ReadTimeout:  5000,
	}
	var sids []string
	sids = append(sids, queueID)
	que.Delete(sids)
}

// StoreIncomingRequest is used to handle single incoming requesst
func StoreIncomingRequest(conn *connections.Connections, botID, sender, destination string, obj dt.FrontendRealMessageQueue, isGroup string) error {
	transid := lib.GetTransactionid(true)

	// push IN message to websocket
	// send "in" struct to websocket
	var hm = ""
	if obj.Data.HasMedia {
		hm = "1"
	}
	var wsmap = make(map[string]interface{})
	wsmap["from"] = sender
	wsmap["bot_id"] = botID
	wsmap["message"] = obj.Data.Body
	wsmap["direction"] = "in"
	wsmap["appid"] = config.Param.AppID
	wsmap["id"] = obj.Data.ID
	wsmap["is_group"] = isGroup
	wsmap["has_media"] = hm
	wsmap["filename"] = obj.Data.Filename
	wsmap["mimetype"] = obj.Data.Mimetype
	wsmap["author"] = obj.Data.Author
	wsmap["reply_to"] = obj.Data.ReplyTo

	wsresp, wscode := lib.CallRestAPIOnBehalf(wsmap, config.Param.WebsocketEndpoint+"/push", "POST", "", 4*time.Second)
	log.Infof("Hit websocket endpoint IN MODE. (HTTP Code %d) : Request : %+v - Response : %s", wscode, wsmap, wsresp)

	// translate incoming request to dt.RequestDataStruct so can be queued
	var incomingStruct dt.RequestDataStruct
	var incomingSub dt.RequestBackendStruct
	incomingSub.ID = obj.Data.ID
	incomingSub.HasMedia = obj.Data.HasMedia
	incomingSub.BotID = botID
	incomingSub.Destination = destination
	incomingSub.Source = sender
	incomingSub.MsgType = "text"
	incomingSub.Filename = obj.Data.Filename
	incomingSub.Mimetype = obj.Data.Mimetype
	incomingSub.TransID = transid
	incomingSub.Direction = "in"
	incomingSub.Message = obj.Data.Body
	incomingSub.IsGroup = isGroup
	incomingSub.Author = obj.Data.Author
	incomingSub.ReplyTo = obj.Data.ReplyTo
	incomingStruct.Data = incomingSub

	// store to msgdata & msgstate as new queue
	return StoreRequestToMessageData(conn, incomingStruct, "delivered", "in")
}
