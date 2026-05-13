CREATE TABLE "tbl_task_workflow_chat" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "workflow_id" INTEGER NOT NULL DEFAULT 0,
    "session_id" TEXT NOT NULL DEFAULT '',
    "prompt" TEXT NOT NULL DEFAULT '',
    "status" TEXT NOT NULL DEFAULT 'running',
    "raw_output" TEXT NOT NULL DEFAULT '',
    "created_at" TEXT NOT NULL DEFAULT '',
    "updated_at" TEXT NOT NULL DEFAULT ''
);
