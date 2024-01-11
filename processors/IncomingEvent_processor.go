package processors

import (
	"encoding/json"
	"skeleton/config"
	"skeleton/connections"
	dt "skeleton/datastruct"
	"skeleton/lib"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func IncomingEventListener(conn *connections.Connections, wg *sync.WaitGroup, closeFlag <-chan struct{}) {
	wg.Add(1)

	defer wg.Done()
	// create group & consumer reader
	err := lib.RedisCreateStreamReader(conn.FrontendEvent, config.Param.RedisPrefix+config.Param.FrontendEvent.EventStream, config.Param.RedisPrefix+config.Param.FrontendEvent.GroupName)
	if err != nil {
		log.Error("[IncomingEventListener] Failed to create stream reader : " + err.Error())
	}
	readNew := false

	for {
		select {
		case <-closeFlag:
			return
		default:
			que := lib.RedisQueue{
				Db:           conn.FrontendEvent,
				Name:         config.Param.RedisPrefix + config.Param.FrontendEvent.EventStream,
				GroupName:    config.Param.RedisPrefix + config.Param.FrontendEvent.GroupName,
				ConsumerName: config.Param.RedisPrefix + config.Param.FrontendEvent.ConsumerName,
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
				log.Info("FrontendEvent Redis queue is empty in old mode. Change frontendEvent redis stream read mode from OLD to NEW mode")
				readNew = true
			}

			// format yang dikirim dari nodejs agak kacau.. jadi perlu double decode json -_-
			for _, item := range res {
				var objmap dt.FrontendEventStruct
				var sjson dt.EventSingleDataStruct
				jsonslice := []byte(item.Json)
				// log.Infof("RAW DATA : %s", item.Json)
				if err := json.Unmarshal(jsonslice, &sjson); err != nil {
					log.Error(err)
					DeleteFrontendEventQueue(conn, item.ID)
					continue
				}

				subjsonslice := []byte(sjson.Data)
				if err := json.Unmarshal(subjsonslice, &objmap); err != nil {
					log.Error(err)
					DeleteFrontendEventQueue(conn, item.ID)
					continue
				}
				log.Infof("EVENT DATA AFTER DECODE: %+v", objmap)

				// handle event by Key
				// hit APIHandlerEndpoint to update the bot event status
				eventUpdateURL := config.Param.APIHandlerEndpoint + "/update-event"
				var pass dt.APIHandlerEventStruct
				pass.Key = objmap.Key
				pass.BotID = objmap.BotID

				if objmap.Key == "QRReceived" {
					pass.Data = objmap.Data.QRString
				} else if objmap.Key == "QRScanFailed" || objmap.Key == "AuthFailure" {
					pass.Data = strconv.Itoa(objmap.FailedIteration)
				} else if objmap.Key == "SendMsg" {
					PushStatusUpdate(conn, objmap)
					var reqdata dt.RequestDataStruct
					rdid, ok := objmap.Data.ID["id"].(string)
					if !ok {
						logrus.Infof("BROKEN FE QUEUE : %+v", objmap)
						DeleteFrontendEventQueue(conn, item.ID)
						continue
					}
					reqdata.Data.ID = rdid

					reqdata.Data.TransID = objmap.TransID
					reqdata.Data.BotID = objmap.BotID
					reqdata.Data.AppID = config.Param.AppID
					reqdata.Data.RequestTime = time.Now().Format("2006-01-02 15:04:05")
					reqdata.Data.Source = objmap.Data.From
					reqdata.Data.Destination = objmap.Data.To
					reqdata.Data.HasMedia = objmap.Data.HasMedia
					reqdata.Data.Message = objmap.Data.Body
					reqdata.Key = objmap.Key

					var status = "failed"
					if objmap.Data.ACK == 0 {
						status = "pending"
					} else if objmap.Data.ACK == 1 {
						status = "sent"
					} else if objmap.Data.ACK == 2 {
						status = "received"
					} else if objmap.Data.ACK >= 3 {
						status = "read"
					}

					log.Infof("ACK is %d", objmap.Data.ACK)
					if objmap.Data.ACK >= 1 {
						log.Infof("CALL StoreMsgState from IncomingEventListener when ack is received/read.")
						StoreMsgState(conn, reqdata, status, "out")
					} else {
						log.Infof("SKIP STOREMSGSTATE FROM IncomingEventListener")
					}
				}

				if len(pass.Data) > 0 {
					// only push event to API Handler when there is data to be sent
					log.Infof("Pass param event to API Handler : %+v", pass)
					body, httpStatus := lib.CallRestAPIOnBehalf(pass, eventUpdateURL, "POST", "", 5*time.Second)
					if httpStatus < 200 || httpStatus >= 400 {
						log.Errorf("IncomingEventListener : Failed to push event to API Handler. Will retry later. (Body : %s)", body)
						// request failed. retry hit again after sleep 1 or 2 second
						time.Sleep(1 * time.Second)
						continue
					}
					log.Info("IncomingEventListener : Bot event has been pushed to API Handler")
				}

				DeleteFrontendEventQueue(conn, item.ID)
				continue
			}

		}
	}
}

// DeleteFrontendEventQueue is used to delete single queue based by ID
func DeleteFrontendEventQueue(conn *connections.Connections, queueID string) {
	que := lib.RedisQueue{
		Db:           conn.FrontendEvent,
		Name:         config.Param.RedisPrefix + config.Param.FrontendEvent.EventStream,
		GroupName:    config.Param.RedisPrefix + config.Param.FrontendEvent.GroupName,
		ConsumerName: config.Param.RedisPrefix + config.Param.FrontendEvent.ConsumerName,
		ReadTimeout:  5000,
	}
	var sids []string
	sids = append(sids, queueID)
	que.Delete(sids)
}
