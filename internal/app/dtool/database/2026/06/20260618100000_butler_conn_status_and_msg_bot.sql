-- 机器人配置表新增连接状态字段
ALTER TABLE "tbl_butler_bot_config" ADD COLUMN "conn_status" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "tbl_butler_bot_config" ADD COLUMN "conn_status_at" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "tbl_butler_bot_config" ADD COLUMN "conn_error" TEXT NOT NULL DEFAULT '';

-- 消息表新增 bot_config_id 字段，用于关联机器人配置
ALTER TABLE "tbl_butler_message" ADD COLUMN "bot_config_id" INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS "idx_butler_msg_bot_config" ON "tbl_butler_message"("bot_config_id", "id");
