CREATE TABLE IF NOT EXISTS "tbl_smart_link_directory_mapping" (
    "id"                INTEGER PRIMARY KEY AUTOINCREMENT,
    "mapping_key"       TEXT NOT NULL DEFAULT '',
    "smart_link_id"     INTEGER NOT NULL DEFAULT 0,
    "label"             TEXT NOT NULL DEFAULT '',
    "account_key"       TEXT NOT NULL DEFAULT '',
    "user_data_index"   INTEGER NOT NULL DEFAULT 0,
    "create_time"       INTEGER NOT NULL DEFAULT 0,
    "update_time"       INTEGER NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS "idx_tbl_smart_link_directory_mapping_mapping_key"
    ON "tbl_smart_link_directory_mapping" ("mapping_key");

CREATE UNIQUE INDEX IF NOT EXISTS "idx_tbl_smart_link_directory_mapping_user_data_index"
    ON "tbl_smart_link_directory_mapping" ("user_data_index");
