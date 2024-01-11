package routes

import (
	"net/http"
	"skeleton/connections"
	"skeleton/services"
	"skeleton/transport"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// HistoryRoute is used for
func HistoryRoute(conn *connections.Connections) {
	var svcHistory services.HistoryServices
	svcHistory = services.HistoryService{}

	callhistory := mux.NewRouter()
	callhistory.Methods("POST").Handler(httptransport.NewServer(
		transport.CallHistoryEndpoint(svcHistory, conn),
		transport.HistoryDecodeRequest,
		transport.GlobalEncodeResponse,
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
		httptransport.ServerBefore(transport.GetRequestInformation),
	))
	http.Handle("/call-history", callhistory)

	storehistory := mux.NewRouter()
	storehistory.Methods("POST").Handler(httptransport.NewServer(
		transport.StoreHistoryEndpoint(svcHistory, conn),
		transport.HistoryDecodeRequest,
		transport.GlobalEncodeResponse,
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
		httptransport.ServerBefore(transport.GetRequestInformation),
	))
	http.Handle("/store-history", storehistory)

}
