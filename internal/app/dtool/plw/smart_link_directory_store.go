package plw

import (
	"dev_tool/internal/app/dtool/common"
	"errors"
	"strings"
	"time"

	"github.com/spf13/cast"
)

var ErrSmartLinkDirectoryIndexOccupied = errors.New("smart-link directory index occupied")

// SmartLinkDirectoryStore hides where smart-link fixed directory mappings are persisted.
// Server mode uses sqlite; agent mode can inject another implementation later.
type SmartLinkDirectoryStore interface {
	GetByMappingKey(mappingKey string) (int, error)
	ExistsUserDataIndex(userDataIndex int) (bool, error)
	UpsertMapping(mappingKey string, smartLinkID int, label, accountKey string, userDataIndex int) error
}

type dbSmartLinkDirectoryStore struct{}

// NewDBSmartLinkDirectoryStore 创建服务端默认的 sqlite 固定目录映射存储实现。
func NewDBSmartLinkDirectoryStore() SmartLinkDirectoryStore {
	return dbSmartLinkDirectoryStore{}
}

// GetByMappingKey 从主库读取固定映射对应的目录索引。
func (dbSmartLinkDirectoryStore) GetByMappingKey(mappingKey string) (int, error) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		return 0, nil
	}
	sql := `select * from tbl_smart_link_directory_mapping where mapping_key = ? `
	row, err := common.DbMain.Client.QueryBySql(sql, mappingKey).One()
	if err != nil {
		return 0, err
	}
	return cast.ToInt(row[`user_data_index`]), nil
}

// ExistsUserDataIndex 从主库判断目录索引是否已被固定映射占用。
func (dbSmartLinkDirectoryStore) ExistsUserDataIndex(userDataIndex int) (bool, error) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		return false, nil
	}
	sql := `select * from tbl_smart_link_directory_mapping where user_data_index = ? `
	row, err := common.DbMain.Client.QueryBySql(sql, userDataIndex).One()
	if err != nil {
		return false, err
	}
	return len(row) > 0, nil
}

// UpsertMapping 在主库记录 smart-link 固定目录映射关系。
func (dbSmartLinkDirectoryStore) UpsertMapping(mappingKey string, smartLinkID int, label, accountKey string, userDataIndex int) error {
	if common.DbMain == nil || common.DbMain.Client == nil {
		return nil
	}
	row, err := common.DbMain.Client.QueryBySql(
		`select * from tbl_smart_link_directory_mapping where mapping_key = ?`,
		mappingKey,
	).One()
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	if len(row) > 0 {
		existingUserDataIndex := cast.ToInt(row[`user_data_index`])
		if existingUserDataIndex != userDataIndex {
			occupied, occupiedErr := isDirectoryIndexOccupiedByOtherMapping(mappingKey, userDataIndex)
			if occupiedErr != nil {
				return occupiedErr
			}
			if occupied {
				return ErrSmartLinkDirectoryIndexOccupied
			}
		}
		_, err = common.DbMain.Client.QuickUpdate(`tbl_smart_link_directory_mapping`, map[string]any{
			`mapping_key`: mappingKey,
		}, map[string]any{
			`smart_link_id`:   smartLinkID,
			`label`:           label,
			`account_key`:     accountKey,
			`user_data_index`: userDataIndex,
			`update_time`:     now,
		}).Exec()
		if err != nil && strings.Contains(err.Error(), `UNIQUE constraint failed: tbl_smart_link_directory_mapping.user_data_index`) {
			return ErrSmartLinkDirectoryIndexOccupied
		}
		return err
	}

	occupied, occupiedErr := isDirectoryIndexOccupiedByOtherMapping(mappingKey, userDataIndex)
	if occupiedErr != nil {
		return occupiedErr
	}
	if occupied {
		return ErrSmartLinkDirectoryIndexOccupied
	}

	_, err = common.DbMain.Client.QuickCreate(`tbl_smart_link_directory_mapping`, map[string]any{
		`mapping_key`:     mappingKey,
		`smart_link_id`:   smartLinkID,
		`label`:           label,
		`account_key`:     accountKey,
		`user_data_index`: userDataIndex,
		`create_time`:     now,
		`update_time`:     now,
	}).Exec()
	if err != nil && strings.Contains(err.Error(), `UNIQUE constraint failed: tbl_smart_link_directory_mapping.user_data_index`) {
		return ErrSmartLinkDirectoryIndexOccupied
	}
	return err
}

func isDirectoryIndexOccupiedByOtherMapping(mappingKey string, userDataIndex int) (bool, error) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		return false, nil
	}
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
