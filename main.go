package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	conf "skeleton/config"
	"skeleton/connections"
	lib "skeleton/lib"
	"skeleton/processors"
	"skeleton/routes"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

// WaitTimeout to wait with timeout
func WaitTimeout(WaitG *sync.WaitGroup, timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		defer close(done)
		WaitG.Wait()
	}()
	select {
	case <-done:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

// GetLogFieldValues is  use for Set Log Fields
func GetLogFieldValues(ctx context.Context, Module string) log.Fields {
	var fields log.Fields = make(log.Fields)
	fields["Node"] = lib.GetHostName()
	fields["Module"] = Module
	if ctx != nil {
		fields["TxID"] = ctx.Value(lib.ContextTransactionID).(string)
	}
	return fields
}

// LoadConfig to load config from config.yml file
func LoadConfig(configPath, logFileName, logLevel string, logFile *os.File) {
	configFile := flag.String("config", configPath, "main configuration file")
	conf.LoadConfig(configFile)
	flag.Parse()
	log.Infof("Reads configuration from %s", *configFile)
	lib.InitLog(logFileName, logLevel, logFile)
}

// PingHandler Liveness
func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

// VersionHandler for Version
func VersionHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprint("{\"Version\":\"" + version + "\",\"BuildDate\":\"" + builddate + "\"}")))
}

// TokenHandler for JWT TOken
func TokenHandler(conn *connections.Connections) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ipaddr := lib.GetRemoteIPAddress(r)
		bearer, err := lib.GenerateToken("APP", ipaddr, conn.JWTSecretKey, 1)
		if err != nil {
			w.Write([]byte(fmt.Sprint("{\"error\":\"" + err.Error() + "\"}")))
		}
		w.Write([]byte(fmt.Sprint("{\"bearer_token\":\"" + bearer + "\"}")))
	}
	return fn
}

// initHandlers handle all route requests
func initHandlers(conn *connections.Connections) {
	http.HandleFunc("/ping", PingHandler)
	http.HandleFunc("/ver", VersionHandler)
	http.HandleFunc("/get-token", TokenHandler(conn))

	// load all routes from Init_route
	routes.InitRoutes(conn)
}

// log incoming request
func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Errorf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		// var logBody string
		// if len(body) > config.Param.MaxBodyLogLength {
		// 	// chunk body log request
		// 	logBody = "(Chunked %d first chars) " + string(body[:config.Param.MaxBodyLogLength])
		// } else {
		// 	logBody = string(body)
		// }

		// Work / inspect body. You may even modify it!
		// And now set a new body, which will simulate the same data we read:
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		log.WithFields(log.Fields{"ip": r.RemoteAddr, "method": r.Method, "url": r.URL}).Info("request")
		handler.ServeHTTP(w, r)

	})
}

var version = "0.0.0"
var builddate = ""

func main() {
	log.Infof("Running on version %s, build date : %s", version, builddate)

	// closeFlag := make(chan struct{})
	var logFile *os.File
	LoadConfig("config/config.yml", conf.Param.Log.FileNamePrefix, conf.Param.Log.Level, logFile)
	conn := connections.InitiateConnections(conf.Param)
	defer lib.CloseLog(logFile)
	flag.Parse()
	// var err error

	initHandlers(conn)
	// start queue listener

	// lib.RegisterExitSignalHandler(func() {
	// 	log.Info("Shutting down gracefully with exit signal")
	// 	close(closeFlag)
	// })

	log.Info("System Ready")

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGKILL)
	server := &http.Server{Addr: conf.Param.ListenPort, Handler: logRequest(http.DefaultServeMux)}

	go func() {
		<-exit
		//initiate gracefullShutDown
		log.Info("Initiate gracefully shutdown with exit signal")
		close(lib.GracefullShutdownChan)
		// WaitTimeout(&lib.WaitG, 10*time.Second)
		log.Info("Shutting down gracefully with exit signal")
		server.Shutdown(context.TODO())
	}()

	go processors.BotTypeMapping(conn, &lib.WaitG)
	go processors.ManageNodesHealth(conn, &lib.WaitG, lib.GracefullShutdownChan)
	go processors.BotAutoReconnect(conn, true, &lib.WaitG)
	go processors.QueueListener(conn, &lib.WaitG, lib.GracefullShutdownChan)
	go processors.IncomingMessageListener(conn, &lib.WaitG, lib.GracefullShutdownChan)
	go processors.IncomingEventListener(conn, &lib.WaitG, lib.GracefullShutdownChan)

	log.Info(server.ListenAndServe())
	WaitTimeout(&lib.WaitG, 10*time.Second)
	//err = http.ListenAndServe(conf.Param.ListenPort, nil)

	// if err != nil {
	// 	log.Errorf("Unable to start the server %v", err)
	// 	os.Exit(1)
	// }
}
