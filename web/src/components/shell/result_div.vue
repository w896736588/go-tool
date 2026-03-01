<template>
  <el-scrollbar id="showShellResult" :style="scrollbarStyle">
    <div
      class="sticky-textarea-div"
      v-html="shellShowResult"
      :style="contentStyle"
    ></div>
  </el-scrollbar>
</template>

<script setup>
/* global defineProps */
import { computed, nextTick, watch, onMounted, onBeforeUnmount } from 'vue'

const props = defineProps({
  shellShowResult: { type: String, default: '' },
  divHeight: { type: Number, default: 200 },
  useContainerHeight: { type: Boolean, default: false }
})

const scrollbarStyle = computed(() => {
  if (props.useContainerHeight) {
    return { height: '100%' }
  }
  return { height: props.divHeight - 17 + 'px' }
})

const contentStyle = computed(() => {
  if (props.useContainerHeight) {
    return { minHeight: '100%' }
  }
  return { minHeight: props.divHeight - 25 + 'px' }
})

const scrollThreshold = 10
let autoScroll = true
let wrapEl = null
let rafLock = false

function getWrap() {
  const sb = document.getElementById('showShellResult')
  return sb?.parentNode
}

function scrollToBottom() {
  if (!autoScroll || !wrapEl) return
  wrapEl.scrollTop = wrapEl.scrollHeight
}

function onScroll() {
  if (rafLock) return
  rafLock = true
  window.requestAnimationFrame(() => {
    const distance = wrapEl.scrollHeight - wrapEl.scrollTop - wrapEl.clientHeight
    autoScroll = distance <= scrollThreshold
    rafLock = false
  })
}

watch(
  () => props.shellShowResult,
  () => nextTick(scrollToBottom),
  { flush: 'post' }
)

onMounted(() => {
  nextTick(() => {
    wrapEl = getWrap()
    if (!wrapEl) return
    scrollToBottom()
    wrapEl.addEventListener('scroll', onScroll, { passive: true })
  })
})

onBeforeUnmount(() => {
  if (wrapEl) wrapEl.removeEventListener('scroll', onScroll)
})
</script>

<style scoped>
.sticky-textarea-div {
  background: #eef3ea;
  color: #435244;
  white-space: pre-wrap;
  word-break: break-all;
  padding: 16px;
  border-radius: 8px;
  border-left: 3px solid #8fae92;
  font-family: 'JetBrains Mono', 'Fira Code', 'SF Mono', 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.7;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  font-weight: 400;
  box-shadow: 0 1px 6px rgba(62, 86, 62, 0.08);
  overflow-y: auto;
  overflow-x: hidden;
  display: block;
  height: 100%;
  transition: all 0.2s ease;
}

#showShellResult {
  height: 100%;
}

::deep(.el-scrollbar__thumb) {
  background: #a4b7a3 !important;
  border-radius: 4px !important;
  opacity: 0.7 !important;
  transition: opacity 0.2s ease;
}

::deep(.el-scrollbar__thumb:hover) {
  background: #8fa48f !important;
  opacity: 1 !important;
}

::deep(.el-scrollbar__bar) {
  background: #dfe8da !important;
  border-radius: 4px;
}
</style>
