#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""查询任务清单中状态为「自测中」的任务"""
import json
import sys
import io

# 解决 Windows 编码问题
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

from urllib import request

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
task_list = data.get("task_list", [])

print(f"任务总数: {len(task_list)}")
print("=" * 80)

# 查找所有 task_status 中带有 "自测中" 的任务
self_testing_tasks = []
for task in task_list:
    status = task.get("task_status", "")
    if status and "自测中" in str(status):
        self_testing_tasks.append(task)

if self_testing_tasks:
    print(f"\n状态为【自测中】的任务共 {len(self_testing_tasks)} 个:\n")
    for t in self_testing_tasks:
        print(f"  ID: {t.get('id')}")
        print(f"  名称: {t.get('name')}")
        print(f"  状态: {t.get('task_status')}")
        print(f"  创建时间: {t.get('create_time_desc', '')}")
        print(f"  更新时间: {t.get('update_time_desc', '')}")
        print(f"  GitID: {t.get('git_ids')}")
        print(f"  API集合: {t.get('api_collection_id')}")
        print("-" * 60)
else:
    print("\n没有找到状态为【自测中】的任务")

# 输出所有任务的名称和状态供参考
print("\n所有任务状态一览:")
print("=" * 80)
for t in task_list:
    status = t.get("task_status", "未知")
    print(f"  [{status}] {t.get('name')} (ID:{t.get('id')})")
