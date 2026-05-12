CREATE TABLE IF NOT EXISTS "tbl_memory_config"
(
    "id"          integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "key"         text    NOT NULL DEFAULT '',
    "value"       text    NOT NULL DEFAULT '',
    "name"        text    NOT NULL DEFAULT '',
    "desc"        text    NOT NULL DEFAULT '',
    "create_time" integer NOT NULL DEFAULT 0,
    "update_time" integer NOT NULL DEFAULT 0
);
