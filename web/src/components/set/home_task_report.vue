<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">工作日报 AI 设置</h3>
      <p class="set-config-desc">这里维护任务清单右侧"AI 生成工作日报"按钮使用的模型和提示词。</p>
    </div>

    <div class="set-config-table-card">
      <el-form label-width="120px" class="memory-config-form">
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
            :rows="5"
            placeholder="请输入首页任务工作日报提示词"
          />
        </el-form-item>
        <el-form-item>
          <pl-button type="primary" @click="saveConfig">保存工作日报配置</pl-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="set-config-header" style="margin-top: 24px;">
      <h3 class="set-config-title">知识片段提示词模板</h3>
      <p class="set-config-desc">
        新建任务时自动创建的知识片段，会将此模板内容写入片段开头。支持以下内置变量：
      </p>
      <el-table :data="fragmentPromptVariables" border size="small" class="fragment-prompt-var-table">
        <el-table-column prop="name" label="变量" width="140" />
        <el-table-column prop="desc" label="说明" />
      </el-table>
    </div>

    <div class="set-config-table-card">
      <el-form label-width="120px" class="memory-config-form">
        <el-form-item label="提示词模板">
          <MdEditor
            v-model="form.home_task_fragment_prompt"
            :preview="true"
            :toolbarsExclude="['github', 'htmlPreview', 'catalog', 'save']"
            placeholder="留空则使用默认行为（仅生成标题）。支持变量：{tapd_url}、{api_host}、{api_token}"
            style="height: 360px;"
          />
        </el-form-item>
        <el-form-item>
          <pl-button type="primary" @click="saveFragmentPrompt">保存提示词模板</pl-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script>
import set from '@/utils/base/git_set'
import AiSetApi from '@/utils/base/ai_set'
import { MdEditor } from 'md-editor-v3'

// DEFAULT_MEMORY_ARRANGE_PROMPT 作为透传字段保留知识片段整理配置。
const DEFAULT_MEMORY_ARRANGE_PROMPT = '帮我把当前 markdown 进行整理格式，让它看起来更顺畅清晰，注意禁止修改内容'
// DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT 定义首页任务日报默认提示词。
const DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT = '请基于当前活跃任务生成中文工作日报，按已完成、进行中、风险与阻塞三个部分总结，输出 Markdown，禁止编造未提供的信息。'

export default {
  name: 'HomeTaskReportSetting',
  emits: ['changed'],
  components: { MdEditor },
  data() {
    return {
      aiModelList: [],
      fragmentPromptVariables: [
        { name: '{tapd_url}', desc: '新建任务时填写的 TAPD 需求地址' },
        { name: '{api_host}', desc: '当前 API 请求地址（如 http://192.168.1.100:8080）' },
        { name: '{api_token}', desc: '当前认证 token' },
      ],
      form: {
        memory_arrange_model_id: null,
        memory_arrange_prompt: DEFAULT_MEMORY_ARRANGE_PROMPT,
        home_task_daily_report_model_id: null,
        home_task_daily_report_prompt: DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT,
        home_task_fragment_prompt: '',
      },
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
        this.form.memory_arrange_model_id = response.Data.memory_arrange_model_id || null
        this.form.memory_arrange_prompt = response.Data.memory_arrange_prompt || DEFAULT_MEMORY_ARRANGE_PROMPT
        this.form.home_task_daily_report_model_id = response.Data.home_task_daily_report_model_id || null
        this.form.home_task_daily_report_prompt = response.Data.home_task_daily_report_prompt || DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT
        this.form.home_task_fragment_prompt = response.Data.home_task_fragment_prompt || ''
      })
    },
    saveConfig() {
      const payload = {
        memory_arrange_model_id: this.form.memory_arrange_model_id,
        memory_arrange_prompt: this.form.memory_arrange_prompt,
        home_task_daily_report_model_id: this.form.home_task_daily_report_model_id,
        home_task_daily_report_prompt: this.form.home_task_daily_report_prompt,
        home_task_fragment_prompt: this.form.home_task_fragment_prompt,
      }
      set.MemoryConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('配置已保存')
          this.$emit('changed')
        }
      })
    },
    saveFragmentPrompt() {
      const payload = {
        memory_arrange_model_id: this.form.memory_arrange_model_id,
        memory_arrange_prompt: this.form.memory_arrange_prompt,
        home_task_daily_report_model_id: this.form.home_task_daily_report_model_id,
        home_task_daily_report_prompt: this.form.home_task_daily_report_prompt,
        home_task_fragment_prompt: this.form.home_task_fragment_prompt,
      }
      set.MemoryConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('提示词模板已保存')
          this.$emit('changed')
        }
      })
    },
  },
}
</script>

<style scoped>
@import "@/css/set_module_unified.css";

.memory-config-form {
  margin-bottom: 0;
}

.fragment-prompt-var-table {
  margin: 12px 0;
}
</style>
