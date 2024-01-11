package lib

import (
	"context"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

type contextKey string

var (
	//ContextTransactionID for context transaction ID
	ContextTransactionID = contextKey("IMSTransactionID")
)

// GetLogFieldValues is  use for Set Log Fields
func GetLogFieldValues(ctx context.Context, Module string) log.Fields {
	var fields log.Fields = make(log.Fields)
	fields["Node"] = GetHostName()
	fields["Module"] = Module
	if ctx != nil {
		fields["TxID"] = ctx.Value(ContextTransactionID).(string)
	}
	return fields
}

// CloseLog to close file handler
func CloseLog(logFile *os.File) {
	logFile.Close()
}

// InitLog to Initiate log file
func InitLog(logFilename, level string, logFile *os.File) {
	if len(logFilename) == 0 {
		log.Infof("Startup, assigning log to %s ,loglevel:%s", "stdout", level)
	} else {
		log.Infof("Startup, assigning log to %s ,loglevel:%s", logFilename, level)
	}
	LevelMap := map[string]log.Level{"panic": log.PanicLevel, "fatal": log.FatalLevel, "error": log.ErrorLevel, "warning": log.WarnLevel, "info": log.InfoLevel, "debug": log.DebugLevel, "trace": log.TraceLevel}
	LogLevel, ok := LevelMap[level]
	if ok {
		log.SetLevel(LogLevel)
	}
	writers := []io.Writer{
		os.Stdout,
	}
	log.SetReportCaller(true)
	formatter := &log.TextFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyTime:  "t",
			log.FieldKeyLevel: "l",
			log.FieldKeyMsg:   "m"}}
	if strings.Trim(logFilename, " \t\\/") != "" {
		var err error
		logFile, err = os.OpenFile(logFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

		if logFile == nil {
			log.Error("Startup, Opening log file fail ", err)
			os.Exit(1)
		}

		writers = append(writers, logFile)
	}

	log.SetFormatter(formatter)
	log.SetOutput(io.MultiWriter(writers...))
}
