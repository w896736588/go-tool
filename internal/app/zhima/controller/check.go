package controller

import (
	"errors"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
	"xkf_tool/base_module"
)

func GetGlobal(reqMap map[string]*gstool.GsCons) (*base_module.Global, error) {
	if reqMap[`Unikey`] == nil || reqMap[`Unikey`].ToStr() == `` {
		return nil, errors.New(`缺少Unikey参数`)
	}
	global := base_module.GetGlobal(cast.ToString(reqMap[`Unikey`]))
	if global == nil {
		return nil, errors.New(`请登录`)
	}
	return global, nil
}
