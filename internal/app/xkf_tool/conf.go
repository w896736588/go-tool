package xkf_tool

import (
	"database/sql"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"redis_manager/internal/pkg/lib_db"
	"redis_manager/internal/pkg/lib_tool"
)

var RedisList []lib_db.RedisConfig
var ConfigViper *viper.Viper
var EncryptMain *lib_tool.Encrypt //加密
var RedisRunList map[string]*redis.Client
var XkfDevMysql *sql.DB
var AppurlDevMysql *sql.DB
var Logger *lib_tool.LibLog
var ProducerMap map[string]*lib_tool.NsqStruct

func InitConfig() {
	Logger = lib_tool.CreateLogger(lib_tool.DirUpNum(4), `logs/xkf_tool`)
	ConfigViper = viper.New()
	ConfigViper.AddConfigPath(`config`)
	ConfigViper.SetConfigName(`config`)
	ConfigViper.SetConfigType(`ini`)
	RedisRunList = make(map[string]*redis.Client)
	ProducerMap = make(map[string]*lib_tool.NsqStruct)
	if err := ConfigViper.ReadInConfig(); err != nil {
		panic(`读取配置失败 config/config.ini`)
	}
	EncryptMain = &lib_tool.Encrypt{
		Key: ConfigViper.GetString(`encrypt.key`),
		Iv:  ConfigViper.GetString(`encrypt.iv`),
	}
	Logger.Debugf(`初始化完成`)
}

//拿到生产者
func GetProducer(host, port, topic string) *lib_tool.NsqStruct {
	checkKey := host + port + topic
	if producer, ok := ProducerMap[checkKey]; ok {
		return producer
	}
	producer := lib_tool.NsqInit(topic)
	err := producer.CreateProducer(lib_tool.NsqConfig{
		Host: host,
		Port: port,
	})
	if err != nil {
		Logger.Errorf(`GetProducer ` + err.Error())
		return nil
	}
	ProducerMap[checkKey] = producer
	return producer
}
