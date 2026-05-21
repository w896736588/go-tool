-- chat 表直接存储 prompt_type 和 cli_type，不再通过 tbl_task_workflow 的 JSON 列映射
ALTER TABLE "tbl_task_workflow_chat" ADD COLUMN "prompt_type" TEXT NOT NULL DEFAULT '';
ALTER TABLE "tbl_task_workflow_chat" ADD COLUMN "cli_type" TEXT NOT NULL DEFAULT '';

-- 删除 tbl_task_workflow 中废弃的 chat_session_ids JSON 列
ALTER TABLE "tbl_task_workflow" DROP COLUMN "prompt_plain_text_requirement_chat_session_ids";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "prompt_requirement_chat_session_ids";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "prompt_design_plan_requirement_chat_session_ids";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "prompt_design_chat_session_ids";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "prompt_api_dev_chat_session_ids";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "prompt_code_review_chat_session_ids";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "prompt_browser_test_chat_session_ids";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "prompt_api_test_chat_session_ids";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "prompt_issue_fix_chat_session_ids";
