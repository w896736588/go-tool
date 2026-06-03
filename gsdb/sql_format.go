package gsdb

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

type SqlFormat struct {
	tableMap       map[string]map[string]string                   //table的字段映射列表
	boolTransType  bool                                           //自动转换类型
	dbType         string                                         //mysql sqlite pgsql
	tableFieldFunc func(string string) (map[string]string, error) //获取表字段映射列表方法
	lock           sync.Mutex
}

func NewSqlFormat(boolTransType bool, dbType string, tableFieldFunc func(string string) (map[string]string, error)) *SqlFormat {
	return &SqlFormat{
		boolTransType:  boolTransType,
		tableFieldFunc: tableFieldFunc,
		dbType:         dbType,
		tableMap:       make(map[string]map[string]string),
	}
}

// FormatQuery 仅支持quick 系列
func (h *SqlFormat) FormatQuery(tableName string, where map[string]any, valueList *[]any) (string, error) {
	sqlWhere := ``
	placeholderList := make([]string, 0)
	whereList := make([]string, 0)
	for field, value := range where {
		valueType := reflect.String
		if value != nil {
			valueType = gstool.ReflectGetType(value)
		}
		switch valueType {
		case reflect.Slice, reflect.Array:
			if customs, ok := gstool.ArrayToSliceAny(value); ok {
				formatError := h.formatCustom(tableName, field, customs, &whereList, valueList)
				if formatError != nil {
					return ``, formatError
				}
			} else {
				return ``, gstool.Error(`自定义查询格式错误 %#v`, value)
			}
			break
		default:
			whereList = append(whereList, h.getField(field)+` = `+h.getPlaceholder(&placeholderList))
			transValue, transErr := h.transValue(tableName, field, value)
			if transErr != nil {
				return ``, transErr
			}
			*valueList = append(*valueList, transValue)
			break
		}
	}
	sqlWhere = strings.Join(whereList, ` and `)
	if sqlWhere != `` {
		sqlWhere = ` where ` + sqlWhere
	}
	return sqlWhere, nil
}

func (h *SqlFormat) getPlaceholder(placeholderList *[]string) string {
	if h.dbType == DbTypePgsql {
		*placeholderList = append(*placeholderList, ``)
		return fmt.Sprintf(`$%d`, len(*placeholderList))
	} else {
		return `?`
	}
}

func (h *SqlFormat) checkReplacePlaceholderNum(sqlStr string, params []any) error {
	var quesNum = 0
	if h.dbType == DbTypePgsql {
		quesNum = gstool.StringCountSubstrRegex(sqlStr, `\$\d+`)
	} else {
		quesNum = gstool.StringCountSubstr(sqlStr, `?`)
	}
	if quesNum != len(params) {
		return gstool.Error(`参数数量不一致 %s %#v`, sqlStr, params)
	}
	return nil
}

func (h *SqlFormat) getField(field string) string {
	if h.dbType == DbTypePgsql {
		return `"` + field + `"`
	}
	return "`" + field + "`"
}

// FormatInsert 仅支持quick create
func (h *SqlFormat) FormatInsert(tableName string, insert map[string]any) (string, string, []any, error) {
	placeholderList := make([]string, 0)
	insertFieldList := make([]string, 0)
	insertValueList := make([]any, 0)
	quesList := make([]string, 0)
	for field, value := range insert {
		insertFieldList = append(insertFieldList, h.getField(field))
		quesList = append(quesList, h.getPlaceholder(&placeholderList))
		transValue, transErr := h.transValue(tableName, field, value)
		if transErr != nil {
			return ``, ``, nil, transErr
		}
		insertValueList = append(insertValueList, transValue)
	}
	return strings.Join(insertFieldList, `,`), strings.Join(quesList, `,`), insertValueList, nil
}

// FormatUpdate 仅支持quick update
func (h *SqlFormat) FormatUpdate(tableName string, update map[string]any, valueList *[]any) (string, error) {
	sqlUpdate := ``
	placeholderList := make([]string, 0)
	updateList := make([]string, 0)
	for field, value := range update {
		updateList = append(updateList, h.getField(field)+` = `+h.getPlaceholder(&placeholderList))
		transValue, transErr := h.transValue(tableName, field, value)
		if transErr != nil {
			return ``, transErr
		}
		*valueList = append(*valueList, transValue)
	}
	sqlUpdate = strings.Join(updateList, ` , `)
	return sqlUpdate, nil
}

// 自定义查询类型处理
func (h *SqlFormat) formatCustom(tableName, field string, customList []any, whereList *[]string, valueList *[]any) error {
	opType, paramList, err := h.getOpTypeParams(field, customList)
	if err != nil {
		return err
	}
	placeholderList := make([]string, 0)
	switch opType {
	case `in`, `not in`, `between`:
		qusList := make([]string, 0)
		for _, customValue := range paramList {
			qusList = append(qusList, h.getPlaceholder(&placeholderList))
			transValue, transErr := h.transValue(tableName, field, customValue)
			if transErr != nil {
				return transErr
			}
			*valueList = append(*valueList, transValue)
		}
		if opType == `between` {
			*whereList = append(*whereList, h.getField(field)+` between `+qusList[0]+` and `+qusList[1])
		} else {
			*whereList = append(*whereList, h.getField(field)+` `+opType+`(`+strings.Join(qusList, `,`)+`)`)
		}
	case `like`, `>`, `<`, `>=`, `<=`, `<>`, `=`:
		transValue, transErr := h.transValue(tableName, field, paramList[0])
		if transErr != nil {
			return transErr
		}
		*valueList = append(*valueList, transValue)
		*whereList = append(*whereList, h.getField(field)+` `+opType+` `+h.getPlaceholder(&placeholderList))
	case `rawsql`: //raw sql 不支持转换
		if !gstool.ReflectIsString(paramList[0]) {
			return gstool.Error(`rawsql参数错误 %s %#v`, field, paramList)
		}
		*whereList = append(*whereList, cast.ToString(paramList[0]))
		if len(paramList) == 2 && paramList[1] != nil {
			rawValues, ok := gstool.ArrayToSliceAny(paramList[1])
			if ok {
				*valueList = append(*valueList, rawValues...)
			} else {
				return gstool.Error(`rawsql参数错误 %s %#v`, field, paramList)
			}
		}
	default:
		return nil
	}
	return nil
}

// 快捷查询格式化
func (h *SqlFormat) getOpTypeParams(field string, customList []any) (string, []any, error) {
	length := len(customList)
	//'app_id' : [] //错误
	if length == 0 {
		return ``, nil, gstool.Error(`错误的参数格式 %s %#v`, field, customList)
	}
	//'app_id' : ['xx'] 转换为=
	if length == 1 { //一个参数 那么就自动转为=
		return `=`, customList, nil
	}
	//'app_id' : [1 , ...params] 数字类型 肯定不是自定义操作符 转为in
	param1Type := gstool.ReflectGetType(customList[0])
	if param1Type != reflect.String { //第一位非字符串只能是默认为in 因为我们的操作符都是字符串形式的 不存在整数的
		return `in`, customList, nil
	}
	//'app_id' : ['' , []] 第二个值为数组切片 那么第一个操作符就只能是in或者not in
	param2Type := gstool.ReflectGetType(customList[1])
	param1 := cast.ToString(customList[0])
	if param2Type == reflect.Array || param2Type == reflect.Slice { //第二位数组 那么只支持in 和not in
		switch param1 {
		case `in`, `not in`, `between`:
			valueList, ok := gstool.ArrayToSliceAny(customList[1])
			if !ok {
				return ``, nil, gstool.Error(`参数错误 %#v`, customList)
			}
			if param1 == `between` && len(valueList) != 2 {
				return ``, nil, gstool.Error(`between参数错误 %#v`, customList)
			}
			return param1, valueList, nil
		default:
			return ``, nil, gstool.Error(`错误的格式 %#v`, customList)
		}
	}
	otherOpTypeList := []string{`like`, `>`, `<`, `>=`, `<=`, `<>`, `=`, `rawsql`}
	if gstool.ArrayExistValue(&otherOpTypeList, param1) {
		return param1, customList[1:], nil
	} else { //都不属于 那么就认为是in
		return `in`, customList, nil
	}

}

// 拿到表字段类型列表
func (h *SqlFormat) getTableFieldTypeList(tableName string) error {
	if _, ok := h.tableMap[tableName]; !ok {
		var err error
		h.tableMap[tableName], err = h.tableFieldFunc(tableName)
		if err != nil {
			return err
		}
		if len(h.tableMap[tableName]) == 0 {
			return gstool.Error(`表字段类型获取失败`)
		}
		for key, val := range h.tableMap[tableName] {
			h.tableMap[tableName][strings.ToLower(key)] = strings.ToLower(val)
		}
	}
	return nil
}

// 转换类型
func (h *SqlFormat) transValue(tableName, fieldName string, value any) (any, error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if !h.boolTransType {
		return value, nil
	}
	if _, ok := h.tableMap[tableName]; !ok {
		getError := h.getTableFieldTypeList(tableName)
		if getError != nil {
			return nil, getError
		}
	}
	if _, ok := h.tableMap[tableName][strings.ToLower(fieldName)]; !ok {
		delete(h.tableMap, tableName)
		getError := h.getTableFieldTypeList(tableName)
		if getError != nil {
			return nil, getError
		}
	}
	if fieldType, ok := h.tableMap[tableName][strings.ToLower(fieldName)]; ok {
		switch fieldType {
		case "int", "tinyint", "smallint", "mediumint", "bigint", "integer":
			return cast.ToInt(value), nil
		case "float", "double", "decimal":
			return cast.ToFloat64(value), nil
		case "date", "datetime", "timestamp":
			return cast.ToString(value), nil
		case "char", "varchar", "tinytext", "text", "mediumtext", "longtext":
			return cast.ToString(value), nil
		default:
			return value, nil
		}
	} else { //字段更新过 返回原值
		return value, nil
	}
}
