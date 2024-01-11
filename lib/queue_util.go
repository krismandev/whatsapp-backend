package lib

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	redis "github.com/go-redis/redis/v7"
)

type JsonRedisItem struct {
	ID           string
	Json         string
	IsCompressed bool
}

func ConverToJsonList(stream []redis.XMessage) ([]JsonRedisItem, error) {
	result := make([]JsonRedisItem, 0)

	for _, message := range stream {
		compressedValue, ok := message.Values[compressedFieldName]
		var jsonValue []byte
		var err error

		if !ok {
			// non compressed field
			if len(message.Values) == 0 {
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

	if len(result) == 0 {
		return nil, nil
	}
	return result, nil

}

// need to validate this function
func getNextRedisStreamID(id string) string {
	var err error
	// id containing time and counter
	parts := strings.Split(id, "-")
	millisec, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return "0"
	}

	var seq int64 = 0
	if len(parts) > 1 {
		seq, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return strconv.FormatInt(millisec, 10)
		}
	}
	seq++
	return fmt.Sprintf("%d-%d", millisec, seq)

}
