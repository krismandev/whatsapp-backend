package transport

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"skeleton/config"
	conf "skeleton/config"
	connections "skeleton/connections"
	dt "skeleton/datastruct"
	er "skeleton/error"
	"skeleton/services"

	"github.com/go-kit/kit/endpoint"
	log "github.com/sirupsen/logrus"
)

// HistoryDecodeRequest is use for ...
func HistoryDecodeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	log.WithFields(dt.GetLogFieldValues(ctx, "HistoryDecodeRequest"))
	var request dt.HistoryChatRequest
	var body []byte

	//decode request body
	body, err := ioutil.ReadAll(r.Body)
	var logBody string
	if len(body) > config.Param.MaxBodyLogLength {
		// chunk body log request
		logBody = "(Chunked %d first chars) " + string(body[:config.Param.MaxBodyLogLength])
	} else {
		logBody = string(body)
	}

	log.Infof("Received Request %s", logBody)
	if err != nil {
		return er.Errorc(dt.ErrInvalidFormat).Rem("Unable to read request body"), nil
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return er.Error(err, dt.ErrInvalidFormat).Rem("Failed decoding json message"), nil
	}

	return request, nil
}

// HistoryEncodeResponse is use for ...
func HistoryEncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	log.WithFields(dt.GetLogFieldValues(ctx, "HistoryEncodeResponse"))
	var body []byte

	body, err := json.Marshal(&response)
	log.Infof("Send Response %s", body[:])
	if err != nil {
		return err
	}

	var e = response.(dt.GlobalJSONResponse).ResponseCode
	w = WriteHTTPResponse(w, e)
	_, err = w.Write(body)

	return err
}

// CallHistoryEndpoint is use for
func CallHistoryEndpoint(svc services.HistoryServices, conn *connections.Connections) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		log.WithFields(dt.GetLogFieldValues(ctx, "CallHistoryEndpoint"))
		if conf.Param.UseJWT {
			errNoJWT, errJWT := HandleJWT(ctx)
			if errJWT != nil {
				return dt.GlobalJSONResponse{ResponseCode: errNoJWT, ResponseDesc: errJWT.Error()}, nil
			}
		}

		if req, ok := request.(dt.HistoryChatRequest); ok {
			return svc.CallHistoryGeneration(ctx, req, conn), nil
		}
		log.Error("Unhandled error occured: request is in unknown format")
		return dt.GlobalJSONResponse{ResponseCode: dt.ErrOthers, ResponseDesc: dt.DescOthers}, nil
	}
}

// StoreHistoryEndpoint is use for
func StoreHistoryEndpoint(svc services.HistoryServices, conn *connections.Connections) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		log.WithFields(dt.GetLogFieldValues(ctx, "StoreHistoryEndpoint"))
		if conf.Param.UseJWT {
			errNoJWT, errJWT := HandleJWT(ctx)
			if errJWT != nil {
				return dt.GlobalJSONResponse{ResponseCode: errNoJWT, ResponseDesc: errJWT.Error()}, nil
			}
		}

		if req, ok := request.(dt.HistoryChatRequest); ok {
			return svc.StoreHistoryGeneration(ctx, req, conn), nil
		}
		log.Error("Unhandled error occured: request is in unknown format")
		return dt.GlobalJSONResponse{ResponseCode: dt.ErrOthers, ResponseDesc: dt.DescOthers}, nil
	}
}
