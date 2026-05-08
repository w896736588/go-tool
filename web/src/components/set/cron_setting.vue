<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">定时任务</h3>
      <p class="set-config-desc">设置每天固定时间自动触发功能，服务运行期间后台自动执行。</p>
    </div>

    <div class="set-config-table-card">
      <el-table :data="taskList" style="width: 100%;">
        <el-table-column label="任务名称" prop="name" width="200" />
        <el-table-column label="启用" width="80" align="center">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" :active-value="1" :inactive-value="0" />
          </template>
        </el-table-column>
        <el-table-column label="触发时间" width="200">
          <template #default="{ row }">
            <el-time-picker
              v-model="row._timeDate"
              format="HH:mm"
              placeholder="选择时间"
              :disabled="row.enabled !== 1"
              style="width: 160px;"
              @change="(val) => onTimeChange(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column label="上次触发" width="180">
          <template #default="{ row }">
            <span v-if="row.last_trigger_time > 0">{{ formatTime(row.last_trigger_time) }}</span>
            <span v-else class="cron-field-help">从未触发</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="saveTask(row)">保存</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="addDialogVisible" title="新增定时任务" width="440px" :close-on-click-modal="false">
      <el-form label-width="80px">
        <el-form-item label="任务类型">
          <el-select v-model="addForm.type" placeholder="请选择任务类型" style="width: 100%;">
            <el-option
              v-for="item in availableTypes"
              :key="item.type"
              :label="item.name"
              :value="item.type"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="addForm.enabled" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="触发时间">
          <el-time-picker
            v-model="addForm._timeDate"
            format="HH:mm"
            placeholder="选择时间"
            style="width: 100%;"
            @change="(val) => addForm.trigger_time = val ? `${String(val.getHours()).padStart(2,'0')}:${String(val.getMinutes()).padStart(2,'0')}` : ''"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="addTask">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import set from '@/utils/base/git_set'

export default {
  name: 'CronSetting',
  emits: ['changed'],
  data() {
    return {
      taskList: [],
      allTypes: [],
      addDialogVisible: false,
      addForm: { type: '', enabled: 0, trigger_time: '', _timeDate: null },
    }
  },
  mounted() {
    this.loadConfig()
  },
  computed: {
    availableTypes() {
      const existing = new Set(this.taskList.map((t) => t.type))
      return this.allTypes.filter((t) => !existing.has(t.type))
    },
  },
  methods: {
    loadConfig() {
      set.CronConfigGet((response) => {
        if (response.ErrCode !== 0 || !Array.isArray(response.Data)) {
          return
        }
        this.taskList = response.Data.map((row) => {
          const item = { ...row, _timeDate: null }
          if (row.trigger_time) {
            const [h, m] = row.trigger_time.split(':').map(Number)
            const d = new Date()
            d.setHours(h, m, 0, 0)
            item._timeDate = d
          }
          return item
        })
      })
      set.CronConfigTypes((response) => {
        if (response.ErrCode === 0 && Array.isArray(response.Data)) {
          this.allTypes = response.Data
        }
      })
    },
    onTimeChange(row, val) {
      if (val) {
        const h = String(val.getHours()).padStart(2, '0')
        const m = String(val.getMinutes()).padStart(2, '0')
        row.trigger_time = `${h}:${m}`
      } else {
        row.trigger_time = ''
      }
    },
    saveTask(row) {
      if (row.enabled === 1 && !row.trigger_time) {
        this.$helperNotify.error('启用定时任务时触发时间不能为空')
        return
      }
      const payload = {
        type: row.type,
        enabled: row.enabled,
        trigger_time: row.trigger_time,
      }
      set.CronConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('定时任务配置已保存')
          this.$emit('changed')
          this.loadConfig()
        } else {
          this.$helperNotify.error(response.ErrMsg || '保存失败')
        }
      })
    },
    openAddDialog() {
      if (this.availableTypes.length === 0) {
        this.$helperNotify.info('没有可新增的定时任务类型')
        return
      }
      this.addForm = { type: '', enabled: 0, trigger_time: '', _timeDate: null }
      this.addDialogVisible = true
    },
    addTask() {
      if (!this.addForm.type) {
        this.$helperNotify.error('请选择任务类型')
        return
      }
      if (this.addForm.enabled === 1 && !this.addForm.trigger_time) {
        this.$helperNotify.error('启用定时任务时触发时间不能为空')
        return
      }
      set.CronConfigSave({
        type: this.addForm.type,
        enabled: this.addForm.enabled,
        trigger_time: this.addForm.trigger_time,
      }, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('定时任务已添加')
          this.addDialogVisible = false
          this.$emit('changed')
          this.loadConfig()
        } else {
          this.$helperNotify.error(response.ErrMsg || '添加失败')
        }
      })
    },
    formatTime(ts) {
      const d = new Date(ts * 1000)
      const pad = (n) => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
    },
  },
}
</script>

<style scoped src="@/css/components/set/cron_setting.css"></style>
