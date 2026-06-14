-- 工作流程模板表
CREATE TABLE "tbl_workflow_template" (
    "id"          INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"        TEXT NOT NULL DEFAULT '',
    "description" TEXT NOT NULL DEFAULT '',
    "is_default"  INTEGER NOT NULL DEFAULT 0,
    "sort_order"  INTEGER NOT NULL DEFAULT 0,
    "create_time" INTEGER NOT NULL DEFAULT 0,
    "update_time" INTEGER NOT NULL DEFAULT 0
);
