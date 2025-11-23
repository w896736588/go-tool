CREATE TABLE "tbl_api_env"
(
    "id"          integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "collection_id" integer NOT NULL DEFAULT 0,
    "name"        text    NOT NULL DEFAULT '',
    "desc"         text    NOT NULL DEFAULT '',
    "create_time" integer NOT NULL DEFAULT 0,
    "update_time" integer NOT NULL DEFAULT 0
);

CREATE TABLE "tbl_api_env_item"
(
    "id"          integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "collection_id" integer NOT NULL DEFAULT 0,
    "env_id"      integer NOT NULL DEFAULT 0,
    "key"         text    NOT NULL DEFAULT '',
    "value"       text    NOT NULL DEFAULT '',
    "desc"        text    NOT NULL DEFAULT '',
    "create_time" integer NOT NULL DEFAULT 0,
    "update_time" integer NOT NULL DEFAULT 0
);