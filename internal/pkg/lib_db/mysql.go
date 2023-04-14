package lib_db

import (
	"database/sql"
	"fmt"
	"github.com/spf13/cast"
	"time"
)

//使用方法 http://events.jianshu.io/p/8963188a4fee

//创建连接
func MysqlCreateConn(dbConfig MysqlConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(`%s:%s@tcp(%s:%s)/%s`, dbConfig.Username, dbConfig.Password, dbConfig.Host, cast.ToString(dbConfig.Port), dbConfig.Dbname)
	db, err := sql.Open(`mysql`, dsn) // 不会校验用户名和密码是否正确
	if err != nil {                   // dsn 格式不正确的时候会报错
		return nil, err
	}
	maxOpenConn := cast.ToInt(dbConfig.MaxOpenConns)
	MaxIdleConns := cast.ToInt(dbConfig.MaxIdleConns)
	maxLifeTimeSecond := cast.ToInt(dbConfig.MaxLifetimeSecond)
	if maxOpenConn == 0 {
		maxOpenConn = 1
	}
	if MaxIdleConns == 0 {
		MaxIdleConns = 1
	}
	if maxLifeTimeSecond == 0 || maxLifeTimeSecond < 30 {
		maxLifeTimeSecond = 60
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Minute * time.Duration(maxLifeTimeSecond))
	return db, nil
}
