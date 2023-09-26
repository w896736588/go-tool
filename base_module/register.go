package base_module

import "gitee.com/Sxiaobai/gs/gstool"

func Register(global *Global, register *RegisterStruct) {
	if len(register.RedisConfigList) > 0 {
		for _, value := range register.RedisConfigList {
			gstool.FmtPrintlnLog(`value %#v`, value)
			global.RedisSetConfig(value)
		}
	}
	if len(register.MysqlConfigList) > 0 {
		for _, value := range register.MysqlConfigList {
			global.MysqlSetConfig(value)
		}
	}
	if len(register.ShellConfigList) > 0 {
		for _, value := range register.ShellConfigList {
			global.ShellSetConfig(value)
		}
	}
	if register.EncryptIv != `` && register.EncryptKey != `` {
		global.SetEncrypt(register.EncryptKey, register.EncryptIv)
	}
}
