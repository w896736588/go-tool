package xkf_tool

import (
	"database/sql"
	"gitee.com/Sxiaobai/gs/gsnsq"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"runtime"
)

var ConfigViper *viper.Viper
var EncryptMain *gstool.Encrypt //加密
var RedisRunList map[string]*redis.Client
var XkfDevMysql *sql.DB
var AppurlDevMysql *sql.DB
var Logger *gstool.LibLog
var ProducerMap map[string]*gsnsq.NsqStruct
var RootPath string

func InitConfig() {
	_, RootPath, _, _ = runtime.Caller(0)
	RootPath = gstool.DirUpNum(RootPath, 4)
	Logger = gstool.CreateLogger(RootPath, `logs/xkf_tool`)
	ConfigViper = viper.New()
	ConfigViper.AddConfigPath(RootPath + `/config`)
	ConfigViper.SetConfigName(`config`)
	ConfigViper.SetConfigType(`ini`)
	RedisRunList = make(map[string]*redis.Client)
	ProducerMap = make(map[string]*gsnsq.NsqStruct)
	if err := ConfigViper.ReadInConfig(); err != nil {
		panic(`读取配置失败 config/config.ini`)
	}
	EncryptMain = &gstool.Encrypt{
		Key: ConfigViper.GetString(`encrypt.key`),
		Iv:  ConfigViper.GetString(`encrypt.iv`),
	}
	Logger.Debugf(`初始化完成`)
}

// 拿到生产者
func GetProducer(host, port, topic string) *gsnsq.NsqStruct {
	checkKey := host + port + topic
	if producer, ok := ProducerMap[checkKey]; ok {
		return producer
	}
	producer := gsnsq.NsqInit(topic)
	err := producer.CreateProducer(gsnsq.NsqConfig{
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
