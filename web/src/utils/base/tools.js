import base from '@/utils/base'

function ToolPortProcessList(data, callBack) {
  base.BasePost('/api/ToolPortProcessList', data, callBack)
}

function ToolPortProcessKill(data, callBack) {
  base.BasePost('/api/ToolPortProcessKill', data, callBack)
}

export default {
  ToolPortProcessList,
  ToolPortProcessKill,
}
