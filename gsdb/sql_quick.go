package gsdb

import (
	"database/sql"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/spf13/cast"
)

const (
	Query = iota
	Insert
	Update
	Delete
	Exec //仅返回受影响行数 如果是插入 需要使用Insert
)

type SqlQuick struct {
	db        *sql.DB
	format    *SqlFormat
	sql       string              //准备执行的sql
	op        int                 //操作 create  query  update  delete
	params    []any               //参数
	err       error               //错误信息
	debugHook func(string, error) //注册执行日志钩子
	limit     bool                //是否设置过limit
	dbType    string              //mysql sqlite pgsql
	sqlFinal  string              //最终执行的sql
	fields    string              //查询的字段 quick系列
	joins     []string            //join
	groupBy   string              //group by : id ,create_time
}

// Limit 操作多条
func (h *SqlQuick) Limit(limit int) *SqlQuick {
	if h.err != nil {
		return h
	}
	if h.sql == `` {
		h.err = errors.New(`请在执行Quick系列函数之后再执行Limit方法`)
		return h
	}
	h.limit = true
	h.sql = h.sql + ` limit ` + cast.ToString(limit)
	return h
}

func (h *SqlQuick) OffsetLimit(offset, limit int) *SqlQuick {
	if h.err != nil {
		return h
	}
	if h.sql == `` {
		h.err = errors.New(`请在执行Quick系列函数之后再执行Where方法`)
		return h
	}
	h.limit = true
	h.sql = h.sql + ` limit ` + cast.ToString(offset) + `,` + cast.ToString(limit)
	return h
}

func (h *SqlQuick) Order(orderSql string) *SqlQuick {
	if h.err != nil {
		return h
	}
	if h.sql == `` {
		h.err = errors.New(`请在执行Quick系列函数之后再执行Order方法`)
		return h
	}
	h.sql = h.sql + ` order by ` + orderSql
	return h
}

func (h *SqlQuick) setTemp(sql string, params []any, op int, err error) {
	h.sql = sql
	//处理空数据
	if len(params) == 1 && params[0] == nil {
		h.params = make([]any, 0)
	} else {
		h.params = params
	}
	h.op = op
	h.err = err
}

func (h *SqlQuick) QuickQuery(tableName, fields string, where map[string]interface{}) *SqlQuick {
	valueList := make([]any, 0)
	sqlWhere, formatErr := h.format.FormatQuery(tableName, where, &valueList)
	if formatErr != nil {
		h.setTemp(``, nil, Query, formatErr)
		return h
	}
	h.fields = fields
	sqlExec := fmt.Sprintf(`select %s from %s {joins} %s {group_by}`, fields, tableName, sqlWhere)
	h.setTemp(sqlExec, valueList, Query, nil)
	return h
}

func (h *SqlQuick) QuickUpdate(tableName string, where map[string]interface{}, update map[string]interface{}) *SqlQuick {
	if where == nil || len(where) == 0 {
		h.err = errors.New(`where条件不能为空`)
		return h
	}
	valueList := make([]any, 0)
	sqlUpdate, formatErr := h.format.FormatUpdate(tableName, update, &valueList)
	if formatErr != nil {
		h.setTemp(``, nil, Update, formatErr)
		return h
	}
	sqlWhere, formatErr := h.format.FormatQuery(tableName, where, &valueList)
	if formatErr != nil {
		h.setTemp(``, nil, Update, formatErr)
		return h
	}
	sqlExec := fmt.Sprintf(`update %s set %s %s`, tableName, sqlUpdate, sqlWhere)
	h.setTemp(sqlExec, valueList, Update, nil)
	return h
}

func (h *SqlQuick) QuickDelete(tableName string, where map[string]interface{}) *SqlQuick {
	valueList := make([]any, 0)
	sqlWhere, formatErr := h.format.FormatQuery(tableName, where, &valueList)
	if formatErr != nil {
		h.setTemp(``, nil, Delete, formatErr)
		return h
	}
	sqlExec := fmt.Sprintf(`delete from %s %s `, tableName, sqlWhere)
	h.setTemp(sqlExec, valueList, Delete, nil)
	return h
}

func (h *SqlQuick) One(txs ...*sql.Tx) (map[string]any, error) {
	h.replacePlaceholders()
	if h.err != nil {
		return nil, h.err
	}
	if h.sql == `` {
		return nil, errors.New(`请先调用QuickQuery系列方法`)
	}
	if h.op != Query {
		return nil, errors.New(`one仅支持quick query系列方法`)
	}
	if !h.limit {
		h.Limit(1)
	}
	var tx *sql.Tx
	if len(txs) > 0 {
		tx = txs[0]
	}
	dataList, err := h._queryBySql(tx, h.sql, h.params...)
	if err != nil {
		return nil, err
	}
	dataLength := len(dataList)
	if dataLength >= 1 {
		return dataList[0], nil
	} else {
		return make(map[string]any), nil
	}
}

func (h *SqlQuick) Value(key string, txs ...*sql.Tx) (any, error) {
	column, err := h.One(txs...)
	if err != nil {
		return nil, err
	}
	return column[key], nil
}

// All 操作多条
func (h *SqlQuick) All(txs ...*sql.Tx) ([]map[string]any, error) {
	h.replacePlaceholders()
	if h.err != nil {
		return nil, h.err
	}
	if h.sql == `` {
		return nil, errors.New(`请先调用QuickQuery方法`)
	}
	if h.op != Query {
		return nil, errors.New(`all仅支持quick query系列方法`)
	}
	var tx *sql.Tx
	if len(txs) > 0 {
		tx = txs[0]
	}
	return h._queryBySql(tx, h.sql, h.params...)
}

// Exec 执行变更
func (h *SqlQuick) Exec(txs ...*sql.Tx) (int64, error) {
	if h.err != nil {
		return 0, h.err
	}
	if h.sql == `` {
		return 0, errors.New(`请先调用QuickDelete QuickUpdate  QuickCreate方法`)
	}
	allowOpList := []int{Update, Delete, Insert, Exec}
	if !gstool.ArrayExistValue(&allowOpList, h.op) {
		return 0, errors.New(`exec仅支持quick update delete insert系列方法`)
	}
	var tx *sql.Tx
	if len(txs) > 0 {
		tx = txs[0]
	}
	if h.op == Delete || h.op == Update || h.op == Exec {
		return h._execBySql(tx, h.sql, h.params...)
	} else {
		return h._insertBySql(tx, h.sql, h.params...)
	}
}

func (h *SqlQuick) GetSql() string {
	return h.sqlFinal
}

func (h *SqlQuick) debugSql(sql string, params []any) {
	h.sqlFinal = h._getSql(sql, params)
	if h.debugHook != nil {
		h.debugHook(h.sqlFinal, h.err)
	}
}

func (h *SqlQuick) _getSql(sql string, params []any) string {
	placeholderNumErr := h.format.checkReplacePlaceholderNum(sql, params)
	if placeholderNumErr != nil {
		sql = placeholderNumErr.Error()
	} else {
		if h.dbType == DbTypePgsql {
			sql = gstool.StringReplaceRegex(sql, `\$\d+`, `%s`)
		} else {
			sql = strings.ReplaceAll(sql, `?`, `%s`)
		}
		replaceValues := make([]any, 0)
		for _, value := range params {
			if gstool.ReflectIsString(value) {
				replaceValues = append(replaceValues, `'`+cast.ToString(value)+`'`)
			} else {
				replaceValues = append(replaceValues, cast.ToString(value))
			}
		}
		sql = fmt.Sprintf(sql, replaceValues...)
	}
	return sql
}

func (h *SqlQuick) QueryBySql(sqlStr string, params ...interface{}) *SqlQuick {
	h.setTemp(sqlStr, params, Query, nil)
	return h
}

func (h *SqlQuick) _queryBySql(tx *sql.Tx, sqlStr string, params ...interface{}) ([]map[string]any, error) {
	var rows *sql.Rows
	var err error
	var dataList = make([]map[string]any, 0)
	h.debugSql(sqlStr, params)
	if tx != nil {
		rows, err = tx.Query(sqlStr, params...)
	} else {
		rows, err = h.db.Query(sqlStr, params...)
	}
	if err != nil {
		return dataList, err
	}
	defer func(rows *sql.Rows) {
		errClose := rows.Close()
		if errClose != nil {
			gstool.FmtPrintlnLogTime(`关闭查询失败  %s %s %#v`, err.Error(), sqlStr, debug.Stack(), params)
		}
	}(rows)
	var columns []string
	columns, err = rows.Columns()
	if err != nil {
		return dataList, err
	}
	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	// 这里需要初始化为空数组，否则在查询结果为空的时候，返回的会是一个未初始化的指针
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return dataList, err
		}
		dataRow := make(map[string]any)
		for i, col := range values {
			if b, ok := col.([]byte); ok {
				dataRow[columns[i]] = string(b) // 如果是 []byte，强制转 string
			} else {
				dataRow[columns[i]] = col
			}
		}
		dataList = append(dataList, dataRow)
	}
	if err = rows.Err(); err != nil {
		return dataList, err
	}
	return dataList, nil
}

func (h *SqlQuick) ExecBySql(sqlStr string, params ...interface{}) *SqlQuick {
	h.setTemp(sqlStr, params, Exec, nil)
	return h
}

func (h *SqlQuick) _execBySql(tx *sql.Tx, sqlStr string, params ...interface{}) (int64, error) {
	h.debugSql(sqlStr, params)
	var ret sql.Result
	var err error
	if tx != nil {
		ret, err = tx.Exec(sqlStr, params...)
	} else {
		ret, err = h.db.Exec(sqlStr, params...)
	}
	if err != nil {
		return 0, err
	}
	rows, err := ret.RowsAffected()
	if err != nil {
		return 0, err
	}
	return cast.ToInt64(rows), nil
}

func (h *SqlQuick) InsertBySql(sqlStr string, params ...interface{}) *SqlQuick {
	h.setTemp(sqlStr, params, Insert, nil)
	return h
}

func (h *SqlQuick) _insertBySql(tx *sql.Tx, sqlStr string, params ...interface{}) (int64, error) {
	h.debugSql(sqlStr, params)
	var ret sql.Result
	var err error
	if tx != nil {
		ret, err = tx.Exec(sqlStr, params...)
	} else {
		ret, err = h.db.Exec(sqlStr, params...)
	}
	if err != nil {
		return 0, err
	}

	lastId, err := ret.LastInsertId() // 新插入的数据id
	if err != nil {
		return 0, err
	}
	return lastId, nil
}

func (h *SqlQuick) QuickCreate(tableName string, params map[string]interface{}, primaryKey ...string) *SqlQuick {
	sqlField, sqlQues, valueList, formatErr := h.format.FormatInsert(tableName, params)
	if formatErr != nil {
		h.setTemp(``, nil, Insert, formatErr)
		return h
	}
	returning := ``
	if h.dbType == DbTypePgsql && len(primaryKey) > 0 {
		returning = fmt.Sprintf(` returning %s`, primaryKey[0])
	}
	sqlExec := fmt.Sprintf(`insert into %s (%s) values(%s) %s`, tableName, sqlField, sqlQues, returning)
	h.setTemp(sqlExec, valueList, Insert, nil)
	return h
}

func (h *SqlQuick) TableColumnsMap(columnList []map[string]any) (map[string]string, error) {
	tableFieldMap := map[string]string{}
	for _, dbFieldTypeParam := range columnList {
		if h.dbType == DbTypeMysql {
			tableFieldMap[cast.ToString(dbFieldTypeParam[`COLUMN_NAME`])] = cast.ToString(dbFieldTypeParam[`DATA_TYPE`])
		} else if h.dbType == DbTypePgsql {
			tableFieldMap[cast.ToString(dbFieldTypeParam[`column_name`])] = cast.ToString(dbFieldTypeParam[`data_type`])
		} else if h.dbType == DbTypeSqlite {
			tableFieldMap[cast.ToString(dbFieldTypeParam[`name`])] = cast.ToString(dbFieldTypeParam[`type`])
		}
	}
	return tableFieldMap, nil
}

func (h *SqlQuick) TableDetail(tableName string) ([]map[string]any, error) {
	sqlStr := fmt.Sprintf(`SELECT * FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE () AND TABLE_NAME = '%s'`, tableName)
	return h.QueryBySql(sqlStr, nil).All()
}

func (h *SqlQuick) Join(join string, params ...any) *SqlQuick {
	if h.op != Query {
		h.err = errors.New(`请先调用QuickQuery系列方法`)
		return h
	}
	h.joins = append(h.joins, join)
	if len(params) > 0 {
		h.params = append(h.params, params...)
	}
	return h
}

// GroupBy id ,create_time
func (h *SqlQuick) GroupBy(groupBy string) *SqlQuick {
	h.groupBy = groupBy
	return h
}

func (h *SqlQuick) replacePlaceholders() {
	h.sql = strings.ReplaceAll(h.sql, `{joins}`, strings.Join(h.joins, ` `))
	if len(h.groupBy) > 0 {
		h.sql = strings.ReplaceAll(h.sql, `{group_by}`, `group by `+h.groupBy)
	} else {
		h.sql = strings.ReplaceAll(h.sql, `{group_by}`, ``)
	}

}

func (h *SqlQuick) ToMapMap(key string, txs ...*sql.Tx) (map[string]map[string]any, error) {
	if h.err != nil {
		return nil, h.err
	}
	if key == `` {
		return nil, errors.New(`key参数不能为空`)
	}
	list, err := h.All(txs...)
	if err != nil {
		return nil, err
	}
	resultMap := make(map[string]map[string]any)
	for _, item := range list {
		if keyValue, exists := item[key]; exists {
			resultMap[cast.ToString(keyValue)] = item
		}
	}
	return resultMap, nil
}

func (h *SqlQuick) ToGroupSlice(key string, txs ...*sql.Tx) (map[string][]map[string]any, error) {
	if h.err != nil {
		return nil, h.err
	}
	if key == `` {
		return nil, errors.New(`key参数不能为空`)
	}
	list, err := h.All(txs...)
	if err != nil {
		return nil, err
	}
	resultMap := make(map[string][]map[string]any)
	for _, item := range list {
		if keyValue, exists := item[key]; exists {
			if _, exists := resultMap[cast.ToString(keyValue)]; !exists {
				resultMap[cast.ToString(keyValue)] = []map[string]any{}
			}
			resultMap[cast.ToString(keyValue)] = append(resultMap[cast.ToString(keyValue)], item)
		}
	}
	return resultMap, nil
}

func (h *SqlQuick) ToMap(fieldKey, valueKey string, txs ...*sql.Tx) (gstool.MapStringAny, error) {
	if h.err != nil {
		return nil, h.err
	}
	if fieldKey == `` || valueKey == `` {
		return nil, errors.New(`key和valueKey参数不能为空`)
	}
	list, err := h.All(txs...)
	if err != nil {
		return nil, err
	}
	mapData := make(gstool.MapStringAny)
	for _, item := range list {
		if keyValue, exists := item[fieldKey]; exists {
			mapData[cast.ToString(keyValue)] = item[valueKey]
		}
	}
	return mapData, nil
}

func (h *SqlQuick) ToSlice(key string, txs ...*sql.Tx) (gstool.SliceAny, error) {
	if h.err != nil {
		return nil, h.err
	}
	if key == `` {
		return make([]any, 0), nil
	}
	list, err := h.All(txs...)
	if err != nil {
		return nil, err
	}
	slice := make(gstool.SliceAny, len(list))
	for _, item := range list {
		if keyValue, exists := item[key]; exists {
			slice = append(slice, keyValue)
		}
	}
	return slice, nil
}
