package lib_db

import (
	"github.com/go-redis/redis"
	"github.com/spf13/cast"
	"time"
)

//获取redis
func RedisCreateRedisClient(redisConfig *RedisConfig) (*redis.Client, error) {
	redisCli := redis.NewClient(&redis.Options{
		Network:     "tcp", // 连接方式，默认使用tcp，可省略
		DialTimeout: time.Second * 30,
		Addr:        redisConfig.Host,
		Password:    redisConfig.Password,
		DB:          0,
		PoolSize:    cast.ToInt(redisConfig.PoolSize),
	})
	if err := redisCli.Ping().Err(); nil != err {
		return nil, err
	}
	return redisCli, nil
}
