<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="$emit('update:visible', $event)"
    title="文件变更详情"
    width="90%"
    top="3vh"
    :close-on-click-modal="true"
    @close="handleClose"
  >
    <div class="file-changes-detail" v-loading="loading">
      <!-- 工具栏 -->
      <div class="file-changes-detail__toolbar">
        <div class="file-changes-detail__info">
          <span class="file-changes-detail__dir">{{ localDir || '-' }}</span>
          <span class="file-changes-detail__mode">
            <el-radio-group v-model="changeMode" size="small" @change="onChangeMode">
              <el-radio-button value="branch">对比代码</el-radio-button>
              <el-radio-button value="workspace">工作区变更</el-radio-button>
            </el-radio-group>
          </span>
          <span v-if="changeMode === 'branch'" class="file-changes-detail__branch">基分支: {{ parentBranch || '-' }}</span>
          <span v-else class="file-changes-detail__branch">对比 HEAD</span>
          <span class="file-changes-detail__summary">
            <template v-if="summary">
              <span v-if="changeMode === 'branch'" class="file-changes-stat file-changes-stat--committed">{{ summary.committed || 0 }} C</span>
              <span class="file-changes-stat file-changes-stat--staged">{{ summary.staged || 0 }} S</span>
              <span class="file-changes-stat file-changes-stat--modified">{{ summary.modified || 0 }} M</span>
              <span class="file-changes-stat file-changes-stat--untracked">{{ summary.untracked || 0 }} U</span>
              <span class="file-changes-stat file-changes-stat--additions">+{{ totalAdditions }}</span>
              <span class="file-changes-stat file-changes-stat--deletions">-{{ totalDeletions }}</span>
            </template>
          </span>
        </div>
        <div class="file-changes-detail__actions">
          <el-radio-group v-model="diffMode" size="small" @change="renderCurrentDiff" v-if="!isBinary && !isImage">
            <el-radio-button value="side-by-side">横向对比</el-radio-button>
            <el-radio-button value="line-by-line">纵向对比</el-radio-button>
          </el-radio-group>
        </div>
      </div>
      <!-- 文件类型过滤器 -->
      <div class="file-changes-detail__filter-bar" v-if="fileTypeCategories.length > 0">
        <span class="file-changes-detail__filter-label">文件类型:</span>
        <template v-for="cat in fileTypeCategories" :key="cat.label">
          <el-checkbox
            v-for="ext in cat.extensions"
            :key="ext"
            :model-value="fileTypeFilter[ext] !== false"
            :label="ext"
            size="small"
            @change="(val) => toggleFileType(ext, val)"
          >
            {{ ext }}
          </el-checkbox>
          <span v-if="cat !== fileTypeCategories[fileTypeCategories.length - 1]" class="file-changes-detail__filter-sep"></span>
        </template>
      </div>
      <!-- 主体：左侧文件树 + 右侧 diff -->
      <div class="file-changes-detail__body">
        <!-- 左侧文件树 -->
        <div class="file-changes-detail__tree-panel">
          <div class="file-changes-detail__tree-title">改动文件列表 ({{ files.length }})</div>
          <div class="file-changes-detail__tree">
            <template v-if="flatTreeItems.length > 0">
              <div
                v-for="item in flatTreeItems"
                :key="item.key"
                :class="[item.isDir ? 'file-changes-tree__dir' : 'file-changes-tree__file', !item.isDir && item.path === selectedFile ? 'file-changes-tree__file--active' : '']"
                :style="{ paddingLeft: (12 + item.depth * 16) + 'px' }"
                @click="item.isDir ? toggleDir(item.key) : selectFile(item)"
              >
                <template v-if="item.isDir">
                  <span class="file-changes-tree__dir-icon">{{ expandedDirs[item.key] !== false ? '▼' : '▶' }}</span>
                  <span class="file-changes-tree__dir-name">{{ item.name }}/</span>
                  <span class="file-changes-tree__dir-count">({{ item.fileCount }})</span>
                </template>
                <template v-else>
                  <span class="file-changes-tree__file-type" :class="'file-changes-tree__file-type--' + item.type">{{ item.status_code }}</span>
                  <span class="file-changes-tree__file-name" :title="item.path">{{ item.name }}</span>
                </template>
              </div>
            </template>
            <div v-else class="file-changes-detail__tree-empty">暂无文件变更</div>
          </div>
        </div>
        <!-- 右侧 diff 视图 -->
        <div class="file-changes-detail__diff-panel">
          <template v-if="diffLoading">
            <div class="file-changes-detail__diff-placeholder file-changes-detail__diff-loading">
              <i class="el-icon-loading"></i>
              <span>加载中...</span>
            </div>
          </template>
          <template v-else-if="diffError">
            <div class="file-changes-detail__diff-error">{{ diffError }}</div>
          </template>
          <!-- 二进制文件展示 -->
          <template v-else-if="isBinary && selectedFile">
            <div class="file-changes-detail__diff-header">
              <span class="file-changes-detail__diff-file">{{ selectedFile }}</span>
              <span class="file-changes-detail__diff-badge file-changes-detail__diff-badge--binary">二进制文件</span>
              <span v-if="selectedFileInfo" class="file-changes-detail__diff-stats">
                <span class="file-changes-detail__diff-stat--add">+{{ selectedFileInfo.additions || 0 }}</span>
                <span class="file-changes-detail__diff-stat--del">-{{ selectedFileInfo.deletions || 0 }}</span>
              </span>
              <div class="file-changes-detail__nav-btns">
                <el-button size="small" :disabled="!hasPrevFile" @click="goToPrevFile">上一个文件</el-button>
                <el-button size="small" :disabled="!hasNextFile" @click="goToNextFile">下一个文件</el-button>
              </div>
            </div>
            <div class="file-changes-detail__binary-info">
              <div class="file-changes-detail__binary-icon">
                <span class="file-changes-detail__binary-ext">{{ binaryInfo.file_type || '' }}</span>
              </div>
              <div class="file-changes-detail__binary-meta">
                <div class="file-changes-detail__binary-row">
                  <span class="file-changes-detail__binary-label">文件类型</span>
                  <span class="file-changes-detail__binary-value">{{ binaryInfo.file_type || '-' }}</span>
                </div>
                <div class="file-changes-detail__binary-row">
                  <span class="file-changes-detail__binary-label">原始版本大小</span>
                  <span class="file-changes-detail__binary-value file-changes-detail__binary-value--old">{{ formatFileSize(binaryInfo.old_size) }}</span>
                </div>
                <div class="file-changes-detail__binary-row">
                  <span class="file-changes-detail__binary-label">改后版本大小</span>
                  <span :class="sizeChangeClass">{{ formatFileSize(binaryInfo.new_size) }}</span>
                  <span v-if="sizeDiffText" :class="sizeChangeClass">({{ sizeDiffText }})</span>
                </div>
              </div>
              <div class="file-changes-detail__binary-hint">
                二进制文件不支持内容比对，仅显示文件大小变化
              </div>
            </div>
          </template>
          <!-- 图片文件展示 -->
          <template v-else-if="isImage && selectedFile">
            <div class="file-changes-detail__diff-header">
              <span class="file-changes-detail__diff-file">{{ selectedFile }}</span>
              <span class="file-changes-detail__diff-badge file-changes-detail__diff-badge--image">图片</span>
              <span v-if="selectedFileInfo" class="file-changes-detail__diff-stats">
                <span class="file-changes-detail__diff-stat--add">+{{ selectedFileInfo.additions || 0 }}</span>
                <span class="file-changes-detail__diff-stat--del">-{{ selectedFileInfo.deletions || 0 }}</span>
              </span>
              <div class="file-changes-detail__nav-btns">
                <el-button size="small" :disabled="!hasPrevFile" @click="goToPrevFile">上一个文件</el-button>
                <el-button size="small" :disabled="!hasNextFile" @click="goToNextFile">下一个文件</el-button>
              </div>
            </div>
            <div class="file-changes-detail__image-compare">
              <div class="file-changes-detail__image-pane">
                <div class="file-changes-detail__image-title">原始版本</div>
                <div class="file-changes-detail__image-box">
                  <img v-if="imageInfo.old_image" :src="'data:image/' + imageInfo.image_type + ';base64,' + imageInfo.old_image" class="file-changes-detail__image-img" />
                  <span v-else class="file-changes-detail__image-empty">(文件不存在或已删除)</span>
                </div>
              </div>
              <div class="file-changes-detail__image-divider"></div>
              <div class="file-changes-detail__image-pane">
                <div class="file-changes-detail__image-title">改后版本</div>
                <div class="file-changes-detail__image-box">
                  <img v-if="imageInfo.new_image" :src="'data:image/' + imageInfo.image_type + ';base64,' + imageInfo.new_image" class="file-changes-detail__image-img" />
                  <span v-else class="file-changes-detail__image-empty">(文件不存在或已删除)</span>
                </div>
              </div>
            </div>
          </template>
          <template v-else-if="selectedFile">
            <div class="file-changes-detail__diff-header">
              <span class="file-changes-detail__diff-file">{{ selectedFile }}</span>
              <span v-if="selectedFileInfo" class="file-changes-detail__diff-stats">
                <span class="file-changes-detail__diff-stat--add">+{{ selectedFileInfo.additions || 0 }}</span>
                <span class="file-changes-detail__diff-stat--del">-{{ selectedFileInfo.deletions || 0 }}</span>
              </span>
              <div class="file-changes-detail__nav-btns" v-if="selectedFile">
                <el-button size="small" :disabled="!canPrev" @click="jumpToPrevChange">{{ prevBtnText }}</el-button>
                <el-button size="small" :disabled="!canNext" @click="jumpToNextChange">{{ nextBtnText }}</el-button>
              </div>
            </div>
            <div ref="diffContainer" class="file-changes-detail__diff-content"></div>
          </template>
          <template v-else>
            <div class="file-changes-detail__diff-placeholder">
              <span>请从左侧选择文件查看变更详情</span>
            </div>
          </template>
        </div>
      </div>
    </div>
    <template #footer>
      <div class="file-changes-detail__footer">
        <GitActionButton compact variant="info" @click="handleClose">关闭</GitActionButton>
      </div>
    </template>
  </el-dialog>
</template>

<script>
import GitActionButton from '@/components/base/GitActionButton.vue'
import taskWorkflowApi from '@/utils/base/task_workflow'
import * as Diff2Html from 'diff2html'
import 'diff2html/bundles/css/diff2html.min.css'

// CodeMirror MergeView
import CodeMirror from 'codemirror'
import 'codemirror/lib/codemirror.css'
import 'codemirror/addon/merge/merge.js'
import 'codemirror/addon/merge/merge.css'
import 'codemirror/addon/fold/foldgutter.css'
import 'codemirror/addon/fold/foldcode.js'
import 'codemirror/addon/fold/foldgutter.js'

// 语法高亮模式（按需导入）
import 'codemirror/mode/javascript/javascript.js'
import 'codemirror/mode/go/go.js'
import 'codemirror/mode/python/python.js'
import 'codemirror/mode/css/css.js'
import 'codemirror/mode/htmlmixed/htmlmixed.js'
import 'codemirror/mode/xml/xml.js'
import 'codemirror/mode/sql/sql.js'
import 'codemirror/mode/markdown/markdown.js'
import 'codemirror/mode/shell/shell.js'
import 'codemirror/mode/yaml/yaml.js'
import 'codemirror/mode/vue/vue.js'
import 'codemirror/mode/php/php.js'
import 'codemirror/mode/ruby/ruby.js'
import 'codemirror/mode/clike/clike.js'
import 'codemirror/mode/nginx/nginx.js'
import 'codemirror/mode/dockerfile/dockerfile.js'
import 'codemirror/mode/toml/toml.js'
import 'codemirror/mode/jsx/jsx.js'
import 'codemirror/mode/sass/sass.js'
import 'codemirror/mode/powershell/powershell.js'

// diff-match-patch（CodeMirror MergeView 依赖）
import DiffMatchPatch from 'diff-match-patch'
// CodeMirror merge addon 直接引用全局常量，需手动挂载
window.diff_match_patch = DiffMatchPatch
window.DIFF_EQUAL = DiffMatchPatch.DIFF_EQUAL
window.DIFF_INSERT = DiffMatchPatch.DIFF_INSERT
window.DIFF_DELETE = DiffMatchPatch.DIFF_DELETE

/**
 * patchCmPrepareSelection 为 CodeMirror 实例的 display.input.prepareSelection 方法
 * 添加 try-catch 防护，作为兜底方案。
 */
function patchCmPrepareSelection(cm) {
  if (!cm || !cm.display || !cm.display.input) return
  const input = cm.display.input
  if (input._prepareSelectionPatched) return
  const origPrepare = input.prepareSelection.bind(input)
  input.prepareSelection = function () {
    try {
      return origPrepare()
    } catch (e) {
      return { cursors: document.createDocumentFragment(), selection: document.createDocumentFragment() }
    }
  }
  input._prepareSelectionPatched = true
}

// 文件扩展名 → CodeMirror mode 映射
const EXT_MODE_MAP = {
  '.js': 'javascript',
  '.mjs': 'javascript',
  '.cjs': 'javascript',
  '.jsx': 'jsx',
  '.ts': 'javascript',
  '.tsx': 'jsx',
  '.vue': 'vue',
  '.go': 'go',
  '.py': 'python',
  '.css': 'css',
  '.scss': 'sass',
  '.less': 'css',
  '.html': 'htmlmixed',
  '.htm': 'htmlmixed',
  '.xml': 'xml',
  '.svg': 'xml',
  '.sql': 'sql',
  '.md': 'markdown',
  '.markdown': 'markdown',
  '.sh': 'shell',
  '.bash': 'shell',
  '.zsh': 'shell',
  '.yml': 'yaml',
  '.yaml': 'yaml',
  '.php': 'php',
  '.rb': 'ruby',
  '.java': 'text/x-java',
  '.c': 'text/x-csrc',
  '.cpp': 'text/x-c++src',
  '.h': 'text/x-csrc',
  '.hpp': 'text/x-c++src',
  '.cs': 'text/x-csharp',
  '.json': { name: 'javascript', json: true },
  '.toml': 'toml',
  '.ini': 'toml',
  '.cfg': 'toml',
  '.conf': 'nginx',
  '.dockerfile': 'dockerfile',
  '.ps1': 'shell',
  '.bat': 'shell',
  '.cmd': 'shell',
}

function getModeForFile(filePath) {
  if (!filePath) return 'text'
  const fileName = filePath.split('/').pop() || filePath
  // 特殊文件名
  if (fileName === 'Dockerfile' || fileName.startsWith('Dockerfile.')) return 'dockerfile'
  if (fileName === 'Makefile' || fileName === 'Taskfile.yml') return 'yaml'
  if (fileName === '.gitignore' || fileName === '.dockerignore' || fileName === '.eslintignore') return 'shell'
  const dotIdx = fileName.lastIndexOf('.')
  if (dotIdx < 0) return 'text'
  const ext = fileName.substring(dotIdx).toLowerCase()
  return EXT_MODE_MAP[ext] || 'text'
}

export default {
  name: 'FileChangesDialog',
  components: { GitActionButton },
  props: {
    visible: { type: Boolean, default: false },
    localDir: { type: String, default: '' },
    parentBranch: { type: String, default: '' },
    initialSummary: { type: Object, default: null },
    initialFiles: { type: Array, default: () => [] },
  },
  emits: ['update:visible'],
  data() {
    return {
      loading: false,
      summary: null,
      files: [],
      selectedFile: '',
      currentDiff: '',
      oldContent: '',
      newContent: '',
      diffError: '',
      diffLoading: false,
      diffMode: 'side-by-side',
      changeMode: 'workspace', // 'workspace' 工作区变更 | 'branch' 对比代码
      expandedDirs: {},
      mergeView: null,
      // 二进制/图片文件相关
      isBinary: false,
      isImage: false,
      binaryInfo: { file_type: '', old_size: 0, new_size: 0 },
      imageInfo: { image_type: '', old_image: '', new_image: '' },
      // 文件类型过滤器
      fileTypeFilter: {},
      // diff 块跳转相关
      diffChunks: [],
      currentChunkIndex: -1,
    }
  },
  computed: {
    // 当前模式下的有效基分支（工作区模式下为空）
    effectiveParentBranch() {
      return this.changeMode === 'branch' ? this.parentBranch : ''
    },
    treeNodes() {
      return this.buildTree(this.filteredFiles)
    },
    // 所有唯一文件类型（按类别分组）
    fileTypeCategories() {
      const extSet = new Set()
      for (const f of this.files) {
        const ext = this.getFileExt(f.path || '')
        if (ext) extSet.add(ext)
      }
      const catMap = {}
      for (const ext of extSet) {
        const cat = this.fileTypeCategory(ext)
        if (!catMap[cat]) catMap[cat] = []
        catMap[cat].push(ext)
      }
      // 排序：优先 代码、样式、图片、二进制、其他
      const order = ['代码', '样式', '配置', '脚本', '文档', '图片', '二进制', 'SQL', '其他']
      const result = []
      for (const cat of order) {
        if (catMap[cat] && catMap[cat].length > 0) {
          result.push({ label: cat, extensions: catMap[cat].sort() })
        }
      }
      return result
    },
    // 按文件类型过滤后的文件列表
    filteredFiles() {
      if (!this.files || this.files.length === 0) return []
      const activeTypes = Object.keys(this.fileTypeFilter).filter(k => this.fileTypeFilter[k] !== false)
      // 如果还没有初始化filter或者全部选中，返回全部
      if (activeTypes.length === 0 || Object.keys(this.fileTypeFilter).length === 0) return this.files
      return this.files.filter(f => {
        const ext = this.getFileExt(f.path || '')
        return ext && this.fileTypeFilter[ext] !== false
      })
    },
    // 文件大小变化描述
    sizeDiffText() {
      const diff = this.binaryInfo.new_size - this.binaryInfo.old_size
      if (diff === 0) return ''
      const sign = diff > 0 ? '+' : ''
      return sign + this.formatFileSize(Math.abs(diff))
    },
    sizeChangeClass() {
      const diff = this.binaryInfo.new_size - this.binaryInfo.old_size
      if (diff > 0) return 'file-changes-detail__binary-value file-changes-detail__binary-value--increased'
      if (diff < 0) return 'file-changes-detail__binary-value file-changes-detail__binary-value--decreased'
      return 'file-changes-detail__binary-value'
    },
    // 总增加行数
    totalAdditions() {
      if (this.summary && typeof this.summary.additions === 'number') return this.summary.additions
      return this.files.reduce((sum, f) => sum + (f.additions || 0), 0)
    },
    // 总删除行数
    totalDeletions() {
      if (this.summary && typeof this.summary.deletions === 'number') return this.summary.deletions
      return this.files.reduce((sum, f) => sum + (f.deletions || 0), 0)
    },
    // 当前选中文件的 additions/deletions 信息
    selectedFileInfo() {
      if (!this.selectedFile || !this.files || this.files.length === 0) return null
      return this.files.find(f => f.path === this.selectedFile) || null
    },
    // 当前选中文件在文件列表中的索引
    selectedFileIndex() {
      if (!this.selectedFile || !this.filteredFiles || this.filteredFiles.length === 0) return -1
      return this.filteredFiles.findIndex(f => f.path === this.selectedFile)
    },
    // 是否有上一个文件
    hasPrevFile() {
      return this.selectedFileIndex > 0
    },
    // 是否有下一个文件
    hasNextFile() {
      const idx = this.selectedFileIndex
      return idx >= 0 && idx < this.filteredFiles.length - 1
    },
    // 是否有上一处改动（当前索引 > 0 才能往上跳）
    hasPrevChange() {
      return this.diffChunks.length > 0 && this.currentChunkIndex > 0
    },
    // 是否有下一处改动（还有未跳转的 chunk）
    hasNextChange() {
      return this.diffChunks.length > 0 && (this.currentChunkIndex < 0 || this.currentChunkIndex < this.diffChunks.length - 1)
    },
    // 上一处改动按钮是否可用
    canPrev() {
      return this.hasPrevChange || this.hasPrevFile
    },
    // 下一处改动按钮是否可用
    canNext() {
      return this.hasNextChange || this.hasNextFile
    },
    // 上一处改动按钮文案
    prevBtnText() {
      if (this.hasPrevChange) return '上一处改动'
      if (this.hasPrevFile) return '查看上一个文件'
      return '上一处改动'
    },
    // 下一处改动按钮文案
    nextBtnText() {
      if (this.hasNextChange) return '下一处改动'
      if (this.hasNextFile) return '查看下一个文件'
      return '下一处改动'
    },
    flatTreeItems() {
      const result = []
      const flatten = (nodes, depth) => {
        for (const node of nodes) {
          if (node.isDir) {
            result.push({
              isDir: true,
              name: node.name,
              key: node.key,
              path: node.path,
              fileCount: node.fileCount,
              depth,
            })
            if (this.expandedDirs[node.key] !== false) {
              flatten(node.children, depth + 1)
              for (const file of node.files) {
                result.push({
                  isDir: false,
                  name: file.name,
                  key: 'file_' + file.path,
                  path: file.path,
                  status_code: file.status_code,
                  type: file.type,
                  depth: depth + 1,
                })
              }
            }
          } else {
            result.push({
              isDir: false,
              name: node.name,
              key: node.key,
              path: node.path,
              status_code: node.status_code,
              type: node.type,
              depth,
            })
          }
        }
      }
      flatten(this.treeNodes, 0)
      return result
    },
  },
  watch: {
    visible(val) {
      if (val) {
        this.initData()
      } else {
        this.destroyMergeView()
      }
    },
  },
  beforeUnmount() {
    this.destroyMergeView()
  },
  methods: {
    initData() {
      this.summary = this.initialSummary ? { ...this.initialSummary } : null
      this.files = Array.isArray(this.initialFiles) ? [...this.initialFiles] : []
      this.selectedFile = ''
      this.currentDiff = ''
      this.oldContent = ''
      this.newContent = ''
      this.diffError = ''
      this.diffLoading = false
      this.changeMode = 'workspace'
      this.expandedDirs = {}
      this.isBinary = false
      this.isImage = false
      this.binaryInfo = { file_type: '', old_size: 0, new_size: 0 }
      this.imageInfo = { image_type: '', old_image: '', new_image: '' }
      // 初始化文件类型过滤器（默认全选）
      this.fileTypeFilter = {}
      this.loadFileList()
    },
    onChangeMode() {
      // 切换模式时清空选中文件与 diff 视图，重新加载文件列表
      this.selectedFile = ''
      this.currentDiff = ''
      this.oldContent = ''
      this.newContent = ''
      this.diffError = ''
      this.isBinary = false
      this.isImage = false
      this.binaryInfo = { file_type: '', old_size: 0, new_size: 0 }
      this.imageInfo = { image_type: '', old_image: '', new_image: '' }
      this.destroyMergeView()
      this.fileTypeFilter = {}
      this.loadFileList()
    },
    // 在文件列表加载完成后初始化过滤器
    initFileTypeFilter() {
      const filter = {}
      for (const f of this.files) {
        const ext = this.getFileExt(f.path || '')
        if (ext && filter[ext] === undefined) {
          filter[ext] = true
        }
      }
      this.fileTypeFilter = filter
    },
    toggleFileType(ext, val) {
      this.fileTypeFilter = { ...this.fileTypeFilter, [ext]: val }
    },
    // 获取文件扩展名
    getFileExt(filePath) {
      if (!filePath) return ''
      const name = filePath.split('/').pop() || filePath
      const lower = name.toLowerCase()
      // 复合扩展名
      for (const compound of ['.tar.gz', '.tar.bz2', '.tar.xz']) {
        if (lower.endsWith(compound)) return compound
      }
      const dotIdx = lower.lastIndexOf('.')
      if (dotIdx < 0) return '(无扩展名)'
      return lower.substring(dotIdx)
    },
    // 文件类型归类
    fileTypeCategory(ext) {
      const code = new Set(['.js', '.mjs', '.cjs', '.jsx', '.ts', '.tsx', '.vue', '.go', '.py', '.java', '.c', '.cpp', '.h', '.hpp', '.cs', '.rb', '.php', '.rs', '.swift', '.kt', '.scala', '.dart', '.lua', '.r', '.pl', '.groovy'])
      const style = new Set(['.css', '.scss', '.less', '.sass', '.styl'])
      const config = new Set(['.json', '.yml', '.yaml', '.toml', '.ini', '.cfg', '.conf', '.xml', '.env', '.properties'])
      const script = new Set(['.sh', '.bash', '.zsh', '.bat', '.cmd', '.ps1', '.psm1'])
      const doc = new Set(['.md', '.markdown', '.txt', '.rst', '.adoc'])
      const image = new Set(['.png', '.jpg', '.jpeg', '.gif', '.webp', '.bmp', '.svg', '.tiff', '.tif'])
      const binary = new Set(['.exe', '.dll', '.so', '.dylib', '.bin', '.dat', '.zip', '.tar', '.gz', '.7z', '.rar', '.pdf', '.doc', '.docx', '.xls', '.xlsx', '.ppt', '.pptx', '.ttf', '.otf', '.woff', '.woff2', '.eot', '.mp3', '.mp4', '.avi', '.mov', '.mkv', '.webm', '.wav', '.flac', '.ogg', '.o', '.a', '.lib', '.class', '.jar', '.war', '.pyc', '.pyo', '.wasm', '.ico', '.cur', '.db', '.sqlite', '.sqlite3', '.node', '.lock', '.sum', '.whl', '.tgz', '.tar.gz', '.tar.bz2', '.tar.xz', '.bz2', '.xz', '.iso', '.dmg', '.pkg', '.deb', '.rpm', '.apk', '.msi'])
      const sql = new Set(['.sql'])
      if (code.has(ext)) return '代码'
      if (style.has(ext)) return '样式'
      if (config.has(ext)) return '配置'
      if (script.has(ext)) return '脚本'
      if (doc.has(ext)) return '文档'
      if (image.has(ext)) return '图片'
      if (binary.has(ext)) return '二进制'
      if (sql.has(ext)) return 'SQL'
      return '其他'
    },
    // 格式化文件大小
    formatFileSize(bytes) {
      if (bytes == null || isNaN(bytes)) return '-'
      if (bytes < 1024) return bytes + ' B'
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
      if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
      return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
    },
    loadFileList() {
      if (!this.localDir) return
      this.loading = true
      taskWorkflowApi.TaskWorkflowFileChangesDetail(this.localDir, this.effectiveParentBranch, (response) => {
        this.loading = false
        if (response && response.ErrCode === 0 && response.Data) {
          if (response.Data.summary) {
            this.summary = response.Data.summary
          }
          if (Array.isArray(response.Data.files)) {
            this.files = response.Data.files
            // 加载完文件列表后初始化过滤器
            this.$nextTick(() => this.initFileTypeFilter())
          }
          if (response.Data.error) {
            this.diffError = response.Data.error
          }
        }
      })
    },
    buildTree(files) {
      if (!files || files.length === 0) return []

      const dirMap = {}

      const ensureDir = (dirPath) => {
        if (dirMap[dirPath]) return dirMap[dirPath]
        const parts = dirPath.split('/')
        const name = parts[parts.length - 1]
        const parentPath = parts.slice(0, -1).join('/')
        const dirNode = {
          isDir: true,
          name,
          key: 'dir_' + dirPath,
          path: dirPath,
          files: [],
          children: [],
          fileCount: 0,
        }
        dirMap[dirPath] = dirNode
        if (parentPath) {
          ensureDir(parentPath).children.push(dirNode)
        }
        return dirNode
      }

      const rootFiles = []

      for (const file of files) {
        const path = file.path || ''
        const slashIdx = path.lastIndexOf('/')
        const dirPath = slashIdx >= 0 ? path.substring(0, slashIdx) : ''
        const fileName = slashIdx >= 0 ? path.substring(slashIdx + 1) : path
        const fileObj = { ...file, name: fileName }

        if (dirPath === '') {
          rootFiles.push({ ...fileObj, isDir: false, key: 'file_' + file.path })
        } else {
          ensureDir(dirPath)
          dirMap[dirPath].files.push(fileObj)
        }
      }

      const topLevelDirPaths = new Set()
      for (const dirPath of Object.keys(dirMap)) {
        const parts = dirPath.split('/')
        if (parts.length === 1) {
          topLevelDirPaths.add(dirPath)
        }
      }

      const computeFileCount = (dirNode) => {
        let count = dirNode.files.length
        for (const child of dirNode.children) {
          count += computeFileCount(child)
        }
        dirNode.fileCount = count
        return count
      }

      const sortDir = (dirNode) => {
        dirNode.children.sort((a, b) => a.name.localeCompare(b.name))
        dirNode.files.sort((a, b) => a.name.localeCompare(b.name))
        for (const child of dirNode.children) {
          sortDir(child)
        }
      }

      const topDirs = []
      for (const dirPath of topLevelDirPaths) {
        topDirs.push(dirMap[dirPath])
      }

      for (const dir of topDirs) {
        computeFileCount(dir)
        sortDir(dir)
      }

      const result = []
      for (const dir of topDirs.sort((a, b) => a.name.localeCompare(b.name))) {
        result.push(dir)
      }
      for (const file of rootFiles.sort((a, b) => a.name.localeCompare(b.name))) {
        result.push(file)
      }

      return result
    },
    toggleDir(key) {
      this.expandedDirs = {
        ...this.expandedDirs,
        [key]: this.expandedDirs[key] !== false ? false : true,
      }
    },
    selectFile(item) {
      this.selectedFile = item.path
      this.diffError = ''
      this.isBinary = false
      this.isImage = false
      this.binaryInfo = { file_type: '', old_size: 0, new_size: 0 }
      this.imageInfo = { image_type: '', old_image: '', new_image: '' }
      this.currentDiff = ''
      this.oldContent = ''
      this.newContent = ''
      this.diffChunks = []
      this.currentChunkIndex = -1
      this.loadFileDiff(item)
    },
    loadFileDiff(file) {
      this.diffLoading = true
      this.currentDiff = ''
      this.oldContent = ''
      this.newContent = ''
      this.isBinary = false
      this.isImage = false
      this.destroyMergeView()
      taskWorkflowApi.TaskWorkflowFileChangesFileDiff(this.localDir, this.effectiveParentBranch, file.path, (response) => {
        this.diffLoading = false
        if (response && response.ErrCode === 0 && response.Data) {
          const data = response.Data
          // 二进制文件
          if (data.is_binary) {
            this.isBinary = true
            this.binaryInfo = {
              file_type: data.file_type || '',
              old_size: data.old_size || 0,
              new_size: data.new_size || 0,
            }
            return
          }
          // 图片文件
          if (data.is_image) {
            this.isImage = true
            this.imageInfo = {
              image_type: data.image_type || 'png',
              old_image: data.old_image || '',
              new_image: data.new_image || '',
            }
            return
          }
          // 文本文件：优先使用 old_content/new_content（CodeMirror MergeView）
          if (data.old_content !== undefined && data.new_content !== undefined) {
            this.oldContent = data.old_content || ''
            this.newContent = data.new_content || ''
          }
          if (data.diff) {
            this.currentDiff = data.diff
          }
          // 如果没有 old/new 内容但有 diff 文本，降级到 diff2html
          if (!this.oldContent && !this.newContent && this.currentDiff) {
            this.renderDiff2Html(this.currentDiff)
          } else {
            this.renderDiff()
          }
        } else {
          if (file.type === 'untracked') {
            this.diffError = '文件 "' + file.path + '" 是未跟踪的新文件，暂无法对比差异。'
          } else {
            this.diffError = (response && response.ErrMsg) || '获取 diff 失败'
          }
        }
      })
    },
    handleClose() {
      this.$emit('update:visible', false)
      this.selectedFile = ''
      this.currentDiff = ''
      this.oldContent = ''
      this.newContent = ''
      this.isBinary = false
      this.isImage = false
      this.destroyMergeView()
    },
    destroyMergeView() {
      if (this._lineHLTimer) {
        clearTimeout(this._lineHLTimer)
        this._lineHLTimer = null
      }
      if (this.mergeView) {
        try {
          const editor = this.mergeView.editor()
          editor?.toTextArea?.()
        } catch (e) { /* ignore */ }
        this.mergeView = null
      }
      this._hlLines = null
      const container = this.$refs.diffContainer
      if (container) {
        container.innerHTML = ''
      }
    },
    renderDiff() {
      this.$nextTick(() => {
        // 等待浏览器完成布局（el-dialog 可能有 CSS 过渡动画）
        requestAnimationFrame(() => {
          setTimeout(() => {
            const container = this.$refs.diffContainer
            if (!container || !container.isConnected) return
            // 容器必须可见且有尺寸，否则 CodeMirror 初始化会报错
            if (container.offsetHeight === 0 || container.offsetWidth === 0) {
              // 等待布局完成后重试一次
              setTimeout(() => {
                this.renderDiff()
              }, 100)
              return
            }

            this.destroyMergeView()
            container.innerHTML = ''

            const mode = getModeForFile(this.selectedFile)

            if (this.diffMode === 'side-by-side') {
              // 横向对比：左侧原始版本，右侧改后版本
              // showDifferences: false 关闭 CodeMirror 内置的字级/行级标记
              try {
                this.mergeView = CodeMirror.MergeView(container, {
                  value: this.newContent || '',
                  origLeft: this.oldContent || '',
                  lineNumbers: true,
                  mode: mode || null,
                  theme: 'default',
                  collapseIdentical: false,
                  revertButtons: false,
                  readOnly: true,
                  lineWrapping: false,
                  scrollbarStyle: 'native',
                  showDifferences: true,
                  // 一次性渲染全部行，避免虚拟滚动时动态创建/销毁 DOM 导致卡顿
                  viewportMargin: Infinity,
                })
                // 手动应用行级背景色（只标整行，不标字级，仅初始化一次，无需滚动重绘）
                this.$nextTick(() => {
                  try { this.applyLineLevelHighlight() } catch (e) { /* ignore */ }
                  // 更新 diffChunks 用于跳转，并自动定位到第一个改动点
                  this.updateDiffChunks()
                })
              } catch (e) {
                console.error('MergeView 创建失败:', e)
                container.innerHTML = '<div style="padding:20px;color:#e53935;text-align:center;">对比视图加载失败: ' + (e.message || String(e)) + '</div>'
              }
            } else {
              // 纵向对比：使用 CodeMirror 单面板 + diff 高亮
              this.renderInlineDiff(container, mode)
            }
          }, 30)
        })
      })
    },
    renderInlineDiff(container, mode) {
      // 纵向对比：交替显示删除行（红底）和新增行（绿底）
      const Diff = require('diff')
      const changes = Diff.diffLines(this.oldContent || '', this.newContent || '')

      // 构建带类型标注的行列表
      let annotatedLines = []
      for (const change of changes) {
        const text = change.value.replace(/\n$/, '')
        const lines = text.split('\n')
        for (let i = 0; i < lines.length; i++) {
          // 跳过末尾空行（diffLines 会多出一个换行）
          if (i === lines.length - 1 && lines[i] === '' && change.value.endsWith('\n')) continue
          let type = 'unchanged'
          if (change.added) type = 'added'
          else if (change.removed) type = 'removed'
          annotatedLines.push({ text: lines[i], type })
        }
      }

      // 构建 CodeMirror 显示内容（删除行前加 - 前缀，新增行前加 + 前缀）
      const displayText = annotatedLines.map(l => {
        if (l.type === 'removed') return '- ' + l.text
        if (l.type === 'added') return '+ ' + l.text
        return '  ' + l.text
      }).join('\n')

      const editor = CodeMirror(container, {
        value: displayText,
        lineNumbers: true,
        mode: mode,
        readOnly: true,
        lineWrapping: false,
        gutters: ['CodeMirror-linenumbers', 'diff-gutter'],
        // 一次性渲染全部行，避免虚拟滚动时动态创建/销毁 DOM 导致卡顿
        viewportMargin: Infinity,
      })
      // 为编辑器的 prepareSelection 添加 try-catch 防护
      patchCmPrepareSelection(editor)
      // 保存内联编辑器实例，用于跳转滚动
      this._inlineEditor = editor

      // 标记每行的背景色和 gutter 符号
      for (let i = 0; i < annotatedLines.length; i++) {
        const line = annotatedLines[i]
        if (line.type === 'added') {
          editor.addLineClass(i, 'background', 'diff-line-added')
          editor.setGutterMarker(i, 'diff-gutter', (() => {
            const el = document.createElement('div')
            el.className = 'diff-gutter-marker diff-gutter-marker--add'
            el.textContent = '+'
            return el
          })())
        } else if (line.type === 'removed') {
          editor.addLineClass(i, 'background', 'diff-line-removed')
          editor.setGutterMarker(i, 'diff-gutter', (() => {
            const el = document.createElement('div')
            el.className = 'diff-gutter-marker diff-gutter-marker--del'
            el.textContent = '-'
            return el
          })())
        }
      }

      editor.setSize('100%', '100%')

      // 从 annotatedLines 中提取连续的改动块行号，用于跳转
      const inlineChunks = []
      let chunkStartLine = -1
      for (let i = 0; i < annotatedLines.length; i++) {
        if (annotatedLines[i].type !== 'unchanged') {
          if (chunkStartLine < 0) chunkStartLine = i
        } else {
          if (chunkStartLine >= 0) {
            inlineChunks.push({ editLine: chunkStartLine, origLine: chunkStartLine })
            chunkStartLine = -1
          }
        }
      }
      if (chunkStartLine >= 0) inlineChunks.push({ editLine: chunkStartLine, origLine: chunkStartLine })
      this.diffChunks = inlineChunks
      this.currentChunkIndex = inlineChunks.length > 0 ? 0 : -1
      // 自动定位到第一个改动点
      if (inlineChunks.length > 0) {
        this.$nextTick(() => this.scrollToChunk(0))
      }
    },
    renderDiff2Html(diffText) {
      this.destroyMergeView()
      try {
        const diffHtml = Diff2Html.html(diffText, {
          drawFileList: false,
          matching: 'lines',
          outputFormat: 'line-by-line',
          renderNothingWhenEmpty: false,
        })
        const container = this.$refs.diffContainer
        if (container) {
          container.innerHTML = diffHtml
          const codeLines = container.querySelectorAll('.d2h-code-line')
          codeLines.forEach(line => {
            const contentCells = line.querySelectorAll('.d2h-code-line-ctn')
            contentCells.forEach(cell => {
              cell.innerHTML = cell.textContent.replace(/\n/g, '<br>')
            })
          })
        }
      } catch (e) {
        this.diffError = '渲染 diff 失败: ' + (e.message || String(e))
      }
    },
    renderDiff2HtmlFromContent(oldText, newText) {
      // 使用 diff 库生成 unified diff，再用 diff2html 渲染
      try {
        const Diff = require('diff')
        const patch = Diff.createPatch(this.selectedFile, oldText || '', newText || '', '原始版本', '改后版本')
        if (patch && patch.trim()) {
          this.renderDiff2Html(patch)
        } else {
          const container = this.$refs.diffContainer
          if (container) {
            container.innerHTML = '<div style="padding:20px;color:#909399;text-align:center;">文件内容无差异</div>'
          }
        }
      } catch (e) {
        this.diffError = '渲染 diff 失败: ' + (e.message || String(e))
      }
    },
    // 手动应用行级背景色（只用 addLineClass，不用 markText，无波浪线）
    // 仅在 MergeView 创建后调用一次，CodeMirror 虚拟滚动时会自动保持 lineClass
    applyLineLevelHighlight() {
      try {
        const mv = this.mergeView
        if (!mv) return
        const editCm = mv.editor()
        const origCm = mv.leftOriginal()
        if (!editCm || !origCm) return

        // 获取差异化块数据
        const leftDv = mv.left
        if (!leftDv || !leftDv.chunks || leftDv.chunks.length === 0) return

        const chunks = leftDv.chunks

        // 清除旧的行级背景（只清除已记录的行，避免遍历全部行）
        if (this._hlLines) {
          for (const i of this._hlLines.edit) {
            editCm.removeLineClass(i, 'background', 'diff-chunk-inserted')
          }
          for (const i of this._hlLines.orig) {
            origCm.removeLineClass(i, 'background', 'diff-chunk-deleted')
          }
        }
        this._hlLines = { edit: [], orig: [] }

        // 为每个 chunk 添加行级背景
        for (const ch of chunks) {
          // 原始版本（左侧）：标记被删除的行
          if (ch.origFrom < ch.origTo) {
            for (let i = ch.origFrom; i < ch.origTo; i++) {
              origCm.addLineClass(i, 'background', 'diff-chunk-deleted')
              this._hlLines.orig.push(i)
            }
          }
          // 改后版本（右侧）：标记新增的行
          if (ch.editFrom < ch.editTo) {
            for (let i = ch.editFrom; i < ch.editTo; i++) {
              editCm.addLineClass(i, 'background', 'diff-chunk-inserted')
              this._hlLines.edit.push(i)
            }
          }
        }
      } catch (e) { /* ignore highlight errors */ }
    },
    renderCurrentDiff() {
      if (this.selectedFile && (this.oldContent || this.newContent || this.currentDiff)) {
        this.renderDiff()
      }
    },
    // updateDiffChunks 从 MergeView 的 chunks 数据中提取改动行号列表，用于跳转
    // 提取后自动滚动到第一个改动点
    updateDiffChunks() {
      if (this.mergeView) {
        try {
          const leftDv = this.mergeView.left
          if (leftDv && leftDv.chunks && leftDv.chunks.length > 0) {
            this.diffChunks = leftDv.chunks.map(ch => ({
              editLine: ch.editFrom,
              origLine: ch.origFrom,
            }))
          } else {
            this.diffChunks = []
          }
        } catch (e) {
          this.diffChunks = []
        }
      } else {
        this.diffChunks = []
      }
      this.currentChunkIndex = -1
      // 自动定位到第一个改动点
      if (this.diffChunks.length > 0) {
        this.currentChunkIndex = 0
        this.$nextTick(() => this.scrollToChunk(0))
      }
    },
    // jumpToPrevChange 跳转到上一处改动，如果已在最前面则切换上一个文件
    jumpToPrevChange() {
      if (this.hasPrevChange) {
        this.currentChunkIndex--
        this.scrollToChunk(this.currentChunkIndex)
      } else if (this.hasPrevFile) {
        this.goToPrevFile()
      }
    },
    // jumpToNextChange 跳转到下一处改动，如果已在最后面则切换下一个文件
    jumpToNextChange() {
      if (this.hasNextChange) {
        // 第一次点击或从上一个文件切过来时，从 -1 跳到 0
        if (this.currentChunkIndex < 0) {
          this.currentChunkIndex = 0
        } else {
          this.currentChunkIndex++
        }
        this.scrollToChunk(this.currentChunkIndex)
      } else if (this.hasNextFile) {
        this.goToNextFile()
      }
    },
    // scrollToChunk 滚动 CodeMirror 到指定 diff 块的行号
    scrollToChunk(index) {
      const chunks = this.diffChunks
      if (!chunks || index < 0 || index >= chunks.length) return
      const chunk = chunks[index]
      if (this.mergeView) {
        try {
          const editor = this.mergeView.editor()
          const orig = this.mergeView.leftOriginal()
          if (editor) editor.scrollIntoView({ line: chunk.editLine, ch: 0 }, 80)
          if (orig) orig.scrollIntoView({ line: chunk.origLine, ch: 0 }, 80)
        } catch (e) { /* ignore */ }
      }
      if (this._inlineEditor) {
        try {
          this._inlineEditor.scrollIntoView({ line: chunk.editLine, ch: 0 }, 80)
        } catch (e) { /* ignore */ }
      }
    },
    // goToNextFile 切换到文件列表中的下一个文件
    goToNextFile() {
      const idx = this.selectedFileIndex
      if (idx < 0 || idx >= this.filteredFiles.length - 1) return
      const nextFile = this.filteredFiles[idx + 1]
      if (nextFile) this.selectFile(nextFile)
    },
    // goToPrevFile 切换到文件列表中的上一个文件
    goToPrevFile() {
      const idx = this.selectedFileIndex
      if (idx <= 0) return
      const prevFile = this.filteredFiles[idx - 1]
      if (prevFile) this.selectFile(prevFile)
    },

  },
}
</script>

<style scoped>
/* ===== 文件变更弹窗样式 ===== */
.file-changes-detail {
  height: 75vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.file-changes-detail__toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0 12px;
  border-bottom: 1px solid #e8ecf1;
  margin-bottom: 8px;
  flex-shrink: 0;
  gap: 12px;
}

/* ===== 文件类型过滤栏 ===== */
.file-changes-detail__filter-bar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px 8px;
  padding: 6px 0 8px;
  border-bottom: 1px solid #e8ecf1;
  margin-bottom: 8px;
  flex-shrink: 0;
}

.file-changes-detail__filter-label {
  font-size: 12px;
  color: #606266;
  font-weight: 600;
  margin-right: 4px;
  flex-shrink: 0;
}

.file-changes-detail__filter-sep {
  width: 1px;
  height: 14px;
  background: #dcdfe6;
  margin: 0 2px;
  flex-shrink: 0;
}

.file-changes-detail__info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.file-changes-detail__dir {
  font-weight: 600;
  font-size: 13px;
  color: #303133;
  max-width: 350px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-changes-detail__branch {
  font-size: 12px;
  color: #909399;
}

.file-changes-detail__summary {
  display: flex;
  gap: 6px;
}

.file-changes-stat {
  font-size: 12px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 3px;
}

.file-changes-stat--committed {
  background: #dafbe1;
  color: #1a7f37;
}

.file-changes-stat--staged {
  background: #f0f7ff;
  color: #0366d6;
}

.file-changes-stat--modified {
  background: #fff8c5;
  color: #9a6700;
}

.file-changes-stat--untracked {
  background: #f3f4f6;
  color: #656d76;
}

.file-changes-stat--additions {
  background: #dafbe1;
  color: #1a7f37;
}

.file-changes-stat--deletions {
  background: #ffebe9;
  color: #cf222e;
}

.file-changes-detail__body {
  flex: 1;
  display: flex;
  gap: 0;
  overflow: hidden;
  min-height: 0;
}

.file-changes-detail__tree-panel {
  width: 280px;
  min-width: 200px;
  border: 1px solid #e8ecf1;
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  margin-right: 8px;
}

.file-changes-detail__tree-title {
  font-size: 12px;
  font-weight: 600;
  color: #606266;
  padding: 10px 12px;
  border-bottom: 1px solid #e8ecf1;
  background: #f6f8fa;
  flex-shrink: 0;
}

.file-changes-detail__tree {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
}

.file-changes-detail__tree-empty {
  padding: 20px;
  text-align: center;
  color: #c0c4cc;
  font-size: 13px;
}

.file-changes-tree__dir {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
  cursor: pointer;
  font-size: 13px;
  color: #0366d6;
  user-select: none;
}

.file-changes-tree__dir:hover {
  background: #f0f7ff;
}

.file-changes-tree__dir-icon {
  font-size: 10px;
  width: 14px;
  text-align: center;
  flex-shrink: 0;
}

.file-changes-tree__dir-name {
  font-weight: 500;
}

.file-changes-tree__dir-count {
  font-size: 11px;
  color: #8b949e;
}

.file-changes-tree__file {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 3px 12px 3px 32px;
  cursor: pointer;
  font-size: 12px;
  color: #24292e;
  transition: background 0.15s;
}

.file-changes-tree__file:hover {
  background: #f0f7ff;
}

.file-changes-tree__file--active {
  background: #ddf4ff !important;
  font-weight: 600;
}

.file-changes-tree__file-type {
  display: inline-block;
  font-size: 10px;
  font-family: monospace;
  padding: 0 4px;
  border-radius: 3px;
  min-width: 20px;
  text-align: center;
  line-height: 18px;
}

.file-changes-tree__file-type--committed {
  background: #dafbe1;
  color: #1a7f37;
}

.file-changes-tree__file-type--staged {
  background: #f0f7ff;
  color: #0366d6;
}

.file-changes-tree__file-type--modified {
  background: #fff8c5;
  color: #9a6700;
}

.file-changes-tree__file-type--untracked {
  background: #f3f4f6;
  color: #656d76;
}

.file-changes-tree__file-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-changes-detail__diff-panel {
  flex: 1;
  border: 1px solid #e8ecf1;
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

.file-changes-detail__diff-header {
  padding: 8px 12px;
  border-bottom: 1px solid #e8ecf1;
  background: #f6f8fa;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-changes-detail__nav-btns {
  margin-left: auto;
  flex-shrink: 0;
  display: flex;
  gap: 4px;
}

.file-changes-detail__diff-file {
  font-size: 13px;
  font-family: monospace;
  color: #24292e;
}

.file-changes-detail__diff-stats {
  display: inline-flex;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  font-family: monospace;
}

.file-changes-detail__diff-stat--add {
  color: #1a7f37;
}

.file-changes-detail__diff-stat--del {
  color: #cf222e;
}

.file-changes-detail__diff-badge {
  font-size: 11px;
  padding: 1px 8px;
  border-radius: 10px;
  margin-left: 8px;
  font-weight: 600;
}

.file-changes-detail__diff-badge--binary {
  background: #fff3cd;
  color: #856404;
}

.file-changes-detail__diff-badge--image {
  background: #d1ecf1;
  color: #0c5460;
}

/* ===== 二进制文件信息卡片 ===== */
.file-changes-detail__binary-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 30px;
  overflow-y: auto;
}

.file-changes-detail__binary-icon {
  width: 80px;
  height: 80px;
  border-radius: 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
}

.file-changes-detail__binary-ext {
  color: #fff;
  font-size: 18px;
  font-weight: 700;
  font-family: monospace;
  text-transform: uppercase;
}

.file-changes-detail__binary-meta {
  background: #f6f8fa;
  border: 1px solid #e8ecf1;
  border-radius: 8px;
  padding: 16px 24px;
  min-width: 300px;
  margin-bottom: 16px;
}

.file-changes-detail__binary-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  font-size: 13px;
}

.file-changes-detail__binary-row + .file-changes-detail__binary-row {
  border-top: 1px solid #e8ecf1;
}

.file-changes-detail__binary-label {
  color: #606266;
}

.file-changes-detail__binary-value {
  font-weight: 600;
  color: #303133;
  font-family: monospace;
}

.file-changes-detail__binary-value--old {
  color: #909399;
}

.file-changes-detail__binary-value--increased {
  color: #e53935;
}

.file-changes-detail__binary-value--decreased {
  color: #1a7f37;
}

.file-changes-detail__binary-hint {
  font-size: 12px;
  color: #909399;
  text-align: center;
}

/* ===== 图片对比展示 ===== */
.file-changes-detail__image-compare {
  flex: 1;
  display: flex;
  align-items: stretch;
  gap: 8px;
  padding: 12px;
  overflow: auto;
  min-height: 0;
}

.file-changes-detail__image-pane {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  border: 1px solid #e8ecf1;
  border-radius: 6px;
  overflow: hidden;
  background: #fafbfc;
}

.file-changes-detail__image-title {
  padding: 8px 12px;
  font-size: 12px;
  font-weight: 600;
  color: #606266;
  background: #f6f8fa;
  border-bottom: 1px solid #e8ecf1;
  text-align: center;
  flex-shrink: 0;
}

.file-changes-detail__image-box {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 12px;
  overflow: auto;
  min-height: 100px;
}

.file-changes-detail__image-img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  border-radius: 4px;
}

.file-changes-detail__image-empty {
  color: #c0c4cc;
  font-size: 13px;
}

.file-changes-detail__image-divider {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 40px;
  font-size: 20px;
  color: #909399;
}

.file-changes-detail__diff-content {
  flex: 1;
  overflow: hidden;
  min-height: 0;
}

.file-changes-detail__diff-placeholder {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #c0c4cc;
  font-size: 14px;
}

.file-changes-detail__diff-loading {
  flex-direction: column;
  gap: 10px;
}

.file-changes-detail__diff-loading i {
  font-size: 28px;
  color: #409eff;
}

.file-changes-detail__diff-error {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #e53935;
  font-size: 13px;
  padding: 20px;
  text-align: center;
}

.file-changes-detail__footer {
  text-align: right;
}

/* ===== CodeMirror MergeView 适配样式 ===== */
.file-changes-detail__diff-content :deep(.CodeMirror) {
  height: 100%;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 13px;
  line-height: 1.5;
}

.file-changes-detail__diff-content :deep(.CodeMirror-merge) {
  height: 100%;
}

.file-changes-detail__diff-content :deep(.CodeMirror-merge-pane) {
  height: 100%;
}

/* 行级改动背景 — 左面板（原始版本，标红删除行） */
.file-changes-detail__diff-content :deep(.diff-chunk-deleted) {
  background-color: #fecdd3;
}

/* 行级改动背景 — 右面板（改后版本，标绿新增行） */
.file-changes-detail__diff-content :deep(.diff-chunk-inserted) {
  background-color: #bbf7d0;
}

/* 移除 MergeView markText 产生的绿色波浪下划线 */
.file-changes-detail__diff-content :deep(.CodeMirror-merge-l-inserted),
.file-changes-detail__diff-content :deep(.CodeMirror-merge-l-deleted),
.file-changes-detail__diff-content :deep(.CodeMirror-merge-r-inserted),
.file-changes-detail__diff-content :deep(.CodeMirror-merge-r-deleted) {
  text-decoration: none !important;
  background-image: none !important;
}

/* 中间连接线区域 */
.file-changes-detail__diff-content :deep(.CodeMirror-merge-gap) {
  background: #f6f8fa;
  border-left: 1px solid #d0d7de;
  border-right: 1px solid #d0d7de;
  z-index: 2;
  min-width: 50px;
}

/* 连接线路径 — 蓝色填充 + 描边 */
.file-changes-detail__diff-content :deep(.CodeMirror-merge-l-connect),
.file-changes-detail__diff-content :deep(.CodeMirror-merge-r-connect) {
  fill: rgba(3, 102, 214, 0.12);
  stroke: rgba(3, 102, 214, 0.4);
  stroke-width: 1;
  stroke-linejoin: round;
}


.file-changes-detail__diff-content :deep(.CodeMirror-gutters) {
  background: #f6f8fa;
  border-right: 1px solid #e8ecf1;
}

/* ===== 纵向对比 CodeMirror 行标记样式 ===== */
.file-changes-detail__diff-content :deep(.diff-line-added) {
  background-color: #bbf7d0 !important;
}

.file-changes-detail__diff-content :deep(.diff-line-removed) {
  background-color: #fecdd3 !important;
}

.file-changes-detail__diff-content :deep(.diff-gutter-marker) {
  font-weight: 700;
  font-size: 14px;
  width: 16px;
  text-align: center;
  line-height: 1;
  margin-top: 2px;
}

.file-changes-detail__diff-content :deep(.diff-gutter-marker--add) {
  color: #1a7f37;
}

.file-changes-detail__diff-content :deep(.diff-gutter-marker--del) {
  color: #cf222e;
}

/* diff2html 纵向对比适配 */
.file-changes-detail__diff-content :deep(.d2h-wrapper) {
  height: 100%;
  overflow: auto;
}

.file-changes-detail__diff-content :deep(.d2h-code-line-ctn) {
  white-space: pre-wrap;
  word-break: break-all;
}

.file-changes-detail__diff-content :deep(.d2h-file-side-diff) {
  display: flex;
}

.file-changes-detail__diff-content :deep(.d2h-code-line) {
  display: flex;
}

.file-changes-detail__diff-content :deep(.d2h-code-side-linenumber) {
  width: 40px;
  min-width: 40px;
  padding: 0 10px;
  background-color: #f6f8fa;
  color: rgba(27, 31, 35, 0.3);
  text-align: right;
  user-select: none;
}
</style>
