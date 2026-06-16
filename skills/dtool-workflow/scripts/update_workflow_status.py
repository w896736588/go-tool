"""
更新工作流程任务步骤状态。

用法:
    from update_workflow_status import update_workflow_status
    update_workflow_status(base_url='http://localhost:17170', token='temptoken', workflow_id=69, step='requirement', status='running')

step 值说明：
    - step 由工作流模板动态定义，后端会校验合法性
    - 自定义步骤格式: custom_{id} (如 custom_10)
"""

import os
import sys
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '../../dtool-common/scripts'))

from api_common import call_api


def update_workflow_status(base_url, token, workflow_id, step, status):
    """更新工作流程指定步骤的状态。

    Args:
        base_url: dtool 服务地址
        token: 认证令牌
        workflow_id: 工作流程 ID
        step: 步骤 key（由工作流模板定义，自定义步骤 custom_{id} 均支持）
        status: 状态值（后端校验合法性，无需前端校验）
    """
    # 同步全局配置
    import api_common
    api_common.BASE_URL = base_url
    api_common.TOKEN = token

    result = call_api("/api/task/workflow/node-status/update", {
        "workflow_id": workflow_id,
        "step": step,
        "status": status,
    })
    if result.get("code") != 0:
        raise RuntimeError(result.get("msg", "更新节点状态失败"))
    return result
