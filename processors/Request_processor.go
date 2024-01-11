package processors

import (
	"skeleton/config"
	conf "skeleton/config"
	"skeleton/connections"
	dt "skeleton/datastruct"
	"skeleton/lib"
	"time"

	"github.com/sirupsen/logrus"
)

// InsertRequest is used for
func InsertRequest(req dt.RequestJSONRequest, conn *connections.Connections) error {
	var err error
	que := lib.RedisQueue{
		Db:        conn.DBRedis,
		Name:      config.Param.RedisPrefix + conf.Param.Redis.EventStream,
		GroupName: config.Param.RedisPrefix + conf.Param.Redis.GroupName,
	}

	data := dt.RequestDataStruct{
		Key:  req.Key,
		Data: req.Data,
	}
	_, err = que.CompressPutInQueue(data)
	return err
}

// StoreMsgConfig is used to load the backend/frontend msgconfig so API Handler can consume
func StoreMsgConfig(conn *connections.Connections, key, value string) error {
	edur := conf.Param.MsgConfig.KeyExpirationDuration
	// logrus.Infof("MSGCONFIG : Set Device Connection Status "+key+" : %s", value)
	if edur > 0 {
		duration := time.Duration(edur) * time.Second
		return conn.MsgConfig.Set(key, value, duration).Err()
	}
	return conn.MsgConfig.Set(key, value, 0).Err()
}

// GetMsgConfig is used for get dbconfig data from redis
func GetMsgConfig(conn *connections.Connections, key string) (string, error) {
	result, err := conn.MsgConfig.Get(key).Result()
	// logrus.Info("GetMsgConfig - "+key+": %s", result)
	if err != nil {
		logrus.Error("Error GetMsgConfig : " + err.Error())
	}
	return result, err
}

// SetMsgConfig is used for set dbconfig data to redis
func SetMsgConfig(conn *connections.Connections, key string, value string) error {
	err := conn.MsgConfig.Set(key, value, 0).Err()
	if err != nil {
		logrus.Error("Error SetMsgConfig %s : %s", key, err.Error())
	}
	return err
}

func ReloadMsgConfig() {
	reloadURL := config.Param.APIHandlerEndpoint + "/reload"
	lib.CallRestAPI(reloadURL, "GET", "{}", 5*time.Second)
	logrus.Info("FORCE Called APIHandlerEndpoint to reload the changed MSGCONFIG")
}
