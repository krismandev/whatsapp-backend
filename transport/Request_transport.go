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
	lib "skeleton/lib"
	"skeleton/services"

	"github.com/go-kit/kit/endpoint"
	log "github.com/sirupsen/logrus"
)

// RequestDecodeRequest is use for ...
func RequestDecodeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	log.WithFields(dt.GetLogFieldValues(ctx, "RequestDecodeRequest"))
	var request dt.RequestJSONRequest
	var body []byte
	remoteIP := lib.GetRemoteIPAddress(r)

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
	request.IPAddr = remoteIP
	request.OriginalRequest = string(body)
	return request, nil
}

// NodeNotifyDecodeRequest is use for ...
func NodeNotifyDecodeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	log.WithFields(dt.GetLogFieldValues(ctx, "NodeNotifyDecodeRequest"))
	var request dt.NodeNotifyRequest
	var body []byte

	//decode request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return er.Errorc(dt.ErrInvalidFormat).Rem("Unable to read request body"), nil
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return er.Error(err, dt.ErrInvalidFormat).Rem("Failed decoding json message"), nil
	}
	return request, nil
}

// DirectChatDecodeRequest is use for ...
func DirectChatDecodeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	log.WithFields(dt.GetLogFieldValues(ctx, "DirectChatDecodeRequest"))
	var request dt.DirectChatRequest
	var body []byte

	//decode request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return er.Errorc(dt.ErrInvalidFormat).Rem("Unable to read request body"), nil
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return er.Error(err, dt.ErrInvalidFormat).Rem("Failed decoding json message"), nil
	}
	return request, nil
}

// PassMediaDecodeRequest is use for ...
func PassMediaDecodeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	log.WithFields(dt.GetLogFieldValues(ctx, "PassMediaDecodeRequest"))
	var request dt.PassMediaRequest
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

// RequestEncodeResponse is use for ...
func RequestEncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	log.WithFields(dt.GetLogFieldValues(ctx, "RequestEncodeResponse"))
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

// DirectChatEncodeResponse is use for ...
func DirectChatEncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	log.WithFields(dt.GetLogFieldValues(ctx, "DirectChatEncodeResponse"))
	var body []byte

	body, err := json.Marshal(&response)
	log.Infof("Send Response %s", body[:])
	if err != nil {
		return err
	}

	var e = response.(dt.DirectChatResponse).ResponseCode
	w = WriteHTTPResponse(w, e)
	_, err = w.Write(body)

	return err
}

// NameInformationEncodeResponse is use for ...
func NameInformationEncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	log.WithFields(dt.GetLogFieldValues(ctx, "NameInformationEncodeResponse"))
	var body []byte

	body, err := json.Marshal(&response)
	log.Infof("Send Response %s", body[:])
	if err != nil {
		return err
	}

	var e = response.(dt.ContactDetailResponse).Code
	w = WriteHTTPResponse(w, e)
	_, err = w.Write(body)

	return err
}

// GetContactEncodeResponse is use for ...
func GetContactEncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	log.WithFields(dt.GetLogFieldValues(ctx, "GetContactEncodeResponse"))
	var body []byte

	body, err := json.Marshal(&response)
	log.Infof("Send Response %s", body[:])
	if err != nil {
		return err
	}

	var e = response.(dt.GetContactResponse).Code
	w = WriteHTTPResponse(w, e)
	_, err = w.Write(body)

	return err
}

// CreateRequestEndpoint is use for
func CreateRequestEndpoint(svc services.RequestServices, conn *connections.Connections) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		log.WithFields(dt.GetLogFieldValues(ctx, "CreateRequestEndpoint"))
		if conf.Param.UseJWT {
			errNoJWT, errJWT := HandleJWT(ctx)
			if errJWT != nil {
				return dt.GlobalJSONResponse{ResponseCode: errNoJWT, ResponseDesc: errJWT.Error()}, nil
			}
		}

		if req, ok := request.(dt.RequestJSONRequest); ok {
			return svc.CreateRequest(ctx, req, conn), nil
		}
		log.Error("Unhandled error occured: request is in unknown format")
		return dt.GlobalJSONResponse{ResponseCode: dt.ErrOthers, ResponseDesc: dt.DescOthers}, nil
	}
}

// DirectChatEndpoint is use for
func DirectChatEndpoint(svc services.RequestServices, conn *connections.Connections) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		log.WithFields(dt.GetLogFieldValues(ctx, "DirectChatEndpoint"))
		if req, ok := request.(dt.DirectChatRequest); ok {
			return svc.DirectChat(ctx, req, conn), nil
		}
		log.Error("Unhandled error occured: request is in unknown format")
		return dt.DirectChatResponse{ResponseCode: dt.ErrOthers, ResponseDesc: dt.DescOthers}, nil
	}
}

// NodeNotifyEndpoint is use for
func NodeNotifyEndpoint(svc services.RequestServices, conn *connections.Connections) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		if req, ok := request.(dt.NodeNotifyRequest); ok {
			return svc.NodeNotify(ctx, req, conn), nil
		}
		log.Error("Unhandled error occured: request is in unknown format")
		return dt.GlobalJSONResponse{ResponseCode: dt.ErrOthers, ResponseDesc: dt.DescOthers}, nil
	}
}

// NameInformationEndpoint is use for
func NameInformationEndpoint(svc services.RequestServices, conn *connections.Connections) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		log.WithFields(dt.GetLogFieldValues(ctx, "NameInformationEndpoint"))
		if req, ok := request.(dt.RequestJSONRequest); ok {
			return svc.GetNameInformation(ctx, req, conn), nil
		}
		log.Error("Unhandled error occured: request is in unknown format")
		return dt.ContactDetailResponse{Code: dt.ErrOthers, Message: dt.DescOthers}, nil
	}
}

// CallFrontendEndpoint is use for
func CallFrontendEndpoint(svc services.RequestServices, conn *connections.Connections) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		log.WithFields(dt.GetLogFieldValues(ctx, "CallFrontendEndpoint"))
		if conf.Param.UseJWT {
			errNoJWT, errJWT := HandleJWT(ctx)
			if errJWT != nil {
				return dt.GlobalJSONResponse{ResponseCode: errNoJWT, ResponseDesc: errJWT.Error()}, nil
			}
		}

		if req, ok := request.(dt.RequestJSONRequest); ok {
			return svc.CallFrontend(ctx, req, conn), nil
		}
		log.Error("Unhandled error occured: request is in unknown format")
		return dt.GlobalJSONResponse{ResponseCode: dt.ErrOthers, ResponseDesc: dt.DescOthers}, nil
	}
}

// DeviceStatusRequestEndpoint is use for
func DeviceStatusRequestEndpoint(svc services.RequestServices, conn *connections.Connections) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		log.WithFields(dt.GetLogFieldValues(ctx, "DeviceStatusRequestEndpoint"))
		if conf.Param.UseJWT {
			errNoJWT, errJWT := HandleJWT(ctx)
			if errJWT != nil {
				return dt.GlobalJSONResponse{ResponseCode: errNoJWT, ResponseDesc: errJWT.Error()}, nil
			}
		}

		if req, ok := request.(dt.RequestJSONRequest); ok {
			return svc.DeviceStatusRequest(ctx, req, conn), nil
		}
		log.Error("Unhandled error occured: request is in unknown format")
		return dt.GlobalJSONResponse{ResponseCode: dt.ErrOthers, ResponseDesc: dt.DescOthers}, nil
	}
}

// PassMediaEndpoint is use for
func PassMediaEndpoint(svc services.RequestServices, conn *connections.Connections) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		log.WithFields(dt.GetLogFieldValues(ctx, "PassMediaEndpoint"))
		if conf.Param.UseJWT {
			errNoJWT, errJWT := HandleJWT(ctx)
			if errJWT != nil {
				return dt.GlobalJSONResponse{ResponseCode: errNoJWT, ResponseDesc: errJWT.Error()}, nil
			}
		}

		if req, ok := request.(dt.PassMediaRequest); ok {
			return svc.PassMedia(ctx, req, conn), nil
		}
		log.Error("Unhandled error occured: request is in unknown format")
		return dt.GlobalJSONResponse{ResponseCode: dt.ErrOthers, ResponseDesc: dt.DescOthers}, nil
	}
}

// GetContactEndpoint is use for
func GetContactEndpoint(svc services.RequestServices, conn *connections.Connections) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		log.WithFields(dt.GetLogFieldValues(ctx, "GetContactEndpoint"))
		if conf.Param.UseJWT {
			errNoJWT, errJWT := HandleJWT(ctx)
			if errJWT != nil {
				return dt.GlobalJSONResponse{ResponseCode: errNoJWT, ResponseDesc: errJWT.Error()}, nil
			}
		}

		if req, ok := request.(dt.RequestJSONRequest); ok {
			return svc.GetContacts(ctx, req, conn), nil
		}
		log.Error("Unhandled error occured: request is in unknown format")
		return dt.GlobalJSONResponse{ResponseCode: dt.ErrOthers, ResponseDesc: dt.DescOthers}, nil
	}
}

// NodeMetricsEndpoint is use for
func NodeMetricsEndpoint(svc services.RequestServices, conn *connections.Connections) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		log.WithFields(dt.GetLogFieldValues(ctx, "NodeMetricsEndpoint"))
		if conf.Param.UseJWT {
			errNoJWT, errJWT := HandleJWT(ctx)
			if errJWT != nil {
				return dt.GlobalJSONResponse{ResponseCode: errNoJWT, ResponseDesc: errJWT.Error()}, nil
			}
		}

		if req, ok := request.(dt.RequestJSONRequest); ok {
			return svc.NodeMetrics(ctx, req, conn), nil
		}
		log.Error("Unhandled error occured: request is in unknown format")
		return dt.GlobalJSONResponse{ResponseCode: dt.ErrOthers, ResponseDesc: dt.DescOthers}, nil
	}
}

func ManualStoreMessageDecodeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	log.WithFields(dt.GetLogFieldValues(ctx, "ManualStoreMessageDecodeRequest"))
	var request dt.RequestDataStruct
	var body []byte

	//decode request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return er.Errorc(dt.ErrInvalidFormat).Rem("Unable to read request body"), nil
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return er.Error(err, dt.ErrInvalidFormat).Rem("Failed decoding json message"), nil
	}
	return request, nil
}

// NodeNotifyEndpoint is use for
func ManualStoreMessageEndpoint(svc services.RequestServices, conn *connections.Connections) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		if req, ok := request.(dt.RequestDataStruct); ok {
			return svc.ManualStoreMessage(ctx, req, conn), nil
		}
		log.Error("Unhandled error occured: request is in unknown format")
		return dt.GlobalJSONResponse{ResponseCode: dt.ErrOthers, ResponseDesc: dt.DescOthers}, nil
	}
}
