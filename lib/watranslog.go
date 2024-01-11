package lib

import (
	"os"
	"skeleton/config"
	"strings"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type WATranslog struct {
	Data     map[string]string
	Headers  []string
	SavePath string
}

// SetData is to prepare data instance before stored
func (w WATranslog) SetData(data map[string]string) WATranslog {
	w.Data = data
	return w
}

// SetDataField is to update or add a field record in translog
func (w WATranslog) SetDataField(fieldname string, value string) WATranslog {
	w.Data[fieldname] = value
	return w
}

// Store will execute the save translog
func (w WATranslog) Store() {
	headers := w.Headers
	if len(w.Headers) == 0 {
		headers = config.Param.TranslogHeaders
	}
	savepath := w.SavePath
	if len(savepath) == 0 {
		savepath = config.Param.Log.FileNamePrefix + "-translog.log"
	}

	// auto map value
	if len(w.Data["hostname"]) == 0 {
		w.Data["hostname"], _ = os.Hostname()
	}

	var storedData []string
	for _, hname := range headers {
		storedData = append(storedData, w.Data[hname])
	}

	// store this []string to translog
	rowlog := strings.Join(storedData, "|")

	f, err := os.OpenFile(savepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logrus.Errorf("Error translog : %+v", err)
	}
	defer f.Close()
	if _, err := f.WriteString(rowlog + "\n"); err != nil {
		log.Error("Error writing translog : " + err.Error())
	}

}
