-- 老数据迁移：创建默认模板 + 关联老任务 + 迁移 step_prompts JSON
-- 注意：此脚本依赖 tbl_home_task_config 中的旧全局提示词配置
-- 执行时机：在 DDL（建表/加字段）之后执行

-- 1. 创建默认模板
INSERT INTO "tbl_workflow_template" ("name", "description", "is_default", "sort_order", "create_time", "update_time")
VALUES ('默认模板', '系统自动创建的默认工作流程模板（从旧版全局配置迁移）', 1, 0, strftime('%s','now'), strftime('%s','now'));

-- 2. 从全局配置读取旧提示词并创建模板步骤
-- 当前 WORKFLOW_NODES 顺序：task-config → requirement-fetch → requirement → design → api-dev → api-test-fix → code-review → browser-test
-- 固定步骤：task-config、requirement-fetch、issue_fix（issue_fix 放在最后）

-- 2.1 任务配置（固定步骤，无提示词）
INSERT INTO "tbl_workflow_template_step" ("template_id", "name", "step_key", "prompt_content", "sort_order", "is_fixed", "create_time", "update_time")
SELECT 1, '任务配置', 'task-config', '', 0, 1, strftime('%s','now'), strftime('%s','now');

-- 2.2 抓取需求（固定步骤，无提示词）
INSERT INTO "tbl_workflow_template_step" ("template_id", "name", "step_key", "prompt_content", "sort_order", "is_fixed", "create_time", "update_time")
SELECT 1, '抓取需求', 'requirement-fetch', '', 1, 1, strftime('%s','now'), strftime('%s','now');

-- 2.3 需求分析（提示词来自 home_task_prompt_dev）
INSERT INTO "tbl_workflow_template_step" ("template_id", "name", "step_key", "prompt_content", "sort_order", "is_fixed", "create_time", "update_time")
SELECT 1, '需求分析', 'requirement', COALESCE((SELECT "value" FROM "tbl_home_task_config" WHERE "key" = 'home_task_prompt_dev'), ''), 2, 0, strftime('%s','now'), strftime('%s','now');

-- 2.4 开发执行（提示词来自 home_task_prompt_design）
INSERT INTO "tbl_workflow_template_step" ("template_id", "name", "step_key", "prompt_content", "sort_order", "is_fixed", "create_time", "update_time")
SELECT 1, '开发执行', 'design', COALESCE((SELECT "value" FROM "tbl_home_task_config" WHERE "key" = 'home_task_prompt_design'), ''), 3, 0, strftime('%s','now'), strftime('%s','now');

-- 2.5 接口生成（提示词来自 home_task_prompt_api_gen）
INSERT INTO "tbl_workflow_template_step" ("template_id", "name", "step_key", "prompt_content", "sort_order", "is_fixed", "create_time", "update_time")
SELECT 1, '接口生成', 'api-dev', COALESCE((SELECT "value" FROM "tbl_home_task_config" WHERE "key" = 'home_task_prompt_api_gen'), ''), 4, 0, strftime('%s','now'), strftime('%s','now');

-- 2.6 自动化测试+修复（提示词来自 home_task_prompt_api_test）
INSERT INTO "tbl_workflow_template_step" ("template_id", "name", "step_key", "prompt_content", "sort_order", "is_fixed", "create_time", "update_time")
SELECT 1, '自动化测试+修复', 'api-test-fix', COALESCE((SELECT "value" FROM "tbl_home_task_config" WHERE "key" = 'home_task_prompt_api_test'), ''), 5, 0, strftime('%s','now'), strftime('%s','now');

-- 2.7 代码检查（提示词来自 home_task_prompt_code_review）
INSERT INTO "tbl_workflow_template_step" ("template_id", "name", "step_key", "prompt_content", "sort_order", "is_fixed", "create_time", "update_time")
SELECT 1, '代码检查', 'code-review', COALESCE((SELECT "value" FROM "tbl_home_task_config" WHERE "key" = 'home_task_prompt_code_review'), ''), 6, 0, strftime('%s','now'), strftime('%s','now');

-- 2.8 需求核对浏览器测试（提示词来自 home_task_prompt_browser_test）
INSERT INTO "tbl_workflow_template_step" ("template_id", "name", "step_key", "prompt_content", "sort_order", "is_fixed", "create_time", "update_time")
SELECT 1, '需求核对浏览器测试', 'browser-test', COALESCE((SELECT "value" FROM "tbl_home_task_config" WHERE "key" = 'home_task_prompt_browser_test'), ''), 7, 0, strftime('%s','now'), strftime('%s','now');

-- 2.9 问题修改（固定步骤，放在最后）
INSERT INTO "tbl_workflow_template_step" ("template_id", "name", "step_key", "prompt_content", "sort_order", "is_fixed", "create_time", "update_time")
SELECT 1, '问题修改', 'issue_fix', COALESCE((SELECT "value" FROM "tbl_home_task_config" WHERE "key" = 'home_task_prompt_issue_fix'), ''), 8, 1, strftime('%s','now'), strftime('%s','now');

-- 3. 将所有已有任务的 workflow_template_id 设为默认模板ID（1）
UPDATE "tbl_home_task" SET "workflow_template_id" = 1 WHERE "workflow_template_id" = 0;

-- 4. 迁移已有工作流实例的 prompt_xxx 字段值到 step_prompts JSON
-- 使用 SQLite 的 json_object 函数构建 JSON，key 与 step_key 对应
UPDATE "tbl_task_workflow" SET "step_prompts" = json_object(
    'requirement',             COALESCE("prompt_requirement", ''),
    'api-dev',                 COALESCE("prompt_api_dev", ''),
    'api-test-fix',            COALESCE("prompt_api_test", ''),
    'design',                  COALESCE("prompt_design", ''),
    'plain_text_requirement',  COALESCE("prompt_plain_text_requirement", ''),
    'design_plan_requirement', COALESCE("prompt_design_plan_requirement", ''),
    'browser-test',            COALESCE("prompt_browser_test", ''),
    'code-review',             COALESCE("prompt_code_review", ''),
    'issue_fix',               ''
) WHERE "step_prompts" = '';
