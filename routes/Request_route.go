package routes

import (
	"net/http"
	"skeleton/connections"
	"skeleton/services"
	"skeleton/transport"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// RequestRoute is used for
func RequestRoute(conn *connections.Connections) {
	var svcRequest services.RequestServices
	svcRequest = services.RequestService{}

	CreateRequestHandler := httptransport.NewServer(
		transport.CreateRequestEndpoint(svcRequest, conn),
		transport.RequestDecodeRequest,
		transport.GlobalEncodeResponse,
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
		httptransport.ServerBefore(transport.GetRequestInformation),
	)

	DeviceStatusRequestHandler := httptransport.NewServer(
		transport.DeviceStatusRequestEndpoint(svcRequest, conn),
		transport.RequestDecodeRequest,
		transport.GlobalEncodeResponse,
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
		httptransport.ServerBefore(transport.GetRequestInformation),
	)

	CallFrontendHandler := httptransport.NewServer(
		transport.CallFrontendEndpoint(svcRequest, conn),
		transport.RequestDecodeRequest,
		transport.GlobalEncodeResponse,
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
		httptransport.ServerBefore(transport.GetRequestInformation),
	)

	PassMediaHandler := httptransport.NewServer(
		transport.PassMediaEndpoint(svcRequest, conn),
		transport.PassMediaDecodeRequest,
		transport.GlobalEncodeResponse,
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
		httptransport.ServerBefore(transport.GetRequestInformation),
	)

	GetContactHandler := httptransport.NewServer(
		transport.GetContactEndpoint(svcRequest, conn),
		transport.RequestDecodeRequest,
		transport.GetContactEncodeResponse,
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
		httptransport.ServerBefore(transport.GetRequestInformation),
	)

	NameInformationHandler := httptransport.NewServer(
		transport.NameInformationEndpoint(svcRequest, conn),
		transport.RequestDecodeRequest,
		transport.NameInformationEncodeResponse,
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
		httptransport.ServerBefore(transport.GetRequestInformation),
	)

	NodeNotifyHandler := httptransport.NewServer(
		transport.NodeNotifyEndpoint(svcRequest, conn),
		transport.NodeNotifyDecodeRequest,
		transport.GlobalEncodeResponse,
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
		httptransport.ServerBefore(transport.GetRequestInformation),
	)

	NodeMetricHandler := httptransport.NewServer(
		transport.NodeMetricsEndpoint(svcRequest, conn),
		transport.RequestDecodeRequest,
		transport.GlobalEncodeResponse,
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
		httptransport.ServerBefore(transport.GetRequestInformation),
	)

	DirectChatHandler := httptransport.NewServer(
		transport.DirectChatEndpoint(svcRequest, conn),
		transport.DirectChatDecodeRequest,
		transport.DirectChatEncodeResponse,
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
		httptransport.ServerBefore(transport.GetRequestInformation),
	)

	ManualStoreMessageHandler := httptransport.NewServer(
		transport.ManualStoreMessageEndpoint(svcRequest, conn),
		transport.ManualStoreMessageDecodeRequest,
		transport.GlobalEncodeResponse,
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
		httptransport.ServerBefore(transport.GetRequestInformation),
	)

	r := mux.NewRouter()
	r.Methods("POST").Handler(CreateRequestHandler)
	http.Handle("/dummy-request", r)

	rr := mux.NewRouter()
	rr.Methods("POST").Handler(DeviceStatusRequestHandler)
	http.Handle("/update-msgconfig", rr)

	rrr := mux.NewRouter()
	rrr.Methods("GET").Handler(CallFrontendHandler)
	http.Handle("/call-frontend", rrr)

	rrrr := mux.NewRouter()
	rrrr.Methods("POST").Handler(PassMediaHandler)
	http.Handle("/pass-media", rrrr)

	rrrrr := mux.NewRouter()
	rrrrr.Methods("GET").Handler(GetContactHandler)
	http.Handle("/get-contacts", rrrrr)

	rrrrrr := mux.NewRouter()
	rrrrrr.Methods("GET").Handler(NameInformationHandler)
	http.Handle("/name-information", rrrrrr)

	rrrrrrr := mux.NewRouter()
	rrrrrrr.Methods("POST").Handler(NodeNotifyHandler)
	http.Handle("/node-notify", rrrrrrr)

	rrrrrrrr := mux.NewRouter()
	rrrrrrrr.Methods("GET").Handler(NodeMetricHandler)
	http.Handle("/node-metric", rrrrrrrr)

	rrrrrrrrr := mux.NewRouter()
	rrrrrrrrr.Methods("POST").Handler(DirectChatHandler)
	http.Handle("/direct-chat", rrrrrrrrr)

	manualStoreMessageRoute := mux.NewRouter()
	manualStoreMessageRoute.Methods("POST").Handler(ManualStoreMessageHandler)
	http.Handle("/manual-store-message", manualStoreMessageRoute)
}
