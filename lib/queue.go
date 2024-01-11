package lib

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fatih/structs"
	redis "github.com/go-redis/redis/v7"
	"github.com/pkg/errors"
)

const compressedFieldName = "compressed"

type RedisQueue struct {
	Db           *redis.Client
	Name         string
	MaxLen       int64
	GroupName    string
	ConsumerName string
	ReadTimeout  time.Duration
	LastID       string
	m            sync.Mutex
}

/*
	How queue will work :
		Each application instance should only have 1 consumer for the group
		Group will represent the function of reader
		Claiming other consumer will only happen manually by
		- specifying the consumername to be taken or
		- specifying the number of time this consumer was off
		1 queue instance in go will represent 1 consumer only

		how to :
		* Listing which consumer has pending message and how long :
			xinfo consumers testq testgroup
		* Listing 10 items is pending for the consumer :
			xpending test1 testgroup - + 10 consumername


	Readqueue (count, threadnumer)
	InitReader (Numberof thread) :
		when Numberofthread is less than number of consumer :  xinfo consumers testq testgroup
			1) 1) "name"
			2) "testgroup1"
			3) "pending"
			4) (integer) 3
			5) "idle"
			6) (integer) 694568
			2) 1) "name"
			2) "testgroup2"
			3) "pending"
			4) (integer) 1
			5) "idle"

		or using xpending :  xpending testq testgroup
		1) (integer) 4
		2) "1584075230601-0"
		3) "1584075230604-0"
		4) 1) 1) "testgroup1"
			2) "3"
		2) 1) "testgroup2"
			2) "1"


		Notes :
			xreadgroup when supplied with param = > then it will retrieve new message
				when supplied with param = 0 then any pending message will be retrieved

		-- Reread the message :  xreadgroup group testgroup testgroup1 count 1 streams testq 0

		when numberof thread is less then th thread, then claim the pending :

		claim : xclaim testq testgroup testgroup2 0 1584075230601-0 justid
		claim go get to be done first before reading the pending message
*/

func (q *RedisQueue) GetNow() time.Time {
	res := q.Db.Time()
	return res.Val()
}

func (q *RedisQueue) SetLastID(ID string) {
	q.m.Lock()
	defer q.m.Unlock()

	if q.LastID < ID || ID == "0" {
		q.LastID = ID
	}
}

func (q *RedisQueue) IsConsumerActive(maxIdle time.Duration) (bool, error) {
	listConsumers, err := RedisXInfoConsumers(q.Db, q.Name, q.GroupName)
	if err != nil {
		return false, err
	}

	for _, val := range listConsumers {
		if (val.Name == q.ConsumerName) && (val.Idle <= maxIdle) {
			return true, nil
		}
	}

	return false, nil
}

// return list of idle consumers
func (q *RedisQueue) IdleConsumers(minIdle time.Duration) ([]string, error) {
	listConsumers, err := RedisXInfoConsumers(q.Db, q.Name, q.GroupName)
	if err != nil {
		return nil, err
	}

	res := make([]string, len(listConsumers))

	resCount := 0
	for _, val := range listConsumers {
		if val.Idle >= minIdle && val.Pending > 0 && val.Name != q.ConsumerName {
			res[resCount] = val.Name
			resCount++
		}
		//fmt.Printf("Consumer %s idle for %s", val.Name, val.Idle)
	}
	return res[0:resCount], nil
}

func (q *RedisQueue) ClaimIdleConsumers(minIdle time.Duration) (int, error) {
	claimedCount := 0

	// get number of consumer for this consumer group
	idleConsumers, err := q.IdleConsumers(minIdle)
	if err != nil {
		return 0, err
	}

	for _, idleConsumer := range idleConsumers {
		param := redis.XPendingExtArgs{Stream: q.Name, Consumer: idleConsumer,
			Count: 1000000, Start: "-", End: "+", Group: q.GroupName}
		PendingList, err := q.Db.XPendingExt(&param).Result()
		if err != nil {
			return claimedCount, err
		}
		// claim for each ID and distribute to the existing ID

		IDList := make([]string, 0)
		for _, item := range PendingList {
			IDList = append(IDList, item.ID)
		}

		claimArg := redis.XClaimArgs{Consumer: q.ConsumerName,
			Group: q.GroupName, Messages: IDList, MinIdle: 0, Stream: q.Name}
		claimResult, err := q.Db.XClaimJustID(&claimArg).Result()
		if err != nil {
			return claimedCount, err
		}
		claimedCount = claimedCount + len(claimResult)

		// now remove the consumer
		// _, err = q.Db.XGroupDelConsumer(q.Name, q.GroupName, idleConsumer).Result()
		// if err != nil {
		// 	return claimedCount, err
		// }
	}
	return claimedCount, nil
}

func (q *RedisQueue) AckItem(streamid []string) (int64, error) {
	cmdResult := q.Db.XAck(q.Name, q.GroupName, streamid...)
	return cmdResult.Result()
}

func (q *RedisQueue) Delete(streamid []string) bool {
	// first ack, so that it will be removed from distribution
	q.Db.XAck(q.Name, q.GroupName, streamid...)
	// delete SubmissionQueue entry
	ret := q.Db.XDel(q.Name, streamid...)
	return ret.Val() > 0
}

// read queue store result as is
// func (q *RedisQueue) ReadQueueAsRedisMessages(ReadCounter int64, ReadNew bool) ([]redis.XMessage, error) {
// 	var LastID string
// 	if ReadNew {
// 		LastID = ">"
// 	} else {
// 		LastID = "0"
// 	}

// 	Timeout := q.ReadTimeout
// 	if Timeout <= 0 {
// 		Timeout = -1
// 	}

// 	stream, err := q.Db.XReadGroup(&redis.XReadGroupArgs{
// 		Group:    q.GroupName,
// 		Consumer: q.ConsumerName,
// 		Streams:  []string{q.Name, LastID},
// 		Count:    ReadCounter,
// 		Block:    Timeout,
// 	}).Result()

// 	if err != nil {
// 		if strings.Contains(err.Error(), "nil") {
// 			return nil, nil
// 		}
// 		return nil, errors.Wrapf(err, "Cannot read stream %s consumer %s", q.Name, q.ConsumerName)
// 	}
// 	if len(stream) == 0 {
// 		return nil, nil
// 	}
// 	if len(stream[0].Messages) == 0 {
// 		return nil, nil
// 	}

// 	// special treatment redis return message with nil values
// 	if stream[0].Messages[0].Values == nil {
// 		return nil, nil
// 	}

// 	return stream[0].Messages, nil
// }

func (q *RedisQueue) PutInQueue(data interface{}) (string, error) {
	qdata := structs.Map(data)

	sarg := redis.XAddArgs{
		Stream:       q.Name,
		MaxLenApprox: q.MaxLen,
		ID:           "*",
		Values:       qdata,
	}

	return q.Db.XAdd(&sarg).Result()
}

func (q *RedisQueue) CompressPutInQueue(data interface{}) (string, error) {

	qdata, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	redisData := make(map[string]interface{})
	redisData[compressedFieldName] = CompressBytes(qdata)

	sarg := redis.XAddArgs{
		Stream:       q.Name,
		MaxLenApprox: q.MaxLen,
		ID:           "*",
		Values:       redisData,
	}

	return q.Db.XAdd(&sarg).Result()
}

// read queue store result in map of string
func (q *RedisQueue) ReadQueueAsJson(ReadCounter int64, ReadNew bool) ([]JsonRedisItem, error) {

	// q.LastID is only used for reading old message where the counter is not managed by redis
	// it will be advanced by every read of new message, but it will be zero again after retrieving new item
	var LastID string
	if ReadNew {
		LastID = ">"
		q.SetLastID("0")
	} else {
		LastID = q.LastID
		if LastID == "" {
			LastID = "0"
		}
	}

	Timeout := q.ReadTimeout
	if Timeout <= 0 {
		Timeout = -1
	}

	stream, err := q.Db.XReadGroup(&redis.XReadGroupArgs{
		Group:    q.GroupName,
		Consumer: q.ConsumerName,
		Streams:  []string{q.Name, LastID},
		Count:    ReadCounter,
		Block:    Timeout,
	}).Result()

	if err != nil {
		// when stream is empty (newly created)
		if strings.Contains(err.Error(), "nil") {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "Cannot read stream %s consumer %s", q.Name, q.ConsumerName)
	}
	if len(stream) == 0 {
		return nil, errors.Errorf("Stream %s is empty", q.Name)
	}

	// log.Debugf("Read LastID : %s, groupname:%s, consumername:%s, stream:%+v", LastID, q.GroupName, q.ConsumerName, stream)
	MessageCount := len(stream[0].Messages)
	if MessageCount == 0 {
		return nil, nil
	}

	if !ReadNew {
		q.SetLastID(stream[0].Messages[MessageCount-1].ID)
	}
	return ConverToJsonList(stream[0].Messages)
}

func (q *RedisQueue) ConverToJsonList(stream []redis.XMessage) ([]JsonRedisItem, error) {
	result := make([]JsonRedisItem, 0)

	var DeletedItem []string

	for _, message := range stream {
		compressedValue, ok := message.Values[compressedFieldName]

		var jsonValue []byte
		var err error
		if !ok {
			// non compressed field
			if len(message.Values) == 0 {
				DeletedItem = append(DeletedItem, message.ID)
				continue // skip null message coming from deleted stream item
			}
			jsonValue, err = json.Marshal(message.Values)
			if err != nil {
				return nil, errors.Wrap(err, "Error convert redis to Json")
			}
		} else {
			uncompressedValue, err := Decompress(compressedValue.(string))
			if err != nil {
				return nil, errors.Wrap(err, "Error decompressing redis")
			}
			jsonValue = []byte(uncompressedValue)
		}
		result = append(result, JsonRedisItem{ID: message.ID, Json: string(jsonValue), IsCompressed: ok})
	}

	q.Delete(DeletedItem)
	return result, nil
}

func (q *RedisQueue) ReadQueue(ReadCounter int64, ReadNew bool) ([]map[string]string, error) {
	JsonList, err := q.ReadQueueAsJson(ReadCounter, ReadNew)
	if err != nil {
		return nil, err
	}

	if JsonList == nil {
		return nil, nil
	}

	result := make([]map[string]string, 0)

	for _, item := range JsonList {
		var ExtractedVal map[string]interface{}
		err := json.Unmarshal([]byte(item.Json), &ExtractedVal)
		if err != nil {
			return nil, errors.Wrap(err, "Error unmarshal json "+item.Json)
		}

		mapString := make(map[string]string)
		mapString["ID"] = item.ID

		for key, value := range ExtractedVal {
			strValue := fmt.Sprintf("%v", value)
			mapString[key] = strValue
		}
		result = append(result, mapString)
	}

	return result, nil
}

//PutInOtherQueue will put sms in redis queue with specific name
func CompressPutInQueue(Db *redis.Client, MaxLen int64, data interface{}, Qname string) (string, error) {
	qdata, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	//fmt.Println("Json:", string(qdata))

	redisData := make(map[string]interface{})
	redisData[compressedFieldName] = CompressBytes(qdata)

	sarg := redis.XAddArgs{
		Stream:       Qname,
		MaxLenApprox: MaxLen,
		ID:           "*",
		Values:       redisData,
	}

	return Db.XAdd(&sarg).Result()
}
