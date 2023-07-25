package base_module

import (
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gstool"
)

type RegisterStruct struct {
	Unikey          string                `json:"Unikey"`
	MysqlConfigList []*gsdb.MysqlConfig   `json:"MysqlConfigList"`
	RedisConfigList []*gsdb.RedisConfig   `json:"RedisConfigList"`
	ShellConfigList []*gstool.ShellConfig `json:"ShellConfigList"`
}

type LoginStruct struct {
	UserName string `json:"UserName"`
	Password string `json:"Password"`
}
