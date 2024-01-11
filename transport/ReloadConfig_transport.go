package transport

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	conf "skeleton/config"
	connections "skeleton/connections"
	dt "skeleton/datastruct"
	er "skeleton/error"
	lib "skeleton/lib"
	"skeleton/services"
	"strings"
	"time"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	log "github.com/sirupsen/logrus"
)

// ReloadConfigDecodeRequest is use for ...
func ReloadConfigDecodeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	log.WithFields(dt.GetLogFieldValues(ctx, "ReloadConfigDecodeRequest"))
	var request dt.ReloadConfigJSONRequest
	remoteIP := lib.GetRemoteIPAddress(r)
	log.Info("Reload Config Request")
	request.IPAddr = remoteIP
	return request, nil
}

// ReloadConfigEncodeResponse is use for ...
func ReloadConfigEncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	log.WithFields(dt.GetLogFieldValues(ctx, "ReloadConfigEncodeResponse"))
	var body []byte

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Connection", "close")
	body, err := json.Marshal(&response)
	log.Infof("Send Response %s", body[:])
	if err != nil {
		return err
	}

	//var e = response.(dt.ReloadConfigJSONResponse).ResponseCode

	// if e < 8000 {
	// 	w.WriteHeader(http.StatusOK)
	// } else if e < 9000 {
	// 	w.WriteHeader(http.StatusUnauthorized)
	// } else if e <= 9999 {
	// 	w.WriteHeader(http.StatusBadRequest)
	// } else {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// }
	w.WriteHeader(http.StatusOK)
	//_, err = w.Write(body)

	return err
}

// ReloadConfigEndpoint is use for
func ReloadConfigEndpoint(svc services.ReloadConfigServices, conn *connections.Connections) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		log.WithFields(dt.GetLogFieldValues(ctx, "ReloadConfigEndpoint"))
		if conf.Param.UseJWT {
			tokenAuth, ok := ctx.Value(httptransport.ContextKeyRequestAuthorization).(string)
			if !strings.Contains(tokenAuth, "Bearer ") {
				return dt.ReloadConfigJSONResponse{ResponseCode: dt.ErrNotAuthorized, ResponseDesc: dt.DescNotAuthorized}, nil
			}
			tokenAuthKey := strings.Replace(tokenAuth, "Bearer ", "", 1)
			tokenAuthKey = strings.Trim(tokenAuthKey, " ")
			remoteIP, ok := ctx.Value(httptransport.ContextKeyRequestRemoteAddr).(string)
			remoteIP, _, _ = net.SplitHostPort(remoteIP)
			forwardedFor := ctx.Value(httptransport.ContextKeyRequestXForwardedFor).(string)
			log.Infof("["+remoteIP+"] ReceivedRequest : %s ; ForwardedFor : "+forwardedFor, tokenAuthKey)
			if !ok {
				log.Errorf("["+remoteIP+"] Error : %s", "Invalid/Not Found/Expired Auth")
				return dt.ReloadConfigJSONResponse{ResponseCode: dt.ErrNotAuthorized, ResponseDesc: dt.DescNotAuthorized}, nil
			}
			claim, tokenValid := lib.ValidateToken(tokenAuthKey, remoteIP, dt.APPID, conf.Param.JWTSecretKey)

			if !tokenValid {
				log.Errorf("["+remoteIP+"] Error : %s", "Invalid Token")
				return dt.ReloadConfigJSONResponse{ResponseCode: dt.ErrNotAuthorized, ResponseDesc: dt.DescNotAuthorized}, nil
			}
			if claim.ExpiresAt < time.Now().Unix() {
				log.Errorf("["+remoteIP+"] Error : %s", "Token Expired")
				return dt.ReloadConfigJSONResponse{ResponseCode: dt.ErrTokenExpired, ResponseDesc: dt.DescTokenExpired}, nil
			}
		}

		if req, ok := request.(dt.ReloadConfigJSONRequest); ok {
			return svc.ReloadConfig(ctx, req, conn), nil
		}
		switch request.(type) {
		case *er.AppError:
			{
				if request != nil {
					return dt.ReloadConfigJSONResponse{ResponseCode: request.(*er.AppError).ErrCode, ResponseDesc: request.(*er.AppError).Remark}, nil
				}
			}
		}
		log.Error("Unhandled error occured: request is in unknown format")
		return dt.ReloadConfigJSONResponse{ResponseCode: dt.ErrOthers, ResponseDesc: dt.DescOthers}, nil
	}
}
