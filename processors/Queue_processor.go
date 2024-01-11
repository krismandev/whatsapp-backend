package processors

import (
	"encoding/json"
	"errors"
	"io/ioutil"
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

// QueueListener is used for listen the queue from redis stream
func QueueListener(conn *connections.Connections, wg *sync.WaitGroup, closeFlag <-chan struct{}) {
	wg.Add(1)
	defer wg.Done()
	err := lib.RedisCreateStreamReader(conn.DBRedis, config.Param.RedisPrefix+config.Param.Redis.EventStream, config.Param.RedisPrefix+config.Param.Redis.GroupName)
	if err != nil {
		log.Error("[QueueListener A] Failed to create stream reader : " + err.Error())
	}

	err = lib.RedisCreateStreamReader(conn.MsgState, config.Param.RedisPrefix+config.Param.MsgState.EventStream, config.Param.RedisPrefix+config.Param.MsgState.GroupName)
	if err != nil {
		log.Error("[QueueListener B] Failed to create stream reader : " + err.Error())
	}

	maxFailedRequest := config.Param.MaxFailedRequest
	sleepDurationWhenFailed := config.Param.SleepDurationWhenFailed
	failedIteration := 0
	readNew := false //later will be changed to "true" when all the old data has been processed
	var temporaryQueueData dt.RequestDataStruct

	// manual rate limit
	manualTimeControl := make(map[string]int64)

	for {
		select {
		case <-closeFlag:
			return
		default:
			var errHandleData error
			if failedIteration >= maxFailedRequest && len(temporaryQueueData.QueueID) > 0 {
				log.Infof("Will try rerun old queue to frontend later after sleep %d seconds", sleepDurationWhenFailed)
				//sleep, then retry
				time.Sleep(time.Duration(sleepDurationWhenFailed) * time.Second)
				failedIteration, errHandleData = HandleSingleQueueDataRequest(conn, temporaryQueueData, failedIteration)
				if errHandleData != nil && failedIteration == 0 {
					// TODO : send ke service lain

					// ada case iteration = 0, tapi gagal kirim. penyebabnya kemungkinan karena salah format data input
					log.Error("TEMPORARY DATA QUEUE SKIPPED. Failed to run the queue " + temporaryQueueData.QueueID + " because there is data error : " + errHandleData.Error())
				}

				if failedIteration != 0 {
					// masih gagal..
					err := StoreMsgConfig(conn, config.Param.RedisPrefix+"BotStatus-"+temporaryQueueData.Data.Source, "0")
					if err != nil {
						log.Errorf("MSGCONFIG ERROR : %s", err.Error())
					}
					log.Infof("Old queue rerun still failed (Iteration %d)", failedIteration)
					continue
				} else {
					// sudah berhasil! set temporary QueueID ke "" agar tidak diproses di next request
					err := StoreMsgConfig(conn, config.Param.RedisPrefix+"BotStatus-"+temporaryQueueData.Data.Source, "1")
					if err != nil {
						log.Errorf("MSGCONFIG ERROR : %s", err.Error())
					}
					temporaryQueueData.QueueID = ""
				}

			}

			que := lib.RedisQueue{
				Db:           conn.DBRedis,
				Name:         config.Param.RedisPrefix + config.Param.Redis.EventStream,
				GroupName:    config.Param.RedisPrefix + config.Param.Redis.GroupName,
				ConsumerName: config.Param.RedisPrefix + config.Param.Redis.ConsumerName,
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
				log.Info("Redis queue is empty in old mode. Change redis stream read mode from OLD to NEW mode")
				readNew = true
			}

			for _, item := range res {
				var objmap dt.RequestDataStruct
				jsonslice := []byte(item.Json)
				if err := json.Unmarshal(jsonslice, &objmap); err != nil {
					log.Errorf("Cannot decode object from json: %v to Struct:  %v", string(jsonslice), objmap)
					log.Fatal(err)
				}
				objmap.QueueID = item.ID
				if len(objmap.Data.BotID) == 0 {
					// must be skipped or will be run infinitely
					log.Error("Undefined BotID. Request skipped")
				}

				// check rate limit
				now := time.Now()
				lastTimestamp, ok := manualTimeControl[objmap.Data.BotID]
				if !ok {
					// new queue bot id, generate map[string]int with current timestamp first. no sleep required
					manualTimeControl[objmap.Data.BotID] = now.Unix()
				} else {
					// check if the last timestamp vs current timestamp is not too long
					selisih := now.Unix() - lastTimestamp
					botRateKey := config.Param.RedisPrefix + "BotRateLimit-" + objmap.Data.BotID
					botRateLimit, err := GetMsgConfig(conn, botRateKey)
					if err != nil {
						fallbackWaitLimit := config.Param.ConcurrentWaitLimit
						if fallbackWaitLimit < 1 {
							fallbackWaitLimit = 2
						}
						log.Error("Failed to grab bot rate limit via redis msgconfig. Will set the default rate limit %d", fallbackWaitLimit)
						SetMsgConfig(conn, botRateKey, strconv.Itoa(fallbackWaitLimit))
						botRateLimit = strconv.Itoa(fallbackWaitLimit)
					}

					intBotRate, _ := strconv.Atoi(botRateLimit)
					if int(selisih) < intBotRate {
						// throttle request. need to sleep a bit for this bot number.
						sleepDuration := intBotRate - int(selisih)
						time.Sleep(time.Duration(sleepDuration) * time.Second)
					}
					manualTimeControl[objmap.Data.BotID] = now.Unix()
				}

				//handle this single request.
				failedIteration, errHandleData = HandleSingleQueueDataRequest(conn, objmap, failedIteration)
				if failedIteration == 0 && errHandleData != nil {
					// TODO : send ke service lain

					//case ada error, tapi queue tidak perlu diretry. kemungkinan karena format data salah
					log.Error("QUEUE SKIPPED. Failed to run queue " + item.ID + " : " + errHandleData.Error())
					DeleteQueue(conn, objmap.QueueID)
					continue
				}
				if failedIteration > 0 {
					// -- TIDAK USAH RETRY. LANGSUNG SET GAGAL SAJA AGAR QUEUE TIDAK MENUMPUK? (2021-11-29)

					//artinya request gagal. karena kalau berhasil, harus bernilai 0
					// for x := failedIteration; x < maxFailedRequest; x++ {
					// 	failedIteration, errHandleData = HandleSingleQueueDataRequest(conn, objmap, failedIteration)
					// 	if failedIteration == 0 {
					// 		// sudah berhasil konek.
					// 		break
					// 	}
					// }

					log.Error("QUEUE SKIPPED BECAUSE SEND CHAT ERROR TOO MUCH. Failed to run queue " + item.ID + " : " + errHandleData.Error())
					DeleteQueue(conn, objmap.QueueID)
					failedIteration = 0 //reset error iteration ke 0 lagi
					continue

					// -- DISABLE SEMENTARA. MENCEGAH BOT TIBA2 KEDISCONNECT SENDIRI TIBA2
					// //jika sudah diloop berulang2 dan masih gagal, stop proses queue dengan menambahkan data ke variabel temporary
					// temporaryQueueData = objmap
					// // set msgConfig bot status to disconnected
					// err := StoreMsgConfig(conn, config.Param.RedisPrefix+"BotStatus-"+objmap.Data.Source, "0")
					// if err != nil {
					// 	log.Errorf("MSGCONFIG ERROR : %s", err.Error())
					// }

				}
			}
		}
	}
}

// HandleSingleQueueDataRequest is used for
func HandleSingleQueueDataRequest(conn *connections.Connections, item dt.RequestDataStruct, failedIteration int) (int, error) {
	var err error
	phoneTarget := item.Data.Source
	if len(phoneTarget) == 0 {
		return 0, errors.New("Incomplete request. Make sure the Source, Destination, and Message parameter is exists")
	}

	translog := InitTranslogFromQueueRequest(item.Data)

	if CheckBotConnected(conn, phoneTarget) {
		//hit Frontend AP
		var pass dt.FrontendSendMessageRequest
		pass.Phone = item.Data.Source
		pass.Receiver = item.Data.Destination
		pass.Message = item.Data.Message
		pass.ReplyTo = item.Data.ReplyTo
		pass.TransID = item.Data.TransID

		if len(item.Data.Buttons) > 0 {
			for _, btn := range item.Data.Buttons {
				if len(btn.Label) > 0 {
					var singleBtn dt.WAButtonStruct
					// untuk saat ini, hanya parameter Body saja yg dibutuhkan
					singleBtn.Body = btn.Label
					singleBtn.ID = btn.CallbackData
					pass.Buttons = append(pass.Buttons, singleBtn)
				}
			}
		}

		var httpResponse dt.FrontendDeviceInfoResponse
		var bodyResp string
		var httpCode int

		var hitURL string

		// store ke msgdata & msgstate dulu. nanti update msgstatenya
		err := StoreRequestToMessageData(conn, item, "pending", "out")
		if err != nil {
			translog.SetDataField("status", "failed")
			translog.SetDataField("error_detail", "ERR_WABACKEND_QUEUE="+err.Error())
			translog.Store()
			log.Errorf("Error StoreRequestToMessageData : ", err)
			return (failedIteration + 1), err
		}

		// if item.Data.BotID != config.Param.BotID {
		// 	return 0, errors.New("Invalid BotID parameter. This queue data is not for whatsapp apps. ")
		// }
		if len(item.Data.BotID) == 0 || len(item.Data.Source) == 0 || len(item.Data.Destination) == 0 {
			translog.SetDataField("status", "failed")
			translog.SetDataField("error_detail", "INCOMPLETE_PARAMETER")
			translog.Store()
			// StoreMsgState(conn, item, "failed")
			return 0, errors.New("Incomplete request. Make sure the Source, Destination, and Message parameter is exists")
		}
		if len(item.Data.Message) == 0 && item.Data.MsgType == "text" {
			translog.SetDataField("status", "failed")
			translog.SetDataField("error_detail", "INCOMPLETE_PARAMETER")
			translog.Store()
			// StoreMsgState(conn, item, "failed")
			return 0, errors.New("You need to define the Message parameter")
		}

		if len(item.Data.MsgType) != 0 && item.Data.MsgType != "text" {
			imgPath := config.Param.PublicStoragePath + "/" + item.Data.BotID + "/media-data/" + item.Data.TransID + ".data"
			storedStream, err := ioutil.ReadFile(imgPath)
			if err != nil {
				translog.SetDataField("status", "failed")
				translog.SetDataField("error_detail", "ERR_WABACKEND_READFILE="+err.Error())
				translog.Store()
				// cannot send image : failed to open image file
				log.Errorf("Error ioutil.Readfile() : " + err.Error())
				// StoreMsgState(conn, item, "failed")
				return 0, err
			}

			var passmedia dt.PassMediaRequest
			err = json.Unmarshal(storedStream, &passmedia)
			if err != nil {
				translog.SetDataField("status", "failed")
				translog.SetDataField("error_detail", "ERR_UNMARSHAL_DOCUMENT="+err.Error())
				translog.Store()
				log.Errorf("Failed to unmarshal the stored document : " + err.Error())
				// StoreMsgState(conn, item, "failed")
				return 0, err
			}

			if item.Data.MsgType == "image" {
				pass.Image = passmedia.Data
				hitURL = "/app/send-image"
			} else if item.Data.MsgType == "audio" {
				pass.Filename = passmedia.Filename
				pass.Mimetype = passmedia.Mimetype
				pass.Data = passmedia.Data
				hitURL = "/app/send-audio"
			} else if item.Data.MsgType == "video" {
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

		podaddr, err := NodeSelector(conn, item.Data.Source, false)

		port, _ := conn.DBRedis.Get("NODE-POD-" + podaddr + "-PORT").Result()
		if len(port) == 0 {
			port = config.Param.WAClientPort
		}
		finalURL := "http://" + podaddr + config.Param.WAClientNamespace + port + hitURL

		bodyResp, httpCode = lib.CallRestAPIOnBehalf(pass, finalURL, "POST", "", 5*time.Second)
		log.Infof("Hit Frontend : %s . HTTP Code %d - %s", finalURL, httpCode, bodyResp)
		errorResult := "Failed to send data to frontend"
		var status = "failed"
		if httpCode >= 200 && httpCode < 400 {
			//request success (setidaknya ada body)
			err := json.Unmarshal([]byte(bodyResp), &httpResponse)
			log.Infof("Frontend Body Parsed Requests : %+v", httpResponse)
			if err != nil {
				translog.SetDataField("status", "failed")
				translog.SetDataField("error_detail", "WAQUEUE_WRONG_RESPONSE="+err.Error())
				translog.Store()

				// StoreMsgState(conn, item, "failed")
				return (failedIteration + 1), err
			}

			// GRAB NEW ID FROM frontendRequest.Message
			rawMessage := strings.Split(httpResponse.Message, "|")
			var newMsgID string
			if len(rawMessage) > 1 {
				newMsgID = rawMessage[len(rawMessage)-1]
				item.Data.ID = newMsgID
			}

			if httpResponse.Type != "success" && len(httpResponse.Message) > 0 {
				errorResult = httpResponse.Message
			} else {
				translog.SetDataField("messageid", newMsgID)
				translog.Store()

				status = "success"
				PushChatToWebsocketAfterSend(item, status, "")
				// request berhasil! hapus queue
				// StoreMsgState(conn, item, "success")
				DeleteQueue(conn, item.QueueID)
				return 0, nil
			}
		}

		jsonpass, err := json.Marshal(pass)
		log.Infof("Err hit frontend send chat (Request : %s) : %s", jsonpass, errorResult)

		// will trigger OCS Refund if ChargeData exists
		if item.Data.ChargeData.Amount > 0 {
			go RefundChatRequest(item.Data.ChargeData)
		}

		// only push to websocket if the request is not have a DBIgnore=true.
		if !item.Data.DBIgnore {
			PushChatToWebsocketAfterSend(item, status, errorResult)
		} else {
			logrus.Infof("Request has a DBIgnore parameter, so request is not saved to websocket. %+v", item)
		}

		// only mark request as failed when HTTP status is bad or server error
		if httpCode >= 400 {
			translog.SetDataField("status", "failed")
			translog.SetDataField("error_detail", "ERR_WAHTTP_RESP="+errorResult)
			translog.Store()
			StoreMsgState(conn, item, "failed", "out")
		} else {
			translog.SetDataField("status", "unknown")
			translog.SetDataField("error_detail", "ERR_WAHTTP_RESP="+errorResult)
			translog.Store()
			StoreMsgState(conn, item, "unknown", "out")
		}

		//hit API gagal : langsung hapus dari queue
		DeleteQueue(conn, item.QueueID)
		return 0, nil
		// return (failedIteration + 1), errors.New(errorResult)
	}

	// LAST UPDATE :
	// jika bot disconnected, langsung hapus queue saja biar ga numpuk.
	if err != nil {
		log.Infof("Err : " + err.Error())
	}
	DeleteQueue(conn, item.QueueID)
	return 0, nil

	// log.Infof("Bot is not connected. (Iteration %d)", failedIteration)
	// return (failedIteration + 1), err
}

func RefundChatRequest(result lib.ChargeResult) (err error) {
	var cr = lib.RefundRequest{
		AccountID:    result.AccountID,
		AmountRefund: result.Amount,
		BalanceType:  result.BalanceType,
		RefID:        result.TransactionID,
		AppID:        config.Param.AppID,
	}
	lib.SetChargeUrl(config.Param.OCSURL)
	refundResult, err := lib.Refund(cr)
	if err != nil {
		logrus.Errorf("Error Charging (Request=%+v) : %+v", cr, err)
	}
	logrus.Infof("Trigger RefundChatRequest (Request=%+v) : %+v", cr, refundResult)
	return
}

func PushChatToWebsocketAfterSend(item dt.RequestDataStruct, status string, error_result string) {
	// push outgoing message to websocket
	// send "out" struct to websocket
	var wsmap = make(map[string]interface{})
	var hm = ""
	if item.Data.HasMedia {
		hm = "1"
	}

	wsmap["appid"] = config.Param.AppID
	wsmap["userapi_id"] = item.Data.UserID
	wsmap["login_id"] = item.Data.LoginID
	wsmap["from"] = item.Data.Destination
	wsmap["transid"] = item.Data.TransID
	wsmap["bot_id"] = item.Data.BotID
	wsmap["message"] = item.Data.Message
	wsmap["direction"] = "out"
	wsmap["id"] = item.Data.ID
	wsmap["is_group"] = item.Data.IsGroup
	wsmap["has_media"] = hm
	wsmap["filename"] = item.Data.Filename
	wsmap["mimetype"] = item.Data.Mimetype
	wsmap["author"] = item.Data.Author
	wsmap["reply_to"] = item.Data.ReplyTo
	wsmap["status"] = status
	wsmap["error_result"] = error_result
	wsmap["requestid"] = item.Data.RequestID
	wsmap["referenceid"] = item.Data.ReferenceID

	wsresp, wscode := lib.CallRestAPIOnBehalf(wsmap, config.Param.WebsocketEndpoint+"/push", "POST", "", 4*time.Second)
	log.Infof("Hit websocket endpoint OUT MODE. (HTTP Code %d) : Request : %+v - Response : %s", wscode, wsmap, wsresp)

}

// DeleteQueue is used to delete single queue based by ID
func DeleteQueue(conn *connections.Connections, queueID string) {
	que := lib.RedisQueue{
		Db:           conn.DBRedis,
		Name:         config.Param.RedisPrefix + config.Param.Redis.EventStream,
		GroupName:    config.Param.RedisPrefix + config.Param.Redis.GroupName,
		ConsumerName: config.Param.RedisPrefix + config.Param.Redis.ConsumerName,
		ReadTimeout:  5000,
	}
	var sids []string
	sids = append(sids, queueID)
	que.Delete(sids)
}

// CheckBotConnected is used for check the phone number is connected or not
func CheckBotConnected(conn *connections.Connections, phoneNumber string) bool {
	podaddr, err := NodeSelector(conn, phoneNumber, false)
	port, _ := conn.DBRedis.Get("NODE-POD-" + podaddr + "-PORT").Result()
	if len(port) == 0 {
		port = config.Param.WAClientPort
	}
	hitURL := "http://" + podaddr + config.Param.WAClientNamespace + port + "/app/get-device-information?phone=" + phoneNumber

	var blank dt.FrontendSendMessageRequest
	blank.Phone = phoneNumber
	bodyResp, _ := lib.CallRestAPIOnBehalf(blank, hitURL, "GET", "", 5*time.Second)
	log.Infof("Check device information %s : %s", hitURL, bodyResp)

	var resp dt.FrontendDeviceInfoResponse
	err = json.Unmarshal([]byte(bodyResp), &resp)
	if err != nil {
		return false
	}

	return resp.Detail.IsActive
}
