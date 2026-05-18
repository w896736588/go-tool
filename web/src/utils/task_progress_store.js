// task_progress_store.js - Claude Code 任务进度集中管理
import { reactive, computed } from 'vue'

// 任务状态常量
const STATUS = {
  STARTED: 'started',
  RUNNING: 'running',
  COMPLETED: 'completed',
  FAILED: 'failed',
}

const state = reactive({
  _map: {},
})

function _makeKey(taskId) {
  // taskId 可能是字符串或数字，统一为字符串
  return String(taskId)
}

function _cleanMsg(msg) {
  const entry = {}
  if (msg.taskId !== undefined) entry.taskId = msg.taskId
  if (msg.description !== undefined) entry.description = msg.description
  if (msg.status !== undefined) entry.status = msg.status
  if (msg.usage) entry.usage = { ...msg.usage }
  if (msg.lastToolName !== undefined) entry.lastToolName = msg.lastToolName
  if (msg.uuid !== undefined) entry.uuid = msg.uuid
  if (msg.sessionId !== undefined) entry.sessionId = msg.sessionId
  return entry
}

function _get(key) {
  if (!state._map[key]) {
    state._map[key] = {}
  }
  return state._map[key]
}

const summary = computed(() => {
  let total = 0
  let running = 0
  let completed = 0
  let failed = 0
  for (const key of Object.keys(state._map)) {
    total++
    const t = state._map[key]
    const s = t.status || ''
    if (s === STATUS.STARTED || s === STATUS.RUNNING) running++
    else if (s === STATUS.COMPLETED) completed++
    else if (s === STATUS.FAILED) failed++
  }
  return { total, running, completed, failed }
})

const tasks = computed(() => {
  return Object.keys(state._map).map(k => state._map[k])
})

function updateFromMessage(msg) {
  const key = _makeKey(msg.taskId)
  if (!key) return
  const entry = _get(key)
  const updates = _cleanMsg(msg)
  // 状态优先级：running > completed > failed > started — 不降级
  if (updates.status) {
    const old = entry.status || ''
    const order = { [STATUS.STARTED]: 1, [STATUS.RUNNING]: 2, [STATUS.COMPLETED]: 3, [STATUS.FAILED]: 3 }
    if (!old || (order[updates.status] || 0) >= (order[old] || 0)) {
      entry.status = updates.status
    }
  }
  // usage 始终覆盖（progress 持续更新）
  if (updates.usage) {
    entry.usage = updates.usage
  }
  if (updates.lastToolName !== undefined) {
    entry.lastToolName = updates.lastToolName
  }
  if (updates.description !== undefined && updates.description) {
    entry.description = updates.description
  }
  if (updates.uuid !== undefined) entry.uuid = updates.uuid
  if (updates.sessionId !== undefined) entry.sessionId = updates.sessionId
  // msgIndex: 记录最新消息索引，供点击定位用
  if (msg._msgIndex !== undefined) {
    entry._msgIndex = msg._msgIndex
  }
}

function reset() {
  for (const k of Object.keys(state._map)) {
    delete state._map[k]
  }
}

export default {
  STATUS,
  state,
  summary,
  tasks,
  updateFromMessage,
  reset,
}
