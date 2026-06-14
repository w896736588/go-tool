<template>
  <div class="wf-template-manager">
    <!-- 左侧模板列表 -->
    <div class="wf-template-manager__left">
      <div class="wf-template-manager__section-title">
        <span>模板列表</span>
        <el-button type="primary" size="small" @click="addTemplate">
          <el-icon><Plus /></el-icon> 新建
        </el-button>
      </div>
      <div class="wf-template-manager__template-list">
        <div
          v-for="tpl in templates"
          :key="tpl.id"
          :class="['wf-template-manager__template-item', {
            'wf-template-manager__template-item--active': selectedTemplateId === tpl.id,
            'wf-template-manager__template-item--default': tpl.is_default === 1
          }]"
          @click="selectTemplate(tpl)"
        >
          <div class="wf-template-manager__template-name">
            {{ tpl.name }}
            <el-tag v-if="tpl.is_default === 1" size="small" type="success" effect="plain">默认</el-tag>
          </div>
          <div class="wf-template-manager__template-desc">{{ tpl.description || '无描述' }}</div>
          <div class="wf-template-manager__template-actions" v-if="tpl.is_default !== 1">
            <el-button text size="small" type="danger" @click.stop="deleteTemplateConfirm(tpl)">删除</el-button>
          </div>
        </div>
        <div v-if="templates.length === 0" class="wf-template-manager__empty">
          暂无模板，请点击"新建"创建
        </div>
      </div>
    </div>

    <!-- 右侧步骤编辑区域 -->
    <div class="wf-template-manager__right" v-if="selectedTemplate">
      <div class="wf-template-manager__section-title">
        <span>模板信息</span>
        <el-button type="primary" size="small" :loading="savingTemplate" @click="saveTemplate">保存模板</el-button>
      </div>
      <div class="wf-template-manager__template-form">
        <el-input
          v-model="editTemplateName"
          placeholder="模板名称（必填）"
          maxlength="50"
          show-word-limit
          class="wf-template-manager__form-item"
        />
        <el-input
          v-model="editTemplateDesc"
          placeholder="模板描述（选填）"
          maxlength="200"
          show-word-limit
          type="textarea"
          :rows="2"
          class="wf-template-manager__form-item"
        />
      </div>

      <!-- 步骤列表 -->
      <div class="wf-template-manager__section-title" style="margin-top: 16px;">
        <span>步骤列表（拖拽排序）</span>
        <el-button type="success" size="small" @click="addStep">
          <el-icon><Plus /></el-icon> 添加步骤
        </el-button>
      </div>

      <div class="wf-template-manager__step-list" ref="stepListRef">
        <draggable
          v-model="editingSteps"
          item-key="id"
          handle=".wf-template-manager__step-drag-handle"
          ghost-class="wf-template-manager__step-ghost"
          @end="onStepDragEnd"
        >
          <template #item="{ element, index }">
            <div :class="['wf-template-manager__step-item', {
              'wf-template-manager__step-item--fixed': element.is_fixed === 1,
              'wf-template-manager__step-item--expanded': expandedStepId === element._key,
            }]">
              <div class="wf-template-manager__step-header">
                <!-- 拖拽手柄 -->
                <span class="wf-template-manager__step-drag-handle" v-if="element.is_fixed !== 1">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                    <circle cx="9" cy="5" r="1.5"/><circle cx="15" cy="5" r="1.5"/>
                    <circle cx="9" cy="12" r="1.5"/><circle cx="15" cy="12" r="1.5"/>
                    <circle cx="9" cy="19" r="1.5"/><circle cx="15" cy="19" r="1.5"/>
                  </svg>
                </span>
                <!-- 固定步骤锁图标 -->
                <span class="wf-template-manager__step-lock" v-if="element.is_fixed === 1" title="固定步骤，不可删除">
                  <el-icon><Lock /></el-icon>
                </span>
                <span class="wf-template-manager__step-index">{{ index + 1 }}.</span>
                <el-input
                  v-model="element.name"
                  placeholder="步骤名称"
                  size="small"
                  class="wf-template-manager__step-name"
                  :disabled="element.is_fixed === 1 && element.step_key === 'task-config'"
                  @blur="saveStepSilent(element)"
                />
                <div class="wf-template-manager__step-actions">
                  <el-button
                    text
                    size="small"
                    :type="expandedStepId === element._key ? 'warning' : 'primary'"
                    @click="toggleStepExpand(element)"
                    :disabled="element.is_fixed === 1 && element.step_key === 'task-config'"
                  >
                    {{ expandedStepId === element._key ? '收起' : '提示词' }}
                  </el-button>
                  <el-button
                    v-if="element.is_fixed !== 1"
                    text
                    size="small"
                    type="danger"
                    @click="deleteStepConfirm(element)"
                  >删除</el-button>
                </div>
              </div>

              <!-- 展开的提示词编辑区 -->
              <div v-if="expandedStepId === element._key" class="wf-template-manager__step-editor">
                <UnifiedMdEditor
                  v-if="element.prompt_content !== undefined"
                  :ref="el => stepEditorRefs[element._key] = el"
                  :model-value="element.prompt_content"
                  :toolbars="promptEditorToolbars"
                  class="wf-template-manager__step-md-editor"
                  @update:model-value="val => element.prompt_content = val"
                />
                <div class="wf-template-manager__step-editor-actions">
                  <el-button size="small" type="primary" @click="saveStepPrompt(element)">保存提示词</el-button>
                </div>
              </div>
            </div>
          </template>
        </draggable>

        <div v-if="editingSteps.length === 0" class="wf-template-manager__empty">
          暂无步骤，请点击"添加步骤"
        </div>
      </div>
    </div>

    <!-- 未选择模板时的占位 -->
    <div class="wf-template-manager__right" v-else>
      <div class="wf-template-manager__empty wf-template-manager__empty--full">
        请从左侧选择一个模板，或新建一个模板
      </div>
    </div>
  </div>
</template>

<script>
import { Plus, Lock } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import draggable from 'vuedraggable'
import UnifiedMdEditor from '@/components/base/UnifiedMdEditor.vue'
import workflowTemplateApi from '@/utils/base/workflow_template'

let stepIdCounter = 0

// 提示词编辑器工具栏配置
const PROMPT_EDITOR_TOOLBARS = [
  'bold', 'underline', 'italic', '-', 'strikeThrough', 'title', 'sub', 'sup', 'quote', 'unorderedList', 'orderedList', '-', 'codeRow', 'code', 'link', 'image', 'table',
]

export default {
  name: 'WorkflowTemplateManager',
  components: {
    Plus,
    Lock,
    draggable,
    UnifiedMdEditor,
  },
  emits: ['templates-loaded'],
  data() {
    return {
      templates: [],
      selectedTemplateId: 0,
      selectedTemplate: null,
      editTemplateName: '',
      editTemplateDesc: '',
      editingSteps: [],
      expandedStepId: '',
      savingTemplate: false,
      loading: false,
      promptEditorToolbars: PROMPT_EDITOR_TOOLBARS,
      stepEditorRefs: {},
    }
  },
  mounted() {
    this.loadTemplates()
  },
  methods: {
    // ===== 模板数据加载 =====
    loadTemplates() {
      this.loading = true
      workflowTemplateApi.WorkflowTemplateList((response) => {
        this.loading = false
        if (response && response.ErrCode === 0 && response.Data && response.Data.list) {
          this.templates = response.Data.list
          // 设置步骤的 _key 用于展开/折叠跟踪
          this.templates.forEach(t => {
            if (t.steps) {
              t.steps.forEach(s => {
                s._key = 'step_' + s.id
              })
            }
          })
          this.$emit('templates-loaded', this.templates)
          // 如果有模板且未选中，自动选中第一个
          if (this.templates.length > 0 && this.selectedTemplateId === 0) {
            this.selectTemplate(this.templates[0])
          }
        }
      })
    },

    // ===== 模板选择 =====
    selectTemplate(tpl) {
      this.selectedTemplateId = tpl.id
      this.selectedTemplate = tpl
      this.editTemplateName = tpl.name
      this.editTemplateDesc = tpl.description || ''
      this.editingSteps = (tpl.steps || []).map(s => ({
        ...s,
        _key: 'step_' + (s.id || ++stepIdCounter),
      }))
      this.expandedStepId = ''
    },

    // ===== 模板 CRUD =====
    addTemplate() {
      const newTpl = {
        id: 0,
        name: '新模板',
        description: '',
        is_default: 0,
        steps: [],
      }
      this.templates.push(newTpl)
      this.selectTemplate(newTpl)
    },

    saveTemplate() {
      if (!this.editTemplateName || !this.editTemplateName.trim()) {
        ElMessage.warning('请填写模板名称')
        return
      }
      this.savingTemplate = true
      workflowTemplateApi.WorkflowTemplateSave({
        id: this.selectedTemplateId || 0,
        name: this.editTemplateName.trim(),
        description: this.editTemplateDesc.trim(),
      }, (response) => {
        this.savingTemplate = false
        if (response && response.ErrCode === 0) {
          const saved = response.Data.template
          if (saved && saved.steps) {
            saved.steps.forEach(s => { s._key = 'step_' + s.id })
          }
          // 更新列表中的模板
          const idx = this.templates.findIndex(t => t.id === this.selectedTemplateId)
          if (idx >= 0) {
            this.templates[idx] = saved
          } else if (this.selectedTemplateId === 0) {
            // 新增模板，替换占位项
            const placeholderIdx = this.templates.findIndex(t => t.id === 0 && t.name === '新模板')
            if (placeholderIdx >= 0) {
              this.templates[placeholderIdx] = saved
            }
          }
          this.selectedTemplateId = saved.id
          this.selectedTemplate = saved
          this.editingSteps = (saved.steps || []).map(s => ({
            ...s,
            _key: 'step_' + s.id,
          }))
          ElMessage.success('模板已保存')
          this.$emit('templates-loaded', this.templates)
        } else {
          ElMessage.error(response.ErrMsg || '保存失败')
        }
      })
    },

    deleteTemplateConfirm(tpl) {
      ElMessageBox.confirm(`确定删除模板"${tpl.name}"吗？此操作不可恢复。`, '确认删除', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        this.deleteTemplate(tpl)
      }).catch(() => {})
    },

    deleteTemplate(tpl) {
      workflowTemplateApi.WorkflowTemplateDelete(tpl.id, (response) => {
        if (response && response.ErrCode === 0) {
          this.templates = this.templates.filter(t => t.id !== tpl.id)
          if (this.selectedTemplateId === tpl.id) {
            this.selectedTemplateId = 0
            this.selectedTemplate = null
            this.editingSteps = []
          }
          ElMessage.success('模板已删除')
          this.$emit('templates-loaded', this.templates)
        } else {
          ElMessage.error(response.ErrMsg || '删除失败')
        }
      })
    },

    // ===== 步骤 CRUD =====
    addStep() {
      if (!this.selectedTemplateId) {
        ElMessage.warning('请先保存模板后再添加步骤')
        return
      }
      const newStep = {
        id: 0,
        template_id: this.selectedTemplateId,
        name: '新步骤',
        step_key: '',
        prompt_content: '',
        sort_order: this.editingSteps.length,
        is_fixed: 0,
        _key: 'step_new_' + (++stepIdCounter),
      }
      this.editingSteps.push(newStep)
      this.expandedStepId = newStep._key
    },

    toggleStepExpand(step) {
      this.expandedStepId = this.expandedStepId === step._key ? '' : step._key
    },

    saveStepSilent(step) {
      if (!step.name || !step.name.trim()) return
      // 仅名称变更时静默保存
      workflowTemplateApi.WorkflowTemplateStepSave({
        id: step.id || 0,
        template_id: this.selectedTemplateId,
        name: step.name.trim(),
        step_key: step.step_key || '',
        prompt_content: step.prompt_content || '',
        sort_order: step.sort_order || 0,
      }, (response) => {
        if (response && response.ErrCode === 0 && response.Data && response.Data.step) {
          const saved = response.Data.step
          saved._key = 'step_' + saved.id
          Object.assign(step, saved)
        }
      })
    },

    saveStepPrompt(step) {
      workflowTemplateApi.WorkflowTemplateStepSave({
        id: step.id || 0,
        template_id: this.selectedTemplateId,
        name: step.name.trim(),
        step_key: step.step_key || '',
        prompt_content: step.prompt_content || '',
        sort_order: step.sort_order || 0,
      }, (response) => {
        if (response && response.ErrCode === 0 && response.Data && response.Data.step) {
          const saved = response.Data.step
          saved._key = 'step_' + saved.id
          Object.assign(step, saved)
          ElMessage.success('提示词已保存')
        } else {
          ElMessage.error(response.ErrMsg || '保存失败')
        }
      })
    },

    deleteStepConfirm(step) {
      if (!step.id) {
        // 未保存的步骤直接从列表中移除
        this.editingSteps = this.editingSteps.filter(s => s._key !== step._key)
        return
      }
      ElMessageBox.confirm(`确定删除步骤"${step.name}"吗？`, '确认删除', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        this.deleteStep(step)
      }).catch(() => {})
    },

    deleteStep(step) {
      workflowTemplateApi.WorkflowTemplateStepDelete(step.id, (response) => {
        if (response && response.ErrCode === 0) {
          this.editingSteps = this.editingSteps.filter(s => s.id !== step.id)
          if (this.expandedStepId === step._key) {
            this.expandedStepId = ''
          }
          // 重新排序
          this.saveStepSort()
          ElMessage.success('步骤已删除')
        } else {
          ElMessage.error(response.ErrMsg || '删除失败')
        }
      })
    },

    // ===== 拖拽排序 =====
    onStepDragEnd() {
      // 拖拽结束后，更新每个步骤的 sort_order
      const stepIds = this.editingSteps.map(s => s.id).filter(id => id > 0)
      if (stepIds.length > 0) {
        this.saveStepSort(stepIds)
      }
    },

    saveStepSort(stepIds) {
      if (!stepIds) {
        stepIds = this.editingSteps.map(s => s.id).filter(id => id > 0)
      }
      if (stepIds.length === 0) return
      workflowTemplateApi.WorkflowTemplateStepSort(this.selectedTemplateId, stepIds)
    },
  },
}
</script>

<style scoped>
.wf-template-manager {
  display: flex;
  gap: 16px;
  height: 100%;
  min-height: 500px;
}

/* 左侧模板列表 */
.wf-template-manager__left {
  width: 260px;
  min-width: 260px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  background: var(--el-bg-color);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.wf-template-manager__section-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.wf-template-manager__template-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.wf-template-manager__template-item {
  padding: 10px;
  border-radius: 6px;
  cursor: pointer;
  margin-bottom: 6px;
  transition: all 0.2s;
  border: 1px solid transparent;
  position: relative;
}

.wf-template-manager__template-item:hover {
  background: var(--el-fill-color-light);
}

.wf-template-manager__template-item--active {
  background: rgba(103, 194, 58, 0.08);
  border-color: #67c23a;
}

.wf-template-manager__template-item--default {
  /* 默认模板特殊样式 */
}

.wf-template-manager__template-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  display: flex;
  align-items: center;
  gap: 6px;
}

.wf-template-manager__template-desc {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wf-template-manager__template-actions {
  margin-top: 6px;
  text-align: right;
}

/* 右侧步骤编辑区域 */
.wf-template-manager__right {
  flex: 1;
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  background: var(--el-bg-color);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.wf-template-manager__template-form {
  padding: 8px 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.wf-template-manager__form-item {
  margin-bottom: 8px;
}

.wf-template-manager__step-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.wf-template-manager__step-item {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  margin-bottom: 6px;
  background: var(--el-bg-color);
  transition: all 0.2s;
}

.wf-template-manager__step-item:hover {
  border-color: var(--el-border-color);
}

.wf-template-manager__step-item--fixed {
  background: var(--el-fill-color-light);
  opacity: 0.85;
}

.wf-template-manager__step-item--expanded {
  border-color: #67c23a;
}

.wf-template-manager__step-ghost {
  opacity: 0.4;
  background: rgba(103, 194, 58, 0.1);
}

.wf-template-manager__step-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 10px;
}

.wf-template-manager__step-drag-handle {
  cursor: grab;
  color: var(--el-text-color-placeholder);
  flex-shrink: 0;
}

.wf-template-manager__step-drag-handle:active {
  cursor: grabbing;
}

.wf-template-manager__step-lock {
  color: var(--el-text-color-placeholder);
  flex-shrink: 0;
}

.wf-template-manager__step-index {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
  min-width: 24px;
  text-align: right;
}

.wf-template-manager__step-name {
  flex: 1;
}

.wf-template-manager__step-actions {
  flex-shrink: 0;
  display: flex;
  gap: 4px;
}

.wf-template-manager__step-editor {
  padding: 0 10px 10px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.wf-template-manager__step-md-editor {
  margin-top: 8px;
  height: 300px;
}

.wf-template-manager__step-editor-actions {
  margin-top: 8px;
  display: flex;
  justify-content: flex-end;
}

/* 空状态 */
.wf-template-manager__empty {
  padding: 40px;
  text-align: center;
  color: var(--el-text-color-placeholder);
  font-size: 14px;
}

.wf-template-manager__empty--full {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
}
</style>
