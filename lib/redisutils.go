package lib

import (
	"github.com/go-redis/redis/v7"
	//redis "github.com/go-redis/redis/v7"
)

//RedisData is  use for
type RedisData struct {
	Key			string
	Value 	string	
}

//SearchKey get all key with format
func SearchKey(DBRedis *redis.Client,keyFormat string) (keyList []string){
	iter := DBRedis.Scan(0, keyFormat, 0).Iterator()
	for iter.Next() {
		keyList=append(keyList, iter.Val())
	}
	return keyList
}


//IsKeyFormatExist Check Key Format Exist atleast 1
func IsKeyFormatExist(DBRedis *redis.Client,keyFormat string) (exist bool){
	keys:=DBRedis.Keys(keyFormat).Val()
	if len(keys)>0 {
		return true
	}
	return false
}

//RedisDeleteKey delete all key with format
func RedisDeleteKey(DBRedis *redis.Client, keyFormat string){
	iter := DBRedis.Scan(0, keyFormat, 0).Iterator()
	for iter.Next() {
		DBRedis.Del(iter.Val())
	}
}

//RedisPipelineDeleteKey delete all key with format
func RedisPipelineDeleteKey(DBRedis *redis.Client, keyFormat string){
	iter := DBRedis.Scan(0, keyFormat, 0).Iterator()
	pipe := DBRedis.Pipeline()
	//m := map[string]*redis.StringCmd{}
	for iter.Next() {
		pipe.Del(iter.Val())
	}
    pipe.Exec()
}

//RedisGetDel Get and delete redis in 1 trx
func RedisGetDel(DBRedis *redis.Client, key string) (result string){
	DBRedis.Watch(func(tx *redis.Tx) error {
		result = tx.Get(key).Val()		
		_, err := tx.Pipelined(func (pipe redis.Pipeliner) error {
			pipe.Del(key)
			return nil
		})
		return err
	}, key)
	return result
}


//SearchKeyAndGetValue get all key with format
func SearchKeyAndGetValue(DBRedis *redis.Client,keyFormat string) (keyList []RedisData){
	iter := DBRedis.Scan(0, keyFormat, 0).Iterator()
	pipe := DBRedis.Pipeline()
	var key RedisData
	m := map[string]*redis.StringCmd{}
	for iter.Next() {		
		m[iter.Val()]= pipe.Get(iter.Val())
	}
	pipe.Exec()
	for keyString,result := range m {	
		key.Key=keyString
		key.Value=result.Val()
		keyList=append(keyList, key)
	}

	return keyList
}

//SearchKeyAndGetValueWithKeys get all key with format
func SearchKeyAndGetValueWithKeys(DBRedis *redis.Client, keyFormat string) (keyList []RedisData) {
	pipe := DBRedis.Pipeline()
	listKeys:=DBRedis.Keys(keyFormat).Val()
	m := map[string]*redis.StringCmd{}
	var key RedisData
	for _,localkey:= range listKeys {
		m[localkey]=pipe.Get(localkey)		
	}
	pipe.Exec()
	for keyString,result := range m {	
		key.Key=keyString
		key.Value=result.Val()
		keyList=append(keyList, key)
	}
	return keyList
}