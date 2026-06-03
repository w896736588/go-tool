package gsdb

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type GsSqlite struct {
	DbPath    string
	db        *sql.DB
	format    *SqlFormat
	debugHook func(string, error)
}

// NewSqlite create sqlite
func NewSqlite(dbPath string, autoTransType bool) (*GsSqlite, error) {
	if dbPath == `` {
		return nil, errors.New(`dbPath can not be empty`)
	}
	sqlLite := &GsSqlite{
		DbPath: dbPath,
	}
	sqlLite.format = NewSqlFormat(autoTransType, DbTypeSqlite, func(tableName string) (map[string]string, error) {
		columnList, columnListErr := sqlLite.TableDetail(tableName).All()
		if columnListErr != nil {
			return nil, columnListErr
		}
		return sqlLite.getQuick().TableColumnsMap(columnList)
	})
	createErr := sqlLite.CreateConn()
	if createErr != nil {
		return nil, createErr
	}
	return sqlLite, nil
}

func (h *GsSqlite) RegisterDebugHook(hook func(string, error)) {
	h.debugHook = hook
}

// CreateConn create connection
func (h *GsSqlite) CreateConn() error {
	if h.format == nil {
		h.format = NewSqlFormat(false, DbTypeSqlite, func(tableName string) (map[string]string, error) {
			columnList, columnListErr := h.TableDetail(tableName).All()
			if columnListErr != nil {
				return nil, columnListErr
			}
			return h.getQuick().TableColumnsMap(columnList)
		})
	}
	var err error
	if h.DbPath == `` {
		return errors.New(`sqlite database file address does not exist`)
	}
	h.db, err = sql.Open(`sqlite3`, h.DbPath)
	if err != nil {
		return err
	}
	return nil
}

// QuickQuery 快速查询数据
func (h *GsSqlite) QuickQuery(tableName, fields string, where map[string]interface{}) *SqlQuick {
	return h.getQuick().QuickQuery(tableName, fields, where)
}

func (h *GsSqlite) getQuick() *SqlQuick {
	return &SqlQuick{
		format:    h.format,
		db:        h.db,
		err:       nil,
		op:        0,
		params:    nil,
		sql:       ``,
		debugHook: h.debugHook,
		dbType:    `sqlite`,
		joins:     make([]string, 0),
	}
}

// QueryBySql 获取多行数据
func (h *GsSqlite) QueryBySql(sqlStr string, params ...interface{}) *SqlQuick {
	return h.getQuick().QueryBySql(sqlStr, params...)
}

// QuickUpdate 快速更新
func (h *GsSqlite) QuickUpdate(tableName string, where map[string]interface{}, update map[string]interface{}) *SqlQuick {
	return h.getQuick().QuickUpdate(tableName, where, update)
}

// QuickDelete 快速删除数据
func (h *GsSqlite) QuickDelete(tableName string, where map[string]interface{}) *SqlQuick {
	return h.getQuick().QuickDelete(tableName, where)
}

// ExecBySql 执行
func (h *GsSqlite) ExecBySql(sqlStr string, params ...interface{}) *SqlQuick {
	return h.getQuick().ExecBySql(sqlStr, params...)
}

// TableDetail 查询表的信息
func (h *GsSqlite) TableDetail(tableName string) *SqlQuick {
	sqlStr := fmt.Sprintf(`PRAGMA table_info(%s)`, tableName)
	return h.getQuick().QueryBySql(sqlStr, nil)
}

// InsertBySql 插入
func (h *GsSqlite) InsertBySql(sqlStr string, params ...interface{}) *SqlQuick {
	return h.getQuick().InsertBySql(sqlStr, params...)
}

// QuickCreate 快速根据map插入
func (h *GsSqlite) QuickCreate(tableName string, params map[string]interface{}) *SqlQuick {
	return h.getQuick().QuickCreate(tableName, params)
}

// GetTx 获取事务
func (h *GsSqlite) GetTx() (*sql.Tx, error) {
	if h.db == nil {
		return nil, errors.New(`db is not connected`)
	}
	return h.db.Begin()
}
