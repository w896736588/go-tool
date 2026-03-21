<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">记忆配置</h3>
      <p class="set-config-desc">配置记忆专属仓库目录与 sqlite 文件名。保存后重启应用生效。</p>
      <div class="set-config-actions">
        <el-button type="primary" @click="saveConfig">保存</el-button>
      </div>
    </div>

    <div class="set-config-table-card">
      <el-form label-width="120px" class="memory-config-form">
        <el-form-item label="memory_dir">
          <el-input v-model="form.memory_dir" placeholder="例如 D:\\repo\\memory-data 或 /data/memory" />
        </el-form-item>
        <el-form-item label="memory_db_name">
          <el-input v-model="form.memory_db_name" placeholder="例如 memory.db" />
        </el-form-item>
      </el-form>
      <el-alert
        :closable="false"
        type="info"
        title="启动时会先判断目录是否为 git 仓库；是则先 git pull，之后加载该 sqlite。"
      />
    </div>
  </div>
</template>

<script>
import set from '@/utils/base/git_set'

export default {
  name: 'MemorySet',
  data() {
    return {
      form: {
        memory_dir: '',
        memory_db_name: '',
      }
    }
  },
  mounted() {
    this.loadConfig()
  },
  methods: {
    loadConfig() {
      set.MemoryConfigGet((response) => {
        if (response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.form.memory_dir = response.Data.memory_dir || ''
        this.form.memory_db_name = response.Data.memory_db_name || ''
      })
    },
    saveConfig() {
      set.MemoryConfigSave(this.form, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('记忆配置已保存，重启应用后生效')
        }
      })
    }
  }
}
</script>

<style scoped>
@import "@/css/set_module_unified.css";

.memory-config-form {
  margin-bottom: 16px;
}
</style>
