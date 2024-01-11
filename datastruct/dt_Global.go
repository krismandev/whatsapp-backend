package datastruct

import (
	"context"
	lib "skeleton/lib"

	logger "github.com/sirupsen/logrus"
)

// GlobalJSONResponse is used for
type GlobalJSONResponse struct {
	TransID      string        `json:"trans_id"`
	ResponseCode int           `json:"response_code"`
	ResponseDesc string        `json:"response_desc"`
	ErrorDetail  string        `json:"error_detail"`
	LastInsertID string        `json:"last_insert_id,omitempty"`
	Data         []interface{} `json:"data,omitempty"`
}

// DataTableParam is used for
type DataTableParam struct {
	PerPage  int    `json:"per_page"`
	Page     int    `json:"page"`
	OrderBy  string `json:"order_by"`
	OrderDir string `json:"order_dir"`
}

//GetLogFieldValues is  use for Set Log Fields
func GetLogFieldValues(ctx context.Context, Module string) logger.Fields {
	var fields logger.Fields = make(logger.Fields)
	fields["Node"] = lib.GetHostName()
	fields["Module"] = Module
	if ctx != nil {
		fields["TxID"] = ctx.Value(ContextTransactionID).(string)
	}
	return fields
}

// ResponseHandlerHelper is used for
func ResponseHandlerHelper(ctx context.Context, responseCode int, responseDesc string, err error) GlobalJSONResponse {
	transid := ctx.Value(ContextTransactionID).(string)
	errorDetail := ""
	if err != nil {
		errorDetail = err.Error()
	}
	return GlobalJSONResponse{
		TransID:      transid,
		ResponseCode: responseCode,
		ResponseDesc: responseDesc,
		ErrorDetail:  errorDetail,
	}
}

//ReloadConfigJSONRequest is use for
type ReloadConfigJSONRequest struct {
	OriginalRequest string
	IPAddr          string
}

//ReloadConfigJSONResponse is  use for
type ReloadConfigJSONResponse struct {
	ResponseCode int    `json:"responseCode"`
	ResponseDesc string `json:"responseDesc"`
	IPAddr       string `json:"-"`
}

//ReloadConfigType is  use for
type ReloadConfigType struct {
	Event        string `json:"event,omitempty"`
	DeviceNumber string `json:"device_number,omitempty"`
	Timestamp    string `json:"timestamp,omitempty"`
}
