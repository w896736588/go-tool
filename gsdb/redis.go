package gsdb

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/w896736588/go-tool/gsssh"
)

const RedisKeyString = `string`
const RedisKeyHash = `hash`
const RedisKeyList = `list`
const RedisKeySet = `set`
const RedisKeyZSet = `zset`

type RedisConfig struct {
	Name              string `json:"name"`
	Host              string `json:"host"`
	Port              int64  `json:"port"`
	Password          string `json:"password"`
	Default           int    `json:"default"`
	Username          string `json:"userName"`
	MaxOpenConns      int    `json:"maxOpenConns"`      //最大连接池连接数量
	MaxIdleConns      int    `json:"maxIdleConns"`      //最大可空闲连接数
	MaxLifetimeSecond int64  `json:"maxLifetimeSecond"` //连接的最大生存时间
}

type GsRedis struct {
	RedisConfig *RedisConfig
	Client      *redis.Client
	SshBridge   *gsssh.SshBridge
}

func NewRedis(config *RedisConfig) *GsRedis {
	return &GsRedis{
		RedisConfig: config,
	}
}

// CreateConn create a connection
func (h *GsRedis) CreateConn() error {
	if h.SshBridge != nil {
		return h.createConnSshPasswordAuth()
	} else {
		return h.createConnDirect()
	}
}

// ssh bridge password auth
func (h *GsRedis) createConnSshPasswordAuth() error {
	targetHostPort := fmt.Sprintf(`%s:%d`, h.RedisConfig.Host, h.RedisConfig.Port)
	localHostPort, runError := h.SshBridge.RunBridge(targetHostPort)
	if runError != nil {
		return runError
	}
	options := &redis.Options{
		Network:         "tcp", // 连接方式，默认使用tcp，可省略
		Addr:            localHostPort,
		Password:        h.RedisConfig.Password,
		DB:              h.RedisConfig.Default,
		PoolSize:        h.RedisConfig.MaxOpenConns,
		MaxIdleConns:    h.RedisConfig.MaxIdleConns,
		ConnMaxLifetime: time.Second * time.Duration(h.RedisConfig.MaxLifetimeSecond),
		Username:        h.RedisConfig.Username,
	}
	return h.dbOpen(options)
}

// direct
func (h *GsRedis) createConnDirect() error {
	options := &redis.Options{
		Network:         "tcp", // 连接方式，默认使用tcp，可省略
		Addr:            fmt.Sprintf(`%s:%d`, h.RedisConfig.Host, h.RedisConfig.Port),
		Password:        h.RedisConfig.Password,
		DB:              h.RedisConfig.Default,
		PoolSize:        h.RedisConfig.MaxOpenConns,
		MaxIdleConns:    h.RedisConfig.MaxIdleConns,
		ConnMaxLifetime: time.Second * time.Duration(h.RedisConfig.MaxLifetimeSecond),
		Username:        h.RedisConfig.Username,
	}
	return h.dbOpen(options)
}

// 创建连接
func (h *GsRedis) dbOpen(options *redis.Options) error {
	maxOpenConn := h.RedisConfig.MaxOpenConns
	MaxIdleConns := h.RedisConfig.MaxIdleConns
	maxLifeTimeSecond := h.RedisConfig.MaxLifetimeSecond
	if maxOpenConn == 0 {
		maxOpenConn = 1
	}
	if MaxIdleConns == 0 {
		MaxIdleConns = 1
	}
	if maxLifeTimeSecond == 0 || maxLifeTimeSecond < 30 {
		maxLifeTimeSecond = 60
	}
	options.PoolSize = maxOpenConn
	options.MaxIdleConns = MaxIdleConns
	options.ConnMaxLifetime = time.Second * time.Duration(maxLifeTimeSecond)
	h.Client = redis.NewClient(options)
	if pingErr := h.Client.Ping(context.Background()).Err(); nil != pingErr {
		return pingErr
	}
	return nil
}
