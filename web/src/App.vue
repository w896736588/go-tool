<template>
  <div id="app">
    <router-view/>
    <div v-if="gitPendingTotalCount > 0 && !isEmbedded" class="status-indicator git-pending-indicator" @click="gitDialogVisible = true">
      Git 未提交 {{ gitPendingTotalCount }}
    </div>
    <div
      v-if="sseConnectionCount > 0 && !isEmbedded"
      class="status-indicator sse-connection-indicator"
      :style="{ backgroundColor: sseConnectionColor }"
      :title="'当前 SSE 连接数: ' + sseConnectionCount + '/' + sseConnectionTotal + '（点击查看详情）'"
      @click="showSseDetailDialog"
    >
      SSE {{ sseConnectionCount }}/{{ sseConnectionTotal }}
    </div>
    <el-dialog v-model="sseDetailVisible" title="SSE 连接详情" width="860px" top="8vh">
      <el-table :data="sseDetailList" border size="small" style="width: 100%">
        <el-table-column prop="businessLabel" label="业务类型" width="130" />
        <el-table-column prop="clientId" label="Client ID" min-width="220">
          <template #default="{ row }">
            <span v-if="row.clientId" class="sse-detail-mono">{{ row.clientId }}</span>
            <span v-else style="color: #909399;">（未连接）</span>
          </template>
        </el-table-column>
        <el-table-column prop="url" label="SSE 地址" min-width="340">
          <template #default="{ row }">
            <span v-if="row.url" class="sse-detail-mono sse-detail-url">{{ row.url }}</span>
            <span v-else style="color: #909399;">（未连接）</span>
          </template>
        </el-table-column>
        <el-table-column prop="statusLabel" label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.connected ? 'success' : 'info'" size="small" disable-transitions>
              {{ row.connected ? '已连接' : '未连接' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
    <el-dialog v-model="gitDialogVisible" title="Git 未提交文件" width="720px">
      <div v-if="gitRepos.length === 0">暂无未提交文件</div>
      <div v-for="repo in gitRepos" :key="repo.label + repo.dir" class="git-repo-block">
        <div class="git-repo-head">
          <div>
            <div class="git-repo-title">{{ repo.label }} · {{ repo.count }}</div>
            <div class="git-repo-dir">{{ repo.dir }}</div>
          </div>
          <el-button
            type="primary"
            size="small"
            :loading="commitPushLoadingMap[repo.dir] === true"
            @click="commitPushRepo(repo)"
          >
            commit+push
          </el-button>
        </div>
        <el-table :data="normalizeRepoFiles(repo).map(item => ({ path: item }))" size="small" border max-height="240">
          <el-table-column prop="path" label="文件" />
        </el-table>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import base from '@/utils/base'
import git from '@/utils/base/git'
import sseDistribute from '@/utils/base/sse_distribute'
import sseBusiness from '@/utils/base/sse_business'
import { ElMessage, ElMessageBox } from 'element-plus'

const SseConnectionCountId = 'sse_connection_count'
const GitPendingStatusId = 'git_pending_status'

// SSE_BUSINESS_LABEL_MAP SSE 业务类型中文标签映射
const SSE_BUSINESS_LABEL_MAP = {
  general: '通用 SSE',
  agent_cli: 'Agent CLI',
  task_workflow: 'Task Workflow',
  agent_cli_chat: 'Agent CLI 对话',
  work_flow_chat: 'Workflow 对话',
}

export default {
  name: 'App',
  data() {
    return {
      sseConnectionCount: 0,
      sseConnectionTotal: 0,
      gitPendingTotalCount: 0,
      gitRepos: [],
      gitDialogVisible: false,
      commitPushLoadingMap: {},
      sseConnectionHandler: null,
      gitPendingStatusHandler: null,
      sseDetailVisible: false,
      sseDetailList: [],
    }
  },
  computed: {
    isEmbedded() {
      try {
        return window.self !== window.top
      } catch (e) {
        return true
      }
    },
    sseConnectionColor() {
      const total = this.sseConnectionTotal
      if (!total) return '#67C23A'
      const pct = Math.round((this.sseConnectionCount / total) * 100)
      if (pct >= 100) return '#F56C6C'
      if (pct >= 90) return '#E6A23C'
      return '#67C23A'
    },
  },
  mounted() {
    base.DisableSaveShortcut()
    const favicon = document.querySelector('link[rel="icon"]')
    if (process.env.NODE_ENV === 'production' && favicon) {
      favicon.href = './favicon.ico'
    }
    if (this.isEmbedded) return
    this.registerSseConnectionCount()
    this.registerGitPendingStatus()
  },
  unmounted() {
    sseDistribute.UnRegisterReceive(SseConnectionCountId, this.sseConnectionHandler)
    sseDistribute.UnRegisterReceive(GitPendingStatusId, this.gitPendingStatusHandler)
  },
  methods: {
    registerSseConnectionCount() {
      this.sseConnectionHandler = (data) => {
        if (data && typeof data === 'object') {
          this.sseConnectionCount = data.count || 0
          this.sseConnectionTotal = data.total || 0
        }
      }
      sseDistribute.RegisterReceive(SseConnectionCountId, this.sseConnectionHandler)
    },
    registerGitPendingStatus() {
      this.gitPendingStatusHandler = (data) => {
        if (!data || typeof data !== 'object') {
          this.gitPendingTotalCount = 0
          this.gitRepos = []
          return
        }
        this.gitPendingTotalCount = Number(data.total_count || 0)
        this.gitRepos = this.normalizeGitRepos(data.repos)
      }
      sseDistribute.RegisterReceive(GitPendingStatusId, this.gitPendingStatusHandler)
    },
    normalizeRepoFiles(repo) {
      if (!repo || !Array.isArray(repo.files)) return []
      return repo.files.filter(item => typeof item === 'string' && item.trim() !== '')
    },
    normalizeGitRepos(repos) {
      if (!Array.isArray(repos)) return []
      return repos.map(repo => ({
        ...repo,
        files: this.normalizeRepoFiles(repo),
      }))
    },
    async commitPushRepo(repo) {
      const dir = repo && repo.dir ? String(repo.dir).trim() : ''
      if (!dir) return
      try {
        const result = await ElMessageBox.prompt('请输入 commit message', 'commit+push', {
          confirmButtonText: '提交',
          cancelButtonText: '取消',
          inputValue: `chore: sync pending changes ${new Date().toLocaleString()}`,
          inputPattern: /\S+/,
          inputErrorMessage: 'commit message 不能为空',
        })
        const message = (result.value || '').trim()
        if (!message) return
        this.commitPushLoadingMap = {
          ...this.commitPushLoadingMap,
          [dir]: true,
        }
        git.GitPendingCommitPush({ dir, message }, (response) => {
          this.commitPushLoadingMap = {
            ...this.commitPushLoadingMap,
            [dir]: false,
          }
          if (!response || response.ErrCode !== 0) {
            ElMessage.error((response && response.ErrMsg) || 'commit+push 失败')
            return
          }
          ElMessage.success('commit+push 成功')
          this.refreshGitPendingStatus()
        })
      } catch (err) {
        return
      }
    },
    refreshGitPendingStatus() {
      base.BasePost('/api/GitPendingStatus', {}, (response) => {
        if (!response || response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.gitPendingTotalCount = Number(response.Data.total_count || 0)
        this.gitRepos = this.normalizeGitRepos(response.Data.repos)
      })
    },
    // showSseDetailDialog 点击 SSE 连接指示器时，收集所有 SSE 连接信息并弹窗展示
    showSseDetailDialog() {
      const list = []
      // 1. 本页通用 SSE 连接
      const generalInfo = sseDistribute.GetSseInfo()
      if (generalInfo) {
        list.push({
          businessType: 'general',
          businessLabel: SSE_BUSINESS_LABEL_MAP['general'] || '通用 SSE',
          clientId: generalInfo.clientId || '',
          url: generalInfo.url || '',
          connected: generalInfo.connected,
          source: 'local',
        })
      }
      // 2. 本页业务级 SSE 连接（agent_cli / task_workflow）
      const bizInfos = sseBusiness.GetAllBusinessInfos()
      if (bizInfos && bizInfos.length) {
        for (let i = 0; i < bizInfos.length; i++) {
          const info = bizInfos[i]
          list.push({
            businessType: info.businessType,
            businessLabel: SSE_BUSINESS_LABEL_MAP[info.businessType] || info.businessType,
            clientId: info.clientId || '',
            url: info.url || '',
            connected: info.connected,
            source: 'local',
          })
        }
      }
      this.sseDetailList = list
      this.sseDetailVisible = true
      // 3. 异步从后端拉取服务端所有 SSE 连接详情，追加到列表
      this.fetchServerSseConnections()
    },
    // fetchServerSseConnections 从后端接口拉取服务端所有 SSE 连接，合并进展示列表
    fetchServerSseConnections() {
      base.BasePost('/api/SseConnectionDetails', {}, (response) => {
        if (!response || response.ErrCode !== 0 || !response.Data || !response.Data.connections) {
          return
        }
        const serverConns = response.Data.connections || []
        const localClientIds = new Set()
        // 收集本页已有的 clientId，避免重复
        for (let i = 0; i < this.sseDetailList.length; i++) {
          if (this.sseDetailList[i].clientId) {
            localClientIds.add(this.sseDetailList[i].clientId)
          }
        }
        // 追加本页没有的服务端连接
        const sseHost = base.GetSseApiHost()
        for (let i = 0; i < serverConns.length; i++) {
          const conn = serverConns[i]
          if (localClientIds.has(conn.client_id)) continue
          const connType = conn.type || 'general'
          let routePath = '/sse'
          if (connType === 'agent_cli') {
            routePath = '/sse/agent_cli'
          } else if (connType === 'task_workflow') {
            routePath = '/sse/task_workflow'
          }
          const url = sseHost ? (sseHost + routePath + '?client_id=' + encodeURIComponent(conn.client_id) + '&token=***') : ''
          this.sseDetailList.push({
            businessType: connType,
            businessLabel: SSE_BUSINESS_LABEL_MAP[connType] || connType,
            clientId: conn.client_id || '',
            url: url,
            connected: true,
            source: 'server',
          })
        }
      })
    },
  },
}
</script>

<style>
html,
body,
#app {
  height: 100%;
}

#app {
  font-family: Consolas, Avenir, Helvetica, Arial, sans-serif;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
}

body {
  margin: 0;
}

.status-indicator {
  position: fixed;
  bottom: 16px;
  min-width: 96px;
  box-sizing: border-box;
  padding: 6px 12px;
  border-radius: 10px;
  color: #fff;
  font-size: 12px;
  line-height: 1.2;
  font-weight: 700;
  text-align: center;
  z-index: 1000;
  user-select: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.git-pending-indicator {
  right: 128px;
  background: #e6a23c;
  cursor: pointer;
}

.sse-connection-indicator {
  right: 16px;
  transition: background-color 0.3s;
  cursor: pointer;
}

.sse-connection-indicator:hover {
  opacity: 0.85;
}

.sse-detail-mono {
  font-family: 'Consolas', 'Courier New', monospace;
  font-size: 12px;
  word-break: break-all;
}

.sse-detail-url {
  color: #409EFF;
  user-select: text;
}

.git-repo-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
}

.git-repo-block + .git-repo-block {
  margin-top: 16px;
}

.git-repo-title {
  font-weight: 600;
}

.git-repo-dir {
  margin: 4px 0 8px;
  color: #909399;
  font-size: 12px;
  word-break: break-all;
}
</style>
