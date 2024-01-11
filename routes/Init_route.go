package routes

import (
	"net/http"
	"skeleton/connections"
	"skeleton/services"
	"skeleton/transport"

	httptransport "github.com/go-kit/kit/transport/http"
)

// InitRoutes handle all routes in this application
func InitRoutes(conn *connections.Connections) {
	// You must register your new copied stub module here
	// Example :
	// StubRoute(conn)
	RequestRoute(conn)
	HistoryRoute(conn)

	globalDefaultRoute(conn)
}

func globalDefaultRoute(conn *connections.Connections) {
	var svcReloadConfig services.ReloadConfigServices
	svcReloadConfig = services.ReloadConfigService{}
	ReloadConfigHandler := httptransport.NewServer(
		transport.ReloadConfigEndpoint(svcReloadConfig, conn),
		transport.ReloadConfigDecodeRequest,
		transport.ReloadConfigEncodeResponse,
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
		httptransport.ServerBefore(transport.GetRequestInformation),
	)
	http.Handle("/reload", ReloadConfigHandler)
}
