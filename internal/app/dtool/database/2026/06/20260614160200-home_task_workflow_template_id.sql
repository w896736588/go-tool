-- tbl_home_task 新增 workflow_template_id 字段，关联工作流程模板
ALTER TABLE "tbl_home_task" ADD COLUMN "workflow_template_id" INTEGER NOT NULL DEFAULT 0;
