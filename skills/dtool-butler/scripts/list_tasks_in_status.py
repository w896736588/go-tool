#!/usr/bin/env python3
"""
列出任务清单中处于指定状态的任务（如“自测中”）。

通过调用 dtool HTTP API 实现，支持自定义基地址、Token、API路径和状态名称。
默认使用 /api/TaskList 端点，可覆盖。

用法:
    python list_tasks_in_status.py --token YOUR_TOKEN
    python list_tasks_in_status.py --token TOKEN --status "自测中" --api-path /api/HomeTaskList
"""

import argparse
import json
import sys
from urllib.parse import urljoin

import requests


def parse_args():
    parser = argparse.ArgumentParser(
        description="从任务清单中过滤出处于指定状态的任务"
    )
    parser.add_argument(
        "--server",
        default="http://localhost:17170",
        help="dtool API 基地址（默认 http://localhost:17170）",
    )
    parser.add_argument(
        "--token",
        required=True,
        help="API 鉴权 Token",
    )
    parser.add_argument(
        "--status",
        default="自测中",
        help="要过滤的任务状态（默认：自测中）",
    )
    parser.add_argument(
        "--api-path",
        default="/api/TaskList",
        help="获取任务列表的 API 路径（默认 /api/TaskList）",
    )
    return parser.parse_args()


def main():
    args = parse_args()
    url = urljoin(args.server, args.api_path)
    headers = {"Authorization": f"Bearer {args.token}"}

    try:
        resp = requests.get(url, headers=headers, timeout=30)
        resp.raise_for_status()
        data = resp.json()
    except requests.exceptions.RequestException as e:
        result = {
            "error": True,
            "message": f"HTTP 请求失败: {str(e)}",
        }
        print(json.dumps(result, ensure_ascii=False, indent=2))
        sys.exit(1)
    except json.JSONDecodeError as e:
        result = {
            "error": True,
            "message": f"响应不是有效的 JSON: {str(e)}",
        }
        print(json.dumps(result, ensure_ascii=False, indent=2))
        sys.exit(1)

    # 智能提取任务列表：处理常见的响应结构
    tasks = None
    if isinstance(data, list):
        tasks = data
    elif isinstance(data, dict):
        # 尝试多种可能的 key
        for key in ("data", "results", "items", "list", "records", "taskList"):
            if key in data and isinstance(data[key], list):
                tasks = data[key]
                break
        if tasks is None:
            # 遍历所有值，寻找第一个列表
            for v in data.values():
                if isinstance(v, list):
                    tasks = v
                    break

    if tasks is None:
        result = {
            "error": True,
            "message": "无法从 API 响应中提取任务列表。响应结构："
                       + json.dumps(data, ensure_ascii=False)[:500],
        }
        print(json.dumps(result, ensure_ascii=False, indent=2))
        sys.exit(1)

    # 过滤出指定状态的任务
    # 状态字段通常为 status / state / taskStatus，尝试多个常见命名
    status_keys = ("status", "state", "taskStatus", "task_state", "Status")
    filtered = []
    for task in tasks:
        if not isinstance(task, dict):
            continue
        task_status = None
        for key in status_keys:
            if key in task:
                task_status = task[key]
                break
        if task_status and task_status == args.status:
            filtered.append(task)

    # 输出结果
    result = {
        "total_count": len(tasks),
        "filtered_count": len(filtered),
        "status_filter": args.status,
        "tasks": filtered,
    }
    print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()