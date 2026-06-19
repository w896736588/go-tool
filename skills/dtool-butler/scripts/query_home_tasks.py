#!/usr/bin/env python3
"""查询首页任务清单中指定状态的任务"""
import json
import sys
from urllib import request, error

BASE_URL = "http://localhost:17170"

def call_api(path, payload):
    body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
    req = request.Request(
        url=f"{BASE_URL}{path}",
        data=body,
        headers={"Content-Type": "application/json; charset=utf-8"},
        method="POST",
    )
    with request.urlopen(req, timeout=30) as resp:
        return json.loads(resp.read().decode("utf-8"))

# 获取首页任务列表（完整数据）
result = call_api("/api/HomeTaskList", {})
# 输出到文件避免编码问题
with open("home_task_list_output.json", "w", encoding="utf-8") as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
print("数据已写入 home_task_list_output.json")
