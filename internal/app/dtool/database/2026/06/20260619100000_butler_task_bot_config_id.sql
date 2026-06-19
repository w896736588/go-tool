-- 管家任务记录增加 bot_config_id 列，用于按机器人筛选 Loop 日志
ALTER TABLE tbl_butler_task ADD COLUMN bot_config_id INTEGER NOT NULL DEFAULT 0;
