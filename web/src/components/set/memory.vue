<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">记忆配置</h3>
      <p class="set-config-desc">记忆库目录和 sqlite 文件名改为从配置文件读取，这里仅展示当前生效值；AI 配置仍可在页面保存。</p>
      <div class="set-config-actions">
        <pl-button type="primary" @click="saveConfig">保存 AI 配置</pl-button>
      </div>
    </div>

    <div class="set-config-table-card">
      <el-alert
        v-if="!form.memory_db_configured"
        :closable="false"
        type="warning"
        :title="memoryConfigAlertTitle"
      />
      <el-alert
        v-else
        :closable="false"
        type="info"
        :title="memoryConfigAlertTitle"
      />

      <el-descriptions class="memory-config-display" :column="1" border>
        <el-descriptions-item label="配置文件">
          {{ form.memory_config_file || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="memoryDbPath">
          {{ form.memory_dir || '未配置，请在配置文件中设置' }}
        </el-descriptions-item>
        <el-descriptions-item label="memoryDbFileName">
          {{ form.memory_db_name || '未配置，请在配置文件中设置' }}
        </el-descriptions-item>
      </el-descriptions>

      <el-form label-width="120px" class="memory-config-form">
        <el-divider content-position="left">AI 整理</el-divider>
        <el-form-item label="整理模型">
          <el-select
            v-model="form.memory_arrange_model_id"
            clearable
            filterable
            style="width: 100%;"
            placeholder="请选择用于整理知识片段的 LLM 模型"
          >
            <el-option
              v-for="item in aiModelList"
              :key="item.id"
              :label="buildModelLabel(item)"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="整理提示词">
          <el-input
            v-model="form.memory_arrange_prompt"
            type="textarea"
            :rows="4"
            placeholder="请输入 AI 整理提示词"
          />
        </el-form-item>
        <el-divider content-position="left">工作日报 AI</el-divider>
        <el-form-item label="日报模型">
          <el-select
            v-model="form.home_task_daily_report_model_id"
            clearable
            filterable
            style="width: 100%;"
            placeholder="请选择用于生成工作日报的 LLM 模型"
          >
            <el-option
              v-for="item in aiModelList"
              :key="item.id"
              :label="buildModelLabel(item)"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="日报提示词">
          <el-input
            v-model="form.home_task_daily_report_prompt"
            type="textarea"
            :rows="4"
            placeholder="请输入首页任务工作日报提示词"
          />
        </el-form-item>
      </el-form>
      <el-alert
        :closable="false"
        type="info"
        title="启动时会先判断记忆目录是否为 git 仓库；是则先 git pull，之后加载该 sqlite。修改 memoryDbPath 或 memoryDbFileName 后需要更新配置文件并重启应用。"
      />
    </div>
  </div>
</template>

<script>
import set from '@/utils/base/git_set'
import AiSetApi from '@/utils/base/ai_set'

const DEFAULT_MEMORY_ARRANGE_PROMPT = '帮我把当前markdown进行整理格式，让它看起来更顺畅清晰，注意禁止修改内容'
// DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT 定义首页任务工作日报默认提示词。
const DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT = '请基于当前活跃任务生成中文工作日报，按已完成、进行中、风险与阻塞三个部分总结，输出 Markdown，禁止编造未提供的信息。'

export default {
  name: 'MemorySet',
  data() {
    return {
      aiModelList: [],
      form: {
        memory_dir: '',
        memory_db_name: '',
        memory_db_configured: false,
        memory_config_file: '',
        memory_arrange_model_id: null,
        memory_arrange_prompt: DEFAULT_MEMORY_ARRANGE_PROMPT,
        home_task_daily_report_model_id: null,
        home_task_daily_report_prompt: DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT,
      }
    }
  },
  computed: {
    // memoryConfigAlertTitle 统一生成记忆 db 配置提示 / build memory db config hint text.
    memoryConfigAlertTitle() {
      const configFile = this.form.memory_config_file || '配置文件'
      if (!this.form.memory_db_configured) {
        return `未检测到记忆库配置，请在 ${configFile} 的 [base] 节点中配置 memoryDbPath 和 memoryDbFileName。`
      }
      return `当前记忆库 db 配置来自 ${configFile} 的 [base] 节点`
    }
  },
  mounted() {
    this.loadAiModelList()
    this.loadConfig()
  },
  methods: {
    buildModelLabel(item) {
      const provider = item.provider_name || '未命名服务商'
      const model = item.name || item.model || `模型#${item.id}`
      return `${provider} / ${model}`
    },
    loadAiModelList() {
      AiSetApi.AiModelList({ model_type: 'llm' }, (response) => {
        if (response.ErrCode !== 0) {
          return
        }
        this.aiModelList = Array.isArray(response.Data) ? response.Data : []
      })
    },
    loadConfig() {
      set.MemoryConfigGet((response) => {
        if (response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.form.memory_dir = response.Data.memory_dir || ''
        this.form.memory_db_name = response.Data.memory_db_name || ''
        this.form.memory_db_configured = !!response.Data.memory_db_configured
        this.form.memory_config_file = response.Data.memory_config_file || ''
        this.form.memory_arrange_model_id = response.Data.memory_arrange_model_id || null
        this.form.memory_arrange_prompt = response.Data.memory_arrange_prompt || DEFAULT_MEMORY_ARRANGE_PROMPT
        this.form.home_task_daily_report_model_id = response.Data.home_task_daily_report_model_id || null
        this.form.home_task_daily_report_prompt = response.Data.home_task_daily_report_prompt || DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT
      })
    },
    saveConfig() {
      const payload = {
        memory_arrange_model_id: this.form.memory_arrange_model_id,
        memory_arrange_prompt: this.form.memory_arrange_prompt,
        home_task_daily_report_model_id: this.form.home_task_daily_report_model_id,
        home_task_daily_report_prompt: this.form.home_task_daily_report_prompt,
      }
      set.MemoryConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('AI 配置已保存')
        }
      })
    }
  }
}
</script>

<style scoped>
@import "@/css/set_module_unified.css";

.memory-config-display {
  margin: 16px 0;
}

.memory-config-form {
  margin-bottom: 16px;
}
</style>
