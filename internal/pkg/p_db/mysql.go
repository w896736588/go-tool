package p_db

import (
	"dev_tool/internal/pkg/p_common"
	"errors"
	"sync"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
	"gitee.com/Sxiaobai/gs/v2/gsssh"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

var MysqlClient *TMysql

type TMysql struct {
	MysqlClientMap map[string]*gsdb.GsMysql
	lock           sync.Mutex
}

func InitMysql() {
	MysqlClient = &TMysql{
		MysqlClientMap: make(map[string]*gsdb.GsMysql),
	}
}

func (h *TMysql) GetClient(mysqlConfig map[string]any, call *p_common.Call) (*gsdb.GsMysql, error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	mysqlId := cast.ToString(mysqlConfig[`id`])
	if mysqlId == `` {
		return nil, errors.New(`mysql配置错误`)
	}
	if mysqlCli, ok := h.MysqlClientMap[mysqlId]; ok {
		_, err := mysqlCli.QueryBySql(`select 1`).All()
		if err == nil {
			return mysqlCli, nil
		}
	}
	gsMysql := &gsdb.GsMysql{
		MysqlConfig: &gsdb.MysqlConfig{
			Name:     cast.ToString(mysqlConfig[`name`]),
			Host:     cast.ToString(mysqlConfig[`host`]),
			Port:     cast.ToInt64(mysqlConfig[`port`]),
			Password: cast.ToString(mysqlConfig[`password`]),
			Username: cast.ToString(mysqlConfig[`username`]),
			Dbname:   cast.ToString(mysqlConfig[`dbname`]),
		},
	}
	gsMysql.OpenDebug()
	if cast.ToInt(mysqlConfig[`ssh_id`]) != 0 {
		sshConfig, sshConfigErr := call.GetSshConfig(mysqlConfig[`ssh_id`])
		if sshConfigErr != nil {
			return nil, gstool.Error(`获取ssh配置失败 %s`, sshConfigErr.Error())
		}
		gsMysql.SshBridge = gsssh.NewSshBridge(gsssh.NewSsh(&gsssh.SshConfig{
			Name:     cast.ToString(sshConfig[`name`]),
			Host:     cast.ToString(sshConfig[`host`]),
			Port:     cast.ToString(sshConfig[`port`]),
			UserName: cast.ToString(sshConfig[`username`]),
			Password: cast.ToString(sshConfig[`password`]),
		}))
	}
	connErr := gsMysql.CreateConn()
	if connErr != nil {
		return nil, connErr
	}
	h.MysqlClientMap[mysqlId] = gsMysql
	return gsMysql, nil
}

func (h *TMysql) PingAll() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				//TODO 连接检查
			}
		}
	}()
}

func (h *TMysql) RmClient(mysqlConfig map[string]any) {
	defer h.lock.Unlock()
	h.lock.Lock()
	mysqlId := cast.ToString(mysqlConfig[`id`])
	delete(h.MysqlClientMap, mysqlId)
}
