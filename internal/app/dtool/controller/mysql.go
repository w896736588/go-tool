package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/pkg/p_db"
	"fmt"
	"regexp"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// getMysqlClient 根据mysql_id获取配置和客户端连接
func getMysqlClient(mysqlId string) (map[string]any, *gsdb.GsMysql, error) {
	mysqlConfig, configErr := common.DbMain.GetMysqlConfig(mysqlId)
	if configErr != nil || len(mysqlConfig) == 0 {
		return nil, nil, fmt.Errorf(`未找到id为 "%s" 的MySQL配置`, mysqlId)
	}
	client, clientErr := p_db.MysqlClient.GetClient(mysqlConfig, common.GetCall())
	if clientErr != nil {
		return nil, nil, fmt.Errorf(`获取MySQL连接失败: %s`, clientErr.Error())
	}
	return mysqlConfig, client, nil
}

// MysqlTables 查询MySQL配置对应数据库的所有表
// 参数：mysql_id(MySQL配置ID)
func MysqlTables(c *gin.Context) {
	reqMap := make(map[string]interface{})
	if err := gsgin.GinPostBody(c, &reqMap); err != nil {
		gsgin.GinResponseError(c, `请求参数错误`, nil)
		return
	}

	mysqlId := cast.ToString(reqMap[`mysql_id`])
	if mysqlId == `` {
		gsgin.GinResponseError(c, `mysql_id不能为空`, nil)
		return
	}

	mysqlConfig, mysqlClient, err := getMysqlClient(mysqlId)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	dbname := cast.ToString(mysqlConfig[`dbname`])
	sql := fmt.Sprintf(
		`SELECT TABLE_NAME AS table_name, TABLE_COMMENT AS table_comment FROM information_schema.TABLES WHERE TABLE_SCHEMA = '%s' ORDER BY TABLE_NAME`,
		dbname,
	)
	list, queryErr := mysqlClient.QueryBySql(sql).All()
	if queryErr != nil {
		gsgin.GinResponseError(c, `查询表列表失败: `+queryErr.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: list,
	})
}

// MysqlTableStructure 查询MySQL表结构
// 参数：mysql_id(MySQL配置ID), table_name(表名)
func MysqlTableStructure(c *gin.Context) {
	reqMap := make(map[string]interface{})
	if err := gsgin.GinPostBody(c, &reqMap); err != nil {
		gsgin.GinResponseError(c, `请求参数错误`, nil)
		return
	}

	mysqlId := cast.ToString(reqMap[`mysql_id`])
	tableName := strings.TrimSpace(cast.ToString(reqMap[`table_name`]))

	if mysqlId == `` {
		gsgin.GinResponseError(c, `mysql_id不能为空`, nil)
		return
	}
	if tableName == `` {
		gsgin.GinResponseError(c, `table_name不能为空`, nil)
		return
	}
	if !isSafeTableName(tableName) {
		gsgin.GinResponseError(c, `table_name不合法`, nil)
		return
	}

	_, mysqlClient, err := getMysqlClient(mysqlId)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	sql := fmt.Sprintf(`SHOW FULL COLUMNS FROM %s`, tableName)
	list, queryErr := mysqlClient.QueryBySql(sql).All()
	if queryErr != nil {
		gsgin.GinResponseError(c, `查询表结构失败: `+queryErr.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: list,
	})
}

// MysqlQuery 执行MySQL查询（仅支持SELECT）
// 参数：mysql_id(MySQL配置ID), sql(查询SQL)
func MysqlQuery(c *gin.Context) {
	reqMap := make(map[string]interface{})
	if err := gsgin.GinPostBody(c, &reqMap); err != nil {
		gsgin.GinResponseError(c, `请求参数错误`, nil)
		return
	}

	mysqlId := cast.ToString(reqMap[`mysql_id`])
	sql := strings.TrimSpace(cast.ToString(reqMap[`sql`]))

	if mysqlId == `` {
		gsgin.GinResponseError(c, `mysql_id不能为空`, nil)
		return
	}
	if sql == `` {
		gsgin.GinResponseError(c, `sql不能为空`, nil)
		return
	}
	if !isSelectQuery(sql) {
		gsgin.GinResponseError(c, `仅支持SELECT查询`, nil)
		return
	}

	_, mysqlClient, err := getMysqlClient(mysqlId)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	list, queryErr := mysqlClient.QueryBySql(sql).All()
	if queryErr != nil {
		gsgin.GinResponseError(c, `查询失败: `+queryErr.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: list,
	})
}

// isSafeTableName 校验表名只包含安全字符，防止SQL注入
func isSafeTableName(name string) bool {
	ok, _ := regexp.MatchString(`^[A-Za-z0-9_\.]+$`, name)
	return ok
}

// isSelectQuery 检查SQL是否为SELECT语句
func isSelectQuery(sql string) bool {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	return strings.HasPrefix(upper, `SELECT`)
}
