-- tbl_webhook_config: Webhook 通知配置（钉钉/飞书/企微等）
CREATE TABLE IF NOT EXISTS "tbl_webhook_config" (
    "id"           INTEGER PRIMARY KEY AUTOINCREMENT,
    "name"         TEXT NOT NULL DEFAULT '',
    "type"         TEXT NOT NULL DEFAULT 'dingtalk',
    "webhook_url"  TEXT NOT NULL DEFAULT '',
    "secret"       TEXT NOT NULL DEFAULT '',
    "created_at"   INTEGER NOT NULL DEFAULT 0,
    "updated_at"   INTEGER NOT NULL DEFAULT 0
);
