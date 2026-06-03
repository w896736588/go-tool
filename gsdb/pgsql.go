package gsdb

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/w896736588/go-tool/gsssh"
)

type PgsqlConfig struct {
	Name              string `json:"name"`
	Host              string `json:"host"`
	Port              int64  `json:"port"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	Dbname            string `json:"dbname"`
	PoolSize          int    `json:"poolsize"`
	MaxOpenConns      int    `json:"maxOpenConns"`
	MaxIdleConns      int    `json:"maxIdleConns"`
	MaxLifetimeSecond int    `json:"maxLifetimeSecond"` //链接重置时间 秒
}

type GsPgsql struct {
	PgsqlConfig *PgsqlConfig
	db          *sql.DB
	GsLog       *gstool.GsSlog
	SshBridge   *gsssh.SshBridge
	format      *SqlFormat
	debugHook   func(string, error)
	openFunc    func(db *sql.DB)
}

func NewPgsql(config *PgsqlConfig, autoTransType bool) *GsPgsql {
	gsPgsql := &GsPgsql{
		PgsqlConfig: config,
	}
	gsPgsql.format = NewSqlFormat(autoTransType, DbTypePgsql, func(tableName string) (map[string]string, error) {
		columnList, columnListErr := gsPgsql.TableDetail(tableName).All()
		if columnListErr != nil {
			return nil, columnListErr
		}
		return gsPgsql.getQuick().TableColumnsMap(columnList)
	})
	return gsPgsql
}

func (h *GsPgsql) SetOpenFunc(openFunc func(db *sql.DB)) {
	h.openFunc = openFunc
}

func (h *GsPgsql) SetGsLog(gsLog *gstool.GsSlog) {
	h.GsLog = gsLog
}

func (h *GsPgsql) RegisterDebugHook(hook func(string, error)) {
	h.debugHook = hook
}

// CreateConn create a connection
func (h *GsPgsql) CreateConn() error {
	if h.format == nil {
		h.format = NewSqlFormat(false, DbTypePgsql, func(tableName string) (map[string]string, error) {
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
func (h *GsPgsql) createConnSshPasswordAuth() error {
	pgsqlHostPort := fmt.Sprintf(`%s:%d`, h.PgsqlConfig.Host, h.PgsqlConfig.Port)
	localHostPort, runError := h.SshBridge.RunBridge(pgsqlHostPort)
	if runError != nil {
		return runError
	}
	parts := strings.Split(localHostPort, ":")
	dns := fmt.Sprintf(`host=%s port=%s user=%s password=%s dbname=%s sslmode=disable`,
		parts[0], parts[1], h.PgsqlConfig.Username, h.PgsqlConfig.Password, h.PgsqlConfig.Dbname)
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
func (h *GsPgsql) createConnDirect() error {
	dns := fmt.Sprintf(`host=%s port=%d user=%s password=%s dbname=%s sslmode=disable `,
		h.PgsqlConfig.Host, h.PgsqlConfig.Port, h.PgsqlConfig.Username, h.PgsqlConfig.Password, h.PgsqlConfig.Dbname)
	openError := h.dbOpen(dns)
	if openError != nil {
		return openError
	}
	if h.openFunc != nil {
		h.openFunc(h.db)
	}
	return nil
}

func (h *GsPgsql) dbOpen(dns string) error {
	var dbError error
	h.db, dbError = sql.Open("postgres", dns)
	if dbError != nil {
		return dbError
	}
	maxOpenConn := h.PgsqlConfig.MaxOpenConns
	MaxIdleConns := h.PgsqlConfig.MaxIdleConns
	maxLifeTimeSecond := h.PgsqlConfig.MaxLifetimeSecond
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

func (h *GsPgsql) connInit() error {
	maxOpenConn := h.PgsqlConfig.MaxOpenConns
	MaxIdleConns := h.PgsqlConfig.MaxIdleConns
	maxLifeTimeSecond := h.PgsqlConfig.MaxLifetimeSecond
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
func (h *GsPgsql) QuickQuery(tableName, fields string, where map[string]interface{}) *SqlQuick {
	return h.getQuick().QuickQuery(tableName, fields, where)
}

func (h *GsPgsql) getQuick() *SqlQuick {
	return &SqlQuick{
		format:    h.format,
		db:        h.db,
		err:       nil,
		op:        0,
		params:    nil,
		sql:       ``,
		debugHook: h.debugHook,
		dbType:    `pgsql`,
		joins:     make([]string, 0),
	}
}

// QueryBySql 获取多行数据
func (h *GsPgsql) QueryBySql(sqlStr string, params ...interface{}) *SqlQuick {
	return h.getQuick().QueryBySql(sqlStr, params...)
}

// QuickUpdate 快速更新
func (h *GsPgsql) QuickUpdate(tableName string, where map[string]interface{}, update map[string]interface{}) *SqlQuick {
	return h.getQuick().QuickUpdate(tableName, where, update)
}

// QuickDelete 快速删除数据
func (h *GsPgsql) QuickDelete(tableName string, where map[string]interface{}) *SqlQuick {
	return h.getQuick().QuickDelete(tableName, where)
}

// ExecBySql 执行
func (h *GsPgsql) ExecBySql(sqlStr string, params ...interface{}) *SqlQuick {
	return h.getQuick().ExecBySql(sqlStr, params...)
}

// TableDetail 查询表的信息
func (h *GsPgsql) TableDetail(tableName string) *SqlQuick {
	sqlStr := fmt.Sprintf(`SELECT * FROM information_schema.columns WHERE table_schema = 'public' AND table_name = '%s'`, tableName)
	return h.getQuick().QueryBySql(sqlStr, nil)
}

// InsertBySql 插入
func (h *GsPgsql) InsertBySql(sqlStr string, params ...interface{}) *SqlQuick {
	return h.getQuick().InsertBySql(sqlStr, params...)
}

// QuickCreate 快速根据map插入
func (h *GsPgsql) QuickCreate(tableName string, params map[string]interface{}) *SqlQuick {
	return h.getQuick().QuickCreate(tableName, params)
}

// GetTx 获取事务
// 注意：如果使用了隧道，那么在tx释放之前，所有查询都会被阻塞
func (h *GsPgsql) GetTx() (*sql.Tx, error) {
	if h.db == nil {
		return nil, errors.New(`db is not connected`)
	}
	return h.db.Begin()
}
