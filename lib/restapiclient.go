package lib

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strings"
	"time"
	log "github.com/sirupsen/logrus"
)

//CallRestAPI simplify CallRestAPIOnBehalf
func CallRestAPI (url,method,request string, timeout time.Duration) (responseBody string, httpStatus int){
	var param interface{}
	json.Unmarshal([]byte(request),&param)
	body,httpStatus:=CallRestAPIOnBehalf(param, url,method, "",timeout)
	return body, httpStatus
}

//CallRestAPIStruct simplify CallRestAPIOnBehalf with struct as input
func CallRestAPIStruct(param interface{}, url string, timeout time.Duration) (body string, httpStatus int) {
	return CallRestAPIOnBehalf(param, url,"POST", "",timeout)
}

//CallRestAPIOnBehalf Call Rest API
func CallRestAPIOnBehalf(param interface{}, url,method, IP string, timeout time.Duration) (body string, httpStatus int) {
	jsonValue, _ := json.Marshal(param)

	// create request structure
	req, err := http.NewRequest(method, url, strings.NewReader(string(jsonValue)))
	if err != nil {
		return
	}
	req.Header.Set("Content-type", "application/json")
	if IP != "" {
		req.Header.Set("X-Forwarded-For", IP)
	}
	req.Close = true // this is required to prevent too many files open

	// Create HTTP Connection
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: timeout,
	}

	// Now hit to destionation endpoint
	res, err := client.Do(req)
	if err != nil {
		log.Errorf("Call Rest API Failed : %s",err.Error())			
		if res!=nil {
			buff := new(bytes.Buffer)
			buff.ReadFrom(res.Body)
			body = buff.String()
			httpStatus = res.StatusCode
			log.Errorf("Body : %s",body)	
		}else{
			body="Call Rest API Failed : "+err.Error()
		}
		return
	}
	defer res.Body.Close()

	buff := new(bytes.Buffer)
	buff.ReadFrom(res.Body)
	body = buff.String()
	httpStatus = res.StatusCode
	return
}
