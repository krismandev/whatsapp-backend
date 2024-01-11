package connections

import (
	"context"
	"os"
	conf "skeleton/config"
	"time"

	redis "github.com/go-redis/redis/v7"
	log "github.com/sirupsen/logrus"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/wzshiming/ssdb"
)

// Connections Holds all passing value to functions
type Connections struct {
	MsgConfig                *redis.Client
	FrontendMsg              *redis.Client
	FrontendEvent            *redis.Client
	MsgState                 *redis.Client
	MsgData                  *as.Client
	DBRedis                  *redis.Client
	Context                  context.Context
	SSDB                     *ssdb.Client
	JWTSecretKey             string
	RedisGroupName           string
	RedisConsumerName        string
	MaxConcurrentProcessData int
}

func InitSSDBConnection(param conf.Configuration) *ssdb.Client {
	mySSDB, err := ssdb.Connect(
		ssdb.Addr(param.SSDB.Url),
	)
	if err != nil {
		log.Errorf("Error SSDB Connection : %+v", err)
		os.Exit(1)
	} else {
		errAuth := mySSDB.Auth(param.SSDB.Password)
		if errAuth != nil {
			log.Errorf("Error SSDB Connection : %+v", errAuth)
			os.Exit(1)
		}
		log.Infof("Connected to SSDB Server : %s", param.SSDB.Url)
	}
	return mySSDB
}

// InitiateConnections is for Initiate Connection
func InitiateConnections(param conf.Configuration) *Connections {
	var conn Connections
	var dbredis *redis.Client
	var msgconfig *redis.Client
	var msgdata *as.Client
	var msgstate *redis.Client
	var frontendmsg *redis.Client
	var frontendEvent *redis.Client

	if len(param.SSDB.Url) > 0 {
		conn.SSDB = InitSSDBConnection(param)
	}
	// ssdb, err := lib.ConnectSSDB(param.SSDB.IP, param.SSDB.Port)
	// if err != nil {
	// 	log.Errorf("Cannot connect to ssdb server : %+v", err)
	// 	panic(err)
	// }
	// conn.SSDB = ssdb

	msgdata, err := as.NewClient(param.MsgData.AerospikeURL, param.MsgData.AerospikePort)
	if err != nil {
		log.Errorf("%v", err)
	}
	log.Infof("MsgData : Connected to aerospike database")

	if len(param.MsgConfig.Sentinel.MasterName) > 0 {
		msgconfig = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    param.MsgConfig.Sentinel.MasterName,
			SentinelAddrs: param.MsgConfig.Sentinel.SentinelURL,
			DialTimeout:   time.Duration(param.RequestTimeout) * time.Second,
		})
	} else {
		msgconfig = redis.NewClient(&redis.Options{
			Addr:     param.MsgConfig.RedisURL,
			Password: param.MsgConfig.RedisPassword,
			DB:       param.MsgConfig.DB,
		})
	}
	log.Infof("MsgConfig : Connected to redis server")

	if len(param.MsgState.Sentinel.MasterName) > 0 {
		msgstate = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    param.MsgState.Sentinel.MasterName,
			SentinelAddrs: param.MsgState.Sentinel.SentinelURL,
			DialTimeout:   time.Duration(param.RequestTimeout) * time.Second,
		})
	} else {
		msgstate = redis.NewClient(&redis.Options{
			Addr:     param.MsgState.RedisURL,
			Password: param.MsgState.RedisPassword,
			DB:       param.MsgState.DB,
		})
	}
	log.Infof("MsgState : Connected to redis server")

	if len(param.FrontendMsg.Sentinel.MasterName) > 0 {
		frontendmsg = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    param.FrontendMsg.Sentinel.MasterName,
			SentinelAddrs: param.FrontendMsg.Sentinel.SentinelURL,
			DialTimeout:   time.Duration(param.RequestTimeout) * time.Second,
		})
	} else {
		frontendmsg = redis.NewClient(&redis.Options{
			Addr:     param.FrontendMsg.RedisURL,
			Password: param.FrontendMsg.RedisPassword,
			DB:       param.FrontendMsg.DB,
		})
	}
	log.Infof("FrontendMsg : Connected to redis server")

	if len(param.FrontendEvent.Sentinel.MasterName) > 0 {
		frontendEvent = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    param.FrontendEvent.Sentinel.MasterName,
			SentinelAddrs: param.FrontendEvent.Sentinel.SentinelURL,
			DialTimeout:   time.Duration(param.RequestTimeout) * time.Second,
		})
	} else {
		frontendEvent = redis.NewClient(&redis.Options{
			Addr:     param.FrontendEvent.RedisURL,
			Password: param.FrontendEvent.RedisPassword,
			DB:       param.FrontendEvent.DB,
		})
	}
	log.Infof("FrontendEvent : Connected to redis server")

	redisType := ""
	if param.UseRedisSentinel {
		log.Infof("Connecting Redis Sentinel Master "+param.RedisSentinel.MasterName+" : %v", param.RedisSentinel.SentinelURL)
		redisType = "redis sentinel"
		dbredis = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    param.RedisSentinel.MasterName,
			SentinelAddrs: param.RedisSentinel.SentinelURL,
			DialTimeout:   time.Duration(param.RequestTimeout) * time.Second,
		})

	} else if param.UseRedis {
		log.Infof("Connecting Redis : %v", param.Redis.RedisURL)
		redisType = "redis"
		dbredis = redis.NewClient(&redis.Options{
			Addr:     param.Redis.RedisURL,
			Password: param.Redis.RedisPassword,
			DB:       param.Redis.DB,
		})
	}

	if param.UseRedis || param.UseRedisSentinel {
		dbStatus := dbredis.Ping()
		if dbStatus.Err() != nil {
			log.Errorf("Error connecting to redis : %v", dbStatus.Err().Error())
			log.Errorf("Unable to connect Redis server %v", dbStatus.Err())
			os.Exit(1)
		}
		log.Infof("Connected to " + redisType)
	} else {
		log.Infof("App doesnt use redis")
	}

	conn.DBRedis = dbredis
	conn.MsgConfig = msgconfig
	conn.MsgState = msgstate
	conn.MsgData = msgdata
	conn.FrontendMsg = frontendmsg
	conn.FrontendEvent = frontendEvent
	//conn.log=&log.Entry{}

	conn.JWTSecretKey = param.JWTSecretKey
	conn.RedisGroupName = param.Redis.GroupName
	conn.RedisConsumerName = param.Redis.ConsumerName
	conn.MaxConcurrentProcessData = param.MaxConcurrentProcessData
	return &conn

}
