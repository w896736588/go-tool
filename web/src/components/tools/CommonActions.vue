<template>
  <div class="common-actions">
    <div class="common-actions__header">
      <div class="common-actions__title">常用操作</div>
      <div class="common-actions__desc">可扩展动作面板，当前提供端口占用查询与结束进程。</div>
    </div>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="16">
        <el-card shadow="hover" class="action-card">
          <template #header>
            <div class="action-card__header">
              <div class="action-card__title">端口进程管理</div>
              <div class="action-card__subtitle">输入端口，先查询占用进程，再确认结束。</div>
            </div>
          </template>

          <el-form @submit.prevent>
            <el-row :gutter="12">
              <el-col :xs="24" :sm="14" :md="12">
                <el-form-item label="端口">
                  <el-input
                    v-model.trim="portInput"
                    clearable
                    placeholder="例如 8080"
                    @keyup.enter="queryProcesses"
                  />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="10" :md="12" class="action-card__buttons">
                <el-button type="primary" :loading="queryLoading" @click="queryProcesses">查询占用进程</el-button>
                <el-button :disabled="!lastQueryPort || queryLoading" @click="refreshProcesses">刷新</el-button>
              </el-col>
            </el-row>
          </el-form>

          <el-alert
            title="结束操作会强制终止目标进程，请先确认 PID 和进程名。"
            type="warning"
            :closable="false"
            show-icon
            class="action-card__alert"
          />

          <el-empty v-if="hasSearched && processList.length === 0 && !queryLoading" description="当前端口没有监听进程" />

          <el-table
            v-if="processList.length > 0"
            :data="processList"
            border
            stripe
            class="action-card__table"
          >
            <el-table-column prop="pid" label="PID" width="120" />
            <el-table-column prop="command" label="进程名" min-width="180">
              <template #default="scope">
                {{ scope.row.command || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="protocol" label="协议" width="120" />
            <el-table-column prop="address" label="监听地址" min-width="220" />
            <el-table-column label="操作" width="160" fixed="right">
              <template #default="scope">
                <el-button
                  type="danger"
                  plain
                  size="small"
                  :loading="killingPid === scope.row.pid"
                  @click="confirmKill(scope.row)"
                >
                  结束进程
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="8">
        <el-card shadow="never" class="action-card action-card--muted">
          <template #header>
            <div class="action-card__title">后续可扩展</div>
          </template>
          <div class="placeholder-list">
            <div class="placeholder-item">按端口查占用</div>
            <div class="placeholder-item">结束指定 PID</div>
            <div class="placeholder-item">打开常用目录</div>
            <div class="placeholder-item">清理临时文件</div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import { ElMessageBox } from 'element-plus'
import toolsApi from '@/utils/base/tools'

export default {
  name: 'CommonActions',
  data() {
    return {
      portInput: '',
      processList: [],
      queryLoading: false,
      killingPid: 0,
      hasSearched: false,
      lastQueryPort: 0,
    }
  },
  methods: {
    parsePortValue() {
      const port = Number(this.portInput)
      if (!Number.isInteger(port) || port < 1 || port > 65535) {
        this.$helperNotify.error('请输入 1-65535 之间的端口')
        return 0
      }
      return port
    },
    queryProcesses() {
      const port = this.parsePortValue()
      if (!port) {
        return
      }
      this.queryLoading = true
      toolsApi.ToolPortProcessList({ port }, (response) => {
        this.queryLoading = false
        this.hasSearched = true
        if (response.ErrCode !== 0) {
          return
        }
        this.lastQueryPort = port
        this.processList = Array.isArray(response.Data?.items) ? response.Data.items : []
        if (this.processList.length === 0) {
          this.$helperNotify.warning('当前端口没有监听进程')
          return
        }
        this.$helperNotify.success(`已查询到 ${this.processList.length} 个进程`)
      })
    },
    refreshProcesses() {
      if (!this.lastQueryPort) {
        return
      }
      this.portInput = String(this.lastQueryPort)
      this.queryProcesses()
    },
    async confirmKill(row) {
      try {
        await ElMessageBox.confirm(
          `确认结束 PID ${row.pid}${row.command ? `（${row.command}）` : ''} 吗？`,
          '结束进程确认',
          {
            type: 'warning',
            confirmButtonText: '确认结束',
            cancelButtonText: '取消',
          }
        )
      } catch (error) {
        return
      }

      this.killingPid = row.pid
      toolsApi.ToolPortProcessKill({ pid: row.pid }, (response) => {
        this.killingPid = 0
        if (response.ErrCode !== 0) {
          return
        }
        this.$helperNotify.success(`PID ${row.pid} 已结束`)
        this.refreshProcesses()
      })
    },
  },
}
</script>

<style scoped>
.common-actions {
  padding: 4px 6px 18px;
}

.common-actions__header {
  margin-bottom: 14px;
}

.common-actions__title {
  font-size: 18px;
  font-weight: 600;
  color: #324a34;
}

.common-actions__desc {
  margin-top: 4px;
  color: #66756a;
  font-size: 13px;
}

.action-card {
  border-radius: 12px;
}

.action-card--muted {
  background: linear-gradient(180deg, #fafcf8 0%, #f4f8f1 100%);
}

.action-card__header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.action-card__title {
  font-size: 16px;
  font-weight: 600;
  color: #35553a;
}

.action-card__subtitle {
  color: #708171;
  font-size: 12px;
}

.action-card__buttons {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-card__alert {
  margin-bottom: 16px;
}

.action-card__table {
  margin-top: 8px;
}

.placeholder-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.placeholder-item {
  padding: 12px 14px;
  border: 1px dashed #c8d7c5;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.72);
  color: #55715a;
}

@media (max-width: 768px) {
  .action-card__buttons {
    justify-content: flex-start;
    margin-bottom: 8px;
  }
}
</style>
