package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

const (
	// SafeSessionExpireMinutesDefault 默认会话有效期（分钟）
	SafeSessionExpireMinutesDefault = 120
	// SafeSessionExpireMinutesMin 最小会话有效期（分钟），0 表示永不过期
	SafeSessionExpireMinutesMin = 0
	// SafePasswordVersionSalt 密码版本计算盐值
	SafePasswordVersionSalt = "dtool_safe_login_v1"
)

// SafeConfig Safe 配置结构体
type SafeConfig struct {
	Password             string
	SessionExpireMinutes int
	PasswordVersion      string
}

// BuildSafePasswordVersion 根据密码生成版本标识
// 使用 md5(password + AppName + fixed_salt) 计算
func BuildSafePasswordVersion(password string, appName string) string {
	if appName == "" {
		appName = "dtool"
	}

	data := fmt.Sprintf("%s%s%s", password, appName, SafePasswordVersionSalt)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// NormalizeSafeSessionExpireMinutes 规范化会话过期时间
// 0 表示永不过期，负数按 0 处理，正数按原值返回
func NormalizeSafeSessionExpireMinutes(minutes int) int {
	if minutes < 0 {
		return 0
	}
	return minutes
}
