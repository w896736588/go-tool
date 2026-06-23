#!/usr/bin/env python3
"""
查询首页任务列表中指定状态的任务。

用法:
    python query_tasks_by_status.py
    python query_tasks_by_status.py --status "上线中"
"""

import json
import sys
import io
from urllib.parse import urljoin
from urllib import request

sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

SERVER = "http://localhost:17170"
TOKEN = ""


def call_api(path, body=None):
    if body is None:
        body = {}
    url = urljoin(SERVER, path)
    data = json.dumps(body).encode("utf-8")
    req = request.Request(
        url=url,
        data=data,
        headers={
            "Content-Type": "application/json; charset=utf-8",
            "Token": TOKEN,
        },
        method="POST",
    )
    with request.urlopen(req, timeout=60) as resp:
        return json.loads(resp.read().decode("utf-8"))


def main():
    # 先获取任务数量
    count_resp = call_api("/api/HomeTaskCount")
    active_count = count_resp.get("Data", {}).get("active_count", 0)
    archived_count = count_resp.get("Data", {}).get("archived_count", 0)
    total_count = active_count + archived_count

    print(f"📊 总任务数: {total_count} (活跃: {active_count}, 归档: {archived_count})", file=sys.stderr)

    # 遍历ID范围获取所有任务详情
    # 先取一个较大的范围
    results = []
    for tid in range(1, total_count + 10):
        try:
            resp = call_api("/api/HomeTaskInfo", {"id": tid})
            if resp.get("ErrCode") == 0:
                task = resp.get("Data", {})
                if task and task.get("id"):
                    results.append(task)
        except Exception:
            pass

    # 按状态分组统计
    status_map = {}
    for task in results:
        status = task.get("task_status", "未知")
        if status not in status_map:
            status_map[status] = []
        status_map[status].append(task)

    print(f"\n📋 共找到 {len(results)} 个任务", file=sys.stderr)
    print(f"\n📊 任务状态分布:", file=sys.stderr)
    for status, tasks in sorted(status_map.items(), key=lambda x: -len(x[1])):
        icon = "🟢" if status in ("开发中", "自测中", "测试中", "上线中", "待测试", "待对接", "对接中") else "⚪"
        print(f"  {icon} {status}: {len(tasks)} 个", file=sys.stderr)

    # 输出主要结果：筛选"上线中"状态
    target_status = "上线中"
    filtered = [t for t in results if t.get("task_status") == target_status]

    output = {
        "total_count": len(results),
        "active_count": active_count,
        "archived_count": archived_count,
        "target_status": target_status,
        "filtered_count": len(filtered),
        "tasks": filtered,
    }
    print(json.dumps(output, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
