ALTER TABLE "tbl_home_task"
    ADD COLUMN "git_id" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "tbl_home_task"
    ADD COLUMN "api_dev_enabled" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "tbl_home_task"
    ADD COLUMN "api_collection_id" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "tbl_home_task"
    ADD COLUMN "api_dir_id" INTEGER NOT NULL DEFAULT 0;
