#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""dtool 知识片段相关接口示例"""

import os, sys
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '../../dtool-common/scripts'))

from api_common import call_api


def memory_fragment_update_by_id(id, content, workflow_id):
    """
    通过片段ID更新知识片段内容（不会修改标题）

    id 为知识片段的唯一标识。
    workflow_id （工作流ID）为必传参数，后端会校验片段是否属于该工作流。
    """
    result = call_api("/api/MemoryFragmentSaveById", {
        "workflow_id": int(workflow_id),
        "id": id,
        "content": content,
    })

    if result.get("code") == 0:
        data = result.get("data", {})
        print(f"更新成功: id={data.get('id')}, title={data.get('title')}")
    else:
        print(f"更新失败: {result.get('msg')}")
    return result


if __name__ == "__main__":
    print("=== dtool 知识片段 API 示例 ===\n")
    # memory_fragment_update_by_id(
    #     "uuid-string",
    #     "## 更新后的内容\n\n新的正文...",
    #     1,
    # )
