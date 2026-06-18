import base from "@/utils/base";

// ButlerBotConfigList 查询管家机器人配置列表
function ButlerBotConfigList(callBack) {
    base.BasePost('/api/Set/ButlerBotConfigList', {}, callBack)
}

// ButlerBotConfigAdd 新增或更新管家机器人配置
function ButlerBotConfigAdd(data, callBack) {
    base.BasePost('/api/Set/ButlerBotConfigAdd', data, callBack)
}

// ButlerBotConfigDelete 删除管家机器人配置
function ButlerBotConfigDelete(data, callBack) {
    base.BasePost('/api/Set/ButlerBotConfigDelete', data, callBack)
}

// ButlerBotConfigTest 测试管家机器人配置（发送测试消息）
function ButlerBotConfigTest(data, callBack) {
    base.BasePost('/api/Set/ButlerBotConfigTest', data, callBack)
}

// ButlerMessageList 查询管家机器人消息日志（分页）
function ButlerMessageList(data, callBack) {
    base.BasePost('/api/Set/ButlerMessageList', data, callBack)
}

// ButlerRoleList 查询管家角色列表
function ButlerRoleList(callBack) {
    base.BasePost('/api/Set/ButlerRoleList', {}, callBack)
}

// ButlerRoleAdd 新增或更新管家角色
function ButlerRoleAdd(data, callBack) {
    base.BasePost('/api/Set/ButlerRoleAdd', data, callBack)
}

// ButlerRoleDelete 删除管家角色
function ButlerRoleDelete(data, callBack) {
    base.BasePost('/api/Set/ButlerRoleDelete', data, callBack)
}

// ButlerConfigList 查询管家运行参数列表
function ButlerConfigList(callBack) {
    base.BasePost('/api/Set/ButlerConfigList', {}, callBack)
}

// ButlerConfigAdd 新增或更新管家运行参数
function ButlerConfigAdd(data, callBack) {
    base.BasePost('/api/Set/ButlerConfigAdd', data, callBack)
}

// ButlerConfigDelete 删除管家运行参数
function ButlerConfigDelete(data, callBack) {
    base.BasePost('/api/Set/ButlerConfigDelete', data, callBack)
}

export default {
    ButlerBotConfigList,
    ButlerBotConfigAdd,
    ButlerBotConfigDelete,
    ButlerBotConfigTest,
    ButlerMessageList,
    ButlerRoleList,
    ButlerRoleAdd,
    ButlerRoleDelete,
    ButlerConfigList,
    ButlerConfigAdd,
    ButlerConfigDelete,
}
