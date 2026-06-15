-- tbl_workflow_template_step 新增 remark 字段，存储步骤备注
ALTER TABLE "tbl_workflow_template_step" ADD COLUMN "remark" TEXT NOT NULL DEFAULT '';
