-- 工作流程模板步骤表
CREATE TABLE "tbl_workflow_template_step" (
    "id"             INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "template_id"    INTEGER NOT NULL DEFAULT 0,
    "name"           TEXT NOT NULL DEFAULT '',
    "step_key"       TEXT NOT NULL DEFAULT '',
    "prompt_content" TEXT NOT NULL DEFAULT '',
    "sort_order"     INTEGER NOT NULL DEFAULT 0,
    "is_fixed"       INTEGER NOT NULL DEFAULT 0,
    "create_time"    INTEGER NOT NULL DEFAULT 0,
    "update_time"    INTEGER NOT NULL DEFAULT 0
);
