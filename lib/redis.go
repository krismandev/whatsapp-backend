package lib

import (
	"strconv"
	"strings"
	"time"

	redis "github.com/go-redis/redis/v7"
)

func RedisCreateStreamReader(db *redis.Client, stream, groupreader string) error {
	rdbresult := db.Conn().XGroupCreate(stream, groupreader, "0")
	if rdbresult.Err() != nil {
		if strings.Contains(rdbresult.Err().Error(), "MKSTREAM") {
			rdbresult = db.Conn().XGroupCreateMkStream(stream, groupreader, "0")
		} else {
			return rdbresult.Err()
		}
	}
	return nil
}

type XInfoConsumersResult struct {
	Name    string
	Pending int64
	Idle    time.Duration
}

func RedisIDtoInt64(ID string) uint64 {
	// look for -
	s := strings.Split(ID, "-")
	if len(s) == 0 {
		return 0
	}

	id, err := strconv.ParseUint(s[0], 10, 64)
	if err != nil {
		return 0
	}

	return id
}

func RedisXInfoConsumers(r *redis.Client, stream, group string) ([]XInfoConsumersResult, error) {
	rescmd := r.Do("xinfo", "consumers", stream, group)

	if rescmd.Err() != nil {
		return nil, rescmd.Err()
	}

	arrIntf := func(i interface{}, err error) interface{} { return i }(rescmd.Result()).([]interface{})

	res := make([]XInfoConsumersResult, len(arrIntf))
	for idx, val := range arrIntf {

		arrInfo := val.([]interface{})
		// convert array of interface to struct
		var mapkey string
		for idx2, val2 := range arrInfo {
			if idx2%2 == 0 {
				mapkey = val2.(string)
			} else {
				switch mapkey {
				case "name":
					res[idx].Name = val2.(string)
				case "pending":
					res[idx].Pending = val2.(int64)
				case "idle":
					idle, _ := val2.(int64)
					res[idx].Idle = time.Millisecond * time.Duration(idle)
				}
			}
		}
	}
	return res, nil
}
