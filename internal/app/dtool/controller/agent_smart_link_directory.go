package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// SmartLinkDirectoryForAgent 给本地 agent 提供智能网页固定目录映射的代理接口。
// 该接口把 DB 访问留在服务端，避免 agent 进程直接读取用户配置库。
func SmartLinkDirectoryForAgent(c *gin.Context) {
	var req define.AgentSmartLinkDirectoryRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, "请求参数错误", nil)
		return
	}
	switch req.Action {
	case define.AgentSmartLinkDirectoryActionGetByMappingKey:
		smartLinkDirectoryForAgentGetByMappingKey(c, req)
	case define.AgentSmartLinkDirectoryActionExistsIndex:
		smartLinkDirectoryForAgentExistsIndex(c, req)
	case define.AgentSmartLinkDirectoryActionUpsert:
		smartLinkDirectoryForAgentUpsert(c, req)
	default:
		gsgin.GinResponseError(c, "action不支持", nil)
	}
}

// smartLinkDirectoryForAgentGetByMappingKey 根据 mapping_key 查询固定目录索引。
func smartLinkDirectoryForAgentGetByMappingKey(c *gin.Context, req define.AgentSmartLinkDirectoryRequest) {
	if strings.TrimSpace(req.MappingKey) == "" {
		gsgin.GinResponseSuccess(c, "", define.AgentSmartLinkDirectoryResponse{})
		return
	}
	row, err := common.DbMain.Client.QueryBySql(
		`select * from tbl_smart_link_directory_mapping where mapping_key = ? `,
		req.MappingKey,
	).One()
	if err != nil {
		gsgin.GinResponseSuccess(c, "", define.AgentSmartLinkDirectoryResponse{})
		return
	}
	gsgin.GinResponseSuccess(c, "", define.AgentSmartLinkDirectoryResponse{
		UserDataIndex: cast.ToInt(row[`user_data_index`]),
	})
}

// smartLinkDirectoryForAgentExistsIndex 判断某目录索引是否已被固定映射占用。
func smartLinkDirectoryForAgentExistsIndex(c *gin.Context, req define.AgentSmartLinkDirectoryRequest) {
	if req.UserDataIndex <= 0 {
		gsgin.GinResponseSuccess(c, "", define.AgentSmartLinkDirectoryResponse{})
		return
	}
	row, err := common.DbMain.Client.QueryBySql(
		`select * from tbl_smart_link_directory_mapping where user_data_index = ? `,
		req.UserDataIndex,
	).One()
	if err != nil {
		gsgin.GinResponseSuccess(c, "", define.AgentSmartLinkDirectoryResponse{})
		return
	}
	gsgin.GinResponseSuccess(c, "", define.AgentSmartLinkDirectoryResponse{
		Exists: len(row) > 0,
	})
}

// smartLinkDirectoryForAgentUpsert 写入或更新固定目录映射关系。
func smartLinkDirectoryForAgentUpsert(c *gin.Context, req define.AgentSmartLinkDirectoryRequest) {
	if strings.TrimSpace(req.MappingKey) == "" || req.SmartLinkID <= 0 || req.UserDataIndex <= 0 {
		gsgin.GinResponseError(c, "参数不完整", nil)
		return
	}
	now := time.Now().Unix()
	row, err := common.DbMain.Client.QueryBySql(
		`select * from tbl_smart_link_directory_mapping where mapping_key = ?`,
		req.MappingKey,
	).One()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if len(row) > 0 {
		existingIndex := cast.ToInt(row[`user_data_index`])
		if existingIndex != req.UserDataIndex {
			occupied, occupiedErr := directoryIndexOccupiedByOtherMapping(req.MappingKey, req.UserDataIndex)
			if occupiedErr != nil {
				gsgin.GinResponseError(c, occupiedErr.Error(), nil)
				return
			}
			if occupied {
				gsgin.GinResponseError(c, "目录索引已被占用", nil)
				return
			}
		}
		_, err = common.DbMain.Client.QuickUpdate(`tbl_smart_link_directory_mapping`, map[string]any{
			`mapping_key`: req.MappingKey,
		}, map[string]any{
			`smart_link_id`:   req.SmartLinkID,
			`label`:           req.Label,
			`account_key`:     req.AccountKey,
			`user_data_index`: req.UserDataIndex,
			`update_time`:     now,
		}).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		gsgin.GinResponseSuccess(c, "", define.AgentSmartLinkDirectoryResponse{})
		return
	}

	occupied, occupiedErr := directoryIndexOccupiedByOtherMapping(req.MappingKey, req.UserDataIndex)
	if occupiedErr != nil {
		gsgin.GinResponseError(c, occupiedErr.Error(), nil)
		return
	}
	if occupied {
		gsgin.GinResponseError(c, "目录索引已被占用", nil)
		return
	}

	_, err = common.DbMain.Client.QuickCreate(`tbl_smart_link_directory_mapping`, map[string]any{
		`mapping_key`:     req.MappingKey,
		`smart_link_id`:   req.SmartLinkID,
		`label`:           req.Label,
		`account_key`:     req.AccountKey,
		`user_data_index`: req.UserDataIndex,
		`create_time`:     now,
		`update_time`:     now,
	}).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", define.AgentSmartLinkDirectoryResponse{})
}

// directoryIndexOccupiedByOtherMapping 判断指定目录索引是否被其他 mapping_key 占用。
func directoryIndexOccupiedByOtherMapping(mappingKey string, userDataIndex int) (bool, error) {
	row, err := common.DbMain.Client.QueryBySql(
		`select mapping_key from tbl_smart_link_directory_mapping where user_data_index = ? `,
		userDataIndex,
	).One()
	if err != nil {
		return false, err
	}
	if len(row) == 0 {
		return false, nil
	}
	return cast.ToString(row[`mapping_key`]) != mappingKey, nil
}
