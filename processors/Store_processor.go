package processors

import (
	"skeleton/config"
	"skeleton/connections"
	dt "skeleton/datastruct"
	"skeleton/lib"

	as "github.com/aerospike/aerospike-client-go"
	log "github.com/sirupsen/logrus"
)

// StoreRequestToMessageData is used to store request to redis + aerospike as msgdata
func StoreRequestToMessageData(conn *connections.Connections, req dt.RequestDataStruct, state, direction string) error {
	// MsgData must be stored first. Because the msgstate redis stream is dependent to this data
	var err error
	err = StoreMsgData(conn, req, direction)
	if err != nil {
		return err
	}

	err = StoreMsgState(conn, req, state, "in")
	return err
}

// StoreMsgState store handler
func StoreMsgState(conn *connections.Connections, req dt.RequestDataStruct, state string, direction string) error {
	storeQue := lib.RedisQueue{
		Db:        conn.MsgState,
		Name:      config.Param.RedisPrefix + config.Param.MsgState.EventStream,
		GroupName: config.Param.RedisPrefix + config.Param.MsgState.GroupName,
		// ConsumerName: config.Param.MsgState.ConsumerName,
		ReadTimeout: 5000,
	}
	log.Infof("Used MSGSTATE Redis Stream : %s %s %s", config.Param.RedisPrefix+config.Param.MsgState.EventStream, config.Param.RedisPrefix+config.Param.MsgState.GroupName, config.Param.RedisPrefix+config.Param.MsgState.ConsumerName)

	var pass dt.MsgStateRequestDataStruct
	var passdata dt.MsgStateRequest
	passdata.AppID = req.Data.AppID
	passdata.BotID = req.Data.BotID
	passdata.TransID = req.Data.TransID
	passdata.ID = req.Data.ID
	passdata.State = state
	passdata.Direction = direction
	passdata.RequestTime = req.Data.RequestTime
	passdata.From = req.Data.Source
	passdata.To = req.Data.Destination
	passdata.HasMedia = req.Data.HasMedia
	passdata.Body = req.Data.Message

	pass.Key = req.Key
	pass.Data = passdata
	qid, err := storeQue.CompressPutInQueue(pass)
	log.Info("QueueID : " + qid)

	return err
}

// StoreMsgData is used for
func StoreMsgData(conn *connections.Connections, req dt.RequestDataStruct, direction string) error {
	log.Infof("Used TransID %s as aerospike key", req.Data.TransID)
	msgkey, err := as.NewKey(config.Param.MsgData.Namespace, config.Param.MsgData.Table, req.Data.TransID)
	if err != nil {
		return err
	}
	ttl := uint32(config.Param.MsgData.TTL)
	writepol := as.NewWritePolicy(0, ttl)
	hm := "0"
	if req.Data.HasMedia {
		hm = "1"
	}

	bins := as.BinMap{
		"ID":          req.Data.ID,
		"HasMedia":    hm,
		"AppID":       req.Data.AppID,
		"BotID":       req.Data.BotID,
		"TransID":     req.Data.TransID,
		"MsgType":     req.Data.MsgType,
		"MediaPath":   req.Data.MediaPath,
		"Source":      req.Data.Source,
		"Destination": req.Data.Destination,
		"Message":     req.Data.Message,
		"IsGroup":     req.Data.IsGroup,
		"Direction":   direction,
		"Filename":    req.Data.Filename,
		"Mimetype":    req.Data.Mimetype,
		"Author":      req.Data.Author,
		"ReplyTo":     req.Data.ReplyTo,
	}
	log.Infof("Put MsgData to Aerospike : %+v", bins)
	err = conn.MsgData.Put(writepol, msgkey, bins)

	return err
}
