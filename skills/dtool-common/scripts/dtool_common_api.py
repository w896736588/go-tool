#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
dtool 通用工具 API 调用示例
包含：Git文件上传、MySQL表查询、MySQL表结构查询、MySQL查询

使用前请先向用户确认以下信息，替换下方占位值：
  - base_url: dtool 服务地址（如 http://192.168.1.100:17170）
  - token: 认证令牌
  - git_id: Git 配置 ID（用于文件上传）
  - mysql_id: MySQL 配置 ID（用于 MySQL 查询）
"""

import json
from urllib import request, error

# ============================================================
# 以下四个变量必须向用户确认后填入
# ============================================================
BASE_URL = "http://localhost:17170"  # TODO: 替换为用户提供的地址
TOKEN = ""                           # TODO: 替换为用户提供的 Token
GIT_ID = ""                          # TODO: 替换为用户提供的 Git 配置 ID（用于获取 SSH 远程连接信息和项目路径）
MYSQL_ID = ""                        # TODO: 替换为用户提供的 MySQL 配置 ID


def call_api(path, payload):
    """通用 API 调用函数"""
    body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
    req = request.Request(
        url=f"{BASE_URL}{path}",
        data=body,
        headers={"Content-Type": "application/json; charset=utf-8", "Token": TOKEN},
        method="POST",
    )
    try:
        with request.urlopen(req, timeout=60) as resp:
            result = json.loads(resp.read().decode("utf-8"))
            return result
    except error.HTTPError as exc:
        body_text = exc.read().decode("utf-8", errors="replace")
        return {"code": -1, "msg": f"HTTP {exc.code}", "data": body_text}
    except Exception as exc:
        return {"code": -1, "msg": str(exc), "data": None}


# ============================================================
# 1. 上传文件到远程项目
# 通过 git_id 获取 SSH 远程连接配置和项目路径，将本地文件传输到远程服务器
# ============================================================
def git_upload_file(local_file_path, upload_dir):
    """
    上传本地文件到远程项目目录

    通过 git_id 获取 SSH 远程连接配置（主机、端口、认证信息）和项目路径，
    将当前项目中的文件传输到远程服务器的指定目录。

    参数:
        local_file_path: 当前项目中要上传文件的绝对路径
        upload_dir: 相对于远程项目根目录的上传目录，如 "src/config"、"public/uploads"
    """
    result = call_api("/api/GitUploadFile", {
        "git_id": GIT_ID,
        "local_file_path": local_file_path,
        "upload_dir": upload_dir,
    })
    if result.get("code") == 0:
        data = result.get("data", {})
        print(f"上传成功: {data.get('remote_path')}")
        print(f"  文件名: {data.get('file_name')}")
        print(f"  大小: {data.get('file_size')} 字节")
    else:
        print(f"上传失败: {result.get('msg')}")
    return result


# ============================================================
# 2. 查询 MySQL 所有表
# ============================================================
def mysql_tables():
    """查询 MySQL 配置对应数据库的所有表"""
    result = call_api("/api/MysqlTables", {
        "mysql_id": MYSQL_ID,
    })
    if result.get("code") == 0:
        table_list = result.get("data", {}).get("list", [])
        print(f"共 {len(table_list)} 张表:")
        for t in table_list:
            name = t.get("table_name", "")
            comment = t.get("table_comment", "")
            print(f"  {name}" + (f"  -- {comment}" if comment else ""))
    else:
        print(f"查询失败: {result.get('msg')}")
    return result


# ============================================================
# 3. 查询 MySQL 表结构
# ============================================================
def mysql_table_structure(table_name):
    """
    查询 MySQL 表结构

    参数:
        table_name: 表名
    """
    result = call_api("/api/MysqlTableStructure", {
        "mysql_id": MYSQL_ID,
        "table_name": table_name,
    })
    if result.get("code") == 0:
        fields = result.get("data", {}).get("list", [])
        print(f"\n表 {table_name} 结构 ({len(fields)} 个字段):")
        print(f"  {'字段':<20} {'类型':<20} {'允许空':<6} {'键':<6} {'默认值':<10} {'备注'}")
        print(f"  {'-'*20} {'-'*20} {'-'*6} {'-'*6} {'-'*10} {'-'*20}")
        for f in fields:
            field = f.get("Field", "")
            ftype = f.get("Type", "")
            null = f.get("Null", "")
            key = f.get("Key", "")
            default = str(f.get("Default", ""))
            comment = f.get("Comment", "")
            print(f"  {field:<20} {ftype:<20} {null:<6} {key:<6} {default:<10} {comment}")
    else:
        print(f"查询失败: {result.get('msg')}")
    return result


# ============================================================
# 4. 执行 MySQL 查询（仅 SELECT）
# ============================================================
def mysql_query(sql):
    """
    执行 MySQL SELECT 查询

    参数:
        sql: SELECT 语句

    示例:
        mysql_query("SELECT * FROM users LIMIT 10")
        mysql_query("SELECT COUNT(*) AS total FROM orders")
    """
    result = call_api("/api/MysqlQuery", {
        "mysql_id": MYSQL_ID,
        "sql": sql,
    })
    if result.get("code") == 0:
        rows = result.get("data", {}).get("list", [])
        if not rows:
            print("查询结果为空")
        else:
            # 打印表格式结果
            columns = list(rows[0].keys())
            # 计算每列宽度
            widths = {}
            for col in columns:
                widths[col] = max(
                    len(str(col)),
                    max(len(str(row.get(col, ""))) for row in rows)
                )
            # 表头
            header = " | ".join(str(col).ljust(widths[col]) for col in columns)
            print(f"\n{header}")
            print("-" * len(header))
            # 数据行
            for row in rows:
                line = " | ".join(str(row.get(col, "")).ljust(widths[col]) for col in columns)
                print(line)
            print(f"\n共 {len(rows)} 行")
    else:
        print(f"查询失败: {result.get('msg')}")
    return result


# ============================================================
# 使用示例
# ============================================================
if __name__ == "__main__":
    # 检查配置
    if not TOKEN:
        print("请先设置 TOKEN（向用户确认后填入）")
        exit(1)

    print("=== dtool 通用工具 API 示例 ===\n")

    # 示例1: 上传文件到远程项目（需设置 GIT_ID）
    # git_upload_file("/home/user/config.yaml", "src/config")

    # 示例2: 查询 MySQL 所有表（需设置 MYSQL_ID）
    # mysql_tables()

    # 示例3: 查询表结构（需设置 MYSQL_ID）
    # mysql_table_structure("users")

    # 示例4: 执行 SELECT 查询（需设置 MYSQL_ID）
    # mysql_query("SELECT * FROM users LIMIT 10")
