package xkf_tool

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"redis_manager/internal/pkg/lib_tool"
	"time"
)

// QueryWechatAppid 拿到应用的AppId https://www.cnblogs.com/jasonminghao/p/12386580.html
// @auth frog
// @date 2023-01-17 12:25:31
func QueryWechatAppid(wechatAppId string) TblWechatapp {
	var appInfo = TblWechatapp{}
	db := XkfDevMysql
	err := db.QueryRow(`select app_id,app_type,user_id from tbl_wechatapp where (app_id = ? or _id = ?) and user_id > 0`, wechatAppId, cast.ToInt(wechatAppId)).Scan(&appInfo.Appid, &appInfo.AppType, &appInfo.UserId)
	if err != nil {
		Logger.Errorf(`执行sql出错 %s`, err.Error())
		return appInfo
	}
	return appInfo
}

// GetAdminUserId 拿到用户信息
// @auth frog
// @date 2023-01-17 15:49:51
func GetAdminUserId(account string) TblUser {
	db := AppurlDevMysql
	var userInfo = TblUser{}

	err := db.QueryRow(`select _id,user_name from tbl_user where (user_name = ? or _id = ?)`, account, cast.ToInt(account)).Scan(&userInfo.Id, &userInfo.Username)
	if err != nil {
		log.Errorf(`执行sql出错 %s`, err.Error())
		return userInfo
	}
	return userInfo
}

// UpdateVip 变更vip版本
// @auth frog
// @date 2023-01-17 16:03:11
func UpdateVip(adminUserId, expiredDay, systemType, vipLevel string) string {
	db := AppurlDevMysql
	vipTable := `tbl_kefu_vip`
	if systemType == `1` { //客服系统
		vipTable = `tbl_kefu_vip`
	} else {
		vipTable = `tbl_official_account_vip`
	}
	//时间
	t := time.Now().Unix()
	t += 86400 * cast.ToInt64(expiredDay)
	var ret sql.Result
	expiredTime := time.Unix(t, 0).Format(`2006-01-02 15:04:05`)
	sqlStr := `update ` + vipTable + ` set expired_time = ? , vip_type = ? where user_id =?`
	log.Debugf(`更新vip表 %s %s %s %s %s`, vipTable, sqlStr, vipLevel, expiredTime, adminUserId)
	ret, err := db.Exec(sqlStr, expiredTime, vipLevel, adminUserId)
	if err != nil {
		log.Errorf(`更新%s失败 %s`, vipTable, err.Error())
		return `更新失败`
	}
	n, err := ret.RowsAffected()
	if err != nil {
		log.Errorf(`更新%s失败 %s`, vipTable, err.Error())
		return `更新成功 ` + cast.ToString(n)
	}

	return `成功`
}

// QueryVip 查询VIP信息
// @auth frog
// @date 2023-03-16 09:31:34
func QueryVip(adminUserId, systemType string) *TblVip {
	var vipInfo = &TblVip{}
	db := AppurlDevMysql
	vipTable := `tbl_kefu_vip`
	if systemType == `1` { //客服系统
		vipTable = `tbl_kefu_vip`
	} else {
		vipTable = `tbl_official_account_vip`
	}

	errQuery := db.QueryRow(`select vip_type,expired_time from `+vipTable+` where user_id = ?`, adminUserId).Scan(&vipInfo.VipType, &vipInfo.ExpiredTime)
	if errQuery != nil {
		log.Errorf(`执行sql出错 %s`, errQuery.Error())
		return vipInfo
	}
	return vipInfo
}

// QueryEnvWechatKefuList 查询微信客服
// @auth frog
// @date 2023-04-12 09:12:46
func QueryEnvWechatKefuList(adminUserId string) string {
	appList := make([]TblWechatapp, 0)
	db := XkfDevMysql
	query, err := db.Query(`select app_name,app_id from tbl_wechatapp where user_id = ? and app_type = ?`, adminUserId, `wechat_kefu`)
	if err != nil {
		return fmt.Sprintf(`执行sql出错 %s`, err.Error())
	}
	for query.Next() {
		appInfo := TblWechatapp{}
		err := query.Scan(&appInfo.AppName, &appInfo.Appid)
		if err != nil {
			return fmt.Sprintf(`循环出错 %s`, err.Error())
		} else {
			appList = append(appList, appInfo)
		}
	}
	return lib_tool.JsonEncode(appList)
}
