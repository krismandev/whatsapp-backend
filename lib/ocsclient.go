package lib

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/sony/gobreaker"
)

const (
	ErrOcsBalanceNotEnough          = 1
	ErrOcsBalanceExpire             = 2
	ErrOcsTariffNotFound            = 3
	ErrOcsAccountNotFound           = 4
	ErrOcsPriceNotFound             = 5 // not yet in testunit
	ErrOcsBalanceNotFound           = 6 // not yet in testunit
	ErrOcsDuplicateTransaction      = 7 // not yet in testunit
	ErrOcsParentAccountNotFound     = 8 // not yet in testunit
	ErrOcsParentAccountFailCharging = 9
	ErrOcsMaxAccountDepth           = 10 // not yet in testunit
	ErrOcsInvalidTariffPlan         = 11
	ErrOcsCodeUnknown               = 99
)

var OcsErrMessage = map[int]string{
	ErrOcsBalanceNotEnough:          "insufficient Balance",
	ErrOcsBalanceExpire:             "balance expire",
	ErrOcsTariffNotFound:            "tariff not found",
	ErrOcsAccountNotFound:           "account not found",
	ErrOcsPriceNotFound:             "price not found",
	ErrOcsBalanceNotFound:           "balance record not found",
	ErrOcsCodeUnknown:               "unknown ocs error",
	ErrOcsDuplicateTransaction:      "duplicate transaction",
	ErrOcsParentAccountNotFound:     "parent account not found",
	ErrOcsParentAccountFailCharging: "parent account fail charging",
	ErrOcsInvalidTariffPlan:         "invalid tariffplan",
	ErrOcsMaxAccountDepth:           "account structure max depth exceeded",
}

const (
	TransactTopup        = iota + 1 // 1
	TransactCharge                  // 2
	TransactRefund                  // 3
	TransactDirectCharge            // 4
	TransactAdjustment
)

type HttpError struct {
	httpresultcode int
}

func (e *HttpError) Error() string {
	return "HTTP error " + strconv.Itoa(e.httpresultcode)
}

func CallRestAPIOCS(param interface{}, url string) (body string, err error) {
	jsonValue, _ := json.Marshal(param)

	// create request structure
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonValue)))
	if err != nil {
		return
	}
	req.Header.Set("Content-type", "application/json")
	req.Close = true // this is required to prevent too many files open

	// Create HTTP Connection
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: time.Duration(10) * time.Second,
	}

	// Now hit to destionation endpoint
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	buff := new(bytes.Buffer)
	buff.ReadFrom(res.Body)
	body = buff.String()

	if res.StatusCode != 200 {
		err = &HttpError{res.StatusCode}
	}

	return
}

type CheckBalanceRequest struct {
	AccountID int
}

type CheckBalanceResponse struct {
	ErrorCode int
	Balance   map[string]int
	ErrorDesc string
}

type ChargeRequest struct {
	AccountID int
	Product   string
	Params    []string
	UnitUsed  int
	RefID     string
	AppID     string
}

type ChargeResultWithChecksum struct {
	Hash string           `shortjson:"Hash"`
	List ChargeResultList `shortjson:"List"`
}

type ChargeResultList []ChargeResult

type BalTransResult struct {
	AccountID     int    `shortjson:"AccID"`
	ErrorCode     int    `shortjson:"Code"`
	Amount        int    `shortjson:"Amt" json:"AmountUsed" `
	LastBalance   int    `shortjson:"Bal"`
	BalanceType   string `shortjson:"Type"`
	TransactionID string `shortjson:"TrxID"`
	ErrorDesc     string `shortjson:"-"`
}

type ChargeResult BalTransResult

type DirectChargeRequest struct {
	AccountID   int
	BalanceType string
	AmountUsed  int
	RefID       string
	AppID       string
}
type DirectChargeResult BalTransResult

type RefundRequest struct {
	AccountID    int
	AmountRefund int
	BalanceType  string
	RefID        string
	AppID        string
	ChargeList   ChargeResultWithChecksum
	LastBalance  int
}

type RefundResult BalTransResult
type RefundResultList []RefundResult

type TopupRequest struct {
	AccountID   int
	AmountTopup int
	Amount      int
	BalanceType string
	RefID       string
	AppID       string
	LastBalance int
}

type TopupResult BalTransResult

var AppID string = "DEFAULT"

var ocsCb *gobreaker.CircuitBreaker

var chargeUrl string
var ocsTimeout = 10 * time.Second

func SetChargeUrl(url string) {
	chargeUrl = url
}

func SetOcsTimeout(timeout time.Duration) {
	ocsTimeout = timeout
}

func Charge(req ChargeRequest) (result *ChargeResult, err error) {
	req.AppID = AppID

	httpResult, err := ocsCb.Execute(func() (interface{}, error) {
		apiResult, err := CallRestAPIOCS(req, chargeUrl+"/charge")
		if err != nil {
			return "", err
		}
		return apiResult, nil
	})

	if err != nil {
		return nil, err
	}

	result = &ChargeResult{}
	err = json.Unmarshal([]byte(httpResult.(string)), &result)
	return
}

func CascadeCharge(req ChargeRequest) (result ChargeResultWithChecksum, jsonResult string, err error) {
	req.AppID = AppID

	json := jsoniter.Config{TagKey: "shortjson"}.Froze()

	httpResult, err := ocsCb.Execute(func() (interface{}, error) {
		apiResult, err := CallRestAPIOCS(req, chargeUrl+"/cascadecharge")
		if err != nil {
			return "", err
		}
		return apiResult, nil
	})

	if err != nil {
		return ChargeResultWithChecksum{}, jsonResult, err
	}

	jsonResult = httpResult.(string)
	err = json.Unmarshal([]byte(jsonResult), &result)
	return
}

func DirectCharge(req DirectChargeRequest) (result *DirectChargeResult, err error) {
	req.AppID = AppID

	httpResult, err := ocsCb.Execute(func() (interface{}, error) {
		apiResult, err := CallRestAPIOCS(req, chargeUrl+"/directcharge")
		if err != nil {
			return "", err
		}
		return apiResult, nil
	})

	if err != nil {
		return nil, err
	}

	result = &DirectChargeResult{}
	err = json.Unmarshal([]byte(httpResult.(string)), &result)
	return
}

// func CallRestAPIShortJson(req, chargeUrl+"/cascaderefund", ocsTimeout)
func CallRestAPIShortJson(param interface{}, url, IP string, timeout time.Duration) (body string, err error) {

	json := jsoniter.Config{TagKey: "shortjson"}.Froze()
	jsonValue, _ := json.Marshal(param)

	// create request structure
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonValue)))
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

	// Now hit to destinnation endpoint
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	buff := new(bytes.Buffer)
	buff.ReadFrom(res.Body)
	body = buff.String()

	if res.StatusCode != 200 {
		err = &HttpError{res.StatusCode}
	}

	return
}

func CascadeRefund(req RefundRequest) (result RefundResultList, err error) {
	req.AppID = AppID
	json := jsoniter.Config{TagKey: "shortjson"}.Froze()

	/* Complication because RefundRequest contains chargelist that is in shortjson instead of normal json*/

	httpResult, err := ocsCb.Execute(func() (interface{}, error) {
		apiResult, err := CallRestAPIShortJson(req, chargeUrl+"/cascaderefund", "", ocsTimeout)
		if err != nil {
			return "", err
		}
		return apiResult, nil
	})

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(httpResult.(string)), &result)
	return
}

func Refund(req RefundRequest) (result *RefundResult, err error) {
	req.AppID = AppID

	httpResult, err := ocsCb.Execute(func() (interface{}, error) {
		apiResult, err := CallRestAPIOCS(req, chargeUrl+"/refund")
		if err != nil {
			return "", err
		}
		return apiResult, nil
	})

	if err != nil {
		return nil, err
	}

	result = &RefundResult{}
	err = json.Unmarshal([]byte(httpResult.(string)), result)
	return
}

func CheckBalance(req CheckBalanceRequest) (result *CheckBalanceResponse, err error) {
	httpResult, err := ocsCb.Execute(func() (interface{}, error) {
		apiResult, err := CallRestAPIOCS(req, chargeUrl+"/checkbalance")
		if err != nil {
			return "", err
		}
		return apiResult, nil
	})

	result = &CheckBalanceResponse{}
	if err != nil {
		// generate blank map balance for error response
		return result, err
	}

	err = json.Unmarshal([]byte(httpResult.(string)), &result)
	return
}

func Topup(req TopupRequest) (result *TopupResult, err error) {
	req.AppID = AppID

	httpResult, err := ocsCb.Execute(func() (interface{}, error) {
		apiResult, err := CallRestAPIOCS(req, chargeUrl+"/topup")
		if err != nil {
			return "", err
		}
		return apiResult, nil
	})

	if err != nil {
		return nil, err
	}

	result = &TopupResult{}
	err = json.Unmarshal([]byte(httpResult.(string)), &result)
	return
}

func Adjust(req TopupRequest) (result *TopupResult, err error) {
	req.AppID = AppID

	httpResult, err := ocsCb.Execute(func() (interface{}, error) {
		apiResult, err := CallRestAPIOCS(req, chargeUrl+"/adjust")
		if err != nil {
			return "", err
		}
		return apiResult, nil
	})

	if err != nil {
		return nil, err
	}

	result = &TopupResult{}
	err = json.Unmarshal([]byte(httpResult.(string)), &result)
	return
}

func init() {
	ocsCb = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name: "ocsclient-breaker",
	})
}
