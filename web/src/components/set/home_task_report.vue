<template>
  <div class="set-config-page">
    <!-- 工作日报 AI 设置 -->
    <div v-show="activeTab === 'daily-report' || !activeTab">
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
            <MdEditor
              v-model="form.home_task_daily_report_prompt"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              class="daily-report-editor"
            />
          </el-form-item>
          <el-form-item>
            <pl-button type="primary" @click="saveConfig">保存工作日报配置</pl-button>
            <pl-button @click="showChangeLog">改动记录</pl-button>
          </el-form-item>
        </el-form>
      </div>
    </div>

    <!-- 工作流提示词模板（新版模板管理） -->
    <div v-show="activeTab === 'prompt-template'" class="prompt-template-section">
      <div class="set-config-header">
        <h3 class="set-config-title">工作流模板管理</h3>
        <p class="set-config-desc">
          左侧选择或新建模板，右侧编辑步骤名称与提示词。固定步骤不可删除，自定义步骤支持拖拽排序。
        </p>
      </div>
      <div class="prompt-template-card">
        <WorkflowTemplateManager ref="templateManager" @templates-loaded="onTemplatesLoaded" />
      </div>
    </div>

    <!-- 开发环境 -->
    <div v-show="activeTab === 'dev-environment'" class="dev-environment-section">
      <div class="set-config-header">
        <h3 class="set-config-title">开发环境</h3>
        <p class="set-config-desc">
          描述当前项目的开发环境信息，支持 Markdown 语法。保存后可在工作流提示词中使用
          <code style="color:#3d6b3d;background:#eef4ec;padding:1px 6px;border-radius:3px;">{开发环境}</code>
          占位符引用此内容。
        </p>
      </div>

      <div class="prompt-placeholder-bar">
        <span class="prompt-placeholder-bar__label">可用占位符：</span>
        <span
          v-for="ph in devEnvironmentPlaceholders"
          :key="ph.value"
          :class="['prompt-placeholder-tag', `prompt-placeholder-tag--${ph.group || 'builtin'}`]"
          @click="copyPlaceholder(ph)"
        >
          {{ ph.label }}
          <el-icon class="prompt-placeholder-tag__icon"><CopyDocument /></el-icon>
          <el-tooltip v-if="ph.tip" :content="ph.tip" placement="top">
            <el-icon class="prompt-placeholder-tag__help"><QuestionFilled /></el-icon>
          </el-tooltip>
        </span>
      </div>

      <div class="set-config-table-card dev-environment-card">
        <MdEditor
          v-model="form.home_task_dev_environment"
          preview-theme="github"
          :preview="true"
          :toolbars="promptEditorToolbars"
          class="dev-environment-editor"
        />
        <div class="dev-environment-footer">
          <pl-button type="primary" @click="saveDevEnvironmentConfig">保存开发环境配置</pl-button>
          <pl-button @click="showChangeLog">改动记录</pl-button>
        </div>
      </div>
    </div>

    <!-- 需求抓取配置（自定义模式） -->
    <div v-show="activeTab === 'requirement-fetch'">
      <div class="set-config-header">
        <h3 class="set-config-title">需求抓取配置</h3>
        <p class="set-config-desc">
          维护需求抓取入口和页面解析规则，新建任务后会根据抓取类型选择对应配置执行。
        </p>
      </div>

      <div class="set-config-table-card">
        <div class="fetch-config-toolbar">
          <pl-button type="primary" size="small" @click="openFetchConfigDialog(null)">+ 添加配置</pl-button>
        </div>
        <el-table :data="requirementFetchConfigs" border style="width: 100%; margin-top: 12px;">
          <el-table-column prop="name" label="名称" min-width="120" />
          <el-table-column label="自定义网页" min-width="200">
            <template #default="scope">
              {{ getSmartLinkNameById(scope.row.smart_link_id) || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="css_selector" label="CSS选择器" min-width="140" />
          <el-table-column prop="wait_seconds" label="等待秒数" width="100" align="center" />
          <el-table-column label="操作" width="140" align="center" fixed="right">
            <template #default="scope">
              <pl-button size="small" @click="openFetchConfigDialog(scope.row)">编辑</pl-button>
              <pl-button size="small" @click="deleteRequirementFetchConfig(scope.$index)">删除</pl-button>
            </template>
          </el-table-column>
        </el-table>
        <div v-if="requirementFetchConfigs.length === 0" class="fetch-config-empty">
          暂无抓取配置，点击"添加配置"创建
        </div>
      </div>

      <!-- 编辑/新增配置弹窗 -->
      <el-dialog
        v-model="fetchConfigDialogVisible"
        :title="isFetchConfigEdit ? '编辑抓取配置' : '添加抓取配置'"
        width="560px"
        destroy-on-close
      >
        <el-form label-width="110px">
          <el-form-item label="配置名称" required>
            <el-input v-model="fetchConfigForm.name" placeholder="如：TAPD、禅道、飞书需求" maxlength="50" />
          </el-form-item>
          <el-form-item label="自定义网页" required>
            <el-autocomplete
              v-model="smartLinkSearchKeyword"
              :fetch-suggestions="filterSmartLinks"
              clearable
              style="width: 100%;"
              placeholder="输入关键词搜索自定义网页"
              @select="onSmartLinkSelect"
            >
              <template #default="{ item }">
                <div class="smart-link-option">{{ item.value }}</div>
              </template>
            </el-autocomplete>
          </el-form-item>
          <el-form-item label="CSS选择器" required>
            <el-input v-model="fetchConfigForm.css_selector" placeholder="如 .content-wrapper 或 #main" />
          </el-form-item>
          <el-form-item label="抓取前等待秒数">
            <el-input-number v-model="fetchConfigForm.wait_seconds" :min="1" :max="30" />
          </el-form-item>
        </el-form>
        <template #footer>
          <pl-button @click="fetchConfigDialogVisible = false">取消</pl-button>
          <pl-button type="primary" @click="confirmFetchConfig">确定</pl-button>
        </template>
      </el-dialog>
    </div>

    <!-- 分支名生成提示词 -->
    <div v-show="activeTab === 'branch-name'">
      <div class="set-config-header">
        <h3 class="set-config-title">分支名生成提示词</h3>
        <p class="set-config-desc">
          用于生成分支名称的提示词模板，支持 Markdown 语法。
        </p>
      </div>

      <div class="prompt-placeholder-bar">
        <span class="prompt-placeholder-bar__label">内置占位符：</span>
        <span
          v-for="ph in branchNamePlaceholders"
          :key="ph.value"
          class="prompt-placeholder-tag"
          @click="copyPlaceholder(ph)"
        >
          {{ ph.label }}
          <el-icon class="prompt-placeholder-tag__icon"><CopyDocument /></el-icon>
          <el-tooltip v-if="ph.tip" :content="ph.tip" placement="top">
            <el-icon class="prompt-placeholder-tag__help"><QuestionFilled /></el-icon>
          </el-tooltip>
        </span>
      </div>

      <div class="set-config-table-card">
        <el-form label-width="120px" class="memory-config-form">
          <el-form-item label="生成模型">
            <el-select
              v-model="form.home_task_branch_name_model_id"
              clearable
              filterable
              style="width: 100%;"
              placeholder="请选择用于生成分支名的 LLM 模型"
            >
              <el-option
                v-for="item in aiModelList"
                :key="item.id"
                :label="buildModelLabel(item)"
                :value="item.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="提示词">
            <MdEditor
              v-model="form.home_task_branch_name_prompt"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              class="branch-name-editor"
              style="height: 360px;"
            />
          </el-form-item>
          <el-form-item>
            <pl-button type="primary" @click="saveBranchNameConfig">保存分支名生成提示词</pl-button>
          </el-form-item>
        </el-form>
      </div>
    </div>

  </div>
</template>

<script>
import set from '@/utils/base/git_set'
import AiSetApi from '@/utils/base/ai_set'
import SmartLinkSet from '@/utils/base/smart_link_set'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import { CopyDocument, QuestionFilled } from '@element-plus/icons-vue'
import WorkflowTemplateManager from './WorkflowTemplateManager.vue'
import workflowTemplateApi from '@/utils/base/workflow_template'

const DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT = '请基于当前活跃任务生成中文工作日报，按已完成、进行中、风险与阻塞三个部分总结，输出 Markdown，禁止编造未提供的信息。'

const PROMPT_EDITOR_TOOLBARS = [
  'bold', 'italic', 'strikeThrough', 'title', 'quote',
  'unorderedList', 'orderedList', 'task', 'link', 'code',
  'codeRow', 'table', 'preview', 'fullscreen',
]

export default {
  name: 'HomeTaskReportSetting',
  emits: ['changed'],
  props: {
    activeTab: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      aiModelList: [],
      smartLinkList: [],
      form: {
        home_task_daily_report_model_id: null,
        home_task_daily_report_prompt: DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT,
        home_task_fragment_prompt: '',
        home_task_dev_environment: '',
        home_task_branch_name_prompt: '',
        home_task_branch_name_model_id: null,
      },
      // 需求抓取自定配置
      requirementFetchConfigs: [],
      fetchConfigDialogVisible: false,
      fetchConfigForm: this.createEmptyFetchConfigForm(),
      editingFetchConfigIndex: -1,
      smartLinkSearchKeyword: '',
      promptEditorToolbars: PROMPT_EDITOR_TOOLBARS,
      skillList: [],
    }
  },
  computed: {
    isFetchConfigEdit() {
      return this.editingFetchConfigIndex >= 0
    },
    branchNamePlaceholders() {
      return [
        { label: '需求名', value: '{需求名}', tip: '替换为任务名称' },
        { label: '父分支', value: '{父分支}', tip: '替换为父分支名称' },
        { label: '任务创建日期', value: '{任务创建日期}', tip: '替换为任务创建日期' },
      ]
    },
    devEnvironmentPlaceholders() {
      const base = [
        { label: '{接口开发API地址}', value: '{接口开发API地址}', tip: '替换为当前服务的 API 基地址（scheme://host）', group: 'builtin' },
        { label: '{接口开发API的token}', value: '{接口开发API的token}', tip: '替换为请求的 Authorization token', group: 'builtin' },
        { label: '{开发配置}', value: '{开发配置}', tip: '替换为开发项目配置的 Markdown 列表', group: 'builtin' },
      ]
      const skillPlaceholders = (this.skillList || []).map(name => ({
        label: `{${name}地址}`,
        value: `{${name}地址}`,
        tip: `替换为 skills/${name} 目录的本地路径`,
        group: 'skill',
      }))
      return base.concat(skillPlaceholders)
    },
  },
  mounted() {
    this.loadAiModelList()
    this.loadSmartLinkList()
    this.loadConfig()
    this.loadSkillList()
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
    loadSmartLinkList() {
      SmartLinkSet.SmartLinkItemList((response) => {
        if (response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.smartLinkList = Array.isArray(response.Data.smart_link_list) ? response.Data.smart_link_list : []
      })
    },
    loadConfig() {
      set.HomeTaskConfigGet((response) => {
        if (response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.form.home_task_daily_report_model_id = response.Data.home_task_daily_report_model_id || null
        this.form.home_task_daily_report_prompt = response.Data.home_task_daily_report_prompt || DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT
        this.form.home_task_fragment_prompt = response.Data.home_task_fragment_prompt || ''
        this.form.home_task_dev_environment = response.Data.home_task_dev_environment || ''
        this.form.home_task_branch_name_prompt = response.Data.home_task_branch_name_prompt || ''
        this.form.home_task_branch_name_model_id = response.Data.home_task_branch_name_model_id || null
        // 加载需求抓取自定配置列表
        const configs = response.Data.home_task_requirement_fetch_configs
        this.requirementFetchConfigs = Array.isArray(configs) ? [...configs] : []
      })
    },
    saveConfig() {
      const payload = this.buildFullPayload()
      // 附带当前的抓取配置列表，防止被覆盖为空
      payload.home_task_requirement_fetch_configs = JSON.stringify(this.requirementFetchConfigs)
      set.HomeTaskConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('配置已保存')
          this.$emit('changed')
        }
      })
    },
    // ===== 需求抓取自定配置管理 =====

    createEmptyFetchConfigForm() {
      return {
        name: '',
        type: '',
        smart_link_id: null,
        link_label: '',
        css_selector: '',
        wait_seconds: 5,
      }
    },
    generateFetchConfigType() {
      return 'custom_' + Date.now().toString(36) + '_' + Math.random().toString(36).slice(2, 8)
    },
    openFetchConfigDialog(row) {
      this.fetchConfigForm = this.createEmptyFetchConfigForm()
      this.editingFetchConfigIndex = -1
      this.smartLinkSearchKeyword = ''
      if (row) {
        // 编辑模式：找到现有配置索引并回填
        const idx = this.requirementFetchConfigs.indexOf(row)
        if (idx >= 0) {
          this.editingFetchConfigIndex = idx
          this.fetchConfigForm = { ...row }
          this.smartLinkSearchKeyword = row.link_label || ''
        }
      }
      this.fetchConfigDialogVisible = true
    },
    confirmFetchConfig() {
      const cfg = this.fetchConfigForm
      if (!cfg.name || !cfg.name.trim()) {
        this.$helperNotify.error('请输入配置名称')
        return
      }
      if (!cfg.smart_link_id) {
        this.$helperNotify.error('请选择自定义网页')
        return
      }
      if (!cfg.css_selector || !cfg.css_selector.trim()) {
        this.$helperNotify.error('请输入CSS选择器')
        return
      }
      cfg.name = cfg.name.trim()
      // 从新表 smart_link 自动获取 link_label
      const selected = this.smartLinkList.find(s => s.id === cfg.smart_link_id)
      cfg.link_label = selected ? selected.label : ''
      cfg.css_selector = cfg.css_selector.trim()
      if (!cfg.type) {
        cfg.type = this.generateFetchConfigType()
      }
      if (cfg.wait_seconds <= 0) {
        cfg.wait_seconds = 5
      }
      if (this.editingFetchConfigIndex >= 0) {
        this.requirementFetchConfigs[this.editingFetchConfigIndex] = cfg
      } else {
        this.requirementFetchConfigs.push(cfg)
      }
      this.fetchConfigDialogVisible = false
      // 弹窗确定后直接保存
      this.saveRequirementFetchConfig()
    },
    deleteRequirementFetchConfig(index) {
      this.requirementFetchConfigs.splice(index, 1)
      // 删除后直接保存
      this.saveRequirementFetchConfig()
    },
    getSmartLinkNameById(id) {
      if (!id) return ''
      const item = this.smartLinkList.find(s => s.id === id)
      return item ? item.label : ''
    },
    filterSmartLinks(queryString, cb) {
      const list = this.smartLinkList
        .filter(s => !queryString || s.label.toLowerCase().includes(queryString.toLowerCase()))
        .map(s => ({ value: s.label, id: s.id }))
      cb(list)
    },
    onSmartLinkSelect(item) {
      this.fetchConfigForm.smart_link_id = item.id
    },
    saveRequirementFetchConfig() {
      const payload = this.buildFullPayload()
      // 附带自定义配置列表
      payload.home_task_requirement_fetch_configs = JSON.stringify(this.requirementFetchConfigs)
      set.HomeTaskConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('需求抓取配置已保存')
          this.$emit('changed')
        }
      })
    },
    savePromptConfig() {
      const payload = this.buildFullPayload()
      payload.home_task_requirement_fetch_configs = JSON.stringify(this.requirementFetchConfigs)
      set.HomeTaskConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('提示词模板配置已保存')
          this.$emit('changed')
        }
      })
    },
    saveDevEnvironmentConfig() {
      const payload = this.buildFullPayload()
      payload.home_task_requirement_fetch_configs = JSON.stringify(this.requirementFetchConfigs)
      set.HomeTaskConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('开发环境配置已保存')
          this.$emit('changed')
        }
      })
    },
    saveBranchNameConfig() {
      const payload = this.buildFullPayload()
      payload.home_task_requirement_fetch_configs = JSON.stringify(this.requirementFetchConfigs)
      set.HomeTaskConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('分支名生成提示词已保存')
          this.$emit('changed')
        }
      })
    },
    buildFullPayload() {
      return {
        home_task_daily_report_model_id: this.form.home_task_daily_report_model_id,
        home_task_daily_report_prompt: this.form.home_task_daily_report_prompt,
        home_task_fragment_prompt: this.form.home_task_fragment_prompt,
        home_task_dev_environment: this.form.home_task_dev_environment,
        home_task_branch_name_prompt: this.form.home_task_branch_name_prompt,
        home_task_branch_name_model_id: this.form.home_task_branch_name_model_id,
      }
    },
    copyPlaceholder(placeholder) {
      const text = placeholder.value
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(() => {
          this.$helperNotify.success(`已复制：${text}`)
        }).catch(() => {
          this.fallbackCopy(text)
        })
      } else {
        this.fallbackCopy(text)
      }
    },
    fallbackCopy(text) {
      const textarea = document.createElement('textarea')
      textarea.value = text
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
      this.$helperNotify.success(`已复制：${text}`)
    },
    // 动态加载 skills 列表
    loadSkillList() {
      workflowTemplateApi.WorkflowSkillList((response) => {
        if (response && response.ErrCode === 0 && response.Data && response.Data.list) {
          this.skillList = response.Data.list.map(item => item.name)
        }
      })
    },
    // 模板管理器加载完成后的回调
    onTemplatesLoaded(templates) {
      // 模板列表已加载，可用于后续操作
      this.$emit('changed')
    },
  },
  components: {
    MdEditor,
    CopyDocument,
    WorkflowTemplateManager,
  },
}
</script>

<style scoped src="@/css/components/set/home_task_report.css"></style>

<style>
.set-config-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.prompt-template-section {
  display: flex;
  flex-direction: column;
  flex: 1 1 auto;
  height: 100%;
  min-height: 0;
}

.prompt-template-card {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.prompt-template-card .el-tabs {
  display: flex;
  flex-direction: column;
  flex: 1 1 auto;
  min-height: 0;
}

.prompt-template-card .el-tabs__header {
  /* 二级 Tab 头部必须固定在内容上方，避免被 flex 顺序挤到底部。 */
  /* Keep nested tab headers above the content instead of being pushed to the bottom. */
  order: -1;
  flex-shrink: 0;
}

.prompt-template-card .el-tabs__content {
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
}

.prompt-template-card .el-tab-pane {
  height: 100%;
}

.daily-report-editor,
.prompt-template-editor,
.branch-name-editor {
  height: calc(100vh - 460px);
  min-height: 200px;
  max-height: 800px;
}

.daily-report-editor :deep(.md-editor),
.prompt-template-editor :deep(.md-editor),
.branch-name-editor :deep(.md-editor) {
  --md-font-size: 13px;
  --md-code-font-size: 13px;
  --md-font-family: Consolas, Avenir, Helvetica, Arial, sans-serif;
}

.prompt-template-editor {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.prompt-template-editor :deep(.md-editor) {
  width: 100%;
  height: 100%;
  min-height: 0;
}

.daily-report-editor :deep(.md-editor-input),
.daily-report-editor :deep(.md-editor-preview-wrapper),
.daily-report-editor :deep(.cm-content),
.daily-report-editor :deep(.md-editor-preview),
.prompt-template-editor :deep(.md-editor-input),
.prompt-template-editor :deep(.md-editor-preview-wrapper),
.prompt-template-editor :deep(.cm-content),
.prompt-template-editor :deep(.md-editor-preview),
.branch-name-editor :deep(.md-editor-input),
.branch-name-editor :deep(.md-editor-preview-wrapper),
.branch-name-editor :deep(.cm-content),
.branch-name-editor :deep(.md-editor-preview) {
  font-family: Consolas, Avenir, Helvetica, Arial, sans-serif !important;
  font-size: 13px !important;
  line-height: 1.6;
}

.daily-report-editor :deep(.md-editor-preview p),
.daily-report-editor :deep(.md-editor-preview li),
.daily-report-editor :deep(.md-editor-preview blockquote),
.daily-report-editor :deep(.md-editor-preview table),
.daily-report-editor :deep(.md-editor-preview td),
.daily-report-editor :deep(.md-editor-preview th),
.daily-report-editor :deep(.md-editor-preview code),
.prompt-template-editor :deep(.md-editor-preview p),
.prompt-template-editor :deep(.md-editor-preview li),
.prompt-template-editor :deep(.md-editor-preview blockquote),
.prompt-template-editor :deep(.md-editor-preview table),
.prompt-template-editor :deep(.md-editor-preview td),
.prompt-template-editor :deep(.md-editor-preview th),
.prompt-template-editor :deep(.md-editor-preview code),
.branch-name-editor :deep(.md-editor-preview p),
.branch-name-editor :deep(.md-editor-preview li),
.branch-name-editor :deep(.md-editor-preview blockquote),
.branch-name-editor :deep(.md-editor-preview table),
.branch-name-editor :deep(.md-editor-preview td),
.branch-name-editor :deep(.md-editor-preview th),
.branch-name-editor :deep(.md-editor-preview code) {
  font-size: 13px !important;
}

.prompt-template-footer {
  padding-top: 12px;
  flex-shrink: 0;
}

.dev-environment-section {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.dev-environment-card {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  min-height: 0;
}

.dev-environment-editor {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.dev-environment-editor :deep(.md-editor) {
  --md-font-size: 13px;
  --md-code-font-size: 13px;
  --md-font-family: Consolas, Avenir, Helvetica, Arial, sans-serif;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.dev-environment-editor :deep(.md-editor-input),
.dev-environment-editor :deep(.md-editor-preview-wrapper),
.dev-environment-editor :deep(.cm-content),
.dev-environment-editor :deep(.md-editor-preview) {
  font-family: Consolas, Avenir, Helvetica, Arial, sans-serif !important;
  font-size: 13px !important;
  line-height: 1.6;
}

.dev-environment-editor :deep(.md-editor-preview p),
.dev-environment-editor :deep(.md-editor-preview li),
.dev-environment-editor :deep(.md-editor-preview blockquote),
.dev-environment-editor :deep(.md-editor-preview table),
.dev-environment-editor :deep(.md-editor-preview td),
.dev-environment-editor :deep(.md-editor-preview th),
.dev-environment-editor :deep(.md-editor-preview code) {
  font-size: 13px !important;
}

.dev-environment-footer {
  padding-top: 12px;
  flex-shrink: 0;
}

/* MdEditor 预览区域绿色滚动条 —— 覆盖 md-editor-v3 的 CSS 变量 */
.set-config-page .md-editor {
  --md-scrollbar-bg-color: #edf3e8;
  --md-scrollbar-thumb-color: #9fb39a;
  --md-scrollbar-thumb-hover-color: #869c82;
  --md-scrollbar-thumb-active-color: #7a8f76;
}

/* 同时覆盖原生滚动条样式 */
.set-config-page .md-editor .md-editor-preview ::-webkit-scrollbar {
  width: 10px !important;
  height: 10px !important;
}

.set-config-page .md-editor .md-editor-preview ::-webkit-scrollbar-track {
  background: #edf3e8 !important;
  border-radius: 999px !important;
}

.set-config-page .md-editor .md-editor-preview ::-webkit-scrollbar-thumb {
  background: #9fb39a !important;
  border: 2px solid #edf3e8 !important;
  border-radius: 999px !important;
}

.set-config-page .md-editor .md-editor-preview ::-webkit-scrollbar-thumb:hover {
  background: #869c82 !important;
}

.set-config-page .md-editor .md-editor-preview ::-webkit-scrollbar-corner {
  background: #edf3e8 !important;
}
</style>

<style>
.set-config-page .daily-report-editor .md-editor-preview,
.set-config-page .daily-report-editor .md-editor-preview-wrapper,
.set-config-page .daily-report-editor .md-editor-html,
.set-config-page .prompt-template-editor .md-editor-preview,
.set-config-page .prompt-template-editor .md-editor-preview-wrapper,
.set-config-page .prompt-template-editor .md-editor-html,
.set-config-page .branch-name-editor .md-editor-preview,
.set-config-page .branch-name-editor .md-editor-preview-wrapper,
.set-config-page .branch-name-editor .md-editor-html,
.set-config-page .dev-environment-editor .md-editor-preview,
.set-config-page .dev-environment-editor .md-editor-preview-wrapper,
.set-config-page .dev-environment-editor .md-editor-html {
  font-family: Consolas, Avenir, Helvetica, Arial, sans-serif !important;
  font-size: 13px !important;
  line-height: 1.6 !important;
}

.set-config-page .daily-report-editor .md-editor-preview *,
.set-config-page .daily-report-editor .md-editor-html *,
.set-config-page .prompt-template-editor .md-editor-preview *,
.set-config-page .prompt-template-editor .md-editor-html *,
.set-config-page .branch-name-editor .md-editor-preview *,
.set-config-page .branch-name-editor .md-editor-html *,
.set-config-page .dev-environment-editor .md-editor-preview *,
.set-config-page .dev-environment-editor .md-editor-html * {
  font-size: inherit !important;
}

.set-config-page .daily-report-editor .cm-content,
.set-config-page .daily-report-editor .md-editor-input,
.set-config-page .prompt-template-editor .cm-content,
.set-config-page .prompt-template-editor .md-editor-input,
.set-config-page .branch-name-editor .cm-content,
.set-config-page .branch-name-editor .md-editor-input,
.set-config-page .dev-environment-editor .cm-content,
.set-config-page .dev-environment-editor .md-editor-input {
  font-family: Consolas, Avenir, Helvetica, Arial, sans-serif !important;
  font-size: 13px !important;
  line-height: 1.6 !important;
}
</style>
