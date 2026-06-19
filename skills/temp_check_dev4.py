#!/usr/bin/env python3
"""查找 dev4 相关信息"""
import sys, os
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), 'dtool-common/scripts'))
from api_common import call_api

# 获取全部 Git 列表
result = call_api("/api/Set/GitList", {})
if result.get("code") == 0:
    data = result.get("data", [])
    print(f"=== 共 {len(data)} 个 Git 仓库 ===")
    for item in data:
        name = item.get("name", "")
        if "dev" in name.lower() or "dev4" in name.lower():
            print(f"找到匹配: id={item['id']}, name={name}, ssh_id={item['ssh_id']}, ssh_name={item.get('ssh_name','')}, code_path={item['code_path']}, group={item.get('git_group_name','')}")
    # 也列一下所有不重复的 name
    names = [item.get("name","") for item in data]
    print("\n所有仓库名称:")
    for n in names:
        print(f"  - {n}")
else:
    print(f"查询失败: {result}")

# 也检查一下首页任务
result2 = call_api("/api/HomeTaskList", {})
if result2.get("code") == 0:
    data2 = result2.get("data", {})
    print(f"\n=== 首页任务数据 ===")
    # Check git_list in home task
    git_list = data2.get("git_list", [])
    print(f"首页任务 git_list 数量: {len(git_list)}")
    for item in git_list:
        name = item.get("name", "")
        if "dev" in name.lower():
            print(f"  首页任务匹配: id={item['id']}, name={name}")
