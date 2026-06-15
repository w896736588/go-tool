-- 工作流文档独立表：统一存储 tbl_task_workflow 上所有知识片段引用
CREATE TABLE "tbl_task_workflow_document" (
    "id"                INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "workflow_id"       INTEGER NOT NULL DEFAULT 0,
    "document_id"       TEXT    NOT NULL DEFAULT '',
    "document_name"     TEXT    NOT NULL DEFAULT '',
    "document_type"     TEXT    NOT NULL DEFAULT '',
    "template_id"       INTEGER NOT NULL DEFAULT 0,
    "template_step_id"  INTEGER NOT NULL DEFAULT 0,
    "file_id"           TEXT    NOT NULL DEFAULT '',
    "folder_name"       TEXT    NOT NULL DEFAULT '',
    "placeholder"       TEXT    NOT NULL DEFAULT '',
    "create_time"       INTEGER NOT NULL DEFAULT 0,
    "update_time"       INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX "idx_task_workflow_document_workflow"
    ON "tbl_task_workflow_document" ("workflow_id" ASC);

CREATE INDEX "idx_task_workflow_document_step"
    ON "tbl_task_workflow_document" ("workflow_id" ASC, "template_step_id" ASC);
