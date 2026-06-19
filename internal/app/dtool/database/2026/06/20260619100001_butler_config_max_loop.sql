-- 管家运行参数增加 Loop 次数上限设置，默认 10 次
ALTER TABLE tbl_butler_config ADD COLUMN max_loop INTEGER NOT NULL DEFAULT 10;
