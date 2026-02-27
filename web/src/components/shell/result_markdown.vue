<template>
  <MarkdownRenderer id="showShellResult" :source="shellShowResult" :style="{ height: divHeight + 'px' }"></MarkdownRenderer>
</template>

<style scoped>
/* 护眼配色方案 */
.sticky-textarea-div {
  background: #282c34;
  color: #abb2bf;
  white-space: pre-wrap;
  word-break: break-all;
  padding: 16px;
  border-radius: 8px;
  border-left: 3px solid #5c6370;
  font-family: 'JetBrains Mono', 'Fira Code', 'SF Mono', 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.7;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  font-weight: 400;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  overflow-y: auto;
  overflow-x: hidden;
  display: block;
  height: 100%;
  transition: all 0.2s ease;
}

#showShellResult{
  height : 100%;
}

::deep(.el-scrollbar__thumb) {
  background: #5c6370 !important;
  border-radius: 4px !important;
  opacity: 0.7 !important;
  transition: opacity 0.2s ease;
}

::deep(.el-scrollbar__thumb:hover) {
  background: #6c7280 !important;
  opacity: 1 !important;
}

::deep(.el-scrollbar__bar) {
  background: #21252b !important;
  border-radius: 4px;
}

@keyframes gentle-blink {
  0%, 100% {
    opacity: 0.7;
  }
  50% {
    opacity: 0.3;
  }
}
</style>

<script>
import {
  defineExpose,
  defineComponent,
  inject,
  defineEmits,
  getCurrentInstance,
  reactive,
  computed,
  ref,
  watch
} from 'vue';
import shell from '@/utils/base/shell'
import MarkdownRenderer from "@/components/base/markdown.vue";
import {Close} from '@element-plus/icons-vue'

export default defineComponent({
  components: {MarkdownRenderer, Close},
  props: {
    shellShowResult: {
      type: String
    },
    showModel: {
      type: String
    },
    isRunning: {
      type: Boolean
    },
    divHeight: {
      type: Number,
    }
  },
  setup(props) {
    const proxy = getCurrentInstance().proxy
    const showOk = ref(false)
    /* 1. 计算属性：运行中显示计数，刚结束显示 ok! */
    const btnText = computed(() =>
        showOk.value ? ' run success ! ' : `shell 输出（${props.shellShowResult.length}）`
    )
    /* 2. 监听 isRunning，变 false 时切到 ok!，1.5s 后恢复 */
    watch(() => props.isRunning, val => {
      if (!val) {
        showOk.value = true
        setTimeout(() => showOk.value = false, 1500)
      }
    })
    return {
      btnText,
    }
  }
})
</script>
