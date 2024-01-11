package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	conf "skeleton/config"
	dt "skeleton/datastruct"
	lib "skeleton/lib"
	"strings"
	"time"

	httptransport "github.com/go-kit/kit/transport/http"
	log "github.com/sirupsen/logrus"
)

// GlobalEncodeResponse is used for
func GlobalEncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	// log.WithFields(dt.GetLogFieldValues(ctx, "GlobalEncodeResponse"))
	var body []byte

	body, err := json.Marshal(&response)

	// var logBody string
	// if len(body) > config.Param.MaxBodyLogLength {
	// 	// chunk body log request
	// 	logBody = "(Chunked %d first chars) " + string(body[:config.Param.MaxBodyLogLength])
	// } else {
	// 	logBody = string(body)
	// }

	// log.Infof("Send Response %s", logBody)
	if err != nil {
		return err
	}

	var e = response.(dt.GlobalJSONResponse).ResponseCode
	w = WriteHTTPResponse(w, e)
	_, err = w.Write(body)

	return err
}

// HandleJWT is used for
func HandleJWT(ctx context.Context) (int, error) {
	tokenAuth, ok := ctx.Value(httptransport.ContextKeyRequestAuthorization).(string)
	if !strings.Contains(tokenAuth, "Bearer ") {
		return dt.ErrNotAuthorized, errors.New(dt.DescNotAuthorized)
	}
	tokenAuthKey := strings.Replace(tokenAuth, "Bearer ", "", 1)
	tokenAuthKey = strings.Trim(tokenAuthKey, " ")
	remoteIP, ok := ctx.Value(httptransport.ContextKeyRequestRemoteAddr).(string)
	remoteIP, _, _ = net.SplitHostPort(remoteIP)
	forwardedFor := ctx.Value(httptransport.ContextKeyRequestXForwardedFor).(string)
	log.Infof("["+remoteIP+"] ReceivedRequest : %s ; ForwardedFor : "+forwardedFor, tokenAuthKey)
	if !ok {
		log.Errorf("["+remoteIP+"] Error : %s", "Invalid/Not Found/Expired Auth")
		return dt.ErrNotAuthorized, errors.New(dt.DescNotAuthorized)
	}
	claim, tokenValid := lib.ValidateToken(tokenAuthKey, remoteIP, dt.APPID, conf.Param.JWTSecretKey)

	if !tokenValid {
		log.Errorf("["+remoteIP+"] Error : %s", "Invalid Token")
		return dt.ErrNotAuthorized, errors.New(dt.DescNotAuthorized)
	}
	if claim.ExpiresAt < time.Now().Unix() {
		log.Errorf("["+remoteIP+"] Error : %s", "Token Expired")
		return dt.ErrNotAuthorized, errors.New(dt.DescNotAuthorized)
	}
	return 0, nil
}

// WriteHTTPResponse is used for
func WriteHTTPResponse(w http.ResponseWriter, e int) http.ResponseWriter {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Connection", "close")

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
	return w
}
