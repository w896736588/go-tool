-- tbl_agent_cli: Agent CLI 实例管理
CREATE TABLE IF NOT EXISTS "tbl_agent_cli" (
    "id"              INTEGER PRIMARY KEY AUTOINCREMENT,
    "name"            TEXT NOT NULL DEFAULT '',
    "type"            TEXT NOT NULL DEFAULT 'claude-code-cli',
    "settings_path"   TEXT NOT NULL DEFAULT '',
    "created_at"      INTEGER NOT NULL DEFAULT 0,
    "updated_at"      INTEGER NOT NULL DEFAULT 0
);
