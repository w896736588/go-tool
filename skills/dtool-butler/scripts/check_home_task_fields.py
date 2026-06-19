#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""查询任务清单中状态为「自测中」的任务"""
import json
from urllib import error, request

BASE_URL = "http://localhost:17170"
TOKEN = ""

def call_api(path, payload):
    body = json.dumps(payload).encode("utf-8")
    req = request.Request(
        url=f"{BASE_URL}{path}",
        data=body,
        headers={"Content-Type": "application/json; charset=utf-8", "Token": TOKEN},
        method="POST",
    )
    with request.urlopen(req, timeout=60) as resp:
        return json.loads(resp.read().decode("utf-8"))

result = call_api("/api/HomeTaskList", {})
data = result.get("Data", {})

# 查找所有包含 task 或 home_task 的字段
tasks_found = []
for key, value in data.items():
    if isinstance(value, list) and len(value) > 0:
        # 检查列表中是否有包含 status/状态 字段的对象
        for item in value[:5]:
            if isinstance(item, dict) and any(k in item for k in ["status", "状态", "task_status", "home_task_status", "task_name", "name", "title"]):
                tasks_found.append(key)
                break

# 输出所有 key 名
print("=== 返回数据的顶层字段 ===")
print(json.dumps(list(data.keys()), ensure_ascii=False, indent=2))

# 看看哪些字段可能是任务
for key in data.keys():
    val = data[key]
    if isinstance(val, list):
        if len(val) > 0 and isinstance(val[0], dict):
            print(f"\n=== 字段 [{key}] 的字段列表 ({len(val)} 条) ===")
            print(list(val[0].keys()))
        else:
            print(f"\n=== 字段 [{key}] 的类型 ===")
            print(f"list of {type(val[0]) if val else 'empty'}")
    elif isinstance(val, dict):
        print(f"\n=== 字段 [{key}] 的子字段 ===")
        print(list(val.keys()))
    else:
        print(f"\n=== 字段 [{key}] ===")
        print(val)
