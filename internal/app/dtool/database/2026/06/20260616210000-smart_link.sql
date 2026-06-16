-- 创建新表 smart_link，将老表 tbl_smart_link 中 links JSON 展开为独立行
-- 将老表的 name 作为分组名自动创建 tbl_group
-- 实际的数据迁移（links JSON 展开）由 SmartLinkMigrateOldData Go 函数完成，
-- 因为纯 Go SQLite 实现的 json_each 支持不稳定。
-- Create new smart_link table, data migration handled by Go function SmartLinkMigrateOldData

CREATE TABLE IF NOT EXISTS "smart_link"
(
    "id"                    INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "label"                 TEXT    NOT NULL DEFAULT '',
    "link"                  TEXT    NOT NULL DEFAULT '',
    "smart_link_group_id"   INTEGER NOT NULL DEFAULT 0,
    "account_list"          TEXT    NOT NULL DEFAULT '',
    "browser_auth_username" TEXT    NOT NULL DEFAULT '',
    "browser_auth_password" TEXT    NOT NULL DEFAULT '',
    "cookie"                TEXT    NOT NULL DEFAULT '',
    "headers"               TEXT    NOT NULL DEFAULT '',
    "open_num"              INTEGER,
    "open_type"             TEXT,
    "process"               TEXT,
    "weight"                integer,
    "combine_type"          integer,
    "status"                integer          DEFAULT 1,
    "value"                 TEXT,
    "create_time"           integer,
    "update_time"           integer,
    "download_finds"        TEXT             DEFAULT '',
    "auto_close_second"     integer          DEFAULT 0,
    "channel"               TEXT             DEFAULT '',
    "show_cookies"          TEXT             DEFAULT '',
    "process_id"            INTEGER NOT NULL DEFAULT 0,
    "filter_uris"           TEXT    NOT NULL DEFAULT ''
);

-- 将老表 name 作为分组名创建 tbl_group
-- Create groups from old table names
INSERT OR IGNORE INTO tbl_group (name, type, create_time, update_time)
SELECT DISTINCT old.name, 4, MIN(old.create_time), MIN(old.update_time)
FROM tbl_smart_link old
WHERE old.status = 1
  AND old.name IS NOT NULL
  AND old.name != ''
GROUP BY old.name;
