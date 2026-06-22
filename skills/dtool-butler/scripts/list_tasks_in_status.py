#!/usr/bin/env python3
"""
列出任务清单中处于指定状态的任务（如"自测中"）。

通过调用 dtool HTTP API 实现，支持自定义基地址、Token、API路径和状态名称。
默认使用 /api/HomeTaskList 端点。仅依赖标准库，无需安装第三方包。

用法:
    python list_tasks_in_status.py --token YOUR_TOKEN
    python list_tasks_in_status.py --token TOKEN --status "自测中" --api-path /api/HomeTaskList
"""

import argparse
import json
import sys
import io
from urllib.parse import urljoin
from urllib import request

# 解决 Windows 编码问题
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')


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
        default="",
        help="API 鉴权 Token（本机调用可留空）",
    )
    parser.add_argument(
        "--status",
        default="自测中",
        help="要过滤的任务状态（默认：自测中）",
    )
    parser.add_argument(
        "--api-path",
        default="/api/HomeTaskList",
        help="获取任务列表的 API 路径（默认 /api/HomeTaskList）",
    )
    return parser.parse_args()


def main():
    args = parse_args()
    url = urljoin(args.server, args.api_path)

    try:
        # 使用 POST + Token 头（dtool API 统一鉴权方式）
        body = json.dumps({}).encode("utf-8")
        req = request.Request(
            url=url,
            data=body,
            headers={
                "Content-Type": "application/json; charset=utf-8",
                "Token": args.token,
            },
            method="POST",
        )
        with request.urlopen(req, timeout=60) as resp:
            data = json.loads(resp.read().decode("utf-8"))
    except Exception as e:
        result = {
            "error": True,
            "message": f"HTTP 请求失败: {str(e)}",
        }
        print(json.dumps(result, ensure_ascii=False, indent=2))
        sys.exit(1)

    # 智能提取任务列表：处理 dtool API 常见响应结构
    tasks = None
    if isinstance(data, list):
        tasks = data
    elif isinstance(data, dict):
        # dtool 标准响应: {"ErrCode":0, "Data": {"task_list": [...]}}
        if "Data" in data and isinstance(data["Data"], dict):
            inner = data["Data"]
            # 优先取 task_list
            for key in ("task_list", "list", "tasks", "items", "records"):
                if key in inner and isinstance(inner[key], list):
                    tasks = inner[key]
                    break
        # 兜底：直接在当前层级查找
        if tasks is None:
            for key in ("data", "Data", "task_list", "results", "items", "list", "records"):
                if key in data and isinstance(data[key], list):
                    tasks = data[key]
                    break
        # 最后遍历嵌套 dict 的值
        if tasks is None:
            for v in data.values():
                if isinstance(v, list):
                    tasks = v
                    break
                if isinstance(v, dict):
                    for vv in v.values():
                        if isinstance(vv, list):
                            tasks = vv
                            break
                    if tasks is not None:
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