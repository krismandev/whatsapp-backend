package config

import (
	//"fmt"
	"os"
	//"time"
	parser "skeleton/parser"

	log "github.com/sirupsen/logrus"
)

type TDB struct {
	DBUrl  string `yaml:"dbUrl"`
	DBType string `yaml:"dbType"`
}

// Configuration stores global configuration
type Configuration struct {
	ListenPort               string         `yaml:"listenPort"`
	DBList                   map[string]TDB `yaml:"dblist"`
	PublicEndpointURL        string         `yaml:"publicEndpointURL"`
	APIHandlerEndpoint       string         `yaml:"apiHandlerEndpoint"`
	UseJWT                   bool           `yaml:"useJWT"`
	AppID                    string         `yaml:"appID"`
	JWTSecretKey             string         `yaml:"JWTSecretKey"`
	RequestTimeout           int            `yaml:"requestTimeout"`
	MaxConcurrentProcessData int            `yaml:"maxConcurrentProcessData"`
	ConcurrentWaitLimit      int            `yaml:"concurrentWaitLimit"`
	SuspendAPIUrl            string         `yaml:"suspendAPIUrl"`
	PPOBAPIUrl               string         `yaml:"ppobAPIUrl"`
	SchedulerDuration        int            `yaml:"schedulerDuration"`
	MaxFailedRequest         int            `yaml:"maxFailedRequest"`
	SleepDurationWhenFailed  int            `yaml:"sleepDurationWhenFailed"`
	WAClientEndpoint         string         `yaml:"waClientEndpoint"`
	WAClientPort             string         `yaml:"waClientPort"`
	WAClientNamespace        string         `yaml:"waClientNamespace"`
	PublicStoragePath        string         `yaml:"publicStoragePath"`
	BotID                    string         `yaml:"botID"`
	RedisPrefix              string         `yaml:"redisPrefix"`
	MaxBodyLogLength         int            `yaml:"maxBodyLogLength"`
	WebsocketEndpoint        string         `yaml:"websocketEndpoint"`
	WhatsappEventLogURL      string         `yaml:"whatsappEventLogURL"`
	MaximumTimeoutThreshold  int            `yaml:"maximumTimeoutThreshold"`
	AutoReconnectEvery       int            `yaml:"autoReconnectEvery"`
	OCSURL                   string         `yaml:"ocsURL"`
	TranslogHeaders          []string       `yaml:"translogHeaders"`
	NotRegisteredOnWAKeepDuration int `yaml:"notRegisteredOnWAKeepDuration"`
	SSDB struct {
		Url   string `yaml:"url"`
		Password string    `yaml:"password"`
	} `yaml:"ssdb"`

	MsgDataDriver string `yaml:"msgDataDriver"`
	MsgData       struct {
		AerospikeURL  string `yaml:"aerospikeUrl"`
		AerospikePort int    `yaml:"aerospikePort"`
		Namespace     string `yaml:"namespace"`
		Table         string `yaml:"table"`
		TTL           int    `yaml:"ttl"`
	} `yaml:"msgData"`

	MsgState struct {
		RedisURL              string `yaml:"redisUrl"`
		RedisPassword         string `yaml:"redisPassword"`
		DB                    int    `yaml:"db"`
		KeyExpirationDuration int    `yaml:"keyExpirationDuration"`
		EventStream           string `yaml:"eventStream"`
		GroupName             string `yaml:"groupName"`
		ConsumerName          string `yaml:"consumerName"`
		Sentinel              struct {
			MasterName       string   `yaml:"masterName"`
			SentinelPassword string   `yaml:"sentinelPassword"`
			SentinelURL      []string `yaml:"sentinelUrl"`
		} `yaml:"sentinel"`
	} `yaml:"msgState"`

	FrontendMsg struct {
		RedisURL      string `yaml:"redisUrl"`
		RedisPassword string `yaml:"redisPassword"`
		DB            int    `yaml:"db"`
		EventStream   string `yaml:"eventStream"`
		GroupName     string `yaml:"groupName"`
		ConsumerName  string `yaml:"consumerName"`
		Sentinel      struct {
			MasterName       string   `yaml:"masterName"`
			SentinelPassword string   `yaml:"sentinelPassword"`
			SentinelURL      []string `yaml:"sentinelUrl"`
		} `yaml:"sentinel"`
	} `yaml:"frontendMsg"`

	FrontendEvent struct {
		RedisURL      string `yaml:"redisUrl"`
		RedisPassword string `yaml:"redisPassword"`
		DB            int    `yaml:"db"`
		EventStream   string `yaml:"eventStream"`
		GroupName     string `yaml:"groupName"`
		ConsumerName  string `yaml:"consumerName"`
		Sentinel      struct {
			MasterName       string   `yaml:"masterName"`
			SentinelPassword string   `yaml:"sentinelPassword"`
			SentinelURL      []string `yaml:"sentinelUrl"`
		} `yaml:"sentinel"`
	} `yaml:"frontendEvent"`

	MsgConfig struct {
		RedisURL              string `yaml:"redisUrl"`
		RedisPassword         string `yaml:"redisPassword"`
		DB                    int    `yaml:"db"`
		KeyExpirationDuration int    `yaml:"keyExpirationDuration"`
		Sentinel              struct {
			MasterName       string   `yaml:"masterName"`
			SentinelPassword string   `yaml:"sentinelPassword"`
			SentinelURL      []string `yaml:"sentinelUrl"`
		} `yaml:"sentinel"`
	} `yaml:"msgConfig"`

	UseRedis         bool `yaml:"useRedis"`
	UseRedisSentinel bool `yaml:"useRedisSentinel"`
	RedisSentinel    struct {
		MasterName       string   `yaml:"masterName"`
		SentinelPassword string   `yaml:"sentinelPassword"`
		SentinelURL      []string `yaml:"sentinelUrl"`
	} `yaml:"redisSentinel"`
	Redis struct {
		RedisURL      string `yaml:"redisUrl"`
		RedisPassword string `yaml:"redisPassword"`
		DB            int    `yaml:"db"`
		GroupName     string `yaml:"groupName"`
		ConsumerName  string `yaml:"consumerName"`
		EventStream   string `yaml:"eventStream"`
	} `yaml:"redis"`
	Translog string `yaml:"translog"`
	Log      struct {
		FileNamePrefix string `yaml:"filenamePrefix"`
		FileName       string
		Level          string `yaml:"level"`
	} `yaml:"log"`
}

//Param is use for
var Param Configuration

//LoadConfig is use for
func LoadConfig(fn *string) {

	if err := parser.LoadYAML(fn, &Param); err != nil {
		log.Errorf("LoadConfigFromFile() - Failed opening config file %s\n%s", &fn, err)
		os.Exit(1)
	}
	// //log.Logf("Loaded configs: %v", Param)
	//t := time.Now()
	//sDate := fmt.Sprintf("%d%02d%02d", t.Year(), t.Month(), t.Day())
	Param.Log.FileName = Param.Log.FileNamePrefix + ".log"
}
