package gsdb

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/w896736588/go-tool/gsssh"
)

type MysqlConfig struct {
	Name              string `json:"name"`
	Host              string `json:"host"`
	Port              int64  `json:"port"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	Dbname            string `json:"dbname"`
	MaxOpenConns      int    `json:"maxOpenConns"`      //最大连接池连接数量
	MaxIdleConns      int    `json:"maxIdleConns"`      //最大可空闲连接数
	MaxLifetimeSecond int    `json:"maxLifetimeSecond"` //链接重置时间 秒
}

type GsMysql struct {
	MysqlConfig *MysqlConfig
	db          *sql.DB
	SshBridge   *gsssh.SshBridge
	format      *SqlFormat
	debugHook   func(string, error)
	openFunc    func(db *sql.DB)
}

func NewMysql(config *MysqlConfig, autoTransType bool) *GsMysql {
	gsMysql := &GsMysql{
		MysqlConfig: config,
	}
	gsMysql.format = NewSqlFormat(autoTransType, DbTypeMysql, func(tableName string) (map[string]string, error) {
		columnList, columnListErr := gsMysql.TableDetail(tableName).All()
		if columnListErr != nil {
			return nil, columnListErr
		}
		return gsMysql.getQuick().TableColumnsMap(columnList)
	})
	return gsMysql
}

func (h *GsMysql) SetOpenFunc(openFunc func(db *sql.DB)) {
	h.openFunc = openFunc
}

func (h *GsMysql) RegisterDebugHook(hook func(string, error)) {
	h.debugHook = hook
}

// CreateConn create a connection
func (h *GsMysql) CreateConn() error {
	if h.format == nil {
		h.format = NewSqlFormat(false, DbTypeMysql, func(tableName string) (map[string]string, error) {
			columnList, columnListErr := h.TableDetail(tableName).All()
			if columnListErr != nil {
				return nil, columnListErr
			}
			return h.getQuick().TableColumnsMap(columnList)
		})
	}
	if h.SshBridge != nil {
		return h.createConnSshPasswordAuth()
	} else {
		return h.createConnDirect()
	}
}

// ssh bridge password auth
// 注意：使用隧道时，虽然链接会创建多个，但是因为隧道本身就只开启了一个，所以数据库操作是同步阻塞的
// 注意：线上环境不要使用隧道
func (h *GsMysql) createConnSshPasswordAuth() error {
	mysqlHostPort := fmt.Sprintf(`%s:%d`, h.MysqlConfig.Host, h.MysqlConfig.Port)
	localHostPort, runError := h.SshBridge.RunBridge(mysqlHostPort)
	if runError != nil {
		return runError
	}
	dns := fmt.Sprintf("%s:%s@tcp(%s)/%s",
		h.MysqlConfig.Username,
		h.MysqlConfig.Password,
		localHostPort,
		h.MysqlConfig.Dbname,
	)
	openError := h.dbOpen(dns)
	if openError != nil {
		return openError
	}
	if h.openFunc != nil {
		h.openFunc(h.db)
	}
	return nil
}

// direct
func (h *GsMysql) createConnDirect() error {
	dns := fmt.Sprintf(`%s:%s@tcp(%s:%d)/%s`, h.MysqlConfig.Username, h.MysqlConfig.Password, h.MysqlConfig.Host, h.MysqlConfig.Port, h.MysqlConfig.Dbname)
	openError := h.dbOpen(dns)
	if openError != nil {
		return openError
	}
	if h.openFunc != nil {
		h.openFunc(h.db)
	}
	return nil
}

func (h *GsMysql) dbOpen(dns string) error {
	var dbError error
	h.db, dbError = sql.Open("mysql", dns)
	if dbError != nil {
		return dbError
	}
	maxOpenConn := h.MysqlConfig.MaxOpenConns
	MaxIdleConns := h.MysqlConfig.MaxIdleConns
	maxLifeTimeSecond := h.MysqlConfig.MaxLifetimeSecond
	if maxOpenConn == 0 {
		maxOpenConn = 1
	}
	if MaxIdleConns == 0 {
		MaxIdleConns = 1
	}
	if maxLifeTimeSecond == 0 || maxLifeTimeSecond < 30 {
		maxLifeTimeSecond = 60
	}

	h.db.SetMaxOpenConns(maxOpenConn)
	h.db.SetMaxIdleConns(MaxIdleConns)
	h.db.SetConnMaxLifetime(time.Minute * time.Duration(maxLifeTimeSecond))
	pingErr := h.db.Ping()
	if pingErr != nil {
		return pingErr
	}
	return nil
}

// QuickQuery 快速查询数据
func (h *GsMysql) QuickQuery(tableName, fields string, where map[string]interface{}) *SqlQuick {
	return h.getQuick().QuickQuery(tableName, fields, where)
}

func (h *GsMysql) getQuick() *SqlQuick {
	return &SqlQuick{
		format:    h.format,
		db:        h.db,
		err:       nil,
		op:        0,
		params:    nil,
		sql:       ``,
		debugHook: h.debugHook,
		dbType:    `mysql`,
		joins:     make([]string, 0),
	}
}

// QueryBySql 获取多行数据
func (h *GsMysql) QueryBySql(sqlStr string, params ...interface{}) *SqlQuick {
	return h.getQuick().QueryBySql(sqlStr, params...)
}

// QuickUpdate 快速更新
func (h *GsMysql) QuickUpdate(tableName string, where map[string]interface{}, update map[string]interface{}) *SqlQuick {
	return h.getQuick().QuickUpdate(tableName, where, update)
}

// QuickDelete 快速删除数据
func (h *GsMysql) QuickDelete(tableName string, where map[string]interface{}) *SqlQuick {
	return h.getQuick().QuickDelete(tableName, where)
}

// ExecBySql 执行
func (h *GsMysql) ExecBySql(sqlStr string, params ...interface{}) *SqlQuick {
	return h.getQuick().ExecBySql(sqlStr, params...)
}

// TableDetail 查询表的信息
func (h *GsMysql) TableDetail(tableName string) *SqlQuick {
	sqlStr := fmt.Sprintf(`SELECT * FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE () AND TABLE_NAME = '%s'`, tableName)
	return h.getQuick().QueryBySql(sqlStr, nil)
}

// InsertBySql 插入
func (h *GsMysql) InsertBySql(sqlStr string, params ...interface{}) *SqlQuick {
	return h.getQuick().InsertBySql(sqlStr, params...)
}

// QuickCreate 快速根据map插入
func (h *GsMysql) QuickCreate(tableName string, params map[string]interface{}) *SqlQuick {
	return h.getQuick().QuickCreate(tableName, params)
}

// GetTx 获取事务
// 注意：如果使用了隧道，那么在tx释放之前，所有查询都会被阻塞
func (h *GsMysql) GetTx() (*sql.Tx, error) {
	if h.db == nil {
		return nil, errors.New(`db is not connected`)
	}
	return h.db.Begin()
}
