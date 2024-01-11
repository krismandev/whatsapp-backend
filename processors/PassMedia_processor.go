package processors

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"skeleton/config"
	dt "skeleton/datastruct"

	"github.com/sirupsen/logrus"
)

func StorePassMedia(req dt.PassMediaRequest) (error, bool) {
	var err error
	storepath := config.Param.PublicStoragePath + "/" + req.BotID + "/media-data"
	storeFinalpath := storepath + "/" + req.ID + ".data"

	logrus.Infof("Try to StorePassMedia to : %s", storeFinalpath)

	// store as json decoded file
	jsonData, err := json.Marshal(req)
	if err != nil {
		logrus.Error("Error when marshalling StorePassMedia data : " + err.Error())
		return err, false
	}

	os.MkdirAll(storepath, os.ModePerm)
	err = ioutil.WriteFile(storeFinalpath, jsonData, 0644)
	if err != nil {
		logrus.Error("Error when write file to passmedia : " + err.Error())
		return err, false
	}

	return err, true
}
