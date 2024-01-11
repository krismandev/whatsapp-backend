package transport

import (
	"context"
	"net/http"
	//"time"
	dt "skeleton/datastruct"
	lib "skeleton/lib"

	log "github.com/sirupsen/logrus"

)

var mylog *log.Entry

// GetRequestInformation is use for ...
func GetRequestInformation(ctx context.Context, r *http.Request) context.Context {
	ctx = context.WithValue(ctx, dt.ContextTransactionID, lib.GetTransactionid(true))
	return ctx
}

