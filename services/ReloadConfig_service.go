package services

import (
	"context"
	dt "skeleton/datastruct"
	connections "skeleton/connections"
	//"strings"
	conf "skeleton/config"

	log "github.com/sirupsen/logrus"

)

// ReloadConfigServices provides operations for endpoint
type ReloadConfigServices interface {
	ReloadConfig(context.Context, dt.ReloadConfigJSONRequest, *connections.Connections) dt.ReloadConfigJSONResponse
}

// ReloadConfigService is use for
type ReloadConfigService struct{}

const ()

//ReloadConfigValidateRequest Validate Draw Prize Request
func ReloadConfigValidateRequest(conn *connections.Connections, req dt.ReloadConfigJSONRequest) *dt.ReloadConfigJSONResponse {
	return nil
}

//ReLoadConfig to load config from config.yml file
func ReLoadConfig(configPath string) {
	conf.LoadConfig(&configPath)
	log.Infof("Configuration Reloaded")
}

// ReloadConfig service is use for
func (ReloadConfigService) ReloadConfig(ctx context.Context, req dt.ReloadConfigJSONRequest, conn *connections.Connections) dt.ReloadConfigJSONResponse {
	//transid:=ctx.Value(dt.ContextTransactionID).(string)
	logger := log.WithFields(dt.GetLogFieldValues(ctx, "ReloadConfig"))
	logger.Info("ReloadConfig Triggered")

	ReLoadConfig("config/config.yml")

	return dt.ReloadConfigJSONResponse{
		ResponseCode: dt.ErrSuccess,
		ResponseDesc: dt.DescSuccess,
		IPAddr:       req.IPAddr,
	}
}
