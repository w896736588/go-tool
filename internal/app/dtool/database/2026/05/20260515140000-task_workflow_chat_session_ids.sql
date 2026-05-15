ALTER TABLE "tbl_task_workflow" ADD COLUMN "prompt_plain_text_requirement_chat_session_ids" TEXT NOT NULL DEFAULT '';
ALTER TABLE "tbl_task_workflow" ADD COLUMN "prompt_requirement_chat_session_ids" TEXT NOT NULL DEFAULT '';
ALTER TABLE "tbl_task_workflow" ADD COLUMN "prompt_design_plan_requirement_chat_session_ids" TEXT NOT NULL DEFAULT '';
ALTER TABLE "tbl_task_workflow" ADD COLUMN "prompt_design_chat_session_ids" TEXT NOT NULL DEFAULT '';
ALTER TABLE "tbl_task_workflow" ADD COLUMN "prompt_api_dev_chat_session_ids" TEXT NOT NULL DEFAULT '';
ALTER TABLE "tbl_task_workflow" ADD COLUMN "prompt_code_review_chat_session_ids" TEXT NOT NULL DEFAULT '';
ALTER TABLE "tbl_task_workflow" ADD COLUMN "prompt_browser_test_chat_session_ids" TEXT NOT NULL DEFAULT '';
ALTER TABLE "tbl_task_workflow" ADD COLUMN "prompt_api_test_chat_session_ids" TEXT NOT NULL DEFAULT '';
