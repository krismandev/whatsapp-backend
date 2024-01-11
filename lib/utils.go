package lib

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/sha3"
)

func init() {
	// seeding time for default pseudo random generator
	rand.Seed(time.Now().UnixNano())
}

func isContain(text string, wordlist []string) bool {
	for _, v := range wordlist {
		if strings.Contains(text, v) {
			return true
		}
	}
	return false
}

//ToTitleCase convert string to CamelCase
func ToTitleCase(Str string) string {
	Str = strings.Trim(Str, " ")
	result := ""
	capital := true
	for _, charr := range Str {
		if charr >= 'A' && charr <= 'Z' {
			result += string(charr)
		}
		if charr >= '0' && charr <= '9' {
			result += string(charr)
		}
		if charr >= 'a' && charr <= 'z' {
			if capital {
				result += strings.ToUpper(string(charr))
			} else {
				result += string(charr)
			}
		}
		if charr == '_' || charr == ' ' || charr == '-' {
			result += string(charr)
			capital = true
		} else {
			capital = false
		}
	}
	return result
}

//GetTransactionid Generate Transaction ID
func GetTransactionid(useip bool) string {
	var s string

	tmNow := time.Now()
	// second as 32 bit (8 hex digit)
	sec := tmNow.Unix()
	// microsecond as 5 hex digit
	usec := (tmNow.UnixNano() / 1000) - (sec * 1000000)
	// random numbers as 4 hex digit
	rnd1 := rand.Intn(0xFFFF)
	rnd2 := rand.Intn(0xFFFF)
	if useip {
		ip := GetIP()
		s = fmt.Sprintf("%08x%08x%05x%04x%04x", ip, sec, usec, rnd1, rnd2)
	} else {
		s = fmt.Sprintf("%08x%05x%04x%04x", sec, usec, rnd1, rnd2)
	}
	return s
}

//GenerateCode Generate Random Code
func GenerateCode() string {
	var s string
	tmNow := time.Now()
	// second as 32 bit (8 hex digit)
	sec := tmNow.Unix()
	// microsecond as 5 hex digit
	usec := (tmNow.UnixNano() / 1000) - (sec * 1000000)
	// random numbers as 4 hex digit
	rnd1 := rand.Intn(0xFFF)
	rnd2 := rand.Intn(0xFFFF)
	s = fmt.Sprintf("%03x%08x%05x%04x", rnd1, sec, usec, rnd2)
	return s
}

//SHA256 Hash SHA256
func SHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

//MD5 Hash MD5
func MD5(data string) string {
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

//CRC32 Hash CRC32
func CRC32(data string) string {
	crc32q := crc32.MakeTable(0xEDB88320)
	//crc32q := crc32.MakeTable(0x04C11DB7)
	result := crc32.Checksum([]byte(data), crc32q)
	return fmt.Sprintf("%08X", result)
}

//SHA3256 Hash SHA3256
func SHA3256(data string) string {
	hash := sha3.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

//FindSlice Find string on slice
func FindSlice(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

//DeleteSlice Delete String slice
func DeleteSlice(slice []string, index int) []string {
	slice[index] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}

//DeleteSliceInt Delete String slice
func DeleteSliceInt(slice []int, index int) []int {
	slice[index] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}

//DeleteIntSlice Delete Int slice
func DeleteIntSlice(slice []int, index int) []int {
	//slice[index]=slice[len(slice)-1]
	slice[len(slice)-1], slice[index] = slice[index], slice[len(slice)-1]
	return slice[:len(slice)-1]
}

//StructToMap to Convert Struct to map[string]interface
func StructToMap(item interface{}) (map[string]interface{}, error) {
	var res map[string]interface{}

	inrec, errMarshall := json.Marshal(&item)
	sString := string(inrec[:])
	sString = strings.TrimPrefix(sString, "[")
	sString = strings.TrimSuffix(sString, "]")
	if errMarshall != nil {
		return nil, errMarshall
	}
	errUnMarshall := json.Unmarshal([]byte(sString), &res)
	if errUnMarshall != nil {
		return nil, errUnMarshall
	}

	return res, nil
}

//MapToStruct to Convert map[string]interface to Struct
func MapToStruct(value map[string]interface{}, item interface{}) error {
	data, errMarshall := json.Marshal(value)
	if errMarshall != nil {
		return errMarshall
	}
	err := json.Unmarshal(data, &item)
	return err
}

//StringMapToStruct to Convert map[string]interface to Struct
func StringMapToStruct(value map[string]string, item interface{}) error {
	data, errMarshall := json.Marshal(value)
	if errMarshall != nil {
		return errMarshall
	}
	err := json.Unmarshal(data, &item)
	return err
}

//ExtractKeyFromMap to extract map[string]interface key list
func ExtractKeyFromMap(values interface{}) []string {
	v := reflect.ValueOf(values)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	mapName := []string{}
	if v.Kind() == reflect.Map {
		for _, keyv := range v.MapKeys() {
			mapName = append(mapName, keyv.String())
		}
		return mapName
	}
	return nil
}

//ExtractStructField to extract Struct FieldName list with struct tag
func ExtractStructField(tag string, values interface{}) []string {
	v := reflect.ValueOf(values)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	var field string
	fields := []string{}

	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			field = v.Type().Field(i).Tag.Get(tag)
			if field != "" {
				fields = append(fields, field)
			}
		}
		return fields
	}
	return nil
}

//MapToJSONstr to Convert map[string]interface to Struct
func MapToJSONstr(value map[string]interface{}) (jsonResult string, errMarshall error) {
	data, errMarshall := json.Marshal(value)
	if errMarshall != nil {
		return "", errMarshall
	}
	return string(data), nil
}

//JSONToMap to Convert map[string]interface to Struct
func JSONToMap(jsonStr string) (mapResult map[string]interface{}, errMarshall error) {
	errUnMarshall := json.Unmarshal([]byte(jsonStr), &mapResult)
	if errUnMarshall != nil {
		return nil, errUnMarshall
	}
	return mapResult, nil
}

//CountArrayString count array group by string
func CountArrayString(source []string) (result map[string]int) {
	result = make(map[string]int)
	sort.Strings(source)
	for _, key := range source {
		result[key]++
	}
	return result
}

//SecondOfDay get how many second to day
func SecondOfDay(t time.Time) int {
	return 60*60*t.Hour() + 60*t.Minute() + t.Second()
}

//SecondtoEndOfDay get how many second left to day
func SecondtoEndOfDay(t time.Time) int {
	return 86400 - SecondOfDay(t)
}

// BeginningOfMonth return the begin of the month of t
func BeginningOfMonth(t time.Time) time.Time {
	year, month, _ := t.Date()
	fmt.Print(time.Date(year, month, 1, 0, 0, 0, 0, t.Location()).Format("2006-01-02"))
	return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth return the end of the month of t
func EndOfMonth(t time.Time) time.Time {
	fmt.Print(BeginningOfMonth(t).AddDate(0, 1, -1).Format("2006-01-02"))
	return BeginningOfMonth(t).AddDate(0, 1, -1)
}

//DaysToEndOfMonth Days Count to end of month
func DaysToEndOfMonth(t time.Time) int {
	endofMontth := EndOfMonth(t)
	fmt.Print(endofMontth.Format("2006-01-02"))
	days := endofMontth.Sub(t).Hours() / 24
	days = math.Ceil(days) //round up the day
	return int(days) + 1   //+1 because we count from t
}

//NormalizeMsisdn Normalize msisdn
func NormalizeMsisdn(msisdn string) string {
	re := regexp.MustCompile("[0-9]+")
	arrAllNumber := re.FindAllString(msisdn, -1)
	msisdn = strings.Join(arrAllNumber, "")

	if len(msisdn) < 3 {
		return ""
	}
	if msisdn[0:1] == "0" {
		return "62" + msisdn[1:]
	} else if msisdn[0:2] == "62" {
		return msisdn
	} else if msisdn[0:3] == "+62" {
		return msisdn[1:]
	}
	return ""
}

//GetAlphaNumericOnly Remove Special Char to prevent attacker
func GetAlphaNumericOnly(text string) string {
	text = strings.TrimSpace(text)
	re := regexp.MustCompile("[a-z,A-Z,0-9]+")
	arrAllNumber := re.FindAllString(text, -1)
	text = strings.Join(arrAllNumber, "")
	return text
}

// RegisterExitSignalHandler used in main.go
func RegisterExitSignalHandler(handler func()) {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-exit
		handler()
	}()
}
