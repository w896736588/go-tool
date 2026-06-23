-- 管家配置增加管家类型字段：1=主管家 2=归档管家
ALTER TABLE tbl_butler_config ADD COLUMN butler_type INTEGER NOT NULL DEFAULT 1;

-- 管家归档队列表：主管家任务完成后将文件+对话提交至此，归档管家轮询处理
CREATE TABLE IF NOT EXISTS "tbl_butler_archive" (
    "id"            INTEGER PRIMARY KEY AUTOINCREMENT,
    "config_id"     INTEGER NOT NULL DEFAULT 0,
    "task_id"       INTEGER NOT NULL DEFAULT 0,
    "session_id"    TEXT    NOT NULL DEFAULT '',
    "files"         TEXT    NOT NULL DEFAULT '',
    "conversation"  TEXT    NOT NULL DEFAULT '',
    "status"        TEXT    NOT NULL DEFAULT 'pending',
    "log"           TEXT    NOT NULL DEFAULT '',
    "result"        TEXT    NOT NULL DEFAULT '',
    "result_file"   TEXT    NOT NULL DEFAULT '',
    "result_index"  TEXT    NOT NULL DEFAULT '',
    "created_at"    INTEGER NOT NULL DEFAULT 0,
    "updated_at"    INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS "idx_butler_archive_status" ON "tbl_butler_archive"("status", "id");
