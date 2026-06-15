-- tbl_workflow_template_step 新增 step_documents JSON 字段，存储步骤预生成知识片段配置
ALTER TABLE "tbl_workflow_template_step" ADD COLUMN "step_documents" TEXT NOT NULL DEFAULT '';

-- tbl_task_workflow 新增 step_fragment_refs JSON 字段，存储步骤预生成知识片段的引用
ALTER TABLE "tbl_task_workflow" ADD COLUMN "step_fragment_refs" TEXT NOT NULL DEFAULT '';
