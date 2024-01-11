package lib

import (
	"encoding/base64"
	"encoding/binary"
	"net"
	"net/http"
	"os"
	"strings"
	//"time"
	"io/ioutil"
	"bytes"
	//log "github.com/sirupsen/logrus"
)

//GetRemoteIPAddress Get Remote IP
func GetRemoteIPAddress(r *http.Request) string {
	IP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ""
	}
	return IP
}

// SplitURI return 
func SplitURI(uri string) string {
	splittedURL := strings.Split(uri, "/")
	var newURL string
	for i, v := range splittedURL {
		if i <= 3 {
			newURL += v + "/"
		}
	}
	return newURL
}


// GetIP return IP as 32bit integer
func GetIP() uint32 {
	ifaces, err := net.Interfaces()
	if err != nil {
		return 0
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return 0
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}

			ipint32 := binary.BigEndian.Uint32(ip)
			return ipint32
		}
	}
	return 0
}
//GetHostName return Host Name
func GetHostName() string {
	name,err:=os.Hostname()
	if err != nil {
		return ""
	}
	return name

}


//DecodeBasicAuth to Decode basic auth
func DecodeBasicAuth(auth string) (username, password string, ok bool) {
	authKey := strings.SplitN(auth, " ", 2)
	if len(authKey) != 2 || authKey[0] != "Basic" {
		ok = false
		return "", "", false
	}

	payload, _ := base64.StdEncoding.DecodeString(authKey[1])
	pair := strings.SplitN(string(payload), ":", 2)

	if len(pair) != 2 {
		ok = false
		return "", "", false
	}
	return pair[0], pair[1], true
}

//EncodeBasicAuth to Encode basic auth
func EncodeBasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

//CloneHTTPBody to convert request body to
func CloneHTTPBody(r *http.Request) string {
	var bodystring string

	bodybytes, err := ioutil.ReadAll(r.Body)
	if err == nil {
		r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodybytes))
		bodystring = string(bodybytes)

	}
	return bodystring
}

// //CallRestAPI to call rest api
// func CallRestAPI (url,method,request string, timeout time.Duration) (responseBody string, httpStatus int){	
	
// 	client := &http.Client {Timeout: timeout}
// 	payload:=strings.NewReader(request)
// 	req, errNewRequest := http.NewRequest(method, url, payload)
// 	if errNewRequest != nil {
// 		return "",http.StatusInternalServerError
// 	}
// 	req.Header.Add("Content-Type", "application/json")
// 	req.Header.Add("Connection", "close")

// 	var body []byte =nil
// 	var errReadResponse error

// 	res, errCallRestAPI := client.Do(req)

// 	if errCallRestAPI != nil {
// 		if res!=nil {
// 			body, errReadResponse = ioutil.ReadAll(res.Body)
// 			if errReadResponse != nil {
// 				log.Errorf("Error Read Response Body : %s",errReadResponse.Error())
// 				log.Errorf("Body : %+v",res.Body)
// 				return "",http.StatusInternalServerError
// 			}
// 			log.Errorf("Body : %+v",res.Body)

// 		}
// 		log.Errorf("Call Rest API Failed : %s",errCallRestAPI.Error())			
// 		return "", http.StatusInternalServerError
// 	}
// 	defer res.Body.Close()
// 	body, errReadResponse = ioutil.ReadAll(res.Body)
// 	if errReadResponse != nil {
// 		return "",http.StatusInternalServerError
// 	}
// 	return string(body),http.StatusOK	

// }