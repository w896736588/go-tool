package p_db

import (
	"fmt"
	"gitee.com/Sxiaobai/gs/v2/gsdb"
	"path/filepath"
)

var SqliteClient *gsdb.GsSqlite

func InitSqlite(dbPath, dbName string) {
	var err error
	SqliteClient, err = gsdb.NewSqlite(filepath.Join(dbPath, dbName), true)
	if err != nil {
		panic(fmt.Sprintf(`连接sqlite失败 %s`, err.Error()))
	}
	//sqlite.OpenDebug()
	createErr := SqliteClient.CreateConn()
	if createErr != nil {
		panic(fmt.Sprintf(`打开sqlite失败 %s`, createErr.Error()))
	}
}
