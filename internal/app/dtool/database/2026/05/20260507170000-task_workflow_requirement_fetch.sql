ALTER TABLE "tbl_task_workflow" ADD COLUMN "requirement_fetch_status" TEXT NOT NULL DEFAULT 'idle';
ALTER TABLE "tbl_task_workflow" ADD COLUMN "requirement_fetch_started_at" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "tbl_task_workflow" ADD COLUMN "requirement_fetch_finished_at" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "tbl_task_workflow" ADD COLUMN "requirement_fetch_error" TEXT NOT NULL DEFAULT '';
ALTER TABLE "tbl_task_workflow" ADD COLUMN "requirement_source_url" TEXT NOT NULL DEFAULT '';
