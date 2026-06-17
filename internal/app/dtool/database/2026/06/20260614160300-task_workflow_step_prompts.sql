-- tbl_task_workflow 新增 step_prompts JSON 字段，存储各步骤的提示词实例值
ALTER TABLE "tbl_task_workflow" ADD COLUMN "step_prompts" TEXT NOT NULL DEFAULT '';
