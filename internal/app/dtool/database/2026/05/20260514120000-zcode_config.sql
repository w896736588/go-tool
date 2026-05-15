CREATE TABLE "tbl_zcode_config" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "zcode_dir" TEXT NOT NULL DEFAULT '',
    "created_at" TEXT NOT NULL DEFAULT '',
    "updated_at" TEXT NOT NULL DEFAULT ''
);

CREATE TABLE "tbl_zcode_project_mapping" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "zcode_config_id" INTEGER NOT NULL DEFAULT 0,
    "project_key" TEXT NOT NULL DEFAULT '',
    "workspace_path" TEXT NOT NULL DEFAULT '',
    "settings_path" TEXT NOT NULL DEFAULT ''
);
